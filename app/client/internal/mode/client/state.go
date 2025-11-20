package client

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"github.com/godyy/ggs/app/client/internal/env"
	"github.com/godyy/ggs/app/client/internal/mode/internal/utils"
	"github.com/godyy/ggs/app/login/httpproto"
	"github.com/godyy/ggs/internal/consts"
	authjwt "github.com/godyy/ggs/internal/core/auth/jwt"
	cenv "github.com/godyy/ggs/internal/env"
	"github.com/godyy/ggs/internal/models"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	"github.com/godyy/ggs/internal/utils/ctxutils"
	pkgerrors "github.com/pkg/errors"
)

type stateLogic interface {
	run(c *Client)
}

var stateLogics = [...]stateLogic{
	stateInit:  &stateInitLogic{},
	stateLogin: &stateLoginLogic{},
	statePlay:  &statePlayLogic{},
}

// stateInitLogic stateInit 状态逻辑.
type stateInitLogic struct {
	signKey         any
	onceLoadSignKey sync.Once
}

func (s *stateInitLogic) run(c *Client) {
	// 生成用户userToken
	userToken, err := s.genUserToken()
	if err != nil {
		log.Fatalf("gen user token failed: %v", err)
	}

	// 请求角色列表
	characterList, err := s.getCharacterList(userToken)
	if err != nil {
		log.Fatalf("get character list failed: %v", err)
	}
	log.Printf("character list: %+v", characterList)

	// 选择角色
	var characters []httpproto.CharacterInfo
	var characterId int64
	for _, v := range characterList {
		if v.ServerID == serverId {
			characters = append(characters, v)
		}
	}
	if len(characters) > 0 {
		characterId = characters[0].ID
	} else {
		var err error
		characterId, err = s.createCharacter(userToken, serverId)
		if err != nil {
			log.Fatalf("create character failed: %v", err)
		}
	}

	// 获取登录令牌.
	agentToken, err := s.loginCharacter(userToken, characterId)
	if err != nil {
		log.Fatalf("login character failed: %v", err)
	}

	// 切换到登录状态.
	c.agentToken = agentToken
	c.changeState(stateLogin)
}

// getSignKey 获取签名密钥.
func (s *stateInitLogic) getSignKey() any {
	s.onceLoadSignKey.Do(func() {
		priKey, err := authjwt.LoadPrivKey(env.SignKeyPath)
		if err != nil {
			log.Fatalf("load sign key failed: %v", err)
		}
		s.signKey = priKey
	})
	return s.signKey
}

// genUserToken 生成用户token.
func (s *stateInitLogic) genUserToken() (string, error) {
	info := &models.UserInfo{
		UID: uid,
	}
	sub, err := json.Marshal(info)
	if err != nil {
		return "", pkgerrors.WithMessage(err, "marshal user info")
	}
	return authjwt.SignToken(s.getSignKey(), cenv.All().Stage(), string(sub), 5*time.Minute, time.Now())
}

// getCharacterList 获取角色列表.
func (s *stateInitLogic) getCharacterList(token string) ([]httpproto.CharacterInfo, error) {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	return utils.GetCharacterList(ctx, env.LoginURLRoot, token)
}

// createCharacter 创建角色.
func (s *stateInitLogic) createCharacter(token string, serverId int64) (int64, error) {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	return utils.CreateCharacter(ctx, env.LoginURLRoot, token, serverId)
}

// loginCharacter 登录角色, 返回登录令牌.
func (s *stateInitLogic) loginCharacter(token string, characterId int64) (string, error) {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	return utils.LoginCharacter(ctx, env.LoginURLRoot, token, characterId)
}

// stateLoginLogic stateLogin 状态逻辑.
type stateLoginLogic struct{}

func (s *stateLoginLogic) run(c *Client) {
	// 连接网关.
	if err := c.connectAgent(); err != nil {
		log.Fatalf("connect agent failed: %v", err)
	}
	log.Println("connect agent successfully.")

	// 登录.
	_, err := sendReq[*pbc2s.LoginReq, *pbc2s.LoginResp](c, &pbc2s.LoginReq{
		Token: c.agentToken,
	})
	if err != nil {
		log.Fatalf("login failed, %v", err)
	}
	log.Println("login successfully.")

	c.changeState(statePlay)
}

// statePlayLogic 游玩状态逻辑.
type statePlayLogic struct {
	line          *readline.Instance
	autoCompleter *readline.PrefixCompleter
}

func (s *statePlayLogic) run(c *Client) {
	// 启动心跳协程
	go c.tick()

	// 构造自动补全
	completers := make([]readline.PrefixCompleterInterface, len(cmdList))
	for i, cmd := range cmdList {
		completers[i] = cmd.autoCompleter
	}
	s.autoCompleter = readline.NewPrefixCompleter(completers...)

	// 构造readline配置
	cfg := &readline.Config{
		Prompt:          "client> ",
		HistoryFile:     "./bin/.ggs_client_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		FuncIsTerminal:  nil,
		AutoComplete:    s.autoCompleter,
	}

	// 构造readline实例
	var err error
	s.line, err = readline.NewEx(cfg)
	if err != nil {
		log.Fatalf("int readline failed, %v", err)
	}
	defer s.line.Close()
	s.line.CaptureExitSignal()
	log.SetOutput(s.line.Stderr())

	// 读取命令行输入并执行对应命令
	for {
		line, err := s.line.Readline()
		if err == readline.ErrInterrupt {
			// Ctrl-C: 忽略并继续
			continue
		}
		if err == io.EOF {
			// Ctrl-D: 退出
			return
		}

		if s.exec(c, line) {
			os.Exit(0)
			break
		}
	}
}

// exec 执行一行命令, 返回是否退出。
func (s *statePlayLogic) exec(cli *Client, line string) bool {
	var (
		cmd  string
		args string
	)

	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	// 提取命令和参数部分
	cmd = line
	if n := strings.Index(line, " "); n > 0 {
		cmd = line[:n]
		if n < len(line)-1 {
			args = line[n+1:]
		}
	}

	// 获取并执行命令
	c := cmdMap[cmd]
	if c == nil {
		log.Printf("unknown command: %s", cmd)
		return false
	}
	return c.exec(c, cli, args)
}
