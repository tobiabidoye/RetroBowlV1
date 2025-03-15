package main

import(
    "sync"
    "fmt"
    "log" 
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/google/uuid"
)

//for our current structure players enter the server and then enter a specific room where they can talk with people in the same room
type Room struct{
    roomId string
    players map[string]*Player //map to track each player in each room at the time
    Mu sync.Mutex
}

type Player struct{
    playerId string
    Conn * websocket.Conn
}

const maxPlayers = 10

var(
    playersMu sync.Mutex
    players = make(map[string]*Player) //key value pair player 
    rooms = make(map[string]*Room)
    roomsMu sync.Mutex
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
    roomName := r.URL.Query().Get("room")
    
    if roomName == ""{
        log.Fatal("no room specified")
        return
    }    

    if err := addPlayer(id, conn, roomName); err != nil{
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
            return
        }
            log.Println("received from", id, string(msg))
            broadcastFunc(id, messageType, msg)
    } 


}

//to add players
func addPlayer(id string, conn *websocket.Conn, roomName string) error{

    roomsMu.Lock()
    room,exists := rooms[roomName]
    if !exists{
        //create a new room at this point
        val := make(map[string]*Player)
        room = &Room{roomId: roomName, players: val,Mu: sync.Mutex{} }
        rooms[roomName] = room
    }
    roomsMu.Unlock()


    //check for count of players in a specific room
    room.Mu.Lock()
    defer room.Mu.Unlock()
    playerCount := len(room.players)
    if (playerCount) >= maxPlayers{
        return fmt.Errorf("too many players")
    } 
    
    curPlayer := &Player{playerId:id, Conn:conn}
    //map which stores another map first map stores the room key is roomname value is the roomName
    //store players in the room by storing them in a map with key being id and value being the actual player
    room.players[id] = curPlayer
    
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
