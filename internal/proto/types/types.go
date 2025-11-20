package types

import (
	"errors"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// Protos 用于注册和创建protobuf协议结构体
type Protos struct {
	pid2Type map[uint16]reflect.Type
	type2Pid map[reflect.Type]uint16
	name2Pid map[string]uint16
}

// NewProtos 创建 Protos.
func NewProtos() *Protos {
	return &Protos{
		pid2Type: make(map[uint16]reflect.Type),
		type2Pid: make(map[reflect.Type]uint16),
		name2Pid: make(map[string]uint16),
	}
}

// Register 注册协议类型
// pid: 协议ID
// proto: 协议结构体指针
func (p *Protos) Register(pid uint16, proto proto.Message) error {
	if proto == nil {
		return errors.New("proto is nil")
	}

	typ := reflect.TypeOf(proto)
	if typ.Kind() != reflect.Ptr {
		return errors.New("proto must be pointer")
	}

	if _, exists := p.pid2Type[pid]; exists {
		return fmt.Errorf("pid %d already registered", pid)
	}

	elemTyp := typ.Elem()
	p.pid2Type[pid] = elemTyp
	p.type2Pid[elemTyp] = pid
	p.name2Pid[elemTyp.Name()] = pid
	return nil
}

// GetPid 通过协议类型获取对象的协议ID.
func (p *Protos) GetPid(proto proto.Message) (uint16, bool) {
	typ := reflect.TypeOf(proto)
	if typ.Kind() != reflect.Ptr {
		return 0, false
	}

	elemTyp := typ.Elem()
	pid, exists := p.type2Pid[elemTyp]
	if !exists {
		return 0, false
	}

	return pid, true
}

// GetPidByName 通过协议类型名称获取协议ID.
func (p *Protos) GetPidByName(name string) (uint16, bool) {
	pid, exists := p.name2Pid[name]
	if !exists {
		return 0, false
	}

	return pid, true
}

// Create 通过协议ID创建协议实体
func (p *Protos) Create(pid uint16) (proto.Message, error) {
	typ, exists := p.pid2Type[pid]
	if !exists {
		return nil, fmt.Errorf("pid %d not registered", pid)
	}

	// 创建协议实体
	inst := reflect.New(typ).Interface().(proto.Message)
	return inst, nil
}

// CreateByName 通过协议类型名称创建协议实体
func (p *Protos) CreateByName(name string) (proto.Message, uint16, error) {
	pid, exists := p.name2Pid[name]
	if !exists {
		return nil, 0, fmt.Errorf("%s not registered", name)
	}

	return reflect.New(p.pid2Type[pid]).Interface().(proto.Message), pid, nil
}

// Check 检查协议ID和协议类型是否匹配.
func (p *Protos) Check(pid uint16, proto proto.Message) error {
	pidTyp, exists := p.pid2Type[pid]
	if !exists {
		return fmt.Errorf("pid %d not registered", pid)
	}

	pTyp := reflect.TypeOf(proto)
	if pTyp.Kind() != reflect.Ptr {
		return errors.New("proto must be pointer")
	}
	pTyp = pTyp.Elem()

	if pTyp != pidTyp {
		return errors.New("proto type not match")
	}

	return nil
}
