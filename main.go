package main

import (
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
	return &Application{}
}

func (app *Application) RunApplication() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	ws := NewWebsocket(nc)
	server := NewHttpServer(ws)
	err = server.RunServer()
	if err != nil {
		println(err)
	}
}
