package server

import (
	"log"
	"net/http"
)

const (
	Connected        = "200 Connected to Wheel RPC"
	DefaultRPCPath   = "/_wheelrpc_"
	DefaultDebugPath = "/debug/wheelrpc"
)

func (s *Server) ServeHttp(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("405 must CONNECT\n"))
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Println("rpc hijacking ", req.RemoteAddr, ":", err)
		return
	}
	_, _ = conn.Write([]byte("HTTP/1.0 " + Connected + "\n\n"))
	s.ServeConn(conn)
}

func (s *Server) HandleHTTP() {
	http.Handle(DefaultRPCPath, http.HandlerFunc(s.ServeHttp))
	http.Handle(DefaultDebugPath, debugHTTP{s})
	log.Println("rpc server debug path:", DefaultDebugPath)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
