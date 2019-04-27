package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients []*websocket.Conn

type message struct {
	Id_message  int    `json:"id_message"`
	Author_name string `json:"author_name"`
	Text        string `json:"text"`
	Date        string `json:"date"`
	Is_deleted  bool   `json:"is_deleted"`
}

func SocketFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error", err)
	}
	// defer ws.Close()
	clients = append(clients, ws)
	// fmt.Println("Ws:", ws)
	UpdateChat(ws) // Update chat for socket
	go ReadSocket(ws)

}

func ReadSocket(ws *websocket.Conn) {
	fmt.Println("State of clients:")
	fmt.Println(len(clients))
	for {
		var msg message
		err := websocket.ReadJSON(ws, &msg)
		if err != nil {
			fmt.Println("Client was disconnected")
			CloseAndRemove(ws)
			break
		}
		fmt.Println("Message: ", msg.Text)
		fmt.Println("Author: ", msg.Author_name)
		db, err := sql.Open("mysql", "chat:123456@tcp(192.168.43.245:3306)/chat")
		if err != nil {
			fmt.Println("Database error", err)
		}
		defer db.Close()
		insert, err := db.Query("INSERT INTO Messages (author_name, text) VALUES (?, ?)", msg.Author_name, msg.Text)
		if err != nil {
			fmt.Println("Database insert error", err)
			break
		}
		insert.Close() // Correct closing
		for _, client := range clients {
			if err := websocket.WriteJSON(client, msg); err != nil {
				fmt.Println("Disconnect")
				break
			}
			fmt.Println("Done.")
		}

	}
}

func UpdateChat(ws *websocket.Conn) {
	db, err := sql.Open("mysql", "chat:123456@tcp(192.168.43.245:3306)/chat")
	if err != nil {
		fmt.Println("Database error", err)
	}
	defer db.Close()
	msgs, err := db.Query("SELECT * FROM Messages ORDER BY date DESC")
	if err != nil {
		panic(err)
	}
	defer msgs.Close()
	messages := []message{}

	for msgs.Next() {
		msg := message{}
		err := msgs.Scan(&msg.Id_message, &msg.Author_name, &msg.Text, &msg.Date, &msg.Is_deleted)
		if err != nil {
			fmt.Println("Error reading from database!", err)
			continue
		}
		messages = append(messages, msg)

	}
	for _, msg := range messages {
		fmt.Printf("id: %d, author: %s, text: %s, date: %s\n", msg.Id_message, msg.Author_name, msg.Text, msg.Date)
	}

	if err := ws.WriteJSON(messages); err != nil {
		fmt.Println("Error JSON array encoding", err)
	}

}
func CloseAndRemove(ws *websocket.Conn) {
	index := -1
	defer ws.Close()
	for i, socket := range clients {
		if socket == ws {
			index = i
			break
		}
	}
	if index == -1 {
		fmt.Println("Socket not found!")
		return
	}
	clients = append(clients[:index], clients[index+1:]...)
}
func MainFunc(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]
	// tmpl, _ := template.ParseFiles("./templates/index.html")
	// tmpl.Execute(w, "")
	http.ServeFile(w, r, "index.html")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", MainFunc)
	router.HandleFunc("/socket", SocketFunc)
	http.Handle("/", router)
	// fs := http.FileServer(http.Dir("static"))
	// http.Handle("/static/", http.StripPrefix("/static", fs))
	fmt.Println("Server is listening")

	http.ListenAndServe(":8080", nil)
}
