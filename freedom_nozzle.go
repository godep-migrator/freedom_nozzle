package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var redisServer = flag.String("redisServer", "localhost:6379", "Redis server with port, if not supplied default is localhost:6379")
var rabbitMQServer = flag.String("rabbitMQServer", "localhost", "RabbitMQ server, if not supplied default is localhost")
var rabbitMQExchange = flag.String("rabbitMQExchange", "salesforce_obm", "RabbitMQ exchange to publish messages to, if not supplied default is 'salesforce_obm'")

var responseTemplate = `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:out="http://soap.sforce.com/2005/09/outbound">
<soapenv:Header/><soapenv:Body>
<out:notificationsResponse><out:Ack>%v</out:Ack></out:notificationsResponse>
</soapenv:Body></soapenv:Envelope>`

func respond(c web.C, w http.ResponseWriter, err error) {
	success := true
	if err != nil {
		log.Printf("Error: %s", err)
		success = false
	}

	log.Print("New notifications: ", c.Env["notificationIds"], " Ack: ", success)
	fmt.Fprintf(w, responseTemplate, success)
}

func handleOutboundMessage(c web.C, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	notifications, err := unsoap(body)
	notificationIds := []string{}

	for _, n := range notifications {
		isNew, redisErr := newNotificationId(n.Id)
		err = redisErr

		if isNew {
			notificationIds = append(notificationIds, n.Id)
			err = publishMessage(n)
		}
	}

	c.Env["notificationIds"] = notificationIds
	respond(c, w, err)
}

func main() {
	flag.Parse()

	redisPool = newRedisPool(*redisServer)
	testRedisPool(redisPool)

	exchangeName = *rabbitMQExchange
	rabbitMQConnection = newRabbitMQConnection(*rabbitMQServer)
	defer rabbitMQConnection.Close()

	goji.Post("/", handleOutboundMessage)
	goji.Serve()
}
