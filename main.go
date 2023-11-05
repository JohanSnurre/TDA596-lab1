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
	incoming := strings.Fields(string(buffer))
	request, err := http.NewRequest(incoming[0], incoming[1], nil)

	fmt.Println(request)
	switch request.Method {
	case "GET":
		path := "./files"
		URL := request.URL.String()
		fileName := URL[:strings.LastIndex(URL, ".")]

		//temp := strings.Split(URL, ".")

		fileExt := URL[strings.LastIndex(URL, ".")+1:]
		println("FILENAME: ", fileName)
		println("FILEEXT: ", fileExt)

		switch fileExt {
		case "html":
			dat, err := os.ReadFile(path + "/html" + fileName + "." + fileExt)
			println("length: ", dat)
			if err != nil {
				fmt.Println("error reading")
				status := "HTTP/1.1 500 Internal Server Error"
				body := "Something on the server went wrong"
				headers := make(map[string]string)
				//headers["Content-Length: "] = strconv.Itoa(len(body))
				headers["Content-Type"] = "text:html"

				res := response{status, headers, body}
				connection.Write([]byte(res.String()))

				return
			}

			status := "HTTP/1.1 200 OK"
			body := string(dat)
			headers := make(map[string]string)
			//headers["Content-Length: "] = strconv.Itoa(len(dat))
			headers["Content-Type"] = "text:html"

			res := response{status, headers, body}

			connection.Write([]byte(res.String()))
			println(res.String())

		case "png":
			dat, err := os.ReadFile(path + "/png" + fileName + "." + fileExt)
			if err != nil {
				fmt.Println("error reading")
				status := "HTTP/1.1 500 Internal Server Error"
				body := "Something on the server went wrong"
				headers := make(map[string]string)
				//headers["Content-Length: "] = strconv.Itoa(len(body))
				headers["Content-Type"] = "text:html"

				res := response{status, headers, body}
				connection.Write([]byte(res.String()))

				return
			}

			status := "HTTP/1.1 200 OK"
			body := string(dat)
			headers := make(map[string]string)
			//headers["Content-Length: "] = strconv.Itoa(len(body))
			headers["Content-Type"] = "image:png"

			res := response{status, headers, body}

			connection.Write([]byte(res.String()))

		case "jpg", "jpeg":
			fileExt = "jpg"
			dat, err := os.ReadFile(path + "/jpg" + fileName + "." + fileExt)
			if err != nil {
				fmt.Println("error reading")
				status := "HTTP/1.1 500 Internal Server Error"
				body := "Something on the server went wrong"
				headers := make(map[string]string)
				//headers["Content-Length: "] = strconv.Itoa(len(body))
				headers["Content-Type"] = "text:html"

				res := response{status, headers, body}
				connection.Write([]byte(res.String()))

				return
			}

			status := "HTTP/1.1 200 OK"
			body := string(dat)
			headers := make(map[string]string)
			//headers["Content-Length: "] = strconv.Itoa(len(body))
			headers["Content-Type"] = "image:jpg"

			res := response{status, headers, body}

			connection.Write([]byte(res.String()))

		case "ico":
			dat, err := os.ReadFile(path + "/ico" + fileName + "." + fileExt)
			if err != nil {
				fmt.Println("error reading")
				status := "HTTP/1.1 500 Internal Server Error"
				body := "Something on the server went wrong"
				headers := make(map[string]string)
				//headers["Content-Length: "] = strconv.Itoa(len(body))
				headers["Content-Type"] = "text:html"

				res := response{status, headers, body}
				connection.Write([]byte(res.String()))

				return
			}

			status := "HTTP/1.1 200 OK"
			body := string(dat)
			headers := make(map[string]string)
			//headers["Content-Length: "] = strconv.Itoa(len(body))
			headers["Content-Type"] = "image:ico"

			res := response{status, headers, body}

			connection.Write([]byte(res.String()))

		}

	case "POST":
		connection.Write([]byte("response dog"))

	default:
		status := "HTTP/1.1 400 Bad Request"
		body := "Bad request"
		headers := make(map[string]string)
		//headers["Content-Length: "] = strconv.Itoa(len(body))
		headers["Content-Type"] = "text:html"

		res := response{status, headers, body}

		connection.Write([]byte(res.String()))

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
