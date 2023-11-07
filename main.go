package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
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

	//port := os.Args[1]
	port := "127.0.0.1:8080"

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
	path := "./files"
	reqFile := request.URL.String()

	fileExt := strings.Split(reqFile, ".")

	var res string

	switch fileExt[len(fileExt)-1] {
	case "html":
		res = makeGetResponse(path+"/html"+reqFile, "text/html")
	case "png":
		res = makeGetResponse(path+"/png"+reqFile, "image/png")
	case "jpg":
		res = makeGetResponse(path+"/jpg"+reqFile, "image/jpg")
	case "jpeg":
		res = makeGetResponse(path+"/jpeg"+reqFile, "image/jpeg")
	case "txt":
		res = makeGetResponse(path+"/txt"+reqFile, "text/plain")
	case "gif":
		res = makeGetResponse(path+"/gif"+reqFile, "image/gif")
	case "css":
		res = makeGetResponse(path+"/gif"+reqFile, "image/gif")
	case "ico":
		res = makeGetResponse(path+"/ico"+reqFile, "image/ico")

	default:
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request"
		headers := make(map[string]string)
		//headers["Content-Length: "] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text:html"

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
	path := request.URL.Path
	fileName := path[1:]

	fmt.Println(fileName)
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

	filepath := "./files/jpg/" + handler.Filename
	dst, err := os.Create(filepath)
	if err != nil {
		fmt.Println("error creating file", err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		fmt.Println("error something file", err)
		return
	}
	fmt.Println("uploaded file")

	status := "HTTP/1.1 200 OK"
	body := "File uploaded successfully"
	headers := make(map[string]string)
	//headers["Content-Length"] = strconv.Itoa(len(body))
	headers["Content-Type"] = "text/html"
	res := response{status, headers, body}
	connection.Write([]byte(res.String()))

}

func getFilePath(fileExt string) string {
	return ""
}

func getHeaderType(fileExt string) string {
	return ""
}
