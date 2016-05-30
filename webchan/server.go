package webchan

import (

    "github.com/satori/go.uuid"
    "github.com/gorilla/websocket"
    "net/http"
    "log"
    "sync"
    "time"
)




const (

    writeWait = 10 * time.Second
    pongWait = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
    maxMessageSize = 4096
)


var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
    }


type Connection struct {
    
    ws      *websocket.Conn

    send    chan  []byte

    server  *Server

    id string
}




type Server struct {

    http.Handler

    handlers

    connections     map[string]map[*Connection]struct{}

    rooms           map[*Connection]map[string]struct{}

    connectionLock  sync.RWMutex

    cids            map[string]*Connection
    cidsLock        sync.RWMutex

}


func (s *Server)ServerHTTP(w http.ResponseWriter,r *http.Request){

    // we need to do some auth

    if err := s.OnAuth(r); err != nil {
        
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte("HTTP status code returned!"))

        return
    }

    ws,err := upgrader.Upgrade(w,r,nil)

    if err != nil {
        log.Println(err) 
        return
    }

    c := &Connection{ 
                ws:ws,
                send:make(chan []byte,10),
                id:uuid.NewV4().String()}

    c.server = s

    // run into loop
    go c.writeLoop()
    go c.readLoop()

}


func (s *Server) onConnection(c *Connection){

    s.cidsLock.Lock()
    defer s.cidsLock.Unlock()

    s.cids[c.id] = c

}


func (s *Server) onCleanConnection(c *Connection) {

    s.connectionLock.Lock()
    defer s.connectionLock.Unlock()

    cos := s.connections

    byRoom,ok := s.rooms[c]

    if ok {
        
        for room := range byRoom {
                if curRoom,ok := cos[room];ok {
                    delete(curRoom,c)
                    if len(curRoom) == 0 {
                        delete (cos,room)    
                    }
                }
        }

        delete(s.rooms,c)
    }

    s.cidsLock.Lock()
    defer s.cidsLock.Unlock()

    delete(s.cids,c.id)

    c.ws.Close()
}


func (s *Server)BroadcastTo(room,message string,args interface{}){
    
    
}



func (s *Server)BroadcastToAll(message string,args interface{}){
    
    
    
}



func (s *Server)List(room string) ([]*Connection,error){



    return nil,nil
}




//  


func (c *Connection) writeLoop(){

    ticker := time.NewTicker(pingPeriod)
    
    defer func() {

        ticker.Stop()
        c.ws.Close()

    }()

    for {
        
        select {
        case message,ok := <- c.send:
            if !ok {
                c.write(websocket.CloseMessage,[]byte{})
                continue
            }
            if err := c.write(websocket.TextMessage, message); err != nil {
                return
            }
        case <- ticker.C:
            if err := c.write(websocket.PingMessage,[]byte{}); err != nil {
                continue    
            }
        }
        
    }


}


// for outside
func (c *Connection) Write(message  []byte){

    c.send <- message

}


// for internal
func (c *Connection) write(mt int,payload []byte) error {

    c.ws.SetWriteDeadline(time.Now().Add(writeWait))
    return c.ws.WriteMessage(mt, payload)

}


func (c *Connection) readLoop(){

    defer func(){
        c.server.onCleanConnection(c)
    }()

    c.ws.SetReadLimit(maxMessageSize)
    c.ws.SetReadDeadline(time.Now().Add(pongWait))
    c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

    for {

        // we only surport TextMessage

        _,message,err := c.ws.ReadMessage()

        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway){
                log.Printf(" websocekt read error",err)
            }

            c.server.OnDisconnection(c,err.Error())

            break
        }

        // onMessage
        go c.server.OnMessage(c,message)
    }
}





/**
new server
*/

func NewServer() *Server {
    
    s := Server{}
    s.connections   = make(map[string]map[*Connection]struct{})
    s.rooms         = make(map[*Connection]map[string]struct{})
    s.cids          = make(map[string]*Connection)

    return &s
}
