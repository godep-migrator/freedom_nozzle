package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

	key, err := createKey(n)
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

func createKey(n notification) (key string, err error) {
	objectName := n.SObject.Type
	if len(objectName) == 0 {
		return "", fmt.Errorf("could not create routing key")
	}

	key = strings.ToLower(objectName)

	objectAction := objectAction(n)
	if len(objectAction) > 0 {
		key = key + "." + objectAction
	}

	return key, nil
}

func objectAction(n notification) (action string) {
	objectCreateTime, createErr := parseSalesforceTime(n.SObject.Fields["CreatedDate"])
	objectModifiedTime, modifiedErr := parseSalesforceTime(n.SObject.Fields["LastModifiedDate"])

	switch {
	case createErr != nil || modifiedErr != nil:
		return ""
	case objectModifiedTime.Equal(objectCreateTime):
		return "create"
	case objectModifiedTime.After(objectCreateTime):
		return "update"
	}
	return ""
}

func parseSalesforceTime(timeField interface{}) (t time.Time, err error) {
	if timeField == nil {
		return time.Now(), fmt.Errorf("no time found")
	}

	t, err = time.Parse(time.RFC3339, timeField.(string))
	return
}
