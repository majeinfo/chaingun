package main

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	log.Debug("newHub")
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		log.Debugln("Loop Hub run()")
		log.Debugf("Count of goroutines=%d", runtime.NumGoroutine())
		select {
		case client := <-h.register:
			log.Debugf("Client %v registers", client)
			h.clients[client] = true
		case client := <-h.unregister:
			log.Debug("Client unregisters")
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			log.Debug("broadcast received")
			for client := range h.clients {
				log.Debugf("Broadcast to client %v", client)
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
