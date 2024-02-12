package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
	"time"
)

type Tests func() error

func Output() error {
	pid := os.Getpid()
	log.Printf("Running as child process with pid: %d\n", pid)
	return nil
}

func SocketConnect() error {
	socket, err := net.Dial("tcp", "142.250.182.68:80")
	if err != nil {
		return err
	}
	defer socket.Close()

	fmt.Fprintf(socket, "GET / HTTP/1.0\r\n\r\n")
	written, err := io.Copy(io.Discard, socket)
	if err != nil {
		return err
	}

	log.Printf("Written %d bytes\n", written)
	return nil
}

func DnsResolver() error {
	_, err := net.LookupIP("google.com")
	if err != nil {
		return err
	}
	return nil
}

func DnsResolverCustom() error {
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	ips, err := resolver.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		return err
	}

	log.Println("IP address: ")
	for _, ip := range ips {
		log.Println(" -> :", ip)
	}
	return nil
}

func Ps() error {
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	isNumber := func(s string) bool {
		for _, c := range s {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if !isNumber(dir.Name()) {
			continue
		}

		cmd := "/proc/" + dir.Name() + "/cmdline"
		data, err := os.ReadFile(cmd)
		if err != nil {
			log.Println("ERROR:", err)
		}
		log.Println(dir.Name(), ":", string(data))
	}

	return nil
}

func PrintInterfaces() error {
	ifs, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, interface_ := range ifs {
		log.Println(interface_.Name)
	}
	return nil
}

func OutputMounts() error {
	log.Println("MOUNTS: List all mounts")
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	log.Println(string(data))

	return nil
}

var (
	server_addr *string
)

func TestServerClientConnection() error {
	if *server_addr == "" {
		conn, err := net.Listen("tcp", ":8080")
		if err != nil {
			return err
		}
		defer conn.Close()

		go func() {
			time.Sleep(3 * time.Second)
			conn.Close()
		}()

		newConn, err := conn.Accept()
		if err != nil {
			return err
		}

		log.Println("Accepted connection from ", newConn.RemoteAddr())
		newConn.Write([]byte("Hello"))

		var incommingMessage [1024]byte
		newConn.Read(incommingMessage[:])
		log.Println("Received message: ", string(incommingMessage[:]))
	} else {
		var conn net.Conn
		var err error

		for i := 0; i < 5; i++ {
			conn, err = net.Dial("tcp", *server_addr)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			} else {
				break
			}
		}

		if err != nil {
			log.Fatal("Cant connect to server")
		}
		defer conn.Close()

		conn.Write([]byte("Hello"))
		var incommingMessage [1024]byte
		conn.Read(incommingMessage[:])
	}
	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	programName := os.Getenv("BARB_SERVICE_NAME")
	log.Print("Program Name:", programName)
	log.SetPrefix(programName + ": ")

	server_addr = flag.String("servaddr", "", "Server address")
	flag.Parse()

	tests := []Tests{
		// Output,
		// SocketConnect,
		// DnsResolverCustom,
		// DnsResolver,
		// Ps,
		// PrintInterfaces,
		// OutputMounts,
		TestServerClientConnection,
	}

	GetFunctionName := func(i interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	}

	for _, test := range tests {
		name := GetFunctionName(test)
		log.Println("Running ", name)
		if err := test(); err != nil {
			log.Println("ERROR:", err)
		}
	}
}
