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

const subject = "subject"
const queueMsg1 = "queueMsg"
const queueMsg2 = "queueMsg2"

func TestRunApplication(t *testing.T) {
	//Initialize Application
	app := NewApplication()
	go app.RunApplication()
	time.Sleep(5 * time.Millisecond)
	//Connect websocket client
	c := connectClientToWebsocket()
	defer c.Close()
	receivedMsgChan := receiveMsgFromWebsocket(c)
	//Send subscribe message and expect receive success subscribe message
	sendSubscribeWSMSg(c, subject)
	time.Sleep(5 * time.Millisecond)
	assertReceiveSubscribeSuccessWSMessage(receivedMsgChan)
	//Connect to nats and publish message
	nc := connectOnNats()
	publishQueueMsg(nc, subject, queueMsg1)
	//Expect to receive message
	assertReceiveNatsWSMessage(receivedMsgChan, queueMsg1)
	//Publish and expect receive message
	publishQueueMsg(nc, subject, queueMsg2)
	assertReceiveNatsWSMessage(receivedMsgChan, queueMsg2)
	//Send subscribe message again
	sendSubscribeWSMSg(c, subject)
	//Publish message and expect receibe only 1 message
	publishQueueMsg(nc, subject, queueMsg1)
	assertReceiveNatsWSMessage(receivedMsgChan, queueMsg1)
	assertNotReceiveWSMSg(receivedMsgChan)
	//Send unsubscribe message
	sendUnsubscribeWSMsg(c, subject)
	//Publish and expect not receive message
	publishQueueMsg(nc, subject, queueMsg1)
	assertNotReceiveWSMSg(receivedMsgChan)
}

func connectClientToWebsocket() *websocket.Conn {
	addr := flag.String("addr", "localhost:5000", "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}

func receiveMsgFromWebsocket(c *websocket.Conn) chan []byte {
	receivedMsgChan := make(chan []byte)
	go func(c *websocket.Conn, msgPointer chan []byte) {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			msgPointer <- msg
			println("websocket client receive msg:", string(msg))
		}
	}(c, receivedMsgChan)
	return receivedMsgChan
}

func connectOnNats() *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	return nc
}

func sendUnsubscribeWSMsg(c *websocket.Conn, subject string) {
	wsMsg := WebSocketMessage{Action: ActionUnsubscribe, Message: subject}
	jsonMsg, err := wsMsg.MarshalJSON()
	err = c.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		panic(err)
	}
}

func assertNotReceiveWSMSg(receivedMsgChan chan []byte) {
	select {
	case _ = <-receivedMsgChan:
		panic("client receive msg")
	case <-time.After(500 * time.Millisecond):
	}
}

func sendSubscribeWSMSg(c *websocket.Conn, subject string) {
	wsMsg := WebSocketMessage{Action: ActionSubscribe, Message: subject}
	jsonMsg, err := wsMsg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	err = c.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		panic(err)
	}
}

func assertReceiveNatsWSMessage(receivedMsgChan chan []byte, msg string) {
	receivedMsg := <-receivedMsgChan
	if string(receivedMsg) != msg {
		panic(`receivedMsg is not equal "` + msg + `"`)
	}
}

func assertReceiveSubscribeSuccessWSMessage(receivedMsgChan chan []byte) {
	wsMsg := WebSocketMessage{Action: ActionSubscribeSuccess}
	jsonMsg, err := wsMsg.MarshalJSON()
	if err != nil {
		panic(err)
	}
	receivedMsg := <-receivedMsgChan
	if string(receivedMsg) != string(jsonMsg) {
		panic(`receivedMsg is not equal "` + string(jsonMsg) + `"`)
	}
}

func publishQueueMsg(nc *nats.Conn, subject, msg string) {
	err := nc.Publish(subject, []byte(msg))
	if err != nil {
		panic(err)
	}
}
