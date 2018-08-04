package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type client struct {
	chanIn  <-chan string
	chanOut chan<- string
	conn    websocket.Conn
}

func activateClient(chanOut chan<- string, conn websocket.Conn) chan string {
	//TODO buffer size is a guess. Change that.
	chanIn := make(chan string, 100)
	newClient := client{chanIn, chanOut, conn}
	newClient.run()
	return chanIn
}

func (thisClient *client) run() {
	go thisClient.listenIn()
	go thisClient.listenOut()
}

//TODO rename all these "in" and "out"s they are very confusing: chanIn chanOut listenIn/Out
func (thisClient *client) listenIn() {
	for msg := range thisClient.chanIn {
		log.Println("Client is listening for incoming messages")
		err := thisClient.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println(err)
			//TODO destroy the client rather than simply return
			return
		}
	}
}

func (thisClient *client) listenOut() {
	for {
		log.Println("Client is listening to its own messages")
		msgType, bytes, err := thisClient.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			//TODO destroy the client rather than simply return
			return
		}
		if msgType != websocket.TextMessage {
			log.Println("Recieved non-text message through socket")
			//TODO do i want to destroy the client or just carry on
			return
		} else {
			thisClient.chanOut <- string(bytes)
		}
	}
}
