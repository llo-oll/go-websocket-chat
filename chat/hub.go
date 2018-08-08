package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type hub struct {
	msgChan       chan string
	clientChanMap map[int]chan<- string
}

//TODO maybe refactor hub and client into a separate package and turn this into a public function. This will force the use of this constructor due to the "hub" type being private.

func newHub() *hub {
	newHub := &hub{msgChan: make(chan string), clientChanMap: make(map[int]chan<- string)}
	newHub.log("Constructing new hub")
	go newHub.listen()
	return newHub
}

func (thisHub *hub) addConnection(conn *websocket.Conn) {
	thisHub.log("Received new connection")
	thisHub.log("Creating new client")
	clientId, toClient, fromClient := newClient(conn)
	thisHub.clientChanMap[clientId] = toClient
	go func() {
		for msg := range fromClient {
			thisHub.msgChan <- fmt.Sprintf("Client %d: %s", clientId, msg)
		}
		thisHub.log(fmt.Sprint("Removing Client", clientId))
		close(thisHub.clientChanMap[clientId])
		delete(thisHub.clientChanMap, clientId)
	}()
}

func (thisHub *hub) listen() {
	thisHub.log("Listening for client messages")
	for msg := range thisHub.msgChan {
		for id, ch := range thisHub.clientChanMap {
			thisHub.log(fmt.Sprintf("Sending message to client %d", id))
			ch <- msg
		}
	}
}

func (thisHub *hub) log(entry interface{}) {
	log.Printf("Hub:\t\t%s", entry)
}
