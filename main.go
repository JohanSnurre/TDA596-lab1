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

	fmt.Println(request)
	switch request.Method {
	case "GET":
		path := "./files"
		reqFile := request.URL.String()

		fileExt := strings.Split(reqFile, ".")

		switch fileExt[len(fileExt)-1] {
		case "html":
			dat, err := os.ReadFile(path + "/html" + reqFile)
			if err != nil {
				fmt.Println("error reading")
				return
			}

			res := "HTTP/1.1 200 OK\n" + "Content-Length: " + string(len(dat)) + "\nContent-Type: text:html\n\n" + string(dat)

			connection.Write([]byte(res))

		case "png":
			dat, err := os.ReadFile(path + "/png" + reqFile)
			if err != nil {
				fmt.Println("error reading")
				return
			}

			res := "HTTP/1.1 200 OK\n" + "Content-Length: " + string(len(dat)) + "\nContent-Type: text:html\n\n" + string(dat)

			connection.Write([]byte(res))
		}

	case "POST":
		connection.Write([]byte("response dog"))

	}

	return

	/*var buf [4096]byte

	defer connection.Close()
	connection.Write(([]byte("hello")))

	for {
		n, err := connection.Read(buf[0:])
		if err != nil {
			fmt.Println("<4>", err)
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := connection.Write(buf[0:n])
		if err2 != nil {
			fmt.Println("<5>", err)
			return

		}

	}
	*/

}
