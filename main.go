package main

import (
	"encoding/json"
	"html/template"
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
	Username string
	Balance string
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
				ui := UserInfo{Username: uname, Balance: ubal}
				users = append(users, User{connection: conn, ui: ui})
			}
		}

		log.Printf("%v\n", result["ID"])
 
	}
}
 
func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	var ut []UserInfo

	for _, u :=  range users{
		ut = append(ut, UserInfo{Username: u.ui.Username, Balance: u.ui.Balance})
	}
	
	tUsers := map[string][]UserInfo{
		"Users": ut,
	}
	tmpl.Execute(w, tUsers)
}

func tip(w http.ResponseWriter, r *http.Request){
	log.Println("Tip!")
	time.Sleep(1 * time.Second)
	from := r.PostFormValue("From")
	to := r.PostFormValue("To")
	amt := r.PostFormValue("Amount")

	for _, u := range users{
		if u.ui.Username == from{
			data := map[string]string{
				"ID": "tip",
				"to": to,
				"amt": amt,
			}
			jsonData, _ := json.Marshal(data)
			u.connection.WriteMessage(1, []byte(jsonData))
		}
	}
	w.Write([]byte("Sent tip"))
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// if connect successfull
	
	err = ws.WriteMessage(1, []byte("con"))
	
	if err != nil {
		log.Println(err)
	}
	reader(ws)
}
 
func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/tip", tip)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {	
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}