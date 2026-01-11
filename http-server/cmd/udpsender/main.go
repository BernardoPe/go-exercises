package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("", "localhost:80")
	if err != nil {
		log.Fatal("Error resolving address:", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("Error dialing UDP:", err)
	}

	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading input:", err)
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatal("Error sending data:", err)
		}
	}
}
