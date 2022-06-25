package websocket

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
	database "gitlab.com/mgdi/kongroo-c2/c2/database/mongo"
)

type Message struct {
	Agent   database.AgentInfo `json:"agent"`
	Message string             `json:"message"`
}
type Hub struct {
	Clients   map[*websocket.Conn]bool
	Broadcast chan Message
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Read(hub *Hub, client *websocket.Conn) {
	select {}
	// for {
	// var message Message
	// err := client.ReadJSON(&message)
	// if !errors.Is(err, nil) {
	// 	log.Printf("error occurred: %v", err)
	// 	delete(hub.clients, client)
	// 	break
	// }
	// log.Println(message)

	// // Send a message to hub
	// hub.broadcast <- Message{"Test"}
	// }
}
func NewHub() *Hub {
	return &Hub{
		Clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case message := <-h.Broadcast:
			for client := range h.Clients {
				if err := client.WriteJSON(message); !errors.Is(err, nil) {
					log.Printf("error occurred: %v", err)
				}
			}
		}
	}
}
