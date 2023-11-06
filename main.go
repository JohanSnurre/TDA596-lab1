package main

import (
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

	fmt.Println(string(buffer))

	incoming := strings.Fields(string(buffer))
	request, err := http.NewRequest(incoming[0], incoming[1], nil)

	if err != nil {
		fmt.Println("req error: ", err.Error())
	}

	switch request.Method {
	case "GET":

		getResponse(connection, *request)

	case "POST":
		// Get the uploaded file
		//request.Header.Set("Content-Type", "multipart/form-data")
		postResponse(connection, *request)

		_, _, err := request.FormFile("image")
		if err != nil {
			fmt.Println("Error retrieving the file:", err)
			return
		}

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

}

func postResponse(connection net.Conn, request http.Request) {

	allowedContentTypes := map[string]bool{
		"text/html":  true,
		"text/plain": true,
		"image/gif":  true,
		"image/jpeg": true,
		"image/png":  true,
		"text/css":   true,
	}

	contentType := request.Header.Get("Content-Type")

	// Check if the request is in multipart/form-data format
	if !allowedContentTypes[contentType] {
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