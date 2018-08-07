package main

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
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
//Clients should be created using the activateClient function.
//This runs two goroutines: msgReceiveRoutine and msgSendRoutine,
//which are responsible for dealing with communication over receiveChan and sendChan.
type client struct {
	id          int
	conn        *websocket.Conn
	receiveChan <-chan string
	sendChan    chan<- string
	deathChan   chan<- int
}

var clientId struct {
	nextId  int
	idMutex sync.Mutex
}

//activateClient creates a client.
//
//conn should be a connection to the html client.
//
//sendChan should be the channel on which to send chat messages to the hub
//
//deathChan allows the client to inform the hub that it has died and should be forgotten about.
//
//Returns the id of the new client and a channel over which chat messages can be sent to the client.
func activateClient(conn *websocket.Conn, sendChan chan string, deathChan chan int) (clientId int, clientChannel chan<- string) {
	//TODO buffer size is a guess. Change that.
	receiveChan := make(chan string, 100)
	id := nextId()
	newClient := client{id, conn, receiveChan, sendChan, deathChan}
	newClient.run()
	return id, receiveChan
}

func nextId() int {
	clientId.idMutex.Lock()
	id := clientId.nextId
	clientId.nextId++
	clientId.idMutex.Unlock()
	return id
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
			thisClient.log("Closing WebSocket")
			thisClient.conn.Close()
			break
		}
	}
	thisClient.log("Exiting msgReceiveRoutine")
}

//msgSendRoutine waits for messages coming over the WebSocket and then sends them over the send channel,
// to be picked up by the hub.
func (thisClient *client) msgSendRoutine() {
	thisClient.log("Waiting to send messages")
	for {
		msgType, bytes, err := thisClient.conn.ReadMessage()
		if err != nil {
			thisClient.log(err)
			thisClient.log("Closing WebSocket")
			thisClient.conn.Close()
			break
		} else if msgType != websocket.TextMessage {
			thisClient.log("Received non-text message through socket")
		} else {
			thisClient.log("Sending message")
			thisClient.sendChan <- string(bytes)
		}
	}
	thisClient.log("Exiting msgSendRoutine")
}

func (thisClient *client) log(entry interface{}) {
	log.Printf("Client %d:\t%s", thisClient.id, entry)
}
