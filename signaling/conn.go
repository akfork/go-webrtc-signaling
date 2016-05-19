
package signaling


import (
    "github.com/gorilla/websocket"
    "net/http"
    "log"
    "time"
    "github.com/satori/go.uuid"
)




const (
    // Time allowed to write a message to the peer.
    
    writeWait = 10 * time.Second

     // Time allowed to read the next pong message from the peer.
    pongWait = 60 * time.Second

    // Send pings to peer with this period. Must be less than pongWait.
    pingPeriod = (pongWait * 9) / 10
    
    // Maximum message size allowed from peer.
    maxMessageSize = 4096
)


var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
}


type Connection struct {
    
    ws  *websocket.Conn

    send chan []byte

    id  string

}



func (c *Connection) readPump(){
    
    defer func() {
        h.unregister <- c
        c.ws.Close()
    }()

    c.ws.SetReadLimit(maxMessageSize)
    c.ws.SetReadDeadline(time.Now().Add(pongWait))
    c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })


    var mess Message;

    for {

        if err := c.ws.ReadJSON(&mess); err != nil {
            
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
                log.Printf("error: %v", err)
            }
            break
        }

        h.broadcast <-  mess

    }
    
}



func (c *Connection) write(mt int,payload []byte) error{
    c.ws.SetWriteDeadline(time.Now().Add(writeWait))
    return c.ws.WriteMessage(mt,payload)
}

func (c *Connection) writePump(){

    ticker := time.NewTicker(pingPeriod)
    
    defer func(){
        ticker.Stop()
        c.ws.Close()
    }()

    for {
        
        select {
            
        case message,ok := <- c.send:
            if !ok {
                c.write(websocket.CloseMessage,[]byte{})   
                return
            }

            if err := c.ws.WriteJSON(&message); err != nil {
                return    
            }
        }
        
    }
    
}


func serveWs(w http.ResponseWriter,r *http.Request){
    
    ws,err := upgrader.Upgrade(w,r,nil)

    if err != nil {
        log.Println(err) 
        return
    }

    c := &connection{send: make(Message,256),ws: ws,id:uuid.NewV4().String()}

    h.register <- c

    go c.writePump()

    c.readPump()

}
