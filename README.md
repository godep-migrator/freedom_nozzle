#freedom_nozzle

A server that can listen to Salesforce.com Outbound Messages and insert the contents of any new notifications into RabbitMQ

##Why?

It can occasionally become necessary to interact with data in Salesforce.com and respond programmatically to changes in those data. Using Salesforce.com's built-in tools or related commercial offerings may sometimes be undesirable. freedom_nozzle was created to provide a more comfortable option, by receiving Outbound Messages from Salesforce and publishing the details to RabbitMQ. You can then do the needful using your favorite language that has a RabbitMQ client.

##Dependencies
Redis and RabbitMQ are required.

##Salesforce configuration
Outbound Messaging should be configured in the salesforce org that you want to receive messages from. See the documentation for Outbound Messaging [here](http://www.salesforce.com/us/developer/docs/api/index_Left.htm#CSHID=sforce_api_om_outboundmessaging.htm|StartTopic=Content%2Fsforce_api_om_outboundmessaging.htm|SkinName=webhelp). The endpoint URL when creating the Outbound Message should be the URL where freedom_nozzle is running.

##Messages published to RabbitMQ
For each unique notification received by freedom_nozzle, the contents will be sent to a RabbitMQ topic exchange as json. The exchange is "salesforce_obm" by default but this can be changed, see Usage below. Fields from the message envelope itself are included with each notification, and each notification is published individually to the exchange.

The specific notifications in the Outbound Messages can be sent by Salesforce more than once. freedom_nozzle tracks which notifications have been published to the exchange and ignores notifications that have been successfully processed before. Each unique notification should be published to the exchange exactly once.

####Message body

Given this Outbound Message:
```
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
 <soapenv:Body>
  <notifications xmlns="http://soap.sforce.com/2005/09/outbound">
   <OrganizationId>00Do0000000xxxxxxx</OrganizationId>
   <ActionId>04ko0000000xxxxxxx</ActionId>
   <SessionId>00Do0000xxxxxxx!etc</SessionId>
   <EnterpriseUrl>https://nax.salesforce.com/services/Soap/c/30.0/00Do0000xxxxxxx</EnterpriseUrl>
   <PartnerUrl>https://nax.salesforce.com/services/Soap/u/30.0/00Do0000xxxxxxx</PartnerUrl>
   <Notification>
    <Id>04lo0000000xxxxxxx</Id>
    <sObject xsi:type="sf:Contact" xmlns:sf="urn:sobject.enterprise.soap.sforce.com">
     <sf:Id>003o000000xxxxxxxxx</sf:Id>
     <sf:CreatedDate>2014-06-26T00:05:38.000Z</sf:CreatedDate>
     <sf:FirstName>some</sf:FirstName>
     <sf:LastModifiedDate>2014-06-26T00:05:38.000Z</sf:LastModifiedDate>
     <sf:LastName>dude</sf:LastName>
    </sObject>
   </Notification>
  </notifications>
 </soapenv:Body>
</soapenv:Envelope>
```

This will be the body published to RabbitMQ:
```
{ "ActionId" : "04ko0000000xxxxxxx",
  "EnterpriseUrl" : "https://nax.salesforce.com/services/Soap/c/30.0/00Do0000xxxxxxx",
  "Id" : "04lo0000000xxxxxxx",
  "OrganizationId" : "00Do0000000xxxxxxx",
  "PartnerUrl" : "https://nax.salesforce.com/services/Soap/u/30.0/00Do0000xxxxxxx",
  "SessionId" : "00Do0000xxxxxxx!etc",
  "SObject" : {
      "Type" : "Contact",
      "Fields" : {
            "CreatedDate" : "2014-06-26T00:05:38.000Z",
            "FirstName" : "some",
            "Id" : "003o000000xxxxxxxxx",
            "LastModifiedDate" : "2014-06-26T00:05:38.000Z",
            "LastName" : "dude"
       }
  }
}

```

The exact field contents of the message are controlled by configuring the Outbound Message settings in Salesforce.

####Routing keys:
The routing key used when publishing to the exchange will be based on the SObject type, CreatedDate, and LastModified date. It will be in the form "object" or "objectname.action", where:

* "objectname" is the name of the SObject that triggered the Outbound Message
* if CreatedDate and LastModified fields are not included in the message, "action" won't be included
* if CreatedDate == LastModifiedDate, "action" will be "create"
* if LastModifiedDate > CreatedDate, "action" will be "update"

The exchange used will be a topic exchange, and this key arrangement allows sending various combinations of objects and actions to different combinations of queues using routing key bindings.

Examples:

* SObject is Contact, CreatedDate and LastModifiedDate are present, LastModifiedDate > CreatedDate:

    contact.update

* SObject is Contact, CreatedDate and LastModifiedDate are present and equal:

    contact.create

* SObject is Contact, CreatedDate or LastModifiedDate not present:

    contact

* SObject is MyCustomThing__c, CreatedDate and LastModifiedDate are present and equal:

    mycustomthing__c.create

Note that if a Salesforce.com record is created and then changes before the Outbound Message is sent, it is possible that freedom_nozzle would create an "update" message for a record that never had a "create" message. There is more info about this caveat in the Salesforce documentation linked above.


###Response to Salesforce Outbound Messages
When freedom_nozzle receives a message and is able to publish any new notifications to the exchange, it will Ack the Outbound Message to Salesforce. If there are new notifications in the Outbound Message but they cannot be handled for some reason, a false Ack will be returned to Salesforce.


##Installation

To install from source you must have Go installed, then:

    go get github.com/cobraextreme/freedom_nozzle
    go install github.com/cobraextreme/freedom_nozzle


##Usage

```
$ ./freedom_nozzle -h
Usage of ./freedom_nozzle:
  -bind=":8000": Address to bind on. If this value has a colon, as in ":8000" or
		"127.0.0.1:9001", it will be treated as a TCP address. If it
		begins with a "/" or a ".", it will be treated as a path to a
		UNIX socket. If it begins with the string "fd@", as in "fd@3",
		it will be treated as a file descriptor (useful for use with
		systemd, for instance). If it begins with the string "einhorn@",
		as in "einhorn@0", the corresponding einhorn socket will be
		used. If an option is not explicitly passed, the implementation
		will automatically select among "einhorn@0" (Einhorn), "fd@3"
		(systemd), and ":8000" (fallback) based on its environment.
  -rabbitMQExchange="salesforce_obm": RabbitMQ exchange to publish messages to, if not supplied default is 'salesforce_obm'
  -rabbitMQServer="localhost": RabbitMQ server, if not supplied default is "localhost"
  -redisServer="localhost:6379": Redis server with port, if not supplied default is "localhost:6379"
```

For real life usage you should only expose freedom_nozzle with SSL, as there could be potentially sensitive data in the Outbound Messages from Salesforce.com.

##Tests

Included tests require Redis running on localhost:6379 and RabbitMQ running on localhost. Run tests with

    go test

