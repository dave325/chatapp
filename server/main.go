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
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create the JWT key used to create the signature
var jwtKey = []byte("my-secret-password") // Used for demonstration and github purposes
var addr = flag.String("addr", ":3001", "http service address")
var ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Second)
var client, mongoErr = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:root@mongodb:27017/"))

var roomMap map[string]*Hub = make(map[string]*Hub)

var result struct {
	User string
}

type documents struct {
	CurrentUser string   `json:"currentUser"`
	Users       []string `json:"users"`
}

type Rooms struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Users    []string           `json:"users" bson:"users"`
	Messages []*UserMessage     `json:"messages,omitempty"  bson:"messages,omitempty"`
}

type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	Token    string             `json:"token" bson:"token"`
}

type Response struct {
	Error string `json:"error"`
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
	hub := newHub(true)
	go hub.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		collection := client.Database("chat").Collection("users")

		keys, ok := r.URL.Query()["user"]
		currentUserID := ""

		if !ok {
			log.Println("No user id")
			w.WriteHeader(http.StatusBadRequest)
			response := Response{Error: "User id not provided"}
			json.NewEncoder(w).Encode(response)
			return
		}
		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]

		filter := bson.M{"username": key}
		err := collection.FindOne(ctx, filter).Decode(&result)

		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			response := Response{Error: "User not found"}
			json.NewEncoder(w).Encode(response)
			return
		} else {
			_, err := collection.UpdateOne(
				ctx,
				bson.M{"username": key},
				bson.D{
					{"$set", bson.M{"available": true}},
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			currentUserID = key
		}

		// Fetch avialable users in db
		cur, err := collection.Find(ctx, bson.M{"available": true})
		defer cur.Close(ctx)
		print(cur.RemainingBatchLength())
		totalUser := make([]string, cur.RemainingBatchLength())
		if err != nil {
			log.Fatal(err)
		}
		currentIdx := 0
		for cur.Next(ctx) {
			var currentUser User
			err := cur.Decode(&currentUser)
			if err != nil {
				log.Fatal(err)
			}
			if len(currentUser.Username) > 0 {
				totalUser[currentIdx] = currentUser.Username
			}
			currentIdx++
		}

		totalStruct := documents{CurrentUser: currentUserID, Users: totalUser}
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
				http.Error(w, err.Error(), 500)
				return
			}
		}
		bsonArr := bson.A{}
		for _, user := range currentRoom.Users {
			bsonArr = append(bsonArr, user)
		}

		// Search if a db already has these users in a room
		filter := bson.D{{"users", bson.M{"$all": bsonArr, "$size": len(currentRoom.Users)}}}
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
			collection = client.Database("chat").Collection("messages")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			filterRoom := currentRoom.ID.Hex()
			// log.Println(filterRoom)
			updatedFilter := bson.M{"room": filterRoom}
			cursor, cursorErr := collection.Find(ctx, updatedFilter)
			if cursorErr != nil {
				log.Println(cursorErr)
			}
			defer cursor.Close(ctx)
			currentIdx := 0
			// Fetch messages for the current room
			currentMessages := make([]*UserMessage, cursor.RemainingBatchLength())
			for cursor.Next(ctx) {
				var message UserMessage
				if err = cursor.Decode(&message); err != nil {
					log.Fatal(err)
					continue
				}
				currentMessages[currentIdx] = &message
				currentIdx++
			}
			currentRoom.Messages = currentMessages
			js, err := json.Marshal(&currentRoom)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Return all messages and users associated with this room
			w.Write(js)
		}

	})

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		collection := client.Database("chat").Collection("rooms")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		keys, ok := r.URL.Query()["room"]
		if !ok || len(keys) == 0 {
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
			roomMap[currentRoom] = newHub(false)
			go roomMap[currentRoom].run()
		}

		serveWs(roomMap[currentRoom], w, r, currentRoom)
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
		_, err := collection.UpdateOne(
			ctx,
			bson.M{"username": keys[0]},
			bson.D{
				{"$set", bson.D{{"available", false}}},
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	})

	http.HandleFunc("/sign-up/", func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer r.Body.Close()
		var currentUser User
		err = json.Unmarshal(b, &currentUser)
		currentUser.Token = createJWTToken(currentUser)
		collection := client.Database("chat").Collection("users")
		// Salt and hash the password using the bcrypt algorithm
		// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(currentUser.Password), 8)
		filter := bson.M{"username": currentUser.Username}
		count, _ := collection.CountDocuments(ctx, filter)
		w.Header().Set("Content-Type", "application/json")
		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			response := Response{Error: "User Already Exists"}
			json.NewEncoder(w).Encode(response)
			return
		}
		_, insertErr := collection.InsertOne(context.Background(), bson.M{"username": currentUser.Username, "password": string(hashedPassword), "token": currentUser.Token})
		if insertErr != nil {
			log.Println("Insert error")
			log.Println(insertErr)
			return
		}
		w.WriteHeader(http.StatusOK)
		token := User{Token: currentUser.Token, Username: currentUser.Username}
		json.NewEncoder(w).Encode(token)
	})

	http.HandleFunc("/login/", func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer r.Body.Close()
		var currentUser User
		err = json.Unmarshal(b, &currentUser)
		collection := client.Database("chat").Collection("users")
		var userFromDb User
		err = collection.FindOne(ctx, bson.M{"username": currentUser.Username}).Decode(&userFromDb)

		log.Println(currentUser.Username)
		if err = bcrypt.CompareHashAndPassword([]byte(userFromDb.Password), []byte(currentUser.Password)); err != nil {
			// If the two passwords don't match, return a 401 status
			w.WriteHeader(http.StatusUnauthorized)
			response := Response{Error: "Username and password do not match"}
			json.NewEncoder(w).Encode(response)
			return
		}
		token := User{Token: userFromDb.Token, Username: userFromDb.Username}
		json.NewEncoder(w).Encode(token)
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

func createJWTToken(user User) string {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"password": user.Password,
		"username": user.Username,
		"nbf":      time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(jwtKey)

	fmt.Println(tokenString, err)
	return tokenString
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
