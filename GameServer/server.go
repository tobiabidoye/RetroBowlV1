package main

import(
    "sync"
    "fmt"
    "log" 
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/google/uuid"
)

type Player struct{
    playerId string
    Conn * websocket.Conn
}


const maxPlayers = 10 
var(
    playersMu sync.Mutex
    players = make(map[string]*Player) //key value pair player 
)

func main(){
    //setup handler route
    http.HandleFunc("/connect", handleWebsocket)
    log.Println("started on :8080") 
    if err := http.ListenAndServe(":8080", nil); err != nil{
        log.Fatal(err)
    }
}

var upgrader = websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r * http.Request) bool {return true},

}

func handleWebsocket(w http.ResponseWriter, r * http.Request){    
    conn, err := upgrader.Upgrade(w , r, nil)
    if err != nil{        
        log.Fatal(err)
    }

    defer conn.Close()
    //generate a unique id
    id := uuid.NewString()    
    if err := addPlayer(id, conn); err != nil{
        //error handling for if adding player fails
        conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("error: %v", err)))
        return
    }
    //once function exits remove player 
    defer removePlayer(id)
    for{
        messageType, msg ,err := conn.ReadMessage()
        if err != nil{
            log.Println("received from", id, string(msg))
            broadcastFunc(id, messageType, msg)
        }
    } 


}

func addPlayer(id string, conn *websocket.Conn) error{
    playersMu.Lock()
     
    if len(players) >= maxPlayers{
        return fmt.Errorf("too many players")
    } 

    players[id] = &Player{playerId: id, Conn: conn}
    return nil

}

func removePlayer(id string){
    playersMu.Lock()

    delete(players, id)
    playersMu.Unlock()
}

//returns all players
func getAllPlayers() []*Player{
    playersMu.Lock()
    res := make([]*Player, 0, len(players))
    for _, p := range players{
        res = append(res, p) 
    }
    playersMu.Unlock() 
    return res

}

func broadcastFunc(senderId string, messageType int, data[]byte){   
    playersMu.Lock()
    defer playersMu.Unlock()
    for _, player := range players{
        if player.playerId == senderId{
            continue
        } 
    
        err := player.Conn.WriteMessage(messageType, data)
        
        if err != nil{
            log.Printf("error writinf to %s: %v\n", player.playerId, err)
        }
    }


        }
