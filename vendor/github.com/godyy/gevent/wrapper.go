package gevent

import (
	"reflect"
)

// EventW 包装后的事件.
type EventW[EK, EV comparable, Ctr, Params any] struct {
	EventID EventID[EK, EV] // 事件ID
	LID     LID             // 监听ID
	Creator Ctr             // 构造者
	Params  Params          // 参数
}

// wrapEvent 包装事件.
func wrapEvent[EK, EV comparable, Ctr, Params any](e Event[EK, EV]) (ew EventW[EK, EV, Ctr, Params], err error) {
	creator, ok := e.Creator().(Ctr)
	if !ok {
		err = &wrapperTypeError{
			field:    "creator",
			expected: reflect.TypeFor[Ctr]().Name(),
			actual:   reflect.TypeOf(e.Creator()).Name(),
		}
		return
	}

	var params Params
	vp := reflect.ValueOf(e.Params())
	tp := reflect.TypeFor[Params]()
	if vp.IsValid() {
		if !vp.CanConvert(tp) {
			err = &wrapperTypeError{
				field:    "params",
				expected: tp.Name(),
				actual:   vp.Type().Name(),
			}
			return
		}
		params = vp.Convert(tp).Interface().(Params)
	} else if tp.Kind() != reflect.Chan &&
		tp.Kind() != reflect.Func &&
		tp.Kind() != reflect.Interface &&
		tp.Kind() != reflect.Map &&
		tp.Kind() != reflect.Pointer &&
		tp.Kind() != reflect.Slice {
		err = &wrapperTypeError{
			field:    "params",
			expected: tp.Name(),
			actual:   "invalid type",
		}
		return
	}

	ew = EventW[EK, EV, Ctr, Params]{
		EventID: e.EventID(),
		LID:     e.LID(),
		Creator: creator,
		Params:  params,
	}
	return
}

// Wrapper 包装器.
type Wrapper[EK, EV comparable, Ctr, Params any] struct {
	*Dispatcher[EK, EV]
}

// NewWrapper 创建包装器.
func NewWrapper[EK, EV comparable, Ctr, Params any](d *Dispatcher[EK, EV]) Wrapper[EK, EV, Ctr, Params] {
	return Wrapper[EK, EV, Ctr, Params]{Dispatcher: d}
}

func (w Wrapper[EK, EV, Ctr, Params]) ListenKind(k EK, cb func(e EventW[EK, EV, Ctr, Params]) error, once bool) LID {
	return w.Dispatcher.ListenKind(k, func(e Event[EK, EV]) error {
		ew, err := wrapEvent[EK, EV, Ctr, Params](e)
		if err != nil {
			return err
		}
		return cb(ew)
	}, once)
}

func (w Wrapper[EK, EV, Ctr, Params]) ListenKindOnce(k EK, cb func(e EventW[EK, EV, Ctr, Params]) error) LID {
	return w.ListenKind(k, cb, true)
}

func (w Wrapper[EK, EV, Ctr, Params]) Listen(eid EventID[EK, EV], cb func(e EventW[EK, EV, Ctr, Params]) error, once bool) LID {
	return w.Dispatcher.Listen(eid, func(e Event[EK, EV]) error {
		ew, err := wrapEvent[EK, EV, Ctr, Params](e)
		if err != nil {
			return err
		}
		return cb(ew)
	}, once)
}

func (w Wrapper[EK, EV, Ctr, Params]) ListenOnce(eid EventID[EK, EV], cb func(e EventW[EK, EV, Ctr, Params]) error) LID {
	return w.Listen(eid, cb, true)
}

func (w Wrapper[EK, EV, Ctr, Params]) Dispatch(eid EventID[EK, EV], creator Ctr, params Params) error {
	return w.Dispatcher.Dispatch(eid, creator, params)
}
