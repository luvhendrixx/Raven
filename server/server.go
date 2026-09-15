package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// setup the Websocket buffers and their sizes
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// this rep a single conn user/client sock.
// it holds the sock conn (*websocket.Conn) and its own private mailbox (send chan []byte) for outgoing msgs
type Client struct {
	conn *websocket.Conn // the actual conn thru which this particular client comms with the server
	send chan []byte     // giving each client a private outbound queue/mailbox...basically where the raw bytes they send are "kept" so as to wait going thru for processing by the server
}

// this acts as the central hub for every client...
// it maintains all active conn and routes msgs to the right people/clients (hence the name)
type Hub struct {
	clients    map[*Client]bool // the client
	register   chan *Client     // registration req from clients when they conn
	unregister chan *Client     // unregistration req from client when they disconn
	broadcast  chan []byte      // inbound msgs from clients to broadcast to other clients
}

// ACTUALLY creating a hub (the above was just a struct...a blueprint)
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// run the hub
// the hub is always in a loop listening and serving
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register: // register the client
			h.clients[client] = true // register as True cuz they joined

		case client := <-h.unregister: // unregister the client
			if _, ok := h.clients[client]; ok { // check if the client exists in the map..if so..
				delete(h.clients, client) // delete them
				close(client.send)        // close their broadcast (messaging mailbox)
			}
		case msg := <-h.broadcast:
			// loop thru all connected/registered clients and broadcast the msg/payload to them
			for client := range h.clients {
				select {
				case client.send <- msg:
				// if unable to do the above cuz the chan is full...
				// don't wait...execute default (delete and close in this case)
				default:
					// if client's send buffer is full, treat them as disconnected
					delete(h.clients, client) // delete them
					close(client.send)        // close their messaging box/capabilities
				}
			}
		}
	}
}

// readPump is just a var name that means...
// read = read incoming data, pump = continuosly/move/flow that data somewhere
func (c *Client) readPump(h *Hub) {
	// an annonymus func
	defer func() { // means..create this func but defer its exe until readPump returns
		h.unregister <- c
		c.conn.Close()
	}() // when u define an annonymus func...you call/execute it at the same time/immediately
	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			break // we use break over return so if the loop fails...the other stuff runs without disruption
		}
		h.broadcast <- payload
	}
}

// this CONSTANTLY (hence the loop) takes data from the mailbox and sends that data inside the mailbox and sends it thru every client's websock conn
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func main() {
	hub := newHub()
	go hub.run() // start the Hub

	// init the router
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		// create the Client struct and get the ptr to it at the same time
		client := &Client{conn: conn, send: make(chan []byte, 256)}

		hub.register <- client

		go client.writePump()
		client.readPump(hub) // blocks until disconnect
	})

	// bc this is blocking..we place it outside our logic
	fmt.Println("HTTP -> WS server now running on port 8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
