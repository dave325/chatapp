// websockets.go
package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type UserMessage struct {
	User    string `json:"User"`
	Message string `json:"Message"`
}

func main() {
	// Needed here so go will allow outside origins, currently set to allow everything
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Listens to frontend
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {
			var msg UserMessage
			// Read message from browser
			err := conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Print the message to the console
			//fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteJSON(msg); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websocket.html")
	})

	http.ListenAndServe(":3001", nil)
}
