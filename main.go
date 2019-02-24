package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nats-io/go-nats"
	"gopkg.in/olahol/melody.v1"
)

func main() {
	app := NewApplication()
	app.RunApplication()
}

type Application struct {
	currentClients map[*melody.Session]*nats.Subscription
}

func NewApplication() *Application {
	return &Application{currentClients: make(map[*melody.Session]*nats.Subscription)}
}

func (app *Application) RunApplication() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	activeSessions := make(map[*melody.Session]bool)
	r := gin.Default()
	m := melody.New()
	r.GET("/ws", func(c *gin.Context) {
		err := m.HandleRequest(c.Writer, c.Request)
		if err != nil {
			println("error:", err.Error())
		}
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		println(string(msg), s)
		msgWs := WebSocketMessage{}
		err := msgWs.UnmarshalJSON(msg)
		if err != nil {
			return
		}
		switch msgWs.Action {
		case ActionSubscribe:
			println("websocket client bind to", msgWs.Message)
			subs, err := nc.Subscribe(msgWs.Message, func(msg *nats.Msg) {
				err = s.Write(msg.Data)
				if err != nil {
					panic(err)
				}
			})
			if err != nil {
				panic(err)
			}
			app.currentClients[s] = subs
		case ActionUnsubscribe:
			subs, ok := app.currentClients[s]
			if !ok {
				return
			}
			_ = subs.Unsubscribe()
		}
	})
	m.HandleConnect(func(session *melody.Session) {
		activeSessions[session] = true
	})
	err = r.Run(":5000")
	if err != nil {
		panic(err)
	}
}
