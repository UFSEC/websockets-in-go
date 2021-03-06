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

type client struct {
	id int
	datachan chan []byte
}

var register chan bool = make(chan bool, 10)
var newClient chan client = make(chan client, 10)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Connection received")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}


	// register new client
	register <- true
	// get my data channel
	client := <-newClient

	defer func(){ ws.Close() }()

	for datapoint := range client.datachan {
		if err := ws.WriteMessage(websocket.TextMessage, datapoint); err != nil {
			log.Println(err)
			return
		}
	}
	log.Println("Closing connection")
}	

func collect(data chan []byte, register chan bool, newClient chan client) {
	var clients map[int]client = make(map[int]client)
	var currId = 0
	for {
		select {
			case datapoint := <-data:
				for _, client := range clients {
					client.datachan <- datapoint
				}
			case <-register:
				c := client{currId, make(chan []byte, 20)}
				clients[currId] = c
				newClient <- c
				currId++
			}
	}
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
	data := make(chan []byte, 100)
	go collectTop(data)
	go collectIoStat(data)
	go collect(data, register, newClient)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}