package main

import (
	"testing"
	"time"
)

func Test_UnsoapNotificationValues(t *testing.T) {
	result, err := unsoap([]byte(singleNotificationMessage))
	if err != nil {
		t.Fatalf("Can not unsoap message to test values: %s", err)
	}

	n := result[0]

	if n.ID != "04lo00000" {
		t.Error("Expected notification id to match input")
	}

	if n.OrganizationID != "00Do00" {
		t.Error("Expected organization id to match input")
	}

	if n.PartnerURL != "https://nax.salesforce.com/etc" {
		t.Error("Expected partnerurl to match input")
	}
}

func Test_UnsoapSObjectValues(t *testing.T) {
	result, err := unsoap([]byte(singleNotificationMessage))
	if err != nil {
		t.Fatalf("Can not unsoap message to test values: %s", err)
	}

	n := result[0]

	if n.SObject.Type != "Contact" {
		t.Error("Expected sobject type to be Contact")
	}

	if n.SObject.Fields["LastName"] != "dude" {
		t.Error("Expected sobject field LastName to be dude")
	}
}

func Test_UnsoapMultipleNotifications(t *testing.T) {
	result, _ := unsoap([]byte(multipleNotificationMessage))
	if len(result) != 2 {
		t.Errorf("Expected to find two notifications, found %v", len(result))
	}
}

func Test_UnsoapMultipleNotificationValues(t *testing.T) {
	result, _ := unsoap([]byte(multipleNotificationMessage))
	if len(result) != 2 {
		t.Fatalf("Expected to find two notifications, found %v", len(result))
	}

	firstResult := result[0]
	if firstResult.ID != "04lo00000x" {
		t.Errorf("Expected first notification id to be 04lo00000x, got: %s", firstResult.ID)
	}

	firstResultSObjID := firstResult.SObject.Fields["Id"]
	if firstResultSObjID != "003o0000x" {
		t.Errorf("Expected first sobject id to be 003o0000x, got: %s", firstResultSObjID)
	}

	secondResult := result[1]
	if secondResult.ID != "04lo00000y" {
		t.Errorf("Expected second notification id to be 04lo00000y, got: %s", secondResult.ID)
	}

	secondResultSObjID := secondResult.SObject.Fields["Id"]
	if secondResultSObjID != "003o0000y" {
		t.Errorf("Expected second sobject id to be 003o0000y, got: %s", secondResultSObjID)
	}
}

func Test_RoutingKeyWithoutDates(t *testing.T) {
	key, err := testNotification.routingKey()
	if err != nil {
		t.Errorf("Error creating routing key: %s", err)
	}

	if key != "testtype" {
		t.Errorf("Expected routing key testtype, got: %v", key)
	}
}

func Test_RoutingKeyWithDates(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	dur := time.Duration(-1 * time.Hour)
	before := time.Now().Add(dur).Format(time.RFC3339)

	//when CreatedDate and LastModifiedDate are equal should have key *.created
	testNotification.SObject.Fields = make(map[string]interface{})
	testNotification.SObject.Fields["CreatedDate"] = now
	testNotification.SObject.Fields["LastModifiedDate"] = now

	key, err := testNotification.routingKey()
	if err != nil {
		t.Errorf("Error creating RabbitMQ routing key: %s", err)
	}

	if key != "testtype.create" {
		t.Errorf("Expected routing key testtype.create, got: %v", key)
	}

	//when CreatedDate before LastModifiedDate should have key *.update
	testNotification.SObject.Fields["CreatedDate"] = before

	key, err = testNotification.routingKey()
	if err != nil {
		t.Errorf("Error creating RabbitMQ routing key: %s", err)
	}

	if key != "testtype.update" {
		t.Errorf("Expected routing key testtype.update, got: %v", key)
	}

	//when one date present and other missing should have just objecttype key
	delete(testNotification.SObject.Fields, "LastModifiedDate")

	key, err = testNotification.routingKey()
	if err != nil {
		t.Errorf("Error creating RabbitMQ routing key: %s", err)
	}

	if key != "testtype" {
		t.Errorf("Expected routing key testtype, got: %v", key)
	}
}

var singleNotificationMessage = `
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
 <soapenv:Body>
  <notifications xmlns="http://soap.sforce.com/2005/09/outbound">
   <OrganizationId>00Do00</OrganizationId>
	 <PartnerUrl>https://nax.salesforce.com/etc</PartnerUrl>
   <Notification>
    <Id>04lo00000</Id>
    <sObject xsi:type="sf:Contact" xmlns:sf="urn:sobject.enterprise.soap.sforce.com">
     <sf:Id>003o0000xxxxxxxxxx</sf:Id>
     <sf:CreatedDate>2014-07-10T00:05:38.000Z</sf:CreatedDate>
     <sf:FirstName>some</sf:FirstName>
     <sf:LastModifiedDate>2014-07-10T00:05:38.000Z</sf:LastModifiedDate>
     <sf:LastName>dude</sf:LastName>
    </sObject>
   </Notification>
  </notifications>
 </soapenv:Body>
</soapenv:Envelope>
`
var multipleNotificationMessage = `
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
 <soapenv:Body>
  <notifications xmlns="http://soap.sforce.com/2005/09/outbound">
   <Notification>
    <Id>04lo00000x</Id>
    <sObject xsi:type="sf:Contact" xmlns:sf="urn:sobject.enterprise.soap.sforce.com">
     <sf:Id>003o0000x</sf:Id>
    </sObject>
   </Notification>
   <Notification>
    <Id>04lo00000y</Id>
    <sObject xsi:type="sf:Contact" xmlns:sf="urn:sobject.enterprise.soap.sforce.com">
     <sf:Id>003o0000y</sf:Id>
    </sObject>
   </Notification>
  </notifications>
 </soapenv:Body>
</soapenv:Envelope>
`
