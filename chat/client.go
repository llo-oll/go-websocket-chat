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
//receiveChan is used by the rest of the application to send messages to the client. These are forwarded over the
//WebSocket. It is unique to this client.
//
//sendChan is used to send messages which have been received over the WebSocket to the rest of the application.
//It is shared by all clients.
//
//deathChan is used to inform the hub that this client has died by sending id over the channel.
//deathChan is shared by all clients
//
//Clients should be created using the newClient function.
//This runs two goroutines: msgReceiveRoutine and msgSendRoutine,
//which are responsible for dealing with communication over receiveChan and sendChan.
type client struct {
	id          int
	conn        *websocket.Conn
	receiveChan <-chan string
	sendChan    chan<- string
}

//This channel provides unique ids for clients (0,1,...)
var idChan = func() <-chan int {
	ch := make(chan int)
	id := 0
	go func() {
		for {
			ch <- id
			id++
		}
	}()
	return ch
}()

//newClient creates a client.
//
//conn should be a WebSocket connection to the html client.
//
//Returns the id of the new client and channels for sending and receiving messages.
func newClient(conn *websocket.Conn) (clientId int, toClient chan<- string, fromClient <-chan string) {
	receiveChan := make(chan string)
	sendChan := make(chan string)
	id := <-idChan
	newClient := client{id, conn, receiveChan, sendChan}
	newClient.run()
	return id, receiveChan, sendChan
}

func (thisClient *client) run() {
	go thisClient.msgReceiveRoutine()
	go thisClient.msgSendRoutine()
}

//msgReceiveRoutine waits for incoming messages from the hub and then sends them over the WebSocket.
func (thisClient *client) msgReceiveRoutine() {
	thisClient.log("Listening for incoming messages")
	for msg := range thisClient.receiveChan {
		thisClient.log("Received a message")
		err := thisClient.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			thisClient.log(err)
			thisClient.log("Failed to write to WebSocket")
			break
		}
	}
	thisClient.log("Stopped listening for messages")
	thisClient.Close()
}

//msgSendRoutine waits for messages coming over the WebSocket and then sends them over the send channel,
// to be picked up by the hub.
func (thisClient *client) msgSendRoutine() {
	thisClient.log("Waiting to send messages")
	for {
		msgType, bytes, err := thisClient.conn.ReadMessage()
		if err != nil {
			thisClient.log(err)
			break
		} else if msgType != websocket.TextMessage {
			thisClient.log("Received non-text message through socket")
		} else {
			thisClient.log("Sending message")
			thisClient.sendChan <- string(bytes)
		}
	}
	thisClient.log("Stopped sending messages")
	thisClient.log("Closing send channel")
	close(thisClient.sendChan)
	thisClient.Close()
}

func (thisClient *client) Close() {
	thisClient.log("Closing WebSocket")
	thisClient.conn.Close()
}

func (thisClient *client) log(entry interface{}) {
	log.Printf("Client %d:\t%s", thisClient.id, entry)
}
