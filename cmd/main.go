package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// gotta split this

var clients = make(map[string]map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)                      // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Define our message object
type Message struct {
	Type   string `json:"type"`
	Data   string `json:"data"`
	RoomID string `json:"roomID"`
}

func main() {
	router := http.NewServeMux()

	// Configure websocket route
	router.HandleFunc("/ws/{roomID}", handleConnections)
	router.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>Server Running </h1>")
	})

	http.Handle("/", router)

	// Starting & listening for incoming chat messages
	go handleMessages()

	// Starting the server
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients[msg.RoomID] {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("error %v", err)
				client.Close()
				delete(clients[msg.RoomID], client)
			}
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	roomID := r.PathValue("roomID")

	log.Println("RoomID %s has been created.", roomID)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	if _, ok := clients[roomID]; !ok {
		clients[roomID] = make(map[*websocket.Conn]bool)
	}
	clients[roomID][ws] = true

	for {

		var msg Message

		err := ws.ReadJSON(&msg)

		if err != nil {
			log.Println("error %v", err)
			delete(clients[roomID], ws)
			break
		}
		broadcast <- msg
	}
}
