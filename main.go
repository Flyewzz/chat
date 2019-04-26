package main

import (
	"fmt"
	"log"
	"net/http"

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

func SocketFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error", err)
	}
	// defer ws.Close()
	clients = append(clients, ws)
	go ReadSocket(ws)
}

func ReadSocket(ws *websocket.Conn) {
	fmt.Println("TEST")
	for {
		// _, message, err := ws.ReadMessage()
		type message struct {
			Text   string `json:"text"`
			Author string `json:"author"`
		}
		var msg message
		err := websocket.ReadJSON(ws, &msg)
		if err != nil {
			fmt.Println("Disconnect")
			// ws.Close()
			break
		}
		fmt.Println("Message: ", msg.Text)
		fmt.Println("Author: ", msg.Author)
		for _, client := range clients {
			if err := websocket.WriteJSON(client, msg); err != nil {
				fmt.Println("Disconnect")
				break
			}
			fmt.Println("Done.")
		}

	}
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
