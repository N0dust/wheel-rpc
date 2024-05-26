package service

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type Service struct {
	Name     string
	typ      reflect.Type
	receiver reflect.Value
	Method   map[string]*MethodType
}

func NewService(receiver interface{}) *Service {
	s := new(Service)
	s.receiver = reflect.ValueOf(receiver)
	s.Name = reflect.Indirect(s.receiver).Type().Name()
	s.typ = reflect.TypeOf(receiver)
	if !ast.IsExported(s.Name) {
		log.Fatalf("rpc server: %s is not a valid Service Name", s.Name)
	}
	s.registerMethods()
	return s
}

func (s *Service) registerMethods() {
	s.Method = make(map[string]*MethodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		methodType := method.Type
		// 两个导出或内置类型的入参  反射时为 3 个，第 0 个是自身 返回值有且只有 1 个，类型为 error
		if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
			continue
		}
		if methodType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := methodType.In(1), methodType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.Method[method.Name] = &MethodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.Name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *Service) Call(m *MethodType, argv, replyValue reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.receiver, argv, replyValue})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
