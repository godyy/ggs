// Package kv 提供带类型信息的键值映射.
//
// 键由 K 表示, 携带值类型 V 和一个可读名称; 映射由 M 表示.
// 通过 K 存取时会做运行期类型校验, 类型不匹配则 panic.
//
// 注意: K 的判等基于内部字符串指针的对象身份, 而非指针指向的字符串内容,
// 因此同一个键必须复用同一个 K 实例(通常定义为包级变量), 且只能由 NewK 创建.
//
// M 不是并发安全的, 并发读写需调用方自行加锁.
package kv

import (
	"fmt"
	"maps"
	"reflect"

	"github.com/godyy/gutils/deepcopy"
)

// KI 键的接口, 由 K 实现.
//
// 接口内含未导出方法, 因此只有本包内的类型能实现它(封闭接口), 用于擦除键的泛型信息后统一处理.
type KI interface {
	// value 返回映射实际使用的键值.
	value() any

	// check 校验值 v 的类型是否与键匹配, 不匹配则 panic.
	check(v any)

	// String 返回键的字符串描述, 实现 fmt.Stringer.
	String() string
}

// MI 映射的接口, 由 M 实现.
type MI interface {
	// Set 设置键 k 对应的值为 v, 键不存在时新增.
	Set(k, v any)

	// Get 返回键 k 对应的值与是否存在.
	// 键不存在时返回 nil, false; 若 k 是 K, 还会校验存储值的类型, 不匹配则 panic.
	Get(k any) (any, bool)

	// Must 返回键 k 对应的值, 键不存在时 panic.
	Must(k any) any
}

// K 带类型信息的键, 其值类型为 V.
//
// 内嵌的 *string 用于携带可读名称并区分不同的键实例.
// K 的判等基于该指针的对象身份, 因此用相同名称调用 NewK 两次得到的是不同的键,
// 同一个键必须复用同一个 K 实例(通常定义为包级变量).
//
// K 的零值不可用, 必须通过 NewK 创建, 否则存取时 panic.
type K[V any] struct {
	*string
}

// NewK 创建一个值类型为 V、名称为 s 的键.
// 键的实际名称会被格式化为 "s(类型)" 仅用于错误信息等可读性用途, 不参与判等;
// 每次调用都返回一个新的键.
func NewK[V any](s string) K[V] {
	k := K[V]{string: new(string)}
	*k.string = fmt.Sprintf("%s(%s)", s, reflect.TypeFor[V]())
	return k
}

// mustInit 校验键已由 NewK 创建, 否则 panic.
// 零值 K 的内部指针为 nil, 若不拦截会与其他零值 K 冲突.
func (k K[V]) mustInit() {
	if k.string == nil {
		panic("kv: K: 键未初始化, 请使用 NewK 创建")
	}
}

// value 返回用作映射键的内部标识, 实现 KI.
func (k K[V]) value() any {
	k.mustInit()
	return k.string
}

// check 校验值 v 的类型是否为 V, 实现 KI.
// v 为 nil 时仅当 V 可赋 nil 才通过; 校验失败则 panic.
func (k K[V]) check(v any) {
	k.mustInit()

	// 常见路径: 类型匹配, 只需一次类型断言
	if _, ok := v.(V); ok {
		return
	}
	// v 为 nil 时, 只要 V 可赋 nil 就允许
	if v == nil && canBeNil[V]() {
		return
	}

	panic(fmt.Sprintf("kv: K.check: key %s, value %v is not of type %v", *k.string, v, reflect.TypeFor[V]()))
}

// String 返回键的字符串描述, 实现 fmt.Stringer.
// 零值键未初始化, 调用会 panic.
func (k K[V]) String() string {
	k.mustInit()
	return *k.string
}

// Set 通过键 k 在映射 m 上设置值 v, 键不存在时新增.
func (k K[V]) Set(m MI, v V) {
	m.Set(k, v)
}

// Get 返回键 k 在映射 m 中对应的值.
// exist 表示键 k 是否存在; 当存储的值为 nil 时返回零值与 true.
func (k K[V]) Get(m MI) (v V, exist bool) {
	if vv, exist := m.Get(k); !exist {
		return v, false
	} else if vv == nil {
		return v, true
	} else {
		return vv.(V), true
	}
}

// Must 返回键 k 在映射 m 中对应的值, 键不存在时 panic.
// 与 Get 一致, 存储的值为 nil 时返回零值.
func (k K[V]) Must(m MI) V {
	vv := m.Must(k)
	if vv == nil {
		var v V
		return v
	}
	return vv.(V)
}

