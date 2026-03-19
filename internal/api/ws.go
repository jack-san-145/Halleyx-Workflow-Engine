package api

import (
	"log"
	"net/http"

	"halleyx-workflow-docker/internal/ws"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSocketHandler upgrades HTTP connections and registers clients.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}

	client := &ws.Client{Hub: ws.DefaultHub, Conn: conn, Send: make(chan []byte, 256)}
	ws.DefaultHub.Register <- client

	log.Println("WS client connected")

	go client.WritePump()
	client.ReadPump()

	log.Println("WS client disconnected")
}
