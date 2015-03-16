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
}

var newClient chan bool = make(chan bool, 100)
var newDataChan chan chan[]byte = make(chan chan []byte, 100)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection received")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// register new client
	newClient <- true
	// get my data channel
	datachan := <-newDataChan

	for datapoint := range datachan {
		if err := ws.WriteMessage(websocket.TextMessage, datapoint); err != nil {
			log.Println(err)
			ws.Close()
			return
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

func collect(data chan []byte, newClient chan bool, newDataChan chan chan []byte) {
	var clients []chan []byte
	for {
		select {
		case datapoint := <-data:
			for _, client := range clients {
				client <- datapoint
			}
		case <-newClient:
			newChan := make(chan []byte, 100)
			clients = append(clients, newChan)
			newDataChan <- newChan
		}
	}
}

func main() {
	data := make(chan []byte, 100)
	go collectTop(data)
	go collectIoStat(data)
	go collect(data, newClient, newDataChan)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}