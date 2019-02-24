package main

import (
	"github.com/nats-io/go-nats"
	"gopkg.in/olahol/melody.v1"
	"log"
	"net/http"
)

type IWebsocket interface {
	HandleWSRequest(writer http.ResponseWriter, request *http.Request) error
}

type Websocket struct {
	melody         *melody.Melody
	currentClients map[*melody.Session]*nats.Subscription
	nc             *nats.Conn
}

func NewWebsocket(nc *nats.Conn) *Websocket {
	websocket := Websocket{
		nc:             nc,
		currentClients: make(map[*melody.Session]*nats.Subscription),
		melody:         melody.New(),
	}
	websocket.melody.HandleMessage(websocket.HandleMessage)
	return &websocket
}

func (ws *Websocket) HandleWSRequest(writer http.ResponseWriter, request *http.Request) error {
	return ws.melody.HandleRequest(writer, request)
}

func (ws *Websocket) HandleMessage(s *melody.Session, msg []byte) {
	log.Println("receive ws msg", string(msg))
	msgWs := WebSocketMessage{}
	err := msgWs.UnmarshalJSON(msg)
	if err != nil {
		return
	}
	switch msgWs.Action {
	case ActionSubscribe:
		ws.handleSubscribeMessage(s, msgWs)
	case ActionUnsubscribe:
		ws.handleUnsubscribeMessage(s)
	}
}

func (ws *Websocket) handleUnsubscribeMessage(s *melody.Session) {
	log.Println("is unsubscribe msg")
	subs, ok := ws.currentClients[s]
	if !ok {
		log.Println("client not subscribe")
		return
	}
	err := subs.Unsubscribe()
	if err != nil {
		panic(err)
	}
	delete(ws.currentClients, s)
	log.Println("unsubscribe client")
}

func (ws *Websocket) handleSubscribeMessage(s *melody.Session, msgWs WebSocketMessage) {
	log.Println("is subscribe msg")
	_, ok := ws.currentClients[s]
	if ok {
		return
	}
	subs, err := ws.nc.Subscribe(msgWs.Message, func(msg *nats.Msg) {
		log.Println("subscribe receive a msg")
		err := s.Write(msg.Data)
		if err != nil {
			panic(err)
		}
		log.Println("write msg on ws")
	})
	if err != nil {
		panic(err)
	}
	ws.currentClients[s] = subs
	wsMsg := WebSocketMessage{Action: ActionSubscribeSuccess}
	jsonMsg, err := wsMsg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	err = s.Write(jsonMsg)
	if err != nil {
		panic(err)
	}
	log.Println("register subscribe")
}

func (ws *Websocket) HandleDisconnect(session *melody.Session) {
	_, ok := ws.currentClients[session]
	if ok {
		delete(ws.currentClients, session)
	}
}
