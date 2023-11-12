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
	501: "501 Not Implemented",
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

	switch request.Method {
	case "GET":
		var reqBuf, resBuf bytes.Buffer

		server, err := net.Dial("tcp", serverAddr)
		if err != nil {
			res := response{statusCode[400], jsonType, jsonRes("Error connecting to the server -"+err.Error(), true)}
			connection.Write([]byte(res.String()))
			return
		}

		defer server.Close()

		err = request.Write(&reqBuf)
		if err != nil {
			res := response{statusCode[400], jsonType, jsonRes("Error writing to server -"+err.Error(), true)}
			connection.Write([]byte(res.String()))
			return
		}

		server.Write(reqBuf.Bytes())

		// read
		reader := bufio.NewReader(server)
		res, err := http.ReadResponse(reader, request)
		if err != nil {
			res := response{statusCode[400], jsonType, jsonRes("Error reading response -"+err.Error(), true)}
			connection.Write([]byte(res.String()))
			return
		}

		err = res.Write(&resBuf)
		if err != nil {
			res := response{statusCode[400], jsonType, jsonRes("Error writing response -"+err.Error(), true)}
			connection.Write([]byte(res.String()))
			return

		}

		connection.Write(resBuf.Bytes())
	default:
		response := response{statusCode[501], jsonType, jsonRes("Method not implemented", true)}
		connection.Write([]byte(response.String()))
	}
}
