package main

import (
	"bufio"
	"fmt"
	"mime"
	"net"
	"os"
	"path"
	"strings"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "9999"
	CONN_TYPE = "tcp"
)

func main() {
	lst, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error when listening:", err.Error())
		os.Exit(1)
	}

	defer lst.Close()

	// Start the server
	fmt.Println("Listening on:", CONN_HOST+":"+CONN_PORT)
	for {
		conn, err := lst.Accept()
		if err != nil {
			fmt.Println("Error when accepting:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error when reading request line:", err)
		return
	}

	parts := strings.Fields(requestLine)
	// "GET /index.html HTTP/1.1\n"
	// => []string{"GET", "/index.html", "HTTP/1.1"}
	if len(parts) != 3 {
		fmt.Println("Invalid request line")
		return
	}

	method, url, _ := parts[0], parts[1], parts[2]
	if method != "GET" {
		fmt.Println("Only support GET method!")
		return
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "" {
			break
		}
		// line =>
		/*
			Host: localhost:8080
			User-Agent: curl/7.68.0
			Accept: ...
			"" => stop reading heading
		*/
	}

	filePath := "." + url
	if url == "/" {
		filePath = "./index.html"
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		sendResponse(conn, "404 Not Found", "text/plain", []byte("File not found"))
		return
	}

	ext := path.Ext(filePath)
	contentType := mime.TypeByExtension(ext)

	if contentType == "" {
		return
	}
	sendResponse(conn, "200 OK", contentType, content)
}

func sendResponse(conn net.Conn, status string, contentType string, body []byte) {
	response := fmt.Sprintf("HTTP/1.1 %s\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n", status, contentType, len(body))
	conn.Write([]byte(response))
	conn.Write(body)
}
