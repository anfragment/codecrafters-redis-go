package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}

		parsed, _, err := parseArray(buf[1:], 0)
		if err != nil {
			conn.Write([]byte("-ERR\r\n"))
		}

		command := parsed.Value[0].(RespBulkString).Value

		switch string(command) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			echo := RespArray{Value: parsed.Value[1:]}
			conn.Write(echo.Bytes())
		case "COMMAND":
			conn.Write([]byte("*1\r\n$4\r\nINFO\r\n"))
		default:
			conn.Write([]byte("-ERR\r\n"))
		}
	}
}
