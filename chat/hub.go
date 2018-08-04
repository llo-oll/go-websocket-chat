package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type hub struct {
	msgChan         chan string
	clientChanArray []chan string
}

/*TODO maybe refactor hub and client into a separate package and turn this into
a public function. This will force the use of this constructor due to the "hub"
type being private.
*/
func newHub() *hub {
	log.Println("Constructing new hub")
	newHub := &hub{msgChan: make(chan string), clientChanArray: make([]chan string, 0)}
	go newHub.listen()
	return newHub
}

func (thisHub *hub) addConnection(conn *websocket.Conn) {
	log.Println("Adding connection to hub")
	newChan := activateClient(thisHub.msgChan, conn)
	thisHub.clientChanArray = append(thisHub.clientChanArray, newChan)
}

func (thisHub *hub) listen() {
	log.Println("Hub is listening for client messages")
	for msg := range thisHub.msgChan {
		for _, ch := range thisHub.clientChanArray {
			ch <- msg
		}
	}
}
