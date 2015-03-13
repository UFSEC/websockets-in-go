package main

import (
	"log"
	"net/http"
	"io"
	"os"
	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize: 1024, WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true }}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection received")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("Client sent: %s\n", string(message))
		if err := ws.WriteMessage(websocket.TextMessage, message); err != nil {
			ws.Close()
			break
		}
	}
	log.Println("Closing connection")
	ws.Close()
}	

func homeHandler(w http.ResponseWriter, req *http.Request) {
	file, err := os.Open("index.html")
	if err != nil {
		log.Println(err)
		return
	}
	io.Copy(w, file)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}