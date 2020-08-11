package main

import (
	env "github.com/caarlos0/env/v6"
	"github.com/rabbitmq-golang-messaging-example/producer/pkg/app"
	"github.com/rabbitmq-golang-messaging-example/producer/pkg/message"
)

func main() {

	cfg := app.Config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	messageProducer := message.NewMessageProducer(cfg.Strategy)
	producer := app.App{
		Config:          &cfg,
		MessageProducer: messageProducer,
	}

	producer.Run()

}
