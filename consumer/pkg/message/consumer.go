package message

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/streadway/amqp"
)

type MessageHandler func([]byte)

type Consumer interface {
	Connect(user string, password string, host string, port string) error
	DeclareQueue(queueName string) error
	ConsumeMessagesFromQueue(queueName string, fn MessageHandler) error
}

func NewMessageConsumer(strategy string) Consumer {

	// strategy allows for other messaging solutions to be implemented
	consumers := map[string]func() Consumer{}

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

// DeclareQueue declares the queue that the client should receive messages from
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
func (r *RabbitMQConsumer) ConsumeMessagesFromQueue(queueName string, fn MessageHandler) error {
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
			go fn(d.Body)
		}
	}()
	log.Printf(" [*] Waiting for messages.")
	<-forever
	return nil
}
func newRabbitMQConsumer() Consumer {
	return &RabbitMQConsumer{}
}
