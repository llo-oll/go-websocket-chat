package main

import (
	"github.com/gorilla/websocket"
	"log"
)

//type client is the applications representation of a chat client.
//It is connected to the clients web page by a WebSocket and provides the interface between the socket and the
//application.
//
//A client communicates to the rest of the application over two channels.
//
//sendChan is used to send messages which have been received over the WebSocket to the rest of the application.
//It is shared by all clients.
//
//receiveChan is used by the rest of the application to send messages to the client. These are forwarded over the
//WebSocket. It is unique to this client.
//
//Clients should be created using the activateClient function.
//This runs two goroutines: msgReceiveRoutine and msgSendRoutine,
//which are responsible for dealing with communication over receiveChan and sendChan.
type client struct {
	receiveChan chan string
	sendChan    chan string
	conn        *websocket.Conn
	id          int
}

//activateClient creates a client.
func activateClient(sendChan chan string, conn *websocket.Conn, id int) chan string {
	//TODO buffer size is a guess. Change that.
	receiveChan := make(chan string, 100)
	newClient := client{receiveChan, sendChan, conn, id}
	newClient.run()
	return receiveChan
}

func (thisClient *client) run() {
	go thisClient.msgReceiveRoutine()
	go thisClient.msgSendRoutine()
}

//msgReceiveRoutine waits for incoming messages and then sends them over the WebSocket.
func (thisClient *client) msgReceiveRoutine() {
	log.Println("Client", thisClient.id, "\b: listening for incoming messages")
	for msg := range thisClient.receiveChan {
		log.Println("Client", thisClient.id, "\b: received a message")
		err := thisClient.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println(err)
			log.Println("Client", thisClient.id, "\b: Failed to write to WebSocket")
			log.Println("Client", thisClient.id, "\b: messageReceiveRoutine is closing down")
			thisClient.conn.Close()
			break
		}
	}
	log.Println("Client", thisClient.id, "\b: is exiting msgReceiveRoutine")
}

//msgSendRoutine waits for messages coming over the WebSocket and then sends them over the send channel,
// to be picked up by the hub.
func (thisClient *client) msgSendRoutine() {
	log.Println("Client", thisClient.id, "\b: is waiting to send messages")
	for {
		msgType, bytes, err := thisClient.conn.ReadMessage()
		log.Println("Client", thisClient.id, "\b: sent a message")
		if err != nil {
			log.Println(err)
			log.Println("Client", thisClient.id, "\b: is closing down")
			close(thisClient.receiveChan)
			break
		} else if msgType != websocket.TextMessage {
			log.Println("Client", thisClient.id, "\b: received non-text message through socket")
		} else {
			thisClient.sendChan <- string(bytes)
		}
	}
	log.Println("Client", thisClient.id, "\b: is exiting msgSendRoutine")
}
