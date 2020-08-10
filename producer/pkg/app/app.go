package app

import (
	"fmt"
	"time"

	"github.com/rabbitmq-golang-messaging-example/producer/pkg/message"
)

// Config pulled from environment variables
type Config struct {
	MQHost             string `env:"MQHOST"`
	MQPort             string `env:"MQPORT" `
	Strategy           string `env:"STRATEGY"`
	MQUser             string `env:"MQUSER"`
	MQPassword         string `env:"MQPASSWORD"`
	EveryXMilliseconds int64  `env:"EVERY_X_MILLISECONDS"`
}

type App struct {
	Config          *Config
	MessageProducer message.Producer
}

func (p *App) Run() {
	err := p.MessageProducer.Connect(p.Config.MQUser, p.Config.MQPassword, p.Config.MQHost, p.Config.MQPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("messenger up and running")

	err = p.MessageProducer.DeclareQueue("test-queue")
	if err != nil {
		panic(err)
	}
	var i int
	for range time.Tick(time.Duration(p.Config.EveryXMilliseconds) * time.Millisecond) {
		i = i + 1
		err = p.MessageProducer.SendMessage("test-queue", fmt.Sprintf("message %v", i))
		if err != nil {
			panic(err)
		}
	}
}
