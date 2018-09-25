package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

//This is the hub which connects all the chat clients together, and coordinates the passing around of messages.
var chatHub = newHub()

//main sets up a server providing a web based chat webSocketAdapter on "/"
func main() {
	http.HandleFunc("/", servePage)
	http.HandleFunc("/ws", connectClient)
	http.ListenAndServe(":5000", nil)
}

//connectClient is an http request handler which upgrades the connection to a websocket and adds the connection to
//the chat hub.
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

//servePage is an http request handler which serves the webchat web page to a webSocketAdapter.
func servePage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "chat/webSocketAdapter.html")
}
