package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	rpcClient "wheel-rpc/client"
	rpcServer "wheel-rpc/server"
)

func startServer(addr chan string) {
	var f Foo
	s := rpcServer.NewServer()
	err := s.Register(f)
	if err != nil {
		log.Fatalln(err)
	}

	l, err := net.Listen("tcp", ":48295")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	s.Accept(l)
}

func startServerHttp(addr chan string) {
	var f Foo
	s := rpcServer.NewServer()
	err := s.Register(f)
	if err != nil {
		log.Fatalln(err)
	}
	s.HandleHTTP()

	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("network error:", err)
	}

	log.Println("start http server on", l.Addr())
	addr <- l.Addr().String()
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal("rpc server serve error:", err)
	}
}

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func main() {
	addr := make(chan string)
	// go startServer(addr)

	go call(addr)
	startServerHttp(addr)
}

func call(addrCh chan string) {
	client, err := rpcClient.DialHTTP("tcp", <-addrCh)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
