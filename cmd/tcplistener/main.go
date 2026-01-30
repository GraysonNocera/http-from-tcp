package main

import (
	"fmt"
	"http-from-tcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error opening file", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error listening to connection", err)
			break;
		}
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error getting request from reader", err)
		}
		fmt.Println("Request Reader")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Println("Headers")
		for key, value := range(r.Headers) {
			fmt.Printf("- %s: %s\n", key, value)
		}
    fmt.Println("Body")
    fmt.Printf("%s\n", string(r.Body))
	}
}
