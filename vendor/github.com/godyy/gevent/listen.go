package gevent

import (
	"container/list"
	"errors"
)

// ErrRemAfterDispatch 派发后移除
// 监听专用，返回该 error 即告诉派发器，在本次派发完成后移除它
var ErrRemAfterDispatch = errors.New("remove after dispatch")

// ListenCB 监听回调
type ListenCB[EK, EV comparable] func(Event[EK, EV]) error

// listen 监听
// 记录监听信息
type listen[EK, EV comparable] struct {
	id LID              // 监听ID
	cb ListenCB[EK, EV] // 监听回调
}

func newListen[EK, EV comparable](id LID, cb ListenCB[EK, EV]) *listen[EK, EV] {
	if cb == nil {
		panic("callback nil")
	}
	return &listen[EK, EV]{
		id: id,
		cb: cb,
	}
}

// call 调用监听回调
// 返回监听产生的错误
func (l *listen[EK, EV]) call(evt Event[EK, EV]) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = &listenPanicError{value: rec}
		}
	}()
	evt.lid = l.id
	return l.cb(evt)
}

// reset 重置数据，解除引用
func (l *listen[EK, EV]) reset() {
	l.cb = nil
}

// listenContainer 监听容器
// 每一个独立的事件，都有与之对应的监听容器来维护相关的监听
type listenContainer[EK, EV comparable] struct {
	l           *list.List            // 监听列表
	m           map[LID]*list.Element // 监听 Elem Map
	mOnce       map[LID]struct{}      // 单次监听标记，记录只监听一次的监听
	lPendingRem *list.List            // 挂起移除列表，等待在事件派发完成后被移除的监听
	mPendingRem map[LID]struct{}      // 挂起移除标记，避免重复挂起
	dispatching int                   // 派发状态计数
}

func newListenContainer[EK, EV comparable]() *listenContainer[EK, EV] {
	return &listenContainer[EK, EV]{
		l: list.New(),
		m: map[LID]*list.Element{},
	}
}

// add 添加监听
// 不能重复添加相同ID的监听
func (lc *listenContainer[EK, EV]) add(l *listen[EK, EV], once bool) {
	if lc.dispatching > 0 {
		// 不能在派发事件状态下，嵌套添加监听
		panic("add listen on dispatching")
	}
	if elem, ok := lc.m[l.id]; ok {
		panic("add listen id already exist")
	} else {
		elem = lc.l.PushBack(l)
		lc.m[l.id] = elem
		if once {
			if lc.mOnce == nil {
				lc.mOnce = map[LID]struct{}{}
			}
			lc.mOnce[l.id] = struct{}{}
		}
	}
}

// rem 移除监听
func (lc *listenContainer[EK, EV]) rem(lid LID) {
	elem, ok := lc.m[lid]
	if !ok {
		return
	}
	l := elem.Value.(*listen[EK, EV])
	if lc.dispatching > 0 {
		// 正在派发事件，所有需要移除的监听都需要挂起，等待派发完成后，统一移除
		lc.pendingRem(l, elem)
	} else {
		// 未派发事件，直接移除
		lc.directRem(l, elem)
	}
}

// directRem 直接移除监听
func (lc *listenContainer[EK, EV]) directRem(l *listen[EK, EV], elem *list.Element) {
	lc.l.Remove(elem)
	delete(lc.m, l.id)
	if lc.mOnce != nil {
		delete(lc.mOnce, l.id)
		if len(lc.mOnce) == 0 {
			lc.mOnce = nil
		}
	}
	l.reset()
}

// pendingRem 挂起移除监听
func (lc *listenContainer[EK, EV]) pendingRem(l *listen[EK, EV], elem *list.Element) {
	if lc.isPendingRem(l.id) {
		// 无须重复添加已经等待移除的监听
		return
	}
	if lc.lPendingRem == nil {
		lc.lPendingRem = list.New()
	}
	if lc.mPendingRem == nil {
		lc.mPendingRem = map[LID]struct{}{}
	}
	lc.mPendingRem[l.id] = struct{}{}
	lc.lPendingRem.PushBack(elem)
}

// isPendingRem 返回监听是否已进入挂起移除状态。
func (lc *listenContainer[EK, EV]) isPendingRem(lid LID) bool {
	if lc.mPendingRem == nil {
		return false
	}
	_, ok := lc.mPendingRem[lid]
	return ok
}

// isOnce 返回监听是否只监听一次。
func (lc *listenContainer[EK, EV]) isOnce(lid LID) bool {
	if lc.mOnce == nil {
		return false
	}
	_, ok := lc.mOnce[lid]
	return ok
}

// empty 返回是否没有监听
func (lc *listenContainer[EK, EV]) empty() bool {
	return lc.l.Len() == 0
}

// call 调用监听回调
// 返回监听们产生的错误
func (lc *listenContainer[EK, EV]) call(event Event[EK, EV]) (err error) {
	lc.dispatching++
	defer func() {
		lc.dispatching--
		if lc.dispatching < 0 {
			lc.dispatching = 0
		}

		if lc.dispatching == 0 {
			lc.flushPendingRem()
		}
	}()

	var errs []error
	elem := lc.l.Front()
	for elem != nil {
		l := elem.Value.(*listen[EK, EV])
		if !lc.isPendingRem(l.id) {
			dispatchErr := l.call(event)
			if dispatchErr != nil && dispatchErr != ErrRemAfterDispatch {
				errs = append(errs, dispatchErr)
			}
			if lc.isOnce(l.id) || dispatchErr == ErrRemAfterDispatch {
				lc.pendingRem(l, elem)
			}
		}
		elem = elem.Next()
	}

	if len(errs) > 0 {
		err = &dispatchErrors{errors: errs}
	}

	return err
}

