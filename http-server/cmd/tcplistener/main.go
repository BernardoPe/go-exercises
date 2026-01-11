package main

import (
	"fmt"
	"http_server/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("Server listening on :80")
	for {
		conn, err := listener.Accept()
		fmt.Printf("connection from %s\n", conn.RemoteAddr().String())
		if err != nil {
			log.Fatal("Error accepting connection:", err)
		}

		go func(c net.Conn) {
			defer c.Close()
			req, err := request.FromReader(c)
			if err != nil {
				log.Println("Error reading request:", err)
				return
			}
			fmt.Println("Request Line:", req.Line.Method, req.Line.RequestTarget, req.Line.HttpVersion)
			fmt.Printf("- Method: %s\n", req.Line.Method)
			fmt.Printf("- Target: %s\n", req.Line.RequestTarget)
			fmt.Printf("- Version: %s\n", req.Line.HttpVersion)
			fmt.Println("Headers:")
			for k, v := range req.Headers {
				fmt.Printf("- %s: %s\n", k, v)
			}
			fmt.Println("Body:")
			fmt.Println(string(req.Body.Data))
		}(conn)
	}
}
