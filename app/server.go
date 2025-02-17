package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func returnResponse(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}

func gzipEncode(str string) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(str)); err != nil {
		fmt.Println("Error writing to gzip: ", err.Error())
		os.Exit(1)
	}
	if err := gz.Close(); err != nil {
		fmt.Println("Error closing gzip: ", err.Error())
		os.Exit(1)
	}
	return b.Bytes()
}

func handleTCPRequest(conn net.Conn, dir string) {
	defer conn.Close()
	readBuffer := make([]byte, 1024)
	int_message, err := bufio.NewReader(conn).Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}
	message := string(readBuffer[:int_message])

	splitMessage := strings.Split(message, " ")
	requestType := splitMessage[0]
	path := splitMessage[1]
	if path == "/" {
		returnResponse(conn, "HTTP/1.1 200 OK\r\n\r\n")
	} else if (len(path) > 6) && (path[0:6] == "/echo/") {
		str := path[6:]
		headerArray := strings.Split(strings.Split(message, "\r\n\r\n")[0], "\r\n")[1:]
		encodingType := ""
		for _, header := range headerArray {
			if strings.Contains(header, "Accept-Encoding") {
				encodingType = strings.Split(header, ": ")[1]
			}
		}
		if encodingType != "" {
			if strings.Contains(encodingType, "gzip") {
				zipedStr := gzipEncode(str)
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(zipedStr), zipedStr)

				returnResponse(conn, response)
			} else {
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)
				returnResponse(conn, response)
			}
		} else {
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)
			returnResponse(conn, response)
		}

	} else if path == "/user-agent" {
		headersArray := strings.Split(strings.Split(message, "\r\n\r\n")[0], "\r\n")[1:]
		for _, header := range headersArray {
			if strings.Contains(header, "User-Agent") {
				userAgent := strings.Split(header, ": ")[1]
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)

				returnResponse(conn, response)
			}
		}
	} else if strings.HasPrefix(path, "/files") {
		if requestType == "GET" {
			filename := strings.Split(path, "/")[2]
			file, err := os.Open(dir + filename)
			if err != nil {
				fmt.Println("Error opening file: ", err.Error())
				response := "HTTP/1.1 404 Not Found\r\n\r\n"
				returnResponse(conn, response)
			}
			defer file.Close()
			fileContentArray := make([]byte, 1024)

			fileDataLength, err := file.Read(fileContentArray)
			if err != nil {
				response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				returnResponse(conn, response)
			}
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", fileDataLength, string(fileContentArray[:fileDataLength]))
			returnResponse(conn, response)
		} else if requestType == "POST" {
			filename := strings.Split(path, "/")[2]
			file, err := os.Create(dir + filename)
			if err != nil {
				response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				returnResponse(conn, response)
			}
			defer file.Close()
			headersArray := strings.Split(strings.Split(message, "\r\n\r\n")[0], "\r\n")[1:]
			reponseBody := strings.Split(message, "\r\n\r\n")[1]
			var contentLength int
			for _, header := range headersArray {
				if strings.Contains(header, "Content-Length") {
					contentLength, err = strconv.Atoi(strings.Split(header, ": ")[1])
					if err != nil {
						response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
						returnResponse(conn, response)
					}
				}
			}
			responseArray := []byte(reponseBody)
			if len(responseArray) != contentLength {
				response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				returnResponse(conn, response)
			}
			_, err = file.Write(responseArray)
			if err != nil {
				response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				returnResponse(conn, response)
			}
			response := "HTTP/1.1 201 Created\r\n\r\n"
			returnResponse(conn, response)
		}
	} else {
		returnResponse(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	var dir string
	flag.StringVar(&dir, "directory", "", "Directory to serve files from")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleTCPRequest(conn, dir)
	}

}
