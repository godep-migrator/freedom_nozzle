package main

import "testing"

func Test_UnsoapNotificationValues(t *testing.T) {
	result, err := unsoap([]byte(singleNotificationMessage))
	if err != nil {
		t.Fatalf("Can not unsoap message to test values: %s", err)
	}

	n := result[0]

	if n.Id != "04lo00000" {
		t.Error("Expected notification id to match input")
	}

	if n.OrganizationId != "00Do00" {
		t.Error("Expected organization id to match input")
	}

	if n.PartnerUrl != "https://nax.salesforce.com/etc" {
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
	if firstResult.Id != "04lo00000x" {
		t.Errorf("Expected first notification id to be 04lo00000x, got: %s", firstResult.Id)
	}

	firstResultSObjId := firstResult.SObject.Fields["Id"]
	if firstResultSObjId != "003o0000x" {
		t.Errorf("Expected first sobject id to be 003o0000x, got: %s", firstResultSObjId)
	}

	secondResult := result[1]
	if secondResult.Id != "04lo00000y" {
		t.Errorf("Expected second notification id to be 04lo00000y, got: %s", secondResult.Id)
	}

	secondResultSObjId := secondResult.SObject.Fields["Id"]
	if secondResultSObjId != "003o0000y" {
		t.Errorf("Expected second sobject id to be 003o0000y, got: %s", secondResultSObjId)
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
