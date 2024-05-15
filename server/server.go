package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	rpcSerializer "wheel-rpc/serializer"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int
	SerializerType rpcSerializer.Type
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	SerializerType: rpcSerializer.GobType,
}

type Server struct {
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
	f := rpcSerializer.NewSerializerFuncMap[opt.SerializerType]
	if f == nil {
		log.Println("rpc server:invalid serializer type")
		return
	}
	s.ServeNewConn(f(conn))
}

var invalidRequest = struct{}{}

func (s *Server) ServeNewConn(serializer rpcSerializer.Serializer) {
	sending := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for {
		req, err := s.readRequest(serializer)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(serializer, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handelRequest(serializer, req, sending, wg)
	}
	wg.Wait()
	_ = serializer.Close()
}

type request struct {
	h      *rpcSerializer.Header
	argV   reflect.Value
	replyV reflect.Value
}

func (s *Server) readRequest(serializer rpcSerializer.Serializer) (*request, error) {
	var h rpcSerializer.Header
	if err := serializer.ReadHeader(&h); err != nil {
		log.Println("rpc server:read header error: ", err)
		return nil, err
	}
	req := &request{h: &h}
	req.argV = reflect.New(reflect.TypeOf(""))
	if err := serializer.ReadBody(req.argV.Interface()); err != nil {
		log.Println("rpc server:read body error: ", err)
		return nil, err
	}
	return req, nil
}

func (s *Server) sendResponse(serializer rpcSerializer.Serializer, h *rpcSerializer.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := serializer.Write(h, body); err != nil {
		log.Println("rpc server:write response error: ", err)
	}
}

func (s *Server) handelRequest(serializer rpcSerializer.Serializer, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h.ServiceMethod)
	const msg = "wheel rpc resp　%d"
	req.replyV = reflect.ValueOf(fmt.Sprintf(msg, req.h.Seq))
	s.sendResponse(serializer, req.h, req.replyV.Interface(), sending)
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
