package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type hub struct {
	msgChan       chan string
	clientChanMap map[int]chan string
	nextClientId  int
}

/*TODO maybe refactor hub and client into a separate package and turn this into
a public function. This will force the use of this constructor due to the "hub"
type being private.
*/

func newHub() *hub {
	log.Println("Constructing new hub")
	newHub := &hub{msgChan: make(chan string), clientChanMap: make(map[int]chan string)}
	go newHub.listen()
	return newHub
}

func (thisHub *hub) addConnection(conn *websocket.Conn) {
	log.Println("Adding connection to hub")
	newChan := activateClient(thisHub.msgChan, conn, thisHub.nextClientId)
	thisHub.clientChanMap[thisHub.nextClientId] = newChan
	thisHub.nextClientId++
}

func (thisHub *hub) listen() {
	log.Println("Hub is listening for client messages")
	for msg := range thisHub.msgChan {
		for id, ch := range thisHub.clientChanMap {
			select {
			case ch <- msg:
			default:
				log.Println("Removing client ", id, " from hub")
				delete(thisHub.clientChanMap, id)
			}
		}
	}
}
