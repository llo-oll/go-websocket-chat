package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout:  0,
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	Subprotocols:      nil,
	Error:             nil,
	CheckOrigin:       nil,
	EnableCompression: false,
}

func connectWebSocket(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//Read from stdin and send over the connection
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text())); err != nil {
				log.Println(err)
				return
			}
		}
	}()

	//Print any text messages recieved over the connection
	go func() {
		for {
			msgType, bytes, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if msgType != websocket.TextMessage {
				log.Println("Recieved non-text message through socket!")
			} else {
				fmt.Println(string(bytes))
			}
		}
	}()
}

func servePage(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.Error(writer, "Not Found", http.StatusNotFound)
		return
	}
	http.ServeFile(writer, request, "simples/simple_websocket_client.html")
}

func main() {
	http.HandleFunc("/", servePage)
	http.HandleFunc("/ws", connectWebSocket)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
