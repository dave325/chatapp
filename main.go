// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var addr = flag.String("addr", ":3001", "http service address")

var result struct {
	User string
}

type documents struct {
	Total int64 "json:total"
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

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
	log.Println("Starting")
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	log.Println("Connected to DB")
	defer cancel()
	client, mongoErr := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:root@127.0.0.1:27017"))
	defer func() {
		if mongoErr = client.Disconnect(ctx); mongoErr != nil {
			panic(mongoErr)
		}
	}()

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		collection := client.Database("chat").Collection("users")

		total, colerr := collection.CountDocuments(ctx, bson.M{})
		log.Println(total)
		if colerr != nil {
			log.Println(colerr)
		}
		totalStruct := documents{Total: total}
		js, err := json.Marshal(totalStruct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(js)
	})
	http.HandleFunc("/userList/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		collection := client.Database("chat").Collection("users")

		keys, ok := r.URL.Query()["user"]
		//roomkeys, roomok := r.URL.Query()["room"]

		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'key' is missing")
			return
		}

		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]

		log.Println(key)
		filter := bson.M{"user": key}

		err := collection.FindOne(ctx, filter).Decode(&result)
		log.Println(result)
		if err != nil {
			log.Println("ye")
			if err.Error() == "mongo: no documents in result" {
				_, insertErr := collection.InsertOne(context.Background(), bson.M{"user": key})
				if insertErr != nil {
					log.Println(insertErr)
					return
				}
			}
		}
		serveUserListWs(hub, w, r, key)
	})
	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		serveWs(hub, w, r)
	})
	httperr := http.ListenAndServe(*addr, nil)
	if httperr != nil {
		log.Fatal("ListenAndServe: ", httperr)
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
