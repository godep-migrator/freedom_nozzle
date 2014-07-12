package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

var testExchange = "testing_freedom_nozzle_exchange"
var testQueue = "testing_freedom_nozzle_queue"

var testNotification = notification{
	Id: "testId",
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

	if notification.Id != "testId" {
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

func Test_CreateKeyWithoutDates(t *testing.T) {
	key, err := createKey(testNotification)
	if err != nil {
		t.Errorf("Error creating routing key: %s", err)
	}

	if key != "testtype" {
		t.Errorf("Expected routing key testtype, got: %v", key)
	}
}

func Test_CreateKeyWithDates(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	dur := time.Duration(-1 * time.Hour)
	before := time.Now().Add(dur).Format(time.RFC3339)

	//when CreatedDate and LastModifiedDate are equal should have key *.created
	testNotification.SObject.Fields = make(map[string]interface{})
	testNotification.SObject.Fields["CreatedDate"] = now
	testNotification.SObject.Fields["LastModifiedDate"] = now

	key, err := createKey(testNotification)
	if err != nil {
		t.Errorf("Error creating RabbitMQ routing key: %s", err)
	}

	if key != "testtype.create" {
		t.Errorf("Expected routing key testtype.create, got: %v", key)
	}

	//when CreatedDate before LastModifiedDate should have key *.update
	testNotification.SObject.Fields["CreatedDate"] = before

	key, err = createKey(testNotification)
	if err != nil {
		t.Errorf("Error creating RabbitMQ routing key: %s", err)
	}

	if key != "testtype.update" {
		t.Errorf("Expected routing key testtype.update, got: %v", key)
	}

	//when one date present and other missing should have just objecttype key
	delete(testNotification.SObject.Fields, "LastModifiedDate")

	key, err = createKey(testNotification)
	if err != nil {
		t.Errorf("Error creating RabbitMQ routing key: %s", err)
	}

	if key != "testtype" {
		t.Errorf("Expected routing key testtype, got: %v", key)
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
