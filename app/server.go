package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func returnResponse(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	conn, err := l.Accept()
	defer conn.Close()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	readBuffer := make([]byte, 1024)
	int_message, err := bufio.NewReader(conn).Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}
	message := string(readBuffer[:int_message])
	fmt.Println("Message received: ", message)
	fmt.Println(readBuffer)

	path := strings.Split(message, " ")[1]
	if path == "/" {
		returnResponse(conn, "HTTP/1.1 200 OK\r\n\r\n")
	} else if (len(path) > 6) && (path[0:6] == "/echo/") {
		str := path[6:]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)
		fmt.Println(response)

		returnResponse(conn, response)
	} else if path == "/user-agent" {
		headersArray := strings.Split(strings.Split(message, "\r\n\r\n")[0], "\r\n")[1:]
		for _, header := range headersArray {
			if strings.Contains(header, "User-Agent") {
				userAgent := strings.Split(header, ": ")[1]
				response := fmt.Sprint("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
				returnResponse(conn, response)
			}
		}
	} else {
		returnResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}

}
