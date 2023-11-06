package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type response struct {
	status  string
	headers map[string]string
	body    string
}

func (r response) String() string {
	res := r.status + "\n"
	for k, v := range r.headers {
		res = res + k + ": " + v + "\n"
	}
	res = res + "\n" + r.body
	return res
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: http_server <port>")
		os.Exit(1)
	}

	port := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+port)
	if err != nil {
		fmt.Println("<1>", err)
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("<2>", err)
		os.Exit(1)
	}

	fmt.Printf("Server started at address (%s) and port (%s)\n", listener.Addr().String(), port)
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("<3>", err)
			os.Exit(1)
		}
		go handleClient(connection)
	}

	os.Exit(0)
}

func handleClient(connection net.Conn) {
	fmt.Println("New connection from: ", connection.RemoteAddr())
	defer connection.Close()
	buffer := make([]byte, 4096)
	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("<10>", err)
		if err == io.EOF {
			fmt.Println("Terminating connection from ", connection.RemoteAddr())
		}
		return
	}
	incoming := strings.Fields(string(buffer))
	request, err := http.NewRequest(incoming[0], incoming[1], nil)
	if err != nil {
		fmt.Println("req error: ", err.Error())
	}
	switch request.Method {
	case "GET":
		getResponse(connection, *request)
	case "POST":
		postResponse(connection, *request)
	default:
		//not implemented error
	}

	return
}

func getResponse(connection net.Conn, request http.Request) {
	path := "./files"
	reqFile := request.URL.String()
	fileExt := strings.Split(reqFile, ".")
	var res string
	switch fileExt[len(fileExt)-1] {
	case "html", "txt", "gif", "jpeg", "jpg", "css":
		res = makeGetResponse(path+reqFile, "text/"+fileExt[len(fileExt)-1])
	default:
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request"
		headers := make(map[string]string)
		headers["Content-Length"] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text/html"
		temp := response{status, headers, body}
		res = temp.String()
	}
	connection.Write([]byte(res))
}

func makeGetResponse(path string, header string) string {
	if _, err := os.Stat(path); err != nil {
		status := "HTTP/1.1 404 Not Found"
		body := "404 This file does not exist"
		headers := make(map[string]string)
		headers["Content-Type"] = "text/html"
		res := response{status, headers, body}
		return res.String()
	}
	dat, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("error reading")
		return ""
	}
	status := "HTTP/1.1 200 OK"
	body := string(dat)
	headers := make(map[string]string)
	headers["Content-Length"] = strconv.Itoa(len(body))
	headers["Content-Type"] = header
	res := response{status, headers, body}
	return res.String()
}

func postResponse(connection net.Conn, request http.Request) {
	// Check if the request is in multipart/form-data format
	if request.Header.Get("Content-Type") != "multipart/form-data" {
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request - POST data must be in multipart/form-data format"
		headers := make(map[string]string)
		headers["Content-Length"] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text/html"
		res := response{status, headers, body}
		connection.Write([]byte(res.String()))
		return
	}
	// Retrieve the file from the request
	file, _, err := request.FormFile("image")
	// Check if the file is present
	if err != nil {
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request - Unable to retrieve the file from the POST request"
		headers := make(map[string]string)
		headers["Content-Length"] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text/html"
		res := response{status, headers, body}
		connection.Write([]byte(res.String()))
		return
	}
	defer file.Close()

	// Todo
	// Store the file on the server

	status := "HTTP/1.1 200 OK"
	body := "File uploaded successfully"
	headers := make(map[string]string)
	headers["Content-Length"] = strconv.Itoa(len(body))
	headers["Content-Type"] = "text/html"
	res := response{status, headers, body}
	connection.Write([]byte(res.String()))
}
