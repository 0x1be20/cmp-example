package core

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options
var addr = flag.String("addr", "localhost:8080", "http service address")
var channels = map[string]*websocket.Conn{}
var chLocker = sync.Mutex{}

var locker = sync.Mutex{}

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	id := uuid.New().String()
	chLocker.Lock()
	channels[id] = c
	chLocker.Unlock()
	log.Printf("connect %s", id)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("close channel:", id)
			// 断开链接
			chLocker.Lock()
			delete(channels, id)
			chLocker.Unlock()
			break
		}
		protocolMsg := &protocol.Message{}
		protocolMsg.UnmarshalBinary(message)
		log.Printf("recv msg,from %s,to %s round %v", protocolMsg.From, protocolMsg.To, protocolMsg.RoundNumber)
		locker.Lock()

		chLocker.Lock()
		for _, conn := range channels {
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
		chLocker.Unlock()
		locker.Unlock()

	}
}

func InitServer() {
	http.HandleFunc("/ws", connect)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
