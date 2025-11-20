package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/godyy/gcluster/center"
	"github.com/godyy/glog"
	pkgerrors "github.com/pkg/errors"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// ErrNodeNotFound 节点未找到错误.
var ErrNodeNotFound = errors.New("node not found")

// Node 节点信息.
type Node struct {
	ID   string `json:"id"`   // 节点ID.
	Addr string `json:"addr"` // 节点集群内通信地址.
}

// GetNodeId 节点ID.
func (n *Node) GetNodeId() string {
	return n.ID
}

// GetNodeAddr 获取节点地址.
func (n *Node) GetNodeAddr() string {
	return n.Addr
}

const (
	dialTimeout          = time.Second * 5  // 默认拨号超时
	dialKeepAliveTime    = time.Second * 30 // 默认保持连接活跃间隔时间
	dialKeepAliveTimeout = time.Second * 5  // 默认保持连接活跃超时时间
	leaseTTL             = int64(15)        // 默认租约存活时间(秒)

	checkAliveTimeout = 1 * time.Second  // 检查租约是否存活超时
	opTimeout         = time.Second * 5  // 默认操作超时时间
	syncNodesTimeout  = 10 * time.Second // 默认同步节点超时时间
)

// CenterConfig 集群中心配置.
type CenterConfig struct {
	// EndPoints etcd 节点地址列表.
	EndPoints []string

	// Root etcd 根路径.
	Root string

	// Self 自身节点信息.
	Self *Node

	// 日志工具.
	Log glog.Logger
}

func (c *CenterConfig) init() error {
	if len(c.EndPoints) == 0 {
		return errors.New("no EndPoints")
	}

	if c.Root == "" {
		return errors.New("no Root")
	}

	if c.Self == nil {
		return errors.New("no Self")
	}

	return nil
}

// Center Center 负责管理和维护集群中的节点信息.
// 使用etcd，通过配置根路径，在根路径下注册自身节点,并获取根路径下注册的所有其它节点.
// 同时还会异步监听根路径下产生的节点新增/删除事件,并实时更新本地节点列表.
type Center struct {
	cfg     *CenterConfig
	etcdCli *etcdv3.Client

	mu    sync.RWMutex
	nodes map[string]*Node

	leaseID         etcdv3.LeaseID
	watchCancel     context.CancelFunc
	keepAliveCancel context.CancelFunc
}

func NewCenter(cfg *CenterConfig) *Center {
	if err := cfg.init(); err != nil {
		panic(err)
	}

	return &Center{
		cfg:   cfg,
		nodes: make(map[string]*Node),
	}
}

// Start 启动集群中心，注册自身并同步/监听节点列表。
func (c *Center) Start(ctx context.Context) (err error) {
	defer func() {
		if err == nil {
			return
		}

		if c.leaseID != 0 {
			_, _ = c.etcdCli.Revoke(ctx, c.leaseID)
			c.leaseID = 0
		}

		if c.etcdCli != nil {
			c.etcdCli.Close()
			c.etcdCli = nil
		}

		c.nodes = nil
	}()

	// 创建 etcd 客户端.
	etcdCli, err := etcdv3.New(etcdv3.Config{
		Endpoints:            c.cfg.EndPoints,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    dialKeepAliveTime,
		DialKeepAliveTimeout: dialKeepAliveTimeout,
		Logger:               c.cfg.Log.ZapLogger().Named("etcd"),
	})
	if err != nil {
		return pkgerrors.WithMessage(err, "create etcd client")
	}
	c.etcdCli = etcdCli

	// 获取租约
	if err := c.grantLease(ctx); err != nil {
		return err
	}

	// 注册自身节点
	if err := c.registerSelf(ctx); err != nil {
		return err
	}

	// 同步所有节点信息
	rev, err := c.syncNodes(ctx, false)
	if err != nil {
		return pkgerrors.WithMessage(err, "sync all nodes from etcd")
	}

	// 启动保持连接活跃后台
	kaCtx, kaCancel := context.WithCancel(context.Background())
	c.keepAliveCancel = kaCancel
	go c.keepAlive(kaCtx)

	// 启动监听节点变更后台
	wCtx, wCancel := context.WithCancel(context.Background())
	c.watchCancel = wCancel
	go c.watch(wCtx, rev)

	return nil
}

// Close 关闭集群中心，撤销注册并关闭 etcd 客户端。
func (c *Center) Close(ctx context.Context) error {
	if c.watchCancel != nil {
		c.watchCancel()
	}
	if c.keepAliveCancel != nil {
		c.keepAliveCancel()
	}
	if c.leaseID != 0 {
		_, _ = c.etcdCli.Revoke(ctx, c.leaseID)
	}
	if c.etcdCli != nil {
		c.etcdCli.Close()
	}
	return nil
}

