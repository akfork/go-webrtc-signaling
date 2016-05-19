

package signaling



type hub struct {

    connections map[string]*Connection

    broadcast chan *Message

    register chan *Connection

    unregister chan *Connection
}


var h = hub{

    broadcast:      make(chan *Message,10)
    register:       make(chan *Connection,10)


}
