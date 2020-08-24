package hub

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type standardMessage struct {
	XMLName xml.Name `xml:"email"`
	Title   string   `xml:"title"`
	Body    string   `xml:"body"`
	Comment string   `xml:",comment"`
}
type Hub struct {
	clients map[*Client]bool
}

type NumberMessage struct {
	Number int `json:"number"`
}

func NewHub() *Hub {
	return &Hub{

		clients: make(map[*Client]bool),
	}
}

// serveWs handles websocket requests from the peer.
func (h *Hub) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: h, conn: conn, send: make(chan []byte, 256)}
	h.register(client)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	// go client.readPump()
}

func (h *Hub) register(client *Client) {
	h.clients[client] = true
}

func (h *Hub) unregister(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

func (h *Hub) broadcast(msg []byte) {

	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) HandleStandardMessage(b []byte) {
	var msg standardMessage

	err := xml.Unmarshal(b, &msg)

	if err != nil {
		log.Print("error unmarshaling message")
	} else {
		fmt.Print(msg.Body)
	}
}
