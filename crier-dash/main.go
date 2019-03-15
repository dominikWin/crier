package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var redis_addr string

var g_redis *redis.Client
var g_redisMutex sync.Mutex

func parseArgs() (uint16, string) {

	var port uint16 = 8000
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
	redis_addr = redis_location

	redisClient := redis.NewClient(&redis.Options{
		Addr: redis_location,
		DB:   0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalln("Failed to connect to redis")
	}

	g_redis = redisClient

	bind_addr := fmt.Sprintf("0.0.0.0:%d", port)

	listener, err := net.Listen("tcp", bind_addr)
	if err != nil {
		log.Fatalf("Failed to bind to %s", bind_addr)
	}

	http.HandleFunc("/", handle_index)
	http.HandleFunc("/js/crier.js", handle_crierjs)
	http.HandleFunc("/ws", handle_ws)
	http.HandleFunc("/message/", handle_message)
	http.Serve(listener, nil)
}

func handle_crierjs(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "crier.js")
}

func handle_index(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		w.WriteHeader(404)
		fmt.Fprintln(w, "404")
	} else {
		http.ServeFile(w, r, "index.html")
	}
}

func get_message(id string) []byte {
	g_redisMutex.Lock()
	defer g_redisMutex.Unlock()

	result, err := g_redis.XRange("crier", id, id).Result()
	if err != nil {
		log.Panic(err)
	}

	msg := result[0].Values["message"]

	if msg_bin, ok := msg.(string); ok {
		return []byte(msg_bin)
	} else {
		log.Panic(ok)
		return []byte{}
	}
}

func handle_message(w http.ResponseWriter, r *http.Request) {
	id := r.RequestURI[9:]
	bin := get_message(id)

	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprint(w, string(bin))
}

func handle_ws(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redis_addr,
		DB:   0,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Fatalln("Failed to connect to redis")
	}

	msg_type, msg, err := ws.ReadMessage()
	if err != nil {
		log.Panic(err)
	}
	if msg_type != websocket.TextMessage {
		log.Fatalln("Bad message type")
	}
	last_message_id := string(msg)

	for {
		cmd := redis.XReadArgs{
			Streams: []string{"crier", last_message_id},
			Block:   0 * time.Millisecond,
			Count:   25,
		}

		results, err := redisClient.XRead(&cmd).Result()
		if err != nil {
			log.Panic(err)
		}

		if len(results) != 1 {
			log.Fatalln("Got stream data from bad number of streams")
		}

		result := results[0]
		for i := 0; i < len(result.Messages); i++ {
			message := result.Messages[i]
			last_message_id = message.ID

			msg_host := ""
			if str, ok := message.Values["host"].(string); ok {
				msg_host = str
			} else {
				log.Panic(ok)
			}

			msg_message := ""
			if str, ok := message.Values["message"].(string); ok {
				msg_message = str
			} else {
				log.Panic(ok)
			}

			lines := strings.SplitN(msg_message, "\n", 2)

			add_ellipses := false
			if len(lines) > 1 {
				msg_message = lines[0]
				add_ellipses = true
			}

			if len(msg_message) > 90 {
				msg_message = msg_message[0:87]
				add_ellipses = true
			}

			if add_ellipses {
				msg_message += "..."
			}

			msg_message = html.EscapeString(msg_message)

			ws_msg := map[string]string{"id": message.ID, "host": msg_host, "message_head": msg_message}

			json_string, err := json.Marshal(ws_msg)
			if err != nil {
				log.Panic(err)
			}

			ws.WriteMessage(websocket.TextMessage, []byte(json_string))
		}
	}
}
