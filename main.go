// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":3001", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// // websockets.go
// package main

// import (
// 	"fmt"
// 	"net/http"

// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// type UserMessage struct {
// 	User    string `json:"User"`
// 	Message string `json:"Message"`
// }

// func main() {
// 	// Needed here so go will allow outside origins, currently set to allow everything
// 	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

// 	// Listens to frontend
// 	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
// 		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
// 		fmt.Println(r.Method)

// 		for {
// 			fmt.Println("1")
// 			var msg UserMessage
// 			// Read message from browser
// 			err := conn.ReadJSON(&msg)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			fmt.Println('2')
// 			// Print the message to the console
// 			//fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

// 			// Write message back to browser
// 			if err = conn.WriteJSON(r); err != nil {
// 				return
// 			}
// 		}
// 	})

// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		http.ServeFile(w, r, "websocket.html")
// 	})

// 	http.ListenAndServe(":3001", nil)
// }
