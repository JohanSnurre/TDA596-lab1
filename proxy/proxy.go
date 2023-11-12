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

var severport string
var severhost string

var requesturl string

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

	fmt.Println("Proxy Sever is listening on port", port)

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

	reader := bufio.NewReader(strings.NewReader(string(buffer)))
	request, err := http.ReadRequest(reader)

	if err != nil {
		fmt.Println("req error: ", err.Error())
	}

	urlget(request)

	switch request.Method {
	case "GET":
		gethandle(request, connection)

	default:
		// not implemented error

	}

	return
}

func urlget(request *http.Request) {

	hostuRL := strings.Split(request.Host, ":")
	severport = hostuRL[1]
	severhost = hostuRL[0]

	fmt.Println(severport)
	fmt.Println(severhost)

	requesturl = request.URL.Path
	fmt.Println(requesturl)

}

// func postResponse(connection net.Conn, request http.Request) {

func gethandle(request *http.Request, connection net.Conn) {
	resp, err := http.Get("http://" + severhost + ":" + severport + requesturl)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	fmt.Println(string(body))

	status := resp.Status
	headers := make(map[string]string)
	headers["Content-Type"] = resp.Header.Get("Content-Type")

	response := response{status, headers, string(body)}

	fmt.Println(response)
}
