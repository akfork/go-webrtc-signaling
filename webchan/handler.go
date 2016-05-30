package webchan




type handlers struct{

    OnConnection        func(c *Connection,args interface{})
    OnAuth              func(args interface{}) error
    OnDisconnection     func(c *Connection,message string)
    OnMessage           func(c *Connection,message []byte)

}




