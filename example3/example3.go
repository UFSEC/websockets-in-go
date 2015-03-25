package main

import (
    "log"
    "fmt"
    "net/http"
    "os"
    "io"
    "encoding/json"
    "github.com/gorilla/websocket"
)

type Message struct {
    Type string
    Data interface{}
}

type Position struct {
    Row int
    Col int
}

type Init struct {
    Who string
    Pos Position
}

type GameStart struct {
    CopPos Position
    RobPos Position
}

type GameOver struct {
    Who string
    Pos Position
}

type Update struct {
    Who string
    Pos Position
}

type MoveCountUpdate struct {
    Count int
}

type Player struct {
    pos Position
    update chan Message
}

var upgrader = &websocket.Upgrader{
    ReadBufferSize: 1024, WriteBufferSize: 1024,
}

var updatesInc chan Update = make(chan Update, 100)
var newClient chan bool = make(chan bool, 10)
var newUpdateChan chan chan Message = make(chan chan Message, 10)

func updateAll(msg Message, cop chan Message, rob chan Message, spectators []chan Message) {
    cop <- msg
    rob <- msg
    for _, specChan := range spectators {
        specChan <- msg
    }
}

func stateManager() {
    var cop *Player = nil
    var rob *Player = nil
    var moveCount int = 100
    var spectators []chan Message

    for {
        select {
            case <-newClient:
                if rob == nil {
                    rob = &Player{Position{0,0}, make(chan Message, 100)}
                    newUpdateChan <- rob.update
                    rob.update <- Message{"init", Init{"rob", rob.pos}}
                } else if cop == nil {
                    cop = &Player{Position{9,9}, make(chan Message, 100)}
                    newUpdateChan <- cop.update
                    cop.update <- Message{"init", Init{"cop", cop.pos}}
                    // both players are here, let's start the game!
                    msg := Message{"start", GameStart{cop.pos, rob.pos}}
                    updateAll(msg, cop.update, rob.update, spectators)
                    // rob.update <- msg
                    // cop.update <- msg
                } else { // send a nil to indicate we've already got our 2 players
                    newChan := make(chan Message, 100)
                    newUpdateChan <- newChan
                    newChan <- Message{"init", Init{"spectator", Position{0,0}}}
                    spectators = append(spectators, newChan)
                }
            case u := <-updatesInc:
                if u.Who == "rob"{
                    rob.pos = u.Pos
                } else if u.Who == "cop" {
                    cop.pos = u.Pos
                    moveCount = moveCount-1
                }
                if rob.pos.Row == cop.pos.Row && rob.pos.Col == cop.pos.Col {
                    // the game is over, cop wins!
                    msg := Message{"winner", GameOver{"cop", cop.pos}}
                    // cop.update <- msg
                    // rob.update <- msg
                    updateAll(msg, cop.update, rob.update, spectators)
                    log.Println("Game over!")
                    return
                }
                if moveCount == 0 {
                    // the game is over, robber wins!
                    msg := Message{"winner", GameOver{"rob", rob.pos}}
                    // cop.update <- msg
                    // rob.update <- msg
                    updateAll(msg, cop.update, rob.update, spectators)
                    log.Println("Game over!")
                    return
                }

                // neither has won, so just send an update about their positions        
                if u.Who == "cop" {
                    // if it is a cop who moved, we send
                    // a modified move cout
                    msg := Message{"update", Update{"cop", cop.pos}}
                    // cop.update <- msg
                    // rob.update <- msg
                    updateAll(msg, cop.update, rob.update, spectators)
                    msg = Message{"count", MoveCountUpdate{moveCount}}
                    // cop.update <- msg
                    // rob.update <- msg
                    updateAll(msg, cop.update, rob.update, spectators)
                } else if u.Who == "rob" {
                    msg := Message{"update", Update{"rob", rob.pos}}
                    // cop.update <- msg
                    // rob.update <- msg
                    updateAll(msg, cop.update, rob.update, spectators)
                }
            } 
    }
    
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Connection received")
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    defer func(){ ws.Close() }()

    newClient <- true
    updateOut := <-newUpdateChan
    if updateOut == nil { // we've already got 2 players
        fmt.Sprint(w, "Sorry, this game is full!")
        return
    }

    go func() { // this forever writes data from the updateOut
        for {   // to the websocket
            u, _ := json.Marshal(<-updateOut)
            if err := ws.WriteMessage(websocket.TextMessage, u); err != nil {
                log.Println(err)
                return
            }
        }
    }()

    for { // this forever reads from websocket writing to the update channel
        _, message, err := ws.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }
        var u Update
        if err := json.Unmarshal(message, &u); err == nil {
            updatesInc <- u
        }
    }
    log.Println("Closing connection")
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
    go func(){
        for {
            stateManager()
        }
    }() 

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/ws", wsHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
