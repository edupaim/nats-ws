package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type IHttpServer interface {
}

type HttpServer struct {
	engine    *gin.Engine
	websocket IWebsocket
}

func NewHttpServer(websocket IWebsocket) *HttpServer {
	engine := gin.Default()
	engine.GET("/ws", func(c *gin.Context) {
		err := websocket.HandleWSRequest(c.Writer, c.Request)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}

	})
	server := HttpServer{engine: engine}
	return &server
}

func (server *HttpServer) RunServer() error {
	return server.engine.Run(":5000")
}
