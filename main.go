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
	id_message  int
	author_name string
	text        string
}

func SocketFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error", err)
	}
	// defer ws.Close()
	clients = append(clients, ws)
	// db, err := sql.Open("mysql", "chat:123456@tcp(192.168.43.245:3306)/chat")
	// if err != nil {
	// 	fmt.Println("Database error", err)
	// }
	// defer db.Close()
	// msgs, err := db.Query("SELECT * FROM Messages")
	// if err != nil {
	// 	panic(err)
	// }
	// defer msgs.Close()
	// messages := []message{}

	// for msgs.Next() {
	// 	msg := message{}
	// 	err := msgs.Scan(&msg.id_message, &msg.author_name, &msg.text)
	// 	if err != nil {
	// 		fmt.Println("Error reading from db!", err)
	// 		continue
	// 	}
	// 	messages = append(messages, msg)
	// }
	// for _, message := range messages {
	// 	fmt.Printf("id: %d, author: %s, text: %s\n", message.id_message, message.author_name, message.text)
	// }
	fmt.Println("Ws:", ws)
	go ReadSocket(ws)

}

func ReadSocket(ws *websocket.Conn) {
	fmt.Println("State of clients:")
	fmt.Println(len(clients))
	for {
		type message struct {
			Text   string `json:"text"`
			Author string `json:"author"`
		}
		var msg message
		err := websocket.ReadJSON(ws, &msg)
		if err != nil {
			// _, _, err := ws.ReadMessage()
			// if err != nil {
			fmt.Println("Disconnect")
			CloseAndRemove(ws)
			// ws.Close()

			// }
			fmt.Println("Client was disconnected")
			break
		}
		fmt.Println("Message: ", msg.Text)
		fmt.Println("Author: ", msg.Author)
		db, err := sql.Open("mysql", "chat:123456@tcp(192.168.43.245:3306)/chat")
		if err != nil {
			fmt.Println("Database error", err)
		}
		defer db.Close()
		insert, err := db.Query("INSERT INTO Messages (author_name, text) VALUES (?, ?)", msg.Author, msg.Text)
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
