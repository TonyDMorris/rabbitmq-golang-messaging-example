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
	go log.Fatal(a.Router.Run(":9090"))
}
func (a *App) Run() {
	err := a.MessageConsumer.Connect(a.Config.MQUser, a.Config.MQPassword, a.Config.MQHost, a.Config.MQPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("consumer up and running")

	err := a.MessageConsumer.DeclareQueue("test-queue")
	if err != nil {
		panic(err)
	}

	go log.Fatal(a.MessageConsumer.ConsumeMessagesFromQueue("test-queue", a.Hub.HandleStandardMessage))

}
