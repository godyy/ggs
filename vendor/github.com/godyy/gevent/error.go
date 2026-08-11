package gevent

import (
	"errors"
	"fmt"
	"strings"
)

// listenPanicError 用于封装监听回调触发的 panic。
type listenPanicError struct {
	value any // panic 值
}

func (e *listenPanicError) Error() string {
	return fmt.Sprintf("listen panic: %v", e.value)
}

// wrapperTypeError 用于描述包装器在转换构造者或参数类型时的失败。
type wrapperTypeError struct {
	field    string // 转换失败的字段名
	expected string // 期望类型
	actual   string // 实际类型
}

func (e *wrapperTypeError) Error() string {
	return fmt.Sprintf("wrap event %s type mismatch: expect %s, got %s", e.field, e.expected, e.actual)
}

// dispatchError 用于封装派发事件时监听返回的错误
// 记录派发的类别、派发时的事件ID
type dispatchError[EK, EV comparable] struct {
	dt      string          // 派发的类别
	eventID EventID[EK, EV] // 事件ID
	err     error           // error
}

func (e *dispatchError[EK, EV]) Error() string {
	return fmt.Sprintf("dispatch %s event of id={kind:%v, value:%v}: %v", e.dt, e.eventID.Kind, e.eventID.Value, e.err.Error())
}

func (e *dispatchError[EK, EV]) Is(o error) bool {
	return errors.Is(e.err, o)
}

func (e *dispatchError[EK, EV]) As(o any) bool {
	return errors.As(e.err, o)
}

// dispatchErrors 用于统计派发事件时监听们返回的所有错误
type dispatchErrors struct {
	errors []error
}

func (e *dispatchErrors) Error() string {
	sb := strings.Builder{}
	sb.WriteString("[")
	for i, err := range e.errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	sb.WriteString("]")
	return sb.String()
}

func (e *dispatchErrors) Is(o error) bool {
	for _, err := range e.errors {
		if errors.Is(err, o) {
			return true
		}
	}
	return false
}

func (e *dispatchErrors) As(o any) bool {
	for _, err := range e.errors {
		if errors.As(err, o) {
			return true
		}
	}
	return false
}
