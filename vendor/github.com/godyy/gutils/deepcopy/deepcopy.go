package deepcopy

import (
	"reflect"
	"unsafe"
)

// CopyGenericI 泛型深度拷贝接口
// 实现该接口，深度拷贝会优先调用该接口
type CopyGenericI[T any] interface {
	DeepCopy() T
}

// Copy 通过反射深度拷贝任意类型对象
func Copy(src interface{}) interface{} {
	if src == nil {
		return src
	}

	if dc, ok := src.(CopyGenericI[interface{}]); ok {
		return dc.DeepCopy()
	}

	var (
		srcValue  = reflect.ValueOf(src)
		destValue = reflect.Zero(srcValue.Type())
	)

	deepCopyRecursive(&destValue, &srcValue)

	return destValue.Interface()
}

// CopyGeneric 泛型深度拷贝
func CopyGeneric[T any](src T) T {
	var i interface{} = src

	if i, ok := i.(CopyGenericI[T]); ok && i != nil {
		return i.DeepCopy()
	}

	dest := Copy(src)
	return dest.(T)
}

func deepCopyRecursive(destValue, srcValue *reflect.Value) {
	// 未导出字段的源值带只读限制, 转为可读视图再处理
	src := readable(*srcValue)
	if !src.CanInterface() {
		// 只读且不可寻址, 无法读取, 跳过
		return
	}
	srcValue = &src

	if srcValue.IsZero() {
		return
	}

	if !destValue.CanSet() {
		*destValue = reflect.New(srcValue.Type()).Elem()
	}

	switch srcValue.Kind() {
	case reflect.Struct:
		// 结构体类型
		deepCopyStruct(destValue, srcValue)

	case reflect.Array:
		// 数组类型
		deepCopyArray(destValue, srcValue)

	case reflect.Slice:
		// 切片类型
		deepCopySlice(destValue, srcValue)

	case reflect.Map:
		// 映射类型
		deepCopyMap(destValue, srcValue)

	case reflect.Pointer:
		// 指针类型
		deepCopyPointer(destValue, srcValue)

	case reflect.Interface:
		// 接口类型
		deepCopyInterface(destValue, srcValue)

	default:
		// 基本类型（包括 int, string, bool 等）、chan 直接赋值
		destValue.Set(*srcValue)
	}
}

func deepCopyStruct(destValue, srcValue *reflect.Value) {
	// 源结构体需可寻址, 才能通过 unsafe 读取其未导出字段
	src := addressable(*srcValue)
	dest := *destValue
	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		// 未导出字段不可读写, 通过 unsafe 去掉只读限制
		srcField := readable(src.Field(i))
		destField := writable(dest.Field(i))
		deepCopyRecursive(&destField, &srcField)
	}
}

func deepCopyArray(destValue, srcValue *reflect.Value) {
	for i := 0; i < srcValue.Len(); i++ {
		srcElem := srcValue.Index(i)
		destElem := destValue.Index(i)
		deepCopyRecursive(&destElem, &srcElem)
	}
}

func deepCopySlice(destValue, srcValue *reflect.Value) {
	destValue.Set(reflect.MakeSlice(srcValue.Type(), srcValue.Len(), srcValue.Cap()))
	for i := 0; i < srcValue.Len(); i++ {
		srcElem := srcValue.Index(i)
		destElem := destValue.Index(i)
		deepCopyRecursive(&destElem, &srcElem)
	}
}

func deepCopyMap(destValue, srcValue *reflect.Value) {
	destValue.Set(reflect.MakeMapWithSize(srcValue.Type(), srcValue.Len()))
	mapIter := srcValue.MapRange()
	for mapIter.Next() {
		k := mapIter.Key()
		v := mapIter.Value()
		vv := reflect.Zero(v.Type())
		deepCopyRecursive(&vv, &v)
		destValue.SetMapIndex(k, vv)
	}
}

func deepCopyPointer(destValue, srcValue *reflect.Value) {
	destValue.Set(reflect.New(srcValue.Type().Elem()))
	srcElem := srcValue.Elem()
	destElem := destValue.Elem()
	deepCopyRecursive(&destElem, &srcElem)
}

func deepCopyInterface(destValue, srcValue *reflect.Value) {
	srcElem := srcValue.Elem()
	destElem := reflect.New(srcElem.Type()).Elem()
	deepCopyRecursive(&destElem, &srcElem)
	destValue.Set(destElem)
}

// addressable 返回 v 的可寻址视图:
// v 已可寻址则原样返回, 否则复制一份到新分配的内存, 以便通过 unsafe 读取其未导出字段.
func addressable(v reflect.Value) reflect.Value {
	// 只读(不可 interface)的值无法通过 Set 复制, 原样返回由上层兜底
	if v.CanAddr() || !v.CanInterface() {
		return v
	}
	p := reflect.New(v.Type()).Elem()
	p.Set(v)
	return p
}

// readable 返回 v 的可读视图, 用于绕过未导出字段带来的只读限制.
// v 需可寻址, 否则原样返回.
func readable(v reflect.Value) reflect.Value {
	if v.CanInterface() || !v.CanAddr() {
		return v
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

// writable 返回 v 的可写视图, 用于绕过未导出字段带来的只读限制.
// v 需可寻址, 否则原样返回.
func writable(v reflect.Value) reflect.Value {
	if v.CanSet() || !v.CanAddr() {
		return v
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}
