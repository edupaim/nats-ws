package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nats-io/go-nats"
	"gopkg.in/olahol/melody.v1"
	"net/http"
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
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		msgWs := WebSocketMessage{}
		err := msgWs.UnmarshalJSON(msg)
		if err != nil {
			return
		}
		switch msgWs.Action {
		case ActionSubscribe:
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
			if ok {
				_ = subs.Unsubscribe()
				delete(app.currentClients, s)
			}
		}
	})
	m.HandleConnect(func(session *melody.Session) {
		activeSessions[session] = true
	})
	m.HandleDisconnect(func(session *melody.Session) {
		_, ok := app.currentClients[session]
		if ok {
			delete(app.currentClients, session)
		}
		_, ok = activeSessions[session]
		if ok {
			delete(activeSessions, session)
		}
	})
	err = r.Run(":5000")
	if err != nil {
		panic(err)
	}
}
