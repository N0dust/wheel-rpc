package service

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	_assert(len(s.Method) == 1, "wrong Service Method, expect 1, but got %d", len(s.Method))
	mType := s.Method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	methodType := s.Method["Sum"]

	argv := methodType.NewArgValue()
	replyValue := methodType.NewReplyValue()

	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.Call(methodType, argv, replyValue)

	out := replyValue.Interface().(*int)
	fmt.Println("out is", *out)

	_assert(err == nil && *replyValue.Interface().(*int) == 4 && methodType.NumCalls() == 1, "failed to Call Foo.Sum")
}
