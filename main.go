package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

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
	//defer connection.Close()

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
		res = makeGetResponse(path+"/png"+reqFile, "image/jpeg")
	case "txt":
		res = makeGetResponse(path+"/txt"+reqFile, "text/plain")
	case "gif":
		res = makeGetResponse(path+"/gif"+reqFile, "image/gif")
	case "css":
		res = makeGetResponse(path+"/gif"+reqFile, "image/gif")
	default:
		//error

	}

	connection.Write([]byte(res))

}

func makeGetResponse(path string, header string) string {
	dat, err := os.ReadFile(path)
	if err != nil {
		// return 400
		fmt.Println("error reading")
		return ""
	}

	return "HTTP/1.1 200 OK\n" + "Content-Length: " + fmt.Sprint(len(dat)) + "\nContent-Type: " + header + "\n\n" + string(dat)
}

func postResponse(connection net.Conn, request http.Request) {

}
