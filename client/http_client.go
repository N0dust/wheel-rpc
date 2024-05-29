package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"wheel-rpc/server"
)

func NewHTTPClient(conn net.Conn, opt *server.Option) (*Client, error) {
	_, err := io.WriteString(conn, "CONNECT "+conn.RemoteAddr().String()+" HTTP/1.0\n\n")
	if err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == server.Connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}

func DialHTTP(network, address string, opts ...*server.Option) (*Client, error) {
	if len(opts) == 0 || opts[0] == nil {
		conn, err := net.Dial(network, address)
		if err != nil {
			return nil, err
		}
		return NewHTTPClient(conn, nil)
	}
	conn, err := net.DialTimeout(network, address, opts[0].ConnectTimeout)
	if err != nil {
		return nil, err
	}
	return NewHTTPClient(conn, opts[0])
}

func XDial(rpcAddr string, opts ...*server.Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		// tcp, unix or other transport protocol
		return Dial(protocol, addr, opts...)
	}
}
