package main

import "github.com/gorilla/websocket"

type hub struct {
	msgChan         chan string
	clientChanArray []chan string
}

/*TODO maybe refactor hub and client into a separate package and turn this into
a public function. This will force the use of this constructor due to the "hub"
type being private.
*/
func newHub() *hub {
	newHub := &hub{msgChan: make(chan string), clientChanArray: make([]chan string, 10)}
	return newHub
}

func (thisHub *hub) addConnection(conn websocket.Conn) {
	thisHub.clientChanArray = append(thisHub.clientChanArray, activateClient(thisHub.msgChan, conn))
}
