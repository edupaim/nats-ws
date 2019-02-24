package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nats-io/go-nats"
	"gopkg.in/olahol/melody.v1"
	"log"
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
	r := gin.Default()
	m := melody.New()
	r.GET("/ws", func(c *gin.Context) {
		err := m.HandleRequest(c.Writer, c.Request)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		log.Println("receive ws msg", string(msg))
		msgWs := WebSocketMessage{}
		err := msgWs.UnmarshalJSON(msg)
		if err != nil {
			return
		}
		switch msgWs.Action {
		case ActionSubscribe:
			log.Println("is subscribe msg")
			_, ok := app.currentClients[s]
			if ok {
				return
			}
			subs, err := nc.Subscribe(msgWs.Message, func(msg *nats.Msg) {
				log.Println("subscribe receive a msg")
				err = s.Write(msg.Data)
				if err != nil {
					panic(err)
				}
				log.Println("write msg on ws")
			})
			if err != nil {
				panic(err)
			}
			app.currentClients[s] = subs
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
		case ActionUnsubscribe:
			log.Println("is unsubscribe msg")
			subs, ok := app.currentClients[s]
			if !ok {
				log.Println("client not subscribe")
				return
			}
			err = subs.Unsubscribe()
			if err != nil {
				panic(err)
			}
			delete(app.currentClients, s)
			log.Println("unsubscribe client")
		}
	})
	m.HandleConnect(func(session *melody.Session) {
	})
	m.HandleDisconnect(func(session *melody.Session) {
		_, ok := app.currentClients[session]
		if ok {
			delete(app.currentClients, session)
		}
	})
	err = r.Run(":5000")
	if err != nil {
		panic(err)
	}
}
