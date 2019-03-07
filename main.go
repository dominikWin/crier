package main

import (
	"log"
	"os"
	"strconv"
)

func parseArgs() uint16 {

	var port uint16 = 6782

	args := os.Args[1:]

	if len(args) == 0 {
		return port
	}

	if len(args)%2 != 0 {
		log.Fatalln("Bad number of arguments, should be <key> <value> pairs!")
	}

	switch args[0] {
	case "-p":
		fallthrough
	case "--port":
		i, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalln("Failed to parse port number!")
		}
		if i < 0 || i > 0xffff {
			log.Fatalln("Port number out of range!")
		}
		port = uint16(i)
	default:
		log.Fatalln("Unknown CLI argument given!")
	}

	return port
}

func main() {
	port := parseArgs()
	pid := os.Getpid()

	log.Println("!!! Crier is starting !!!")
	log.Printf("Port: %d  PID: %d", port, pid)
}
