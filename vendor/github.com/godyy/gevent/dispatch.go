package gevent

// Dispatcher 事件派发器
// 用于为特定类型或特定值类型的事件添加监听，并在产生事件时将事件派发给监听
// 当前实现为同步派发，且不是 goroutine-safe。
type Dispatcher[EK, EV comparable] struct {
	kindContainers map[EK]*kindContainer[EK, EV] // 按照事件类型划分的监听容器
	lidIncr        int32                         // 监听ID递增器
	dispatching    int                           // 派发状态计数
}

func NewDispatcher[EK, EV comparable]() *Dispatcher[EK, EV] {
	return &Dispatcher[EK, EV]{
		kindContainers: map[EK]*kindContainer[EK, EV]{},
	}
}

// nextLID 获取下一个监听ID
func (d *Dispatcher[EK, EV]) nextLID() LID {
	if d.lidIncr == maxLID {
		panic("lid overflow")
	}
	d.lidIncr++
	return LID(d.lidIncr)
}

// addORGetKindContainer 添加或获取事件类型监听容器
func (d *Dispatcher[EK, EV]) addORGetKindContainer(evtKind EK) *kindContainer[EK, EV] {
	kc := d.kindContainers[evtKind]
	if kc == nil {
		kc = newKindContainer[EK, EV]()
		d.kindContainers[evtKind] = kc
	}
	return kc
}

// ListenKind 监听事件类型
func (d *Dispatcher[EK, EV]) ListenKind(ek EK, cb ListenCB[EK, EV], once bool) LID {
	if cb == nil {
		panic("listen callback nil")
	}
	lid := d.nextLID()
	l := newListen(lid, cb)
	kc := d.addORGetKindContainer(ek)
	kc.addKind(l, once)
	return lid
}

// ListenKindOnce 监听事件类型一次
func (d *Dispatcher[EK, EV]) ListenKindOnce(ek EK, cb ListenCB[EK, EV]) LID {
	return d.ListenKind(ek, cb, true)
}

// UnlistenKind 取消监听事件类型
func (d *Dispatcher[EK, EV]) UnlistenKind(evtKind EK, lid LID) {
	kc := d.kindContainers[evtKind]
	if kc == nil {
		return
	}
	kc.remKind(lid)
	if kc.empty() {
		delete(d.kindContainers, evtKind)
	}
}

// Listen 监听事件
func (d *Dispatcher[EK, EV]) Listen(evtId EventID[EK, EV], cb ListenCB[EK, EV], once bool) LID {
	if cb == nil {
		panic("listen callback nil")
	}
	lid := d.nextLID()
	l := newListen(lid, cb)
	kc := d.addORGetKindContainer(evtId.Kind)
	kc.addValue(evtId.Value, l, once)
	return lid
}

// ListenOnce 监听事件一次
func (d *Dispatcher[EK, EV]) ListenOnce(evtId EventID[EK, EV], cb ListenCB[EK, EV]) LID {
	return d.Listen(evtId, cb, true)
}

// Unlisten 取消监听事件
func (d *Dispatcher[EK, EV]) Unlisten(evtId EventID[EK, EV], lid LID) {
	kc := d.kindContainers[evtId.Kind]
	if kc == nil {
		return
	}
	kc.remValue(evtId.Value, lid)
	if kc.empty() {
		delete(d.kindContainers, evtId.Kind)
	}
}

// Clear 清理状态，移除所有监听
func (d *Dispatcher[EK, EV]) Clear() {
	if d.dispatching > 0 {
		// 派发过程中不能清理
		return
	}

	for _, v := range d.kindContainers {
		v.clear()
	}
	d.kindContainers = map[EK]*kindContainer[EK, EV]{}
}

// Dispatch 构造事件，派发给 evtID 指定的监听.
func (d *Dispatcher[EK, EV]) Dispatch(evtId EventID[EK, EV], creator any, param any) error {
	kc := d.kindContainers[evtId.Kind]
	if kc == nil {
		return nil
	}

	d.dispatching++
	defer func() {
		if kc.empty() {
			delete(d.kindContainers, evtId.Kind)
		}
		d.dispatching--
		if d.dispatching < 0 {
			d.dispatching = 0
		}
	}()

	evt := Event[EK, EV]{
		eventID: evtId,
		params:   param,
		creator: creator,
	}
	return kc.invoke(evt)
}
