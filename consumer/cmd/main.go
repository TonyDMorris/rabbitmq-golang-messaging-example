package main

import (
	"rabbitmq-golang-messaging-example/consumer/pkg/app"
	"rabbitmq-golang-messaging-example/consumer/pkg/message"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := app.Config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	r := gin.Default()
	messageConsumer := message.NewMessageConsumer(cfg.Strategy)
	consumer := app.App{
		Config:          &cfg,
		MessageConsumer: messageConsumer,
		Router:          r,
	}

	consumer.Run()
}

// MessageConsumer is the interface which controls connecting to a message broker, joining a queue as a consumer, and ultimately consuming messages
