package service

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

type MethodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *MethodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *MethodType) NewArgValue() reflect.Value {
	var argv reflect.Value
	// arg may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *MethodType) NewReplyValue() reflect.Value {
	// reply must be a pointer type
	replyV := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyV.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyV.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	default:
		k := m.ReplyType.Elem().Kind()
		fmt.Println("reply type is", k)
	}
	return replyV
}
