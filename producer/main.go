package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/streadway/amqp"
	"time"
)

type config struct {
	MQHost     string `env:"MQHOST"`
	MQPort     string `env:"MQPORT" `
	Strategy   string `env:"STRATEGY"`
	MQUser     string `env:"MQUSER"`
	MQPassword string `env:"MQPASSWORD"`
}

func main() {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	producer := NewMessageProducer(cfg.Strategy)

	err = producer.Connect(cfg.MQUser, cfg.MQPassword, cfg.MQHost, cfg.MQPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("messenger up and running")

	err = producer.DeclareQueue("test-queue")
	if err != nil {
		panic(err)
	}

	for x := range time.Tick(10 * time.Second) {
		err = producer.SendMessage("test-que", fmt.Sprintf("hello world the time is %v", x))
		if err != nil {
			panic(err)
		}
	}

}

type MessageProducer interface {
	Connect(user string, password string, host string, port string) error
	SendMessage(que string, msg string) error
	DeclareQueue(name string) error
}

func NewMessageProducer(strategy string) MessageProducer {
	producers := map[string]func() MessageProducer{}

	producers["rabbitMQ"] = newRabbitMQProducer

	return producers[strategy]()

}

type RabbitMQProducer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	ques       map[string]*amqp.Queue
}

func (r *RabbitMQProducer) Connect(user string, password string, host string, port string) error {
	cfg := new(tls.Config)

	// see at the top
	cfg.RootCAs = x509.NewCertPool()

	//if ca, err := ioutil.ReadFile("testca/cacert.pem"); err == nil {
	//	cfg.RootCAs.AppendCertsFromPEM(ca)
	//}
	//
	//// Move the client cert and key to a location specific to your application
	//// and load them here.
	//
	//if cert, err := tls.LoadX509KeyPair("client/cert.pem", "client/key.pem"); err == nil {
	//	cfg.Certificates = append(cfg.Certificates, cert)
	//}
	conn, err := amqp.DialTLS(fmt.Sprintf("amqps://%v:%v@%v:%v/", user, password, host, port), cfg)
	if err != nil {
		return err
	}
	r.connection = conn
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	r.channel = ch

	fmt.Print("connected to rabbitMQ")
	return nil
}

func (r *RabbitMQProducer) DeclareQueue(name string) error {
	_, err := r.channel.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQProducer) SendMessage(que string, msg string) error {
	err := r.channel.Publish(
		"",    // exchange
		que,   // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	return err
}

func newRabbitMQProducer() MessageProducer {
	return &RabbitMQProducer{}
}
