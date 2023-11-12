package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

type response struct {
	status  string
	headers map[string]string
	body    string
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

func (r response) String() string {

	res := r.status + "\n"
	for k, v := range r.headers {

		res = res + k + ": " + v + "\n"

	}
	res = res + "\n" + r.body

	return res

}

func main() {
	port := os.Args[1]

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

	//define a channel to control the number of goroutines

	maxGoroutines := 10
	goRoutineSem := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup

	fmt.Printf("Server started at address (%s) and port (%s)\n", listener.Addr().String(), port)
	for {

		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("<3>", err)
			os.Exit(1)
		}

		//add a goroutine to the waitgroup
		wg.Add(1)

		//start a goroutine
		go func(connection net.Conn) {
			defer wg.Done()

			goRoutineSem <- struct{}{}
			handleClient(connection)
			<-goRoutineSem
		}(connection)

	}

	//wait for all goroutines to finish
	wg.Wait()
	close(goRoutineSem)

	os.Exit(0)

}

func handleClient(connection net.Conn) {

	fmt.Println("New connection from: ", connection.RemoteAddr())
	defer connection.Close()

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
		// not implemented error

	}

	return
}

func getResponse(connection net.Conn, request http.Request) {
	var res string
	path := "./files/"

	reqFile := request.URL.String()

	fileExt := getFileExt(reqFile)

	header, err := getHeaderType(fileExt)
	if err != nil {
		fmt.Println(err)
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request"
		headers := make(map[string]string)
		//headers["Content-Length: "] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text:html"

		temp := response{status, headers, body}
		res = temp.String()
		connection.Write([]byte(res))
		return
	}

	filePath := path + fileExt + reqFile

	res = makeGetResponse(filePath, header)

	connection.Write([]byte(res))
}

func makeGetResponse(path string, header string) string {
	if _, err := os.Stat(path); err != nil {
		status := "HTTP/1.1 404 Not Found"
		body := "404 This file does not exist"
		headers := make(map[string]string)
		headers["Content-Type"] = "text:html"
		res := response{status, headers, body}
		return res.String()
	}
	dat, err := os.ReadFile(path)
	if err != nil {
		// return 400
		fmt.Println("error reading")
		return ""
	}

	status := "HTTP/1.1 200 OK"
	body := string(dat)
	headers := make(map[string]string)
	//headers["Content-Length: "] = strconv.Itoa(len(body))
	headers["Content-Type"] = header

	res := response{status, headers, body}

	return res.String()
}

func postResponse(connection net.Conn, request http.Request) {
	request.ParseMultipartForm(10 << 20)

	file, handler, err := request.FormFile("file")

	// Check if the file is present
	if err != nil {
		fmt.Println("<14>file receive error", err)
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request - Unable to retrieve the file from the POST request"
		headers := make(map[string]string)
		//headers["Content-Length"] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text/html"
		res := response{status, headers, body}
		connection.Write([]byte(res.String()))
		return
	}
	defer file.Close()

	fileName := handler.Filename
	fileExt := getFileExt(fileName)

	filePath := "./files/" + fileExt + "/" + handler.Filename

	dst, err := os.Create(filePath)
	if err != nil {
		fmt.Println("error creating file", err)
		return
	}

	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		fmt.Println("error something file", err)
		return
	}

	status := "HTTP/1.1 200 OK"
	body := "File uploaded successfully"
	headers := make(map[string]string)
	//headers["Content-Length"] = strconv.Itoa(len(body))
	headers["Content-Type"] = "text/html"
	res := response{status, headers, body}
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
