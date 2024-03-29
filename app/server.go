package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

var storage = make(map[string]RespBulkString)

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

		throw := func() {
			conn.Write([]byte("-ERR\r\n"))
		}
		if len(parsed.Value) == 0 {
			throw()
			break
		}
		command := parsed.Value[0].(RespBulkString).Value
		switch strings.ToUpper(string(command)) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			echo := RespArray{Value: parsed.Value[1:]}
			conn.Write(echo.Bytes())
		case "SET":
			if len(parsed.Value) < 3 {
				throw()
				break
			}

			key, keyOk := parsed.Value[1].(RespBulkString)
			value, valueOk := parsed.Value[2].(RespBulkString)
			if !keyOk || !valueOk {
				throw()
				break
			}
			storage[string(key.Value)] = value

			if len(parsed.Value) == 5 {
				px, pxOk := parsed.Value[3].(RespBulkString)
				if !pxOk || strings.ToUpper(string(px.Value)) != "PX" {
					throw()
					break
				}
				exp, expOk := parsed.Value[4].(RespBulkString)
				if !expOk {
					throw()
					break
				}
				ms, err := strconv.Atoi(exp.String())
				if err != nil {
					throw()
					break
				}

				go func() {
					timer := time.NewTimer(time.Duration(ms) * time.Millisecond)
					<-timer.C
					v := storage[string(key.Value)]
					v.Value = nil
					storage[string(key.Value)] = v
				}()
			}

			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if len(parsed.Value) != 2 {
				throw()
				continue
			}
			key, ok := parsed.Value[1].(RespBulkString)
			if !ok {
				throw()
				continue
			}
			value, ok := storage[string(key.Value)]
			if !ok {
				throw()
				continue
			}
			conn.Write(value.Bytes())
		default:
			throw()
		}
	}
}
