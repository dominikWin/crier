package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-redis/redis"
)

var activeConnections uint64
var rejectConnections uint32

var redisClient *redis.Client
var redisMutex sync.Mutex

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
	port, redis_location := parseArgs()
	pid := os.Getpid()

	log.Println("!!! Crier is starting !!!")

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGTERM)
	signal.Notify(shutdown, syscall.SIGINT)

	httpServer, httpServerOnclose := startWebServer(port)

	redisClient = redis.NewClient(&redis.Options{
		Addr: redis_location,
		DB:   0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalln("Failed to connect to redis")
	}

	// Quick ref sheet
	fmt.Printf("\n")
	fmt.Printf("    Port: %d\n", port)
	fmt.Printf("    PID: %d\n", pid)
	fmt.Printf("\n")

	<-shutdown
	log.Println("Shutdown signal received! Closing...")

	stopWebServer(httpServer, httpServerOnclose)

	log.Println("Crier reached the end")
}

func stopWebServer(httpServer *http.Server, onclose chan int) {
	atomic.StoreUint32(&rejectConnections, 1)

	// Try not to force shutdown until necessary
	for i := 0; i < 500; i++ {
		if atomic.LoadUint64(&activeConnections) > 0 {
			time.Sleep(time.Millisecond)
		} else {
			break
		}
	}

	err := httpServer.Shutdown(context.Background())
	if err != nil {
		log.Panic(err)
	}

	<-onclose
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Success!\n")
}

type server struct{}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadUint32(&rejectConnections) > 0 {
		w.WriteHeader(503)
	} else {
		atomic.AddUint64(&activeConnections, 1)
		handle(w, r)
		atomic.AddUint64(&activeConnections, ^uint64(0))
	}
}

func runWebServer(svr *http.Server, listener net.Listener, onclose chan int) {
	err := svr.Serve(listener)
	if err == http.ErrServerClosed {
		log.Println("Web server closed")
		onclose <- 1
	} else {
		log.Fatalln("Web server crashed! ", err)
	}
}

func startWebServer(port uint16) (*http.Server, chan int) {
	bind_addr := fmt.Sprintf("0.0.0.0:%d", port)

	listener, err := net.Listen("tcp", bind_addr)
	if err != nil {
		log.Fatalf("Failed to bind to %s", bind_addr)
	}

	h := http.Server{Handler: &server{}}
	onclose := make(chan int)

	go runWebServer(&h, listener, onclose)

	log.Println("Started web server")
	return &h, onclose
}
