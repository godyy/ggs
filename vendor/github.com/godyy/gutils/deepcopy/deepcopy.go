package deepcopy

import "reflect"

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
	srcType := srcValue.Type()
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		if !field.IsExported() {
			continue
		}
		srcField := srcValue.Field(i)
		destField := destValue.Field(i)
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
