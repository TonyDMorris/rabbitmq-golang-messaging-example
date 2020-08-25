package main

import (
	"rabbitmq-golang-messaging-example/consumer/pkg/app"
	"rabbitmq-golang-messaging-example/consumer/pkg/hub"
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
	h := &hub.Hub{}
	r := gin.Default()
	messageConsumer := message.NewMessageConsumer(cfg.Strategy)
	consumer := app.App{
		Config:          &cfg,
		MessageConsumer: messageConsumer,
		Router:          r,
		Hub:             h,
	}
	consumer.Route()
	consumer.Run()
}
