package main


import (
    "./webchan"
    "log"
    "net/http"
)




func main(){

    server := webchan.NewServer()

    
    server.OnAuth = func(args interface{}) error {

            log.Println("OnAuth")

            return nil
        }


    server.OnConnection = func(c *webchan.Connection,ars interface{}){

                log.Println("OnConnection")
            }



    server.OnMessage = func(c *webchan.Connection,message []byte) {

                log.Println("OnMessage")
        }



    server.OnDisconnection = func(c *webchan.Connection,message string){
        
                log.Println("OnDisconnection")
        }




    serveMux := http.NewServeMux()

    serveMux.Handle("/ws",server)


    log.Println("Starting server...")

    http.ListenAndServe(":8000",serveMux)


}
