package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/nats-io/go-nats"
	"log"
	"net/url"
	"testing"
	"time"
)

func TestRunApplication(t *testing.T) {
	app := NewApplication()
	go app.RunApplication()
	time.Sleep(5 * time.Millisecond)
	addr := flag.String("addr", "localhost:5000", "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	wsMsg := WebSocketMessage{Action: ActionSubscribe, Message: "subject"}
	jsonMsg, err := wsMsg.MarshalJSON()
	err = c.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		panic(err)
	}
	receivedMsgChan := make(chan []byte)
	go func(c *websocket.Conn, msgPointer chan []byte) {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			msgPointer <- msg
		}
	}(c, receivedMsgChan)
	nc, _ := nats.Connect(nats.DefaultURL)
	err = nc.Publish("subject", []byte("queueMsg"))
	if err != nil {
		panic(err)
	}
	receivedMsg := <-receivedMsgChan
	if string(receivedMsg) != "queueMsg" {
		panic(`receivedMsg is not equal "queueMsg"`)
	}
	err = nc.Publish("subject", []byte("queueMsg2"))
	if err != nil {
		panic(err)
	}
	receivedMsg = <-receivedMsgChan
	if string(receivedMsg) != "queueMsg2" {
		panic(`receivedMsg is not equal "queueMsg2"`)
	}
	wsMsg = WebSocketMessage{Action: ActionUnsubscribe, Message: "subject"}
	jsonMsg, err = wsMsg.MarshalJSON()
	err = c.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		panic(err)
	}
	select {
	case receivedMsg = <-receivedMsgChan:
		panic("client receive a msg")
	case <-time.After(500 * time.Millisecond):
		return
	}
}