func (c *Center) getLog() glog.Logger {
	return c.cfg.Log
}

// GetSelf 返回自身节点。
func (c *Center) GetSelf() *Node { return c.cfg.Self }

// GetNode 返回指定节点（不含自身）。
func (c *Center) GetNode(nodeID string) (center.Node, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.nodes == nil {
		return nil, errors.New("nodes not initialize")
	}
	if c.cfg.Self.ID == nodeID {
		return c.cfg.Self, nil
	}
	node, ok := c.nodes[nodeID]
	if !ok {
		return nil, ErrNodeNotFound
	}
	return node, nil
}

// keyForNode 返回节点对应的key.
func (c *Center) keyForNode(nodeID string) string {
	return c.cfg.Root + "/" + nodeID
}

// grantLease 申请租约.
func (c *Center) grantLease(ctx context.Context) error {
	ttl := leaseTTL
	leaseResp, err := c.etcdCli.Grant(ctx, ttl)
	if err != nil {
		return pkgerrors.WithMessage(err, "grant etcd lease")
	}
	c.leaseID = leaseResp.ID
	return nil
}

// checkLeaseAlive 检查租约是否活跃.
func (c *Center) checkLeaseAlive(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, checkAliveTimeout)
	defer cancel()
	resp, err := c.etcdCli.TimeToLive(ctx, c.leaseID)
	if err != nil {
		return false, err
	}
	return resp.TTL > 0, nil
}

// registerSelf 使用当前租约将自身节点信息写入 etcd。
func (c *Center) registerSelf(ctx context.Context) error {
	val, err := json.Marshal(c.cfg.Self)
	if err != nil {
		return pkgerrors.WithMessage(err, "marshal self node")
	}
	_, err = c.etcdCli.Put(ctx, c.keyForNode(c.cfg.Self.ID), string(val), etcdv3.WithLease(c.leaseID))
	if err != nil {
		return pkgerrors.WithMessage(err, "put self node to etcd")
	}
	return nil
}

// grantLeaseAndReregisterSelf 重新申请租约并重新注册自身节点。
func (c *Center) grantLeaseAndReregisterSelf(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout*2)
	defer cancel()

	if err := c.grantLease(ctx); err != nil {
		return err
	}
	if err := c.registerSelf(ctx); err != nil {
		return err
	}
	return nil
}

// syncNodes 同步所有节点信息.
func (c *Center) syncNodes(ctx context.Context, withTimeout bool) (int64, error) {
	if withTimeout {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, syncNodesTimeout)
		defer cancel()
	}

	prefix := c.cfg.Root + "/"
	resp, err := c.etcdCli.Get(ctx, prefix, etcdv3.WithPrefix())
	if err != nil {
		return 0, err
	}

	nodes := make(map[string]*Node)
	for _, kv := range resp.Kvs {
		id := strings.TrimPrefix(string(kv.Key), prefix)
		if id == c.cfg.Self.ID {
			continue
		}
		n := &Node{}
		if err := json.Unmarshal(kv.Value, n); err != nil {
			c.getLog().Errorf("unmarshal node %s: %v", id, err)
			continue
		}
		nodes[id] = n
	}

	c.mu.Lock()
	c.nodes = nodes
	c.mu.Unlock()

	return resp.Header.Revision, nil
}

// addNode 添加节点.
func (c *Center) addNode(n *Node) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodes[n.ID] = n
}

// delNode 删除节点.
func (c *Center) delNode(nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.nodes, nodeID)
}

// keepAlive 保持租约活跃.
func (c *Center) keepAlive(ctx context.Context) {
	var (
		alive   = true
		backoff = getRetryBackoff(0) // 重试退避时间
	)

	for {
		// ctx canceled
		if ctx.Err() != nil {
			return
		}

		// 如果租约存活，检查确认租约是否存活.
		// 否则, 重新申请租约.
		if alive {
			if b, err := c.checkLeaseAlive(ctx); err != nil {
				// 表示etcd服务器存在问题, 稍后重试.
				c.getLog().Errorf("keep-alive: check lease alive failed, %v", err)
				backoff = doRetryBackoff(ctx, backoff, 0)
				continue
			} else {
				if b != alive {
					alive = b
					backoff = getRetryBackoff(0)
					if !alive {
						// 租约失效, 稍后重新申请租约
						c.getLog().Error("keep-alive: lease expired or invalid")
						continue
					}
				}
			}
		} else {
			if err := c.grantLeaseAndReregisterSelf(ctx); err != nil {
				// 失败则稍后重试
				c.getLog().Errorf("keep-alive: %v", err)
				backoff = doRetryBackoff(ctx, backoff, 0)
				continue
			}

			alive = true
			backoff = getRetryBackoff(0)
		}

		// 创建保活通道
		ch, err := c.etcdCli.KeepAlive(ctx, c.leaseID)
		if err != nil {
			c.getLog().Errorf("keep-alive: create channel failed: %v", err)
			backoff = doRetryBackoff(ctx, backoff, 0)
			continue
		}
		backoff = getRetryBackoff(0)

	keep_alive_internal:
		for {
			select {
			case <-ctx.Done():
				// ctx canceled
				return
			case resp := <-ch:
				if resp == nil || resp.TTL <= 0 {
					// 租约到期或失效，标记租约失效, 稍后重新申请租约
					c.getLog().Error("keep-alive: channel closed")
					backoff = doRetryBackoff(ctx, backoff, time.Duration(rand.Int63n(100)*10)*time.Millisecond)
					break keep_alive_internal
				}
			}
		}
	}
}

