package server

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
	"wheel-rpc/coder"
	"wheel-rpc/service"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int
	SerializerType coder.Type
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	SerializerType: coder.GobType,
	ConnectTimeout: 10 * time.Second,
}

type Server struct {
	serviceMap sync.Map // 存了哪些注册的服务
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("")
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server:Decode opt: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Println("rpc server:invalid magic number ", opt.MagicNumber)
		return
	}
	f := coder.NewCoderFuncMap[opt.SerializerType]
	if f == nil {
		log.Println("rpc server:invalid coder type")
		return
	}
	s.ServeNewConn(f(conn), opt.HandleTimeout)
}

var invalidRequest = struct{}{}

func (s *Server) ServeNewConn(serializer coder.Coder, handleTimeout time.Duration) {
	sendMutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for {
		req, err := s.readRequest(serializer)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(serializer, req.h, invalidRequest, sendMutex)
			continue
		}
		wg.Add(1)
		go s.handleRequest(serializer, req, sendMutex, wg, handleTimeout)
	}
	wg.Wait()
	_ = serializer.Close()
}

func (s *Server) Register(receiver interface{}) error {
	newService := service.NewService(receiver)
	if _, dup := s.serviceMap.LoadOrStore(newService.Name, newService); dup {
		return errors.New("rpc: service already defined: " + newService.Name)
	}
	return nil
}

func (s *Server) FindService(serviceMethod string) (*service.Service, *service.MethodType, error) {
	splitPoint := strings.Index(serviceMethod, ".")
	if splitPoint < 0 {
		return nil, nil, errors.New("rpc server: serviceMethod has no . or invalid format")
	}
	serviceName, methodName := serviceMethod[:splitPoint], serviceMethod[splitPoint+1:]
	serviceInterface, ok := s.serviceMap.Load(serviceName)
	if !ok {
		return nil, nil, errors.New("rpc server: can't find service " + serviceName)
	}
	foundService := serviceInterface.(*service.Service)
	method, ok := foundService.Method[methodName]
	if !ok {
		return nil, nil, errors.New("rpc server: can't find method " + methodName)
	}
	return foundService, method, nil
}

type request struct {
	h          *coder.Header
	argV       reflect.Value
	replyV     reflect.Value
	methodType *service.MethodType
	service    *service.Service
}

func (s *Server) readRequest(serializer coder.Coder) (*request, error) {
	var h coder.Header
	err := serializer.ReadHeader(&h)
	if err != nil {
		log.Println("rpc server:read header error: ", err)
		return nil, err
	}
	req := &request{h: &h}
	req.service, req.methodType, err = s.FindService(h.ServiceMethod)

	req.argV = req.methodType.NewArgValue()
	req.replyV = req.methodType.NewReplyValue()

	argValueInterface := req.argV.Interface()
	if req.argV.Kind() != reflect.Ptr {
		argValueInterface = req.argV.Addr().Interface()
	}
	err = serializer.ReadBody(argValueInterface)
	if err != nil {
		log.Println("rpc server:read body error: ", err)
	}
	return req, err
}

func (s *Server) handleRequest(c coder.Coder, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan int)
	sent := make(chan int)
	go func() {
		err := req.service.Call(req.methodType, req.argV, req.replyV)
		called <- 1
		if err != nil {
			req.h.Error = err.Error()
			s.sendResponse(c, req.h, invalidRequest, sending)
			sent <- 1
			return
		}
		s.sendResponse(c, req.h, req.replyV.Interface(), sending)
		sent <- 1
	}()
	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Error = "rpc server: request handle timeout"
		s.sendResponse(c, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

func (s *Server) sendResponse(serializer coder.Coder, h *coder.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := serializer.Write(h, body); err != nil {
		log.Println("rpc server:write response error: ", err)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