// flushPendingRem 清理本轮派发结束后等待移除的监听。
func (lc *listenContainer[EK, EV]) flushPendingRem() {
	if lc.lPendingRem == nil {
		return
	}

	elem := lc.lPendingRem.Front()
	for elem != nil {
		remElem := elem.Value.(*list.Element)
		remListen := remElem.Value.(*listen[EK, EV])
		delete(lc.mPendingRem, remListen.id)
		lc.directRem(remListen, remElem)
		next := elem.Next()
		lc.lPendingRem.Remove(elem)
		elem = next
	}
	lc.lPendingRem = nil
	lc.mPendingRem = nil
}

// clear 清理容器，移除所有监听
func (lc *listenContainer[EK, EV]) clear() {
	if lc.dispatching > 0 {
		// 派发过程中无法清理
		return
	}
	lc.m = nil
	lc.mOnce = nil
	if lc.lPendingRem != nil {
		lc.lPendingRem.Init()
		elem := lc.lPendingRem.Front()
		for elem != nil {
			elem.Value = nil
			rem := elem
			elem = elem.Next()
			lc.lPendingRem.Remove(rem)
		}
		lc.lPendingRem = nil
	}
	lc.mPendingRem = nil
	if lc.l != nil {
		elem := lc.l.Front()
		for elem != nil {
			elem.Value.(*listen[EK, EV]).reset()
			elem.Value = nil
			rem := elem
			elem = elem.Next()
			lc.l.Remove(rem)
		}
		lc.l = nil
	}
}

// kindContainer 按事件类型划分的监听容器
type kindContainer[EK, EV comparable] struct {
	kindListens  *listenContainer[EK, EV]        // 类型事件监听
	valueListens map[EV]*listenContainer[EK, EV] // 值类事件监听
	dispatching  int                             // 派发状态计数
}

func newKindContainer[EK, EV comparable]() *kindContainer[EK, EV] {
	return &kindContainer[EK, EV]{}
}

// addKind 添加类型事件监听
func (kc *kindContainer[EK, EV]) addKind(l *listen[EK, EV], once bool) {
	if kc.dispatching > 0 {
		// 不能在派发事件状态下，嵌套添加监听
		panic("add kind listen on dispatching")
	}
	if kc.kindListens == nil {
		kc.kindListens = newListenContainer[EK, EV]()
	}
	kc.kindListens.add(l, once)
}

// remKind 移除类型事件监听
func (kc *kindContainer[EK, EV]) remKind(lid LID) {
	if kc.kindListens == nil {
		return
	}
	kc.kindListens.rem(lid)
	if kc.kindListens.empty() {
		kc.kindListens = nil
	}
}

// addValue 添加值类型事件监听
func (kc *kindContainer[EK, EV]) addValue(value EV, l *listen[EK, EV], once bool) {
	if kc.dispatching > 0 {
		// 不能在派发事件状态下，嵌套添加监听
		panic("add value listen on dispatching")
	}
	if kc.valueListens == nil {
		kc.valueListens = map[EV]*listenContainer[EK, EV]{}
	}
	lc := kc.valueListens[value]
	if lc == nil {
		lc = newListenContainer[EK, EV]()
		kc.valueListens[value] = lc
	}
	lc.add(l, once)
}

// remValue 移除值类型事件监听
func (kc *kindContainer[EK, EV]) remValue(value EV, lid LID) {
	lc := kc.valueListens[value]
	if lc != nil {
		lc.rem(lid)
		if lc.empty() {
			delete(kc.valueListens, value)
		}
		if len(kc.valueListens) == 0 {
			kc.valueListens = nil
		}
	}
}

// empty 返回是否没有监听
func (kc *kindContainer[EK, EV]) empty() bool {
	return (kc.kindListens == nil || kc.kindListens.empty()) && (kc.valueListens == nil || len(kc.valueListens) == 0)
}

// invoke 调用监听回调
// 先派发类型事件，再配发值类事件
func (kc *kindContainer[EK, EV]) invoke(evt Event[EK, EV]) (err error) {
	kc.dispatching++
	defer func() {
		kc.dispatching--
		if kc.dispatching < 0 {
			kc.dispatching = 0
		}
	}()
	var errs []error

	if kc.kindListens != nil {
		if kindErr := kc.kindListens.call(evt); kindErr != nil {
			errs = append(errs, &dispatchError[EK, EV]{
				dt:      "kind",
				eventID: evt.eventID,
				err:     kindErr,
			})
		}
		if kc.kindListens.empty() {
			kc.kindListens = nil
		}
	}

	if kc.valueListens != nil {
		lc := kc.valueListens[evt.eventID.Value]
		if lc != nil {
			if valueErr := lc.call(evt); valueErr != nil {
				errs = append(errs, &dispatchError[EK, EV]{
					dt:      "value",
					eventID: evt.eventID,
					err:     valueErr,
				})
			}
			if lc.empty() {
				delete(kc.valueListens, evt.eventID.Value)
				if len(kc.valueListens) == 0 {
					kc.valueListens = nil
				}
			}
		}
	}

	if len(errs) > 0 {
		err = &dispatchErrors{errors: errs}
	}

	return err
}

// clear 清理容器，移除所有监听
func (kc *kindContainer[EK, EV]) clear() {
	if kc.dispatching > 0 {
		// 派发中不能清理
		return
	}
	if kc.kindListens != nil {
		kc.kindListens.clear()
		kc.kindListens = nil
	}

	for _, v := range kc.valueListens {
		v.clear()
	}
	kc.valueListens = nil
}
