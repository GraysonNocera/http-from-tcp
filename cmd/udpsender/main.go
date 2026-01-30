package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("error resolving udp addr", err)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("error connecting udp", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error reading from stdin", err)
		}
		_, err = conn.Write([]byte(str))
		if err != nil {
			log.Fatal("error writing to udp", err)
		}
	}
}