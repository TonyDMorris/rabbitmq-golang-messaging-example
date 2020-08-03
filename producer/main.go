package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/streadway/amqp"
	"io/ioutil"
	"time"
)

// config pulled from environment variables
type config struct {
	MQHost             string `env:"MQHOST"`
	MQPort             string `env:"MQPORT" `
	Strategy           string `env:"STRATEGY"`
	MQUser             string `env:"MQUSER"`
	MQPassword         string `env:"MQPASSWORD"`
	EveryXMilliseconds int64  `env:"EVERY_X_MILLISECONDS"`
}

func main() {

	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	messageProducer := NewMessageProducer(cfg.Strategy)
	producer := Producer{
		config:          &cfg,
		MessageProducer: messageProducer,
	}

	producer.Run()

}

type Producer struct {
	config          *config
	MessageProducer MessageProducer
}

func (p *Producer) Run() {
	err := p.MessageProducer.Connect(p.config.MQUser, p.config.MQPassword, p.config.MQHost, p.config.MQPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("messenger up and running")

	err = p.MessageProducer.DeclareQueue("test-queue")
	if err != nil {
		panic(err)
	}
	var i int
	for range time.Tick(time.Duration(p.config.EveryXMilliseconds) * time.Millisecond) {
		i = i + 1
		err = p.MessageProducer.SendMessage("test-queue", fmt.Sprintf("message %v", i))
		if err != nil {
			panic(err)
		}
	}
}

// MessageProducer is the interface which controls connecting to a message broker , declaring a queue, and finally sending messages
type MessageProducer interface {
	Connect(user string, password string, host string, port string) error
	SendMessage(queue string, msg string) error
	DeclareQueue(queueName string) error
}

func NewMessageProducer(strategy string) MessageProducer {
	// strategy allows for other messaging solutions to be implemented
	producers := map[string]func() MessageProducer{}

	producers["rabbitMQ"] = newRabbitMQProducer

	return producers[strategy]()

}

// RabbitMQProducer is the concrete implementation used it holds connection info, and channel info
type RabbitMQProducer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// Connect connects to the rabbitMQ instance
func (r *RabbitMQProducer) Connect(user string, password string, host string, port string) error {
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

	fmt.Print("[*] Connected sending messages.")
	return nil
}

func (r *RabbitMQProducer) DeclareQueue(queueName string) error {

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

func (r *RabbitMQProducer) SendMessage(queue string, msg string) error {
	err := r.channel.Publish(
		"",    // exchange
		queue, // routing key
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
