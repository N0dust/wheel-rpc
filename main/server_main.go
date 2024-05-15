package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	rpcClient "wheel-rpc/client"
	"wheel-rpc/server"
)

func startServer(addr chan string) {
	// pick a free port
	l, err := net.Listen("tcp", ":48295")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	server.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	client, _ := rpcClient.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("wheel rpc req %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatalln(err)
			}
		}(i)
	}
	wg.Wait()
}
