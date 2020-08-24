package message

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"github.com/streadway/amqp"
)

type standardMessage struct {
	XMLName xml.Name `xml:"email"`
	Title   string   `xml:"title"`
	Body    string   `xml:"body"`
	Comment string   `xml:",comment"`
}

// Producer is the interface which controls connecting to a message broker , declaring a queue, and finally sending messages
type Producer interface {
	Connect(user string, password string, host string, port string) error
	SendMessage(queue string, msg string) error
	DeclareQueue(queueName string) error
}

func NewMessageProducer(strategy string) Producer {
	// strategy allows for other messaging solutions to be implemented
	producers := map[string]func() Producer{}

	producers["rabbitMQ"] = NewRabbitMQProducer

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
	stdMsg := &standardMessage{Body: msg, Title: "train update"}
	output, err := xml.MarshalIndent(stdMsg, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	err = r.channel.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/xml",
			Body:        output,
		})
	return err
}

func NewRabbitMQProducer() Producer {
	return &RabbitMQProducer{}
}
