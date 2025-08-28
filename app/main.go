package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn, dir string) {
	defer conn.Close()

	for {
		var request = readRequest(conn)
		headers := getHeaders(request)

		if request == "" {
			return
		}
		statusLine, err := getStatusLine(request)
		if err != nil {
			fmt.Printf("Error while parsing connection request: %v\n", err.Error())
			notFound()
		}
		path := statusLine.Endpoint
		method := statusLine.Method

		var response Response
		switch {
		case routeMatches(path, "^$") || routeMatches("^/$", path):
			response = baseCont()
		case routeMatches("^/echo/([^/]+)$", path):
			response = echoCont(request, path)
		case routeMatches("^/user-agent", path):
			response = userAgentCont(request)
		case method == "GET" && routeMatches("^/files/([^/]+)$", path):
			response = filsCont(path, dir)
		case method == "POST" && routeMatches("^/files/([^/]+)$", path):
			response = postFileCont(request, path, dir)
		default:
			response = notFound()
		}

		if strings.ToLower(headers.Headers["Connection"]) == "close" {
			response.RespHeaders["Connection"] = "close"
		}
		var responseStr string = createResponseString(response)
		fmt.Fprintf(conn, "%v", responseStr)

		if strings.ToLower(headers.Headers["Connection"]) == "close" {
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatal("Failed to bind to port 4221")
	}

	var dir string
	args := os.Args
	for idx, arg := range args {
		if arg == "--directory" {
			dir = args[idx+1]
			break
		}
	}

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn, dir)
	}
}
