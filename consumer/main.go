package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
)

// config pulled from environment variables
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
	messageConsumer := NewMessageConsumer(cfg.Strategy)
	consumer := Consumer{
		config:          &cfg,
		MessageConsumer: messageConsumer,
		ConsumerFunc: func(msg string) {
			fmt.Println(msg)
		},
	}

	consumer.Run()
}

type Consumer struct {
	config          *config
	MessageConsumer MessageConsumer
	ConsumerFunc    func(msg string)
}

func (c *Consumer) Run() {
	err := c.MessageConsumer.Connect(c.config.MQUser, c.config.MQPassword, c.config.MQHost, c.config.MQPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("consumer up and running")

	err = c.MessageConsumer.DeclareQueue("test-queue")
	if err != nil {
		panic(err)
	}

	log.Fatal(c.MessageConsumer.ConsumeMessagesFromQueue("test-queue", c.ConsumerFunc))

}

// MessageConsumer is the interface which controls connecting to a message broker, joining a queue as a consumer, and ultimately consuming messages
type MessageConsumer interface {
	Connect(user string, password string, host string, port string) error
	DeclareQueue(queueName string) error
	ConsumeMessagesFromQueue(queueName string, fn func(msg string)) error
}

func NewMessageConsumer(strategy string) MessageConsumer {

	// strategy allows for other messaging solutions to be implemented
	consumers := map[string]func() MessageConsumer{}

	consumers["rabbitMQ"] = newRabbitMQConsumer

	return consumers[strategy]()

}

// RabbitMQConsumer is the concrete implementation used it holds connection info, channel info, and implements the MessageConsumer interface
type RabbitMQConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// Connect connects to the rabbitMQ instance
func (r *RabbitMQConsumer) Connect(user string, password string, host string, port string) error {
	// generate a new tls config
	cfg := new(tls.Config)
	// InsecureSkipVerify allows the connection to succeed even when the certificate is signed by a bad authority
	// which in this case it is as i have not shared the cert files from the rabbit mq instance !! NOT FOR PRODUCTION
	cfg.InsecureSkipVerify = true

	//
	cfg.RootCAs = x509.NewCertPool()
	// if the cert files are shared we can add the cacert to the root trusted authorities and load the key pair into the config
	if ca, err := ioutil.ReadFile("testca/cacert.pem"); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	}

	if cert, err := tls.LoadX509KeyPair("client/cert.pem", "client/key.pem"); err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	}
	// create a connection to rabbitMQ the given params and generated config
	conn, err := amqp.DialTLS(fmt.Sprintf("amqps://%v:%v@%v:%v", user, password, host, port), cfg)
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

// DeclareQueue declares the queue that the client should recceive messages from
func (r *RabbitMQConsumer) DeclareQueue(queueName string) error {

	_, err := r.channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}

	return nil
}

// ConsumeMessagesFromQueue takes a simple function and executes it with each message body within a goroutine
func (r *RabbitMQConsumer) ConsumeMessagesFromQueue(queueName string, fn func(msg string)) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			go fn(string(d.Body))
		}
	}()
	log.Printf(" [*] Waiting for messages.")
	<-forever
	return nil
}
func newRabbitMQConsumer() MessageConsumer {
	return &RabbitMQConsumer{}
}
