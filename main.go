package main

import (
	"encoding/json"
	"fmt"
	"os/user"

	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)


type User struct{
	connection *websocket.Conn
	ui UserInfo
}

type UserInfo struct{
	Username string `json:"Username"`
	Balance string `json:"Balance"`
}

var users []User


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func deleteByConn(conn *websocket.Conn){
	for i, u := range users{
		if u.connection == conn{
			users[i] = users[len(users)-1]
			users = users[:len(users)-1]
		}
	}
}

func updateBalance(username string, newBalance string){
	for _, u := range users{
		if u.ui.Username == username{
			u.ui.Balance = newBalance
		}
	}
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			// user discconect  / other error
			deleteByConn(conn)
			return
		}
		var result map[string]any
		json.Unmarshal(p, &result)

		switch result["ID"]{
			case "new": {
				uname, _ := result["Username"].(string)
				ubal, _ := result["Balance"].(string)
				// log.Printf("New user\n%v\n%v wls\n", uname, ubal)
				ui := UserInfo{Username: uname, Balance: ubal}
				users = append(users, User{connection: conn, ui: ui})
			}
			case "updateBal": {
				uname, _ := result["Username"].(string)
				ubal, _ := result["Balance"].(string)
				for i := range len(users){
					if users[i].ui.Username == uname{
						users[i].ui.Balance = ubal
					}
				}
			}
		}
	}
}

func tip(w http.ResponseWriter, r *http.Request){
	from := r.PathValue("from")
	to := r.PathValue("to")
	amt := r.PathValue("amount")
	queue := r.PathValue("queue")
	time := r.PathValue("time")
	
	for _, u := range user{
		if u.ui.Username == from{
			data := map[string]string{
				"ID": "tip",
				"to": to,
				"amount": amt,
				"queue", queue,
				"time", time,
			}
			jsonData, _ := json.Marshal(data)
			u.connection.WriteMessage(1, []byte(jsonData))
		}
	}
	// fmt.Fprintf(w, "Sent tip request to client %v", from)
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// if connect successfull
	
	err = ws.WriteMessage(1, []byte("Connected to server websocket"))
	
	if err != nil {
		log.Println(err)
	}
	reader(ws)
}

func setupAPI(){
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsEndpoint)
	// GET ALL USER DATA
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request){
		var temp []UserInfo
		for _, u := range users{
			temp = append(temp, u.ui)
		}
		j, _ := json.Marshal(temp)
		fmt.Fprintf(w, string(j))
	})

	// TIP
	mux.HandleFunc("POST /tip/{from}/{to}/{amount}/{queue}/{time}", tip)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func main() {
	setupAPI()
	//log.Fatal(http.ListenAndServe(":8080", nil))
}