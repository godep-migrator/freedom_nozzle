package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/streadway/amqp"
)

var testExchange = "testing_freedom_nozzle_exchange"
var testQueue = "testing_freedom_nozzle_queue"

var testNotification = notification{
	ID: "testId",
	SObject: sObject{
		Type: "testtype",
	},
}

func Test_PublishMessage(t *testing.T) {
	exchangeName = testExchange
	rabbitMQConnection = newRabbitMQConnection("localhost")

	out, err := testReceiveQueue(t)
	if err != nil {
		t.Fatalf("Could not test RabbitMQ publish: %s", err)
	}

	err = publishMessage(testNotification)
	if err != nil {
		fmt.Println(err)
	}

	delivery := <-out
	notification := &notification{}
	err = json.Unmarshal(delivery.Body, notification)
	if err != nil {
		t.Fatalf("Could not retreive test notification from RabbitMQ")
	}

	if notification.ID != "testId" {
		t.Error("Did not retrieve expected test notification from RabbitMQ")
	}

	cleanUp()
}

func Test_CreateMessage(t *testing.T) {
	msg, err := createMessage(testNotification)
	if err != nil {
		t.Errorf("Error creating RabbitMQ message: %s", err)
	}

	if msg.ContentType != "application/json" {
		t.Errorf("Expected message content type application/json, got: %v", msg.ContentType)
	}
}

func testReceiveQueue(t *testing.T) (out <-chan amqp.Delivery, err error) {
	_, err = channel.QueueDeclare(testQueue, false, true, false, false, nil)
	if err != nil {
		return
	}

	err = channel.QueueBind(testQueue, "#", testExchange, false, nil)
	if err != nil {
		return
	}

	out, err = channel.Consume(testQueue, "testreceive", true, false, false, false, nil)
	return
}

func cleanUp() {
	channel.ExchangeDelete(testExchange, false, false)
	channel.QueueDelete(testQueue, false, false, false)
}