// watch 监听节点变更.
func (c *Center) watch(ctx context.Context, startRev int64) {
	var (
		prefix  = c.cfg.Root + "/"   // 监听前缀
		lastRev = startRev           // 最后一次同步数据的修订好号
		backoff = getRetryBackoff(0) // 重试退避时间
		sync    = false              // 是否需要同步数据
	)

	for {
		// ctx canceled
		if ctx.Err() != nil {
			return
		}

		// 同步数据，获取最新的修订号
		if sync {
			rev, err := c.syncNodes(ctx, true)
			if err != nil {
				// 同步失败，稍后重试
				c.getLog().Errorf("watch: sync nodes: %v", err)
				backoff = doRetryBackoff(ctx, backoff, 0)
				continue
			}

			lastRev = rev
			sync = false
			backoff = getRetryBackoff(0)
		}

		// 创建监听通道
		wch := c.etcdCli.Watch(ctx, prefix, etcdv3.WithPrefix(), etcdv3.WithRev(lastRev+1))

		// 开始监听
	watch_internal:
		for {
			select {
			case <-ctx.Done():
				// ctx canceled
				return
			case wresp, ok := <-wch:
				if !ok {
					// 通道关闭（断线等），跳出以重建监听通道
					c.getLog().Error("watch: channel closed")
					backoff = doRetryBackoff(ctx, backoff, time.Duration(rand.Int63n(100)*10)*time.Millisecond)
					break watch_internal
				}
				if wresp.Canceled {
					// 监听被取消：可能因为压缩或其他错误
					// 如果出现压缩，或无法获取更多事件，需要重新同步并从最新修订恢复
					if wresp.CompactRevision > 0 || wresp.Err() != nil {
						c.getLog().Errorf("watch: compact revision: %d, err: %v", wresp.CompactRevision, wresp.Err())
						sync = true
					}
					backoff = doRetryBackoff(ctx, backoff, time.Duration(rand.Int63n(100)*10)*time.Millisecond)
					break watch_internal
				}
				// 正常事件，处理并更新最新修订
				lastRev = wresp.Header.Revision
				for _, ev := range wresp.Events {
					nodeID := strings.TrimPrefix(string(ev.Kv.Key), prefix)
					if nodeID == c.cfg.Self.ID {
						continue
					}
					switch ev.Type {
					case etcdv3.EventTypePut:
						n := &Node{}
						if err := json.Unmarshal(ev.Kv.Value, n); err != nil {
							c.getLog().Errorf("watch: unmarshal node %s: %v", nodeID, err)
							continue
						}
						c.addNode(n)
						c.getLog().Infof("watch: add node %s", nodeID)
					case etcdv3.EventTypeDelete:
						c.delNode(nodeID)
						c.getLog().Infof("watch: del node %s", nodeID)
					}
				}
			}
		}
	}
}

const (
	minRetryBackoff = time.Millisecond * 200 // 最小重试退避时间
	maxRetryBackoff = time.Second * 5        // 最大重试退避时间
)

// getRetryBackoff 获取重试退避时间.
// 支持指数退避.
func getRetryBackoff(cur time.Duration) time.Duration {
	if cur < minRetryBackoff {
		return minRetryBackoff
	}
	if cur >= maxRetryBackoff {
		return maxRetryBackoff
	}
	cur = cur << 2
	if cur > maxRetryBackoff {
		cur = maxRetryBackoff
	}
	return cur
}

// doRetryBackoff 执行重试退避.
// cur 当前需要退避的时间. add 为附加延迟时间, 不计入下次退避时间的计算.
// 返回下一次需要退避的时间.
func doRetryBackoff(ctx context.Context, cur time.Duration, add time.Duration) time.Duration {
	select {
	case <-ctx.Done():
	case <-time.After(cur + add):
	}
	return getRetryBackoff(cur)
}