// M 键值映射.
//
// 键可以是 K, 也可以是任意可比较类型; K 类型的键按内部标识存取, 其余类型的键直接用作 map 键.
// 非 K 的键必须是可比较(comparable)类型, 否则存取时 panic.
//
// M 不暴露底层 map, 只能通过本类型的方法访问, 以保证类型校验不被绕过.
// M 不是并发安全的, 并发读写需调用方自行加锁.
// M 的零值可直接使用.
type M struct {
	m map[any]any
}

// NewM 按键值交替的顺序 p0, v0, p1, v1, ... 创建映射.
// pairs 的元素个数必须为偶数, 否则 panic; 重复的键会静默覆盖.
func NewM(pairs ...any) *M {
	m := &M{}
	m.SetPairs(pairs...)
	return m
}

// Len 返回映射中键值对的数量.
func (m *M) Len() int {
	return len(m.m)
}

// Set 设置键 k 对应的值为 v, 键不存在时新增.
// 若 k 是 K, 会先校验 v 的类型, 不匹配则 panic.
func (m *M) Set(k, v any) {
	if ki, ok := k.(KI); ok {
		ki.check(v)
		k = ki.value()
	}
	if m.m == nil {
		m.m = make(map[any]any)
	}
	m.m[k] = v
}

// SetPairs 设置多个键值对.
// pairs 的元素个数必须为偶数, 否则 panic; 重复的键会静默覆盖.
func (m *M) SetPairs(pairs ...any) {
	if len(pairs)%2 != 0 {
		panic("kv: M.SetPairs: pairs must be even")
	}
	if m.m == nil {
		m.m = make(map[any]any, len(pairs)/2)
	}
	for i := 0; i < len(pairs); i += 2 {
		m.Set(pairs[i], pairs[i+1])
	}
}

// Get 返回键 k 对应的值与是否存在.
// 键不存在时返回 nil, false; 若 k 是 K, 还会校验存储值的类型, 不匹配则 panic.
func (m *M) Get(k any) (any, bool) {
	if ki, ok := k.(KI); ok {
		v, exist := m.m[ki.value()]
		if !exist {
			return nil, false
		}
		ki.check(v)
		return v, true
	}
	v, exist := m.m[k]
	return v, exist
}

// Must 返回键 k 对应的值, 键不存在时 panic.
// 若 k 是 K, 还会校验存储值的类型, 不匹配则 panic.
func (m *M) Must(k any) any {
	v, exist := m.Get(k)
	if !exist {
		panic(fmt.Sprintf("kv: M.Must: key %v not found", k))
	}
	return v
}

// Del 删除键 k 对应的键值对.
// 若 k 是 K, 会先校验存储值的类型, 不匹配则 panic.
func (m *M) Del(k any) (any, bool) {
	ki, ok := k.(KI)
	if !ok {
		v, exist := m.m[k]
		if !exist {
			return nil, false
		}
		delete(m.m, k)
		return v, true
	}

	// K 类型的键: 命中后校验类型, 再按内部标识删除
	kk := ki.value()
	v, exist := m.m[kk]
	if !exist {
		return nil, false
	}
	ki.check(v)
	delete(m.m, kk)
	return v, true
}

// Clear 清空映射中的所有键值对, 清空后 Len 为 0, 映射仍可继续使用.
func (m *M) Clear() {
	clear(m.m)
}

// Clone 返回 m 的浅拷贝, 键值对相同但底层 map 相互独立, 修改其一不影响另一个.
// 浅拷贝意味着值为指针、slice、map 等引用类型时, 副本与原映射仍共享同一份底层数据.
func (m *M) Clone() *M {
	return &M{m: maps.Clone(m.m)}
}

// DeepClone 返回 m 的深拷贝: 键与原映射共享, 值则通过 deepcopy.Copy 递归深拷贝, 修改其一不影响另一个.
// 值的动态类型若实现了 deepcopy.CopyGenericI, 会优先调用其 DeepCopy 方法.
func (m *M) DeepClone() *M {
	if m.m == nil {
		return &M{}
	}
	c := &M{m: make(map[any]any, len(m.m))}
	for k, v := range m.m {
		c.m[k] = deepcopy.Copy(v)
	}
	return c
}

// canBeNil 判断类型 T 的零值是否为 nil, 即 T 能否被赋值为 nil.
// 只有 chan、func、interface、map、pointer、slice、unsafe.Pointer 这几种类型可赋 nil.
func canBeNil[T any]() bool {
	switch reflect.TypeFor[T]().Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return true
	default:
		return false
	}
}

// Get 返回键 k 在映射 m 中对应的值, 等价于 k.Get(m).
func Get[V any](m MI, k K[V]) (v V, exist bool) {
	return k.Get(m)
}

// Must 返回键 k 在映射 m 中对应的值, 键不存在时 panic, 等价于 k.Must(m).
func Must[V any](m MI, k K[V]) V {
	return k.Must(m)
}
