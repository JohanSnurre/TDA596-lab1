package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
)

type response struct {
	status string
	header string
	body   string
}

var supportedContentTypes = map[string]string{
	"html": "text/html",
	"txt":  "text/plain",
	"gif":  "image/gif",
	"jpeg": "image/jpeg",
	"jpg":  "image/jpg",
	"css":  "text/css",
	"ico":  "image/ico",
}

var jsonType string = "application/json"

var statusCode = map[int]string{
	200: "200 OK",
	201: "201 Created",
	400: "400 Bad Request",
	404: "404 Not Found",
}

func jsonRes(text string, isError bool) string {
	var field string
	if isError {
		field = "error: "
	} else {
		field = "response: "
	}

	return "{" + field + text + "}"
}

func (res response) String() string {
	statusHeader := "HTTP/1.1 " + res.status + "\r\n"
	contentHeader := "Content-Type:" + res.header + "\r\n\r\n"

	return statusHeader + contentHeader + res.body
}

func main() {
	port := os.Args[1]
	server := os.Args[2]

	tcpAddr, err := net.ResolveTCPAddr("tcp", port)

	if err != nil {
		fmt.Println("<1>", err)
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("<2>", err)
		os.Exit(1)
	}

	fmt.Printf("Proxy started at address (%s) and port (%s)\n", listener.Addr().String(), port)
	for {

		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("<3>", err)
			os.Exit(1)
		}

		go handleClient(connection, server)

	}

	os.Exit(0)

}

func handleClient(connection net.Conn, serverAddr string) {

	fmt.Println("New connection from client: ", connection.RemoteAddr())
	defer connection.Close()

	reader := bufio.NewReader(connection)
	request, err := http.ReadRequest(reader)

	if err != nil {
		fmt.Println("<4> Error reading request: ", err.Error())
		return
	}

	/*server, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Print("proxy error cnnection to server: ", err.Error())
	}
	server.Close()
	*/

	switch request.Method {
	case "GET":
		var reqBuf, resBuf bytes.Buffer

		server, err := net.Dial("tcp", serverAddr)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = request.Write(&reqBuf)
		if err != nil {
			fmt.Print(err.Error())
			return
		}

		server.Write(reqBuf.Bytes())

		// read
		reader := bufio.NewReader(server)
		res, err := http.ReadResponse(reader, request)
		if err != nil {
			fmt.Println("Error reading response:", err.Error())
			return
		}

		err = res.Write(&resBuf)
		if err != nil {
			fmt.Println("Error write:", err.Error())
			return

		}

		connection.Write(resBuf.Bytes())
	default:
		response := response{statusCode[400], jsonType, jsonRes("Method not implemented", true)}
		connection.Write([]byte(response.String()))
	}
}
