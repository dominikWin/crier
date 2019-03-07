package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

func parseArgs() (uint16, string) {

	var port uint16 = 6782
	set_port := false

	var redis_location string = "localhost:6379"
	set_redis_location := false

	args := os.Args[1:]

	if len(args) == 0 {
		return port, redis_location
	}

	if len(args)%2 != 0 {
		log.Fatalln("Bad number of arguments, should be <key> <value> pairs!")
	}

	for i := 0; i < len(args)/2; i++ {
		key := args[i*2]
		value := args[i*2+1]
		switch key {
		case "--port":
			if set_port {
				log.Fatalln("Duplicated argument!")
			}
			set_port = true

			n, err := strconv.Atoi(value)
			if err != nil {
				log.Fatalln("Failed to parse port number!")
			}
			if n < 0 || n > 0xffff {
				log.Fatalln("Port number out of range!")
			}
			port = uint16(n)
		case "--redis":
			if set_redis_location {
				log.Fatalln("Duplicated argument!")
			}
			set_redis_location = true

			redis_location = value
		default:
			log.Fatalf("Unknown CLI argument (%s) given!\n", key)
		}
	}

	return port, redis_location
}

func main() {
	port, _ := parseArgs()
	pid := os.Getpid()

	log.Println("!!! Crier is starting !!!")

	startWebServer(port)

	// Quick ref sheet
	fmt.Printf("\n")
	fmt.Printf("    Port: %d\n", port)
	fmt.Printf("    PID: %d\n", pid)
	fmt.Printf("\n")

	select {}

	log.Println("Crier reached the end")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Success!\n")
}

func runWebServer(listener net.Listener) {
	log.Fatalln("HTTP Server Quit!", http.Serve(listener, nil))
}

func startWebServer(port uint16) {
	bind_addr := fmt.Sprintf("0.0.0.0:%d", port)

	listener, err := net.Listen("tcp", bind_addr)
	if err != nil {
		log.Fatalf("Failed to bind to %s", bind_addr)
	}

	http.HandleFunc("/", handler)

	go runWebServer(listener)

	log.Println("Started web server")
}
