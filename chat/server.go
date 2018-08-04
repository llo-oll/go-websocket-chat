package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var chatHub = newHub()

/**
Sets up a server providing a web based chat client on "/"
*/
func main() {
	http.HandleFunc("/", servePage)
	http.HandleFunc("/ws", connectClient)
	http.ListenAndServe(":5000", nil)
}

func connectClient(writer http.ResponseWriter, request *http.Request) {
	//upgrade connection
	var upgrader = websocket.Upgrader{
		HandshakeTimeout:  0,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		Subprotocols:      nil,
		Error:             nil,
		CheckOrigin:       nil,
		EnableCompression: false,
	}

	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	chatHub.addConnection(conn)

}
func servePage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "chat/client.html")
}
