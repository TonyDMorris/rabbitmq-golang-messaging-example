package app

import (
	"fmt"
	"log"
	"rabbitmq-golang-messaging-example/consumer/pkg/hub"
	"rabbitmq-golang-messaging-example/consumer/pkg/message"

	"github.com/gin-gonic/gin"
)

type App struct {
	Config          *Config
	MessageConsumer message.Consumer
	Router          *gin.Engine
	Hub             *hub.Hub
}

func (a *App) Route() {
	a.Router.GET("/", func(c *gin.Context) {
		c.String(200, "We got Gin")
	})
	a.Router.GET("/ws", func(c *gin.Context) {
		a.Hub.ServeWs(c.Writer, c.Request)
	})

}
func (c *App) Run() {
	err := c.MessageConsumer.Connect(c.Config.MQUser, c.Config.MQPassword, c.Config.MQHost, c.Config.MQPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("consumer up and running")

	err = c.MessageConsumer.DeclareQueue("test-queue")
	if err != nil {
		panic(err)
	}
	c.Route()
	go log.Fatal(c.MessageConsumer.ConsumeMessagesFromQueue("test-queue", c.Hub.HandleStandardMessage))
	go log.Fatal(c.Router.Run(":9090"))
}
