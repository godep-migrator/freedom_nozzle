package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

var channel *amqp.Channel
var rabbitMQConnection *amqp.Connection
var exchangeName string

func newRabbitMQConnection(server string) *amqp.Connection {
	connection, err := amqp.Dial("amqp://" + server)
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %s", err)
	}

	channel, err = connection.Channel()
	if err != nil {
		log.Fatalf("Could not open RabbitMQ channel: %s", err)
	}

	err = channel.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Could not bind to exchange: %s", err)
	}

	return connection
}

func publishMessage(n notification) (err error) {
	msg, err := createMessage(n)
	if err != nil {
		return
	}

	key, err := n.routingKey()
	if err != nil {
		return
	}

	err = channel.Publish(exchangeName, key, false, false, msg)
	return
}

func createMessage(n notification) (msg amqp.Publishing, err error) {
	json, err := json.Marshal(n)
	if err != nil {
		return amqp.Publishing{}, err
	}

	msg = amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         json,
	}

	return
}
