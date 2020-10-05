// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var addr = flag.String("addr", ":3001", "http service address")
var ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
var client, mongoErr = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:root@127.0.0.1:27017"))

var roomMap map[string]*Hub = make(map[string]*Hub)

var result struct {
	User string
}

type documents struct {
	CurrentUser int64    `json:"currentUser"`
	Users       []string `json:"users"`
}

type Rooms struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Users    []string           `json:"users" bson:"users"`
	Messages []*UserMessage     `json:"messages,omitempty"  bson:"messages,omitempty"`
}

var rooms Rooms
var messages []*UserMessage

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
	// Needed here so go will allow outside origins, currently set to allow everything
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	defer cancel()

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
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		collection := client.Database("chat").Collection("users")

		keys, ok := r.URL.Query()["user"]
		//roomkeys, roomok := r.URL.Query()["room"]
		total, colerr := collection.CountDocuments(ctx, bson.M{})
		log.Println(total)
		if colerr != nil {
			log.Println(colerr)
		}
		currentUserID := ""
		if ok && len(keys[0]) > 0 {
			// Query()["key"] will return an array of items,
			// we only want the single item.
			key := keys[0]

			filter := bson.M{"user": key}

			err := collection.FindOne(ctx, filter).Decode(&result)

			if err != nil {
				if err.Error() == "mongo: no documents in result" {
					currentUserIDInt64 := total + 1
					currentUserIDInt := int(currentUserIDInt64)
					currentUserID = strconv.Itoa(currentUserIDInt)
					_, insertErr := collection.InsertOne(context.Background(), bson.M{"user": currentUserID, "available": true})
					if insertErr != nil {
						log.Println(insertErr)
						return
					}
				}
			} else {
				result, err := collection.UpdateOne(
					ctx,
					bson.M{"user": key},
					bson.D{
						{"$set", bson.M{"available": true}},
					},
				)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
			}
		} else {
			currentUserIDInt64 := total + 1
			currentUserIDInt := int(currentUserIDInt64)
			currentUserID = strconv.Itoa(currentUserIDInt)
			_, insertErr := collection.InsertOne(context.Background(), bson.M{"user": currentUserID, "available": true})
			if insertErr != nil {
				log.Println(insertErr)
				return
			}
		}

		totalUser := make([]string, total*2)
		if total > 0 {

			cur, err := collection.Find(ctx, bson.M{"available": true})
			if err != nil {
				log.Fatal(err)
			}
			currentIdx := 0
			for cur.Next(ctx) {
				err := cur.Decode(&user)
				if err != nil {
					log.Fatal(err)
				}
				totalUser[currentIdx] = user.User
				currentIdx++
				// do something with result....
			}
		}

		totalStruct := documents{CurrentUser: total, Users: totalUser}
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

		serveUserListWs(hub, w, r)
	})

	http.HandleFunc("/checkChat/", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		collection := client.Database("chat").Collection("rooms")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer r.Body.Close()
		var currentRoom Rooms
		err = json.Unmarshal(b, &currentRoom)
		if err != nil {
			if err != io.EOF {
				fmt.Println("aklsjflasdjflsj")
				fmt.Println(err)
				http.Error(w, err.Error(), 500)
				return
			}
		}
		bsonArr := bson.A{}
		for _, user := range currentRoom.Users {
			bsonArr = append(bsonArr, user)
		}

		// Search if a db already has these users in a room
		filter := bson.D{{"users", bson.D{{"$all", bsonArr}}}}
		err = collection.FindOne(ctx, filter).Decode(&currentRoom)

		// If the no rooms exist, create an entry in db
		if err != nil {
			insertID, _ := collection.InsertOne(context.Background(), bson.M{"users": currentRoom.Users})
			currentRoom.ID = insertID.InsertedID.(primitive.ObjectID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			js, err := json.Marshal(&currentRoom)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(js)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			filter := bson.M{"room": currentRoom.ID}
			cursor, cursorErr := collection.Find(ctx, filter)
			if cursorErr != nil {
				log.Println(cursorErr)
			}
			defer cursor.Close(ctx)
			for cursor.Next(ctx) {
				var message UserMessage
				if err = cursor.Decode(&message); err != nil {
					log.Fatal(err)
				}
				messages = append(messages, &message)
			}
			currentRoom.Messages = messages
			js, err := json.Marshal(&currentRoom)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			print("yeeee?")
			w.Write(js)
		}

	})

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		collection := client.Database("chat").Collection("rooms")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		keys, ok := r.URL.Query()["room"]
		fmt.Println(keys)
		fmt.Println(ok)
		if !ok || len(keys) == 0 {
			print("here")
			w.Write([]byte("You are not part of a room"))
			return
		}

		currentRoom := keys[0]
		_, exists := roomMap[currentRoom]
		roomObjID, _ := primitive.ObjectIDFromHex(currentRoom)
		filter := bson.M{"_id": roomObjID}
		err := collection.FindOne(ctx, filter).Decode(&rooms)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if !exists {
			fmt.Println(roomMap)
			fmt.Println(currentRoom)
			fmt.Println(roomMap[currentRoom])
			roomMap[currentRoom] = newHub()
			go roomMap[currentRoom].run()
		}
		serveWs(roomMap[currentRoom], w, r)
	})

	http.HandleFunc("/user-unavailable/", func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		collection := client.Database("chat").Collection("users")
		keys, ok := r.URL.Query()["user"]
		if !ok || len(keys) == 0 {
			print("ehhh")
		}
		result, err := collection.UpdateOne(
			ctx,
			bson.M{"user": keys[0]},
			bson.D{
				{"$set", bson.D{{"available", false}}},
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	})

	httperr := http.ListenAndServe(*addr, nil)
	if httperr != nil {
		log.Fatal("ListenAndServe: ", httperr)
	}
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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
