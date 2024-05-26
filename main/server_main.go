package main

import (
	"log"
	"net"
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

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	client, _ := rpcClient.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
