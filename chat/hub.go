package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"regexp"
	"strings"
	"sync"
)

type hub struct {
	msgChan       chan string
	clientChanMap map[int]chan<- string
	nameMap       usernameMap
}

//TODO maybe refactor hub and client into a separate package and turn this into a public function. This will force the use of this constructor due to the "hub" type being private.

func newHub() *hub {
	newHub := &hub{msgChan: make(chan string), clientChanMap: make(map[int]chan<- string), nameMap: newUsernameMap()}
	newHub.log("Constructing new hub")
	go newHub.listenAndSend()
	return newHub
}

func (thisHub *hub) addConnection(conn *websocket.Conn) {
	thisHub.log("Received new connection")
	thisHub.log("Creating new client")
	clientId, toClient, fromClient := newClient(conn)
	thisHub.clientChanMap[clientId] = toClient
	thisHub.nameMap.changeName(clientId, fmt.Sprintf("User%d", clientId))
	go func() {
		for msg := range fromClient {
			if string(msg[0]) == "/" {
				thisHub.log(fmt.Sprintf("Client %d has run a command", clientId))
				thisHub.run_command(clientId, string(msg[1:]))
				continue
			}
			thisHub.msgChan <- fmt.Sprintf("%s: %s", thisHub.nameMap.getName(clientId), msg)
		}
		thisHub.log(fmt.Sprint("Removing Client", clientId))
		close(thisHub.clientChanMap[clientId])
		delete(thisHub.clientChanMap, clientId)
		//removes the client from nameMap
		thisHub.nameMap.changeName(clientId, "")
		thisHub.log(fmt.Sprintf("%s", thisHub.nameMap))
	}()
}

//listenAndSend listens for incoming messages and sends them out to all clients.
func (thisHub *hub) listenAndSend() {
	thisHub.log("Listening for client messages")
	for msg := range thisHub.msgChan {
		for id, ch := range thisHub.clientChanMap {
			thisHub.log(fmt.Sprintf("Sending message to client %d", id))
			ch <- msg
		}
	}
}

func (thisHub *hub) run_command(clientId int, s string) {
	if match, _ := regexp.MatchString("^name", s); match {
		newName := string(s[4:])
		thisHub.nameMap.changeName(clientId, newName)
	}
}

func (thisHub *hub) log(entry interface{}) {
	log.Printf("Hub:\t\t%s", entry)
}

//TODO remove closed connections from usernameMap
type usernameMap struct {
	mutex         sync.Mutex
	id2NameMap    map[int]string
	nameSet       map[string]bool
	maxNameLength int
}

func newUsernameMap() usernameMap {
	var nameMap usernameMap
	nameMap.id2NameMap = make(map[int]string)
	nameMap.nameSet = make(map[string]bool)
	nameMap.maxNameLength = 20
	return nameMap
}

//changeName changes the username of the client if the new name isn't already in use.
//returns true if successful, false otherwise.
//Changing a name to the empty string removes the entry from the map.
//
func (nameMap *usernameMap) changeName(clientId int, name string) bool {
	nameMap.mutex.Lock()
	defer nameMap.mutex.Unlock()

	name = strings.TrimSpace(name)
	nameEnd := strings.Index(name, " ")
	if nameEnd > 0 {
		name = name[:nameEnd]
	}

	if len(name) > nameMap.maxNameLength {
		name = string(name[:nameMap.maxNameLength])
	}

	if nameMap.nameSet[name] {
		return false
	}

	oldName := nameMap.id2NameMap[clientId]
	delete(nameMap.nameSet, oldName)

	if name == "" {
		delete(nameMap.id2NameMap, clientId)
	} else {
		nameMap.id2NameMap[clientId] = name
		nameMap.nameSet[name] = true
	}

	return true
}

//getName is a concurrency safe way of getting a username.
func (nameMap *usernameMap) getName(clientId int) string {
	nameMap.mutex.Lock()
	defer nameMap.mutex.Unlock()

	return nameMap.id2NameMap[clientId]
}
