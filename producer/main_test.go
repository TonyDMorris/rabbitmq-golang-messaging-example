package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestProducer struct {
	t *testing.T
}

func (tp *TestProducer) Connect(user string, password string, host string, port string) error {
	assert.Equal(tp.t, "host", host)
	assert.Equal(tp.t, "mqport", port)
	assert.Equal(tp.t, "user", user)
	assert.Equal(tp.t, "pass", password)
	return nil
}
func (tp *TestProducer) SendMessage(queue string, msg string) error {

	return errors.New("test complete")
}
func (tp *TestProducer) DeclareQueue(queueName string) error {
	assert.Equal(tp.t, "test-queue", queueName)
	return nil
}

func TestRabbitMQProducer_Connect(t *testing.T) {
	producer := Producer{
		config: &config{
			MQHost: "host",
			MQPort: "mqport",

			MQUser:     "user",
			MQPassword: "pass",
		},
		MessageProducer: &TestProducer{t},
	}
	producer.Run()
}
