

package signaling



import (

    "encoding/json"
)





const (
    
)



type Message struct {
    UserId  string      `json:"user_id"` 
    Type    string      `json:"type"`
    Room    string      `json:"room"`
    message interface{} `json:"message"`

}



