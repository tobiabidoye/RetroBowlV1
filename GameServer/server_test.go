package main

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

//test to create a connection
func createTestWebSocketConn(t * testing.T) *websocket.Conn{
   
    mux := http.NewServeMux()
    server := &http.Server{ 
        Handler: mux,
    }

    mux.HandleFunc("/ws", func(w http.ResponseWriter, r * http.Request){
        upgrader := websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {return true},
        }
        _,err := upgrader.Upgrade(w , r, nil)
        
        if err != nil{
            t.Fatalf("failed to upgrade websockets, %v", err)
        } 
    })
    
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil{
        t.Fatalf("failed to start test server %v", err) 
    }
    server.Addr = listener.Addr().String()

    go func() {
        if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
            t.Errorf("websocket server failed to start: %v", err)

        }  
    }()
    
    t.Cleanup(func(){
        server.Close()
    })

    time.Sleep(100 * time.Millisecond)
    conn, _, err := websocket.DefaultDialer.Dial("ws://"+server.Addr+"/ws", nil)
    
    if err != nil{
        t.Fatalf("failed to connect websocket server: %v", err)
    }
    
    return conn
}

func TestAddPlayer(t * testing.T){
    rooms = make(map[string]*Room)
    roomsMu = sync.Mutex{}
    conn := createTestWebSocketConn(t)
    err := addPlayer("player1", conn, "room1")

    if err != nil{
        t.Errorf("error adding player %v", err)   
    }

    if _, exists := rooms["room1"]; !exists{
        t.Errorf("room was not created")
    }

    if _, exists := rooms["room1"].players["player1"]; !exists{
        t.Errorf("player was not added to the room")
    }
} 

//function to test adding a player
func TestRoomLimit(t *testing.T){
 
    rooms = make(map[string]*Room)
    roomsMu = sync.Mutex{}
    conn := createTestWebSocketConn(t)
    roomName := "room1"

    for i := 0; i < maxPlayers; i++{
        err := addPlayer("player"+strconv.Itoa(i), conn, roomName)
        if err != nil{
            t.Errorf("unexpected error %v",err)
        }
    }

    err := addPlayer("player_extra", conn, roomName)
    if err == nil{
        t.Errorf("expected error but got nil, room count surpasses maximum")
    }
}


