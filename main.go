package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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

var maxworker = 10
var currentWorkers = 0

var mu, ma, cond_mu sync.Mutex
var cond sync.Cond

func main() {

	port := os.Args[1]
	cond = *sync.NewCond(&cond_mu)

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

	fmt.Printf("Server started at address (%s) and port (%s)\n", listener.Addr().String(), port)

	k := make(chan string)
	for i := 0; i < 5; i++ {
		fmt.Println("Created test")
		go listen(listener, i, k)
	}

	for {
		fmt.Println(<-k)
	}

	os.Exit(0)

}

func cmp() int {
	mu.Lock()
	ret := 0
	if currentWorkers >= maxworker {
		ret = 1
	}
	mu.Unlock()
	return ret
}

func listen(listener *net.TCPListener, i int, k chan string) {
	for {
		ma.Lock()
		for cmp() == 1 {

			cond.L.Lock()
			k <- strconv.Itoa(i) + " WAITING" + ", Current threads: " + strconv.Itoa(currentWorkers)
			cond.Wait()
			cond.L.Unlock()
		}

		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("<3>", err)
			os.Exit(1)
		}
		currentWorkers = currentWorkers + 1
		ma.Unlock()
		go handleClient(connection, i, k)

	}

}

func handleClient(connection net.Conn, i int, k chan string) {

	//fmt.Println("New connection from: ", connection.RemoteAddr())
	defer connection.Close()
	k <- strconv.Itoa(i) + ", New connection from: " + connection.RemoteAddr().String() + ", Current threads: " + strconv.Itoa(currentWorkers)
	reader := bufio.NewReader(connection)
	request, err := http.ReadRequest(reader)

	if err != nil {
		fmt.Println("<4> Error reading request: ", err.Error())
		return
	}

	switch request.Method {
	case "GET":
		getResponse(connection, *request)

	case "POST":
		postResponse(connection, *request)

	default:
		response := response{statusCode[400], jsonType, jsonRes("Method not implemented", true)}
		connection.Write([]byte(response.String()))
	}

	mu.Lock()
	cond.L.Lock()
	currentWorkers = currentWorkers - 1
	cond.Signal()
	cond.L.Unlock()
	mu.Unlock()
	return
}

func getResponse(connection net.Conn, request http.Request) {
	var res string
	path := "./files/"

	reqFile := request.URL.String()

	fileExt := getFileExt(reqFile)

	header, err := getHeaderType(fileExt)
	if err != nil {
		res := response{statusCode[400], jsonType, jsonRes("Invalid file type", true)}
		connection.Write([]byte(res.String()))
		return
	}

	filePath := path + fileExt + reqFile

	res = makeGetResponse(filePath, header)

	connection.Write([]byte(res))
}

func makeGetResponse(path string, header string) string {
	if _, err := os.Stat(path); err != nil {
		res := response{statusCode[404], jsonType, jsonRes("The file does not exists", true)}
		return res.String()
	}
	dat, err := os.ReadFile(path)
	if err != nil {
		// return 400
		res := response{statusCode[400], jsonType, jsonRes("The file cannot be read", true)}
		return res.String()
	}

	res := response{statusCode[200], header, string(dat)}
	return res.String()
}

func postResponse(connection net.Conn, request http.Request) {
	request.ParseMultipartForm(10 << 20)

	file, handler, err := request.FormFile("file")

	// Check if the file is present
	if err != nil {
		res := response{statusCode[400], jsonType, jsonRes("Unable to retrieve the file from the POST request - "+err.Error(), true)}
		connection.Write([]byte(res.String()))
		return
	}
	defer file.Close()

	fileName := handler.Filename
	fileExt := getFileExt(fileName)

	filePath := "./files/" + fileExt + "/" + handler.Filename

	dst, err := os.Create(filePath)
	if err != nil {
		res := response{statusCode[400], jsonType, jsonRes("Error creating file - "+err.Error(), true)}
		connection.Write([]byte(res.String()))
		return
	}

	defer dst.Close()

	// copy content
	if _, err := io.Copy(dst, file); err != nil {
		res := response{statusCode[400], jsonType, jsonRes("Error copying content - "+err.Error(), true)}
		connection.Write([]byte(res.String()))
		return
	}

	res := response{statusCode[201], jsonType, jsonRes("Item created successfully", false)}
	connection.Write([]byte(res.String()))

}

func getFileExt(reqFile string) string {
	split := strings.Split(reqFile, ".")
	return split[len(split)-1]
}

func getHeaderType(fileExt string) (string, error) {
	header, err := supportedContentTypes[fileExt]

	if !err {
		return "", fmt.Errorf("unsupported header")
	} else {
		return header, nil
	}
}
