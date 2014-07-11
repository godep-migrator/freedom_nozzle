package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Message struct {
	Notifications Notifications `xml:"Body>notifications"`
}

type Notifications struct {
	OrganizationId   string
	ActionId         string
	SessionId        string
	EnterpriseUrl    string
	PartnerUrl       string
	NotificationList []Notification `xml:"Notification"`
}

type Notification struct {
	Id             string
	OrganizationId string
	ActionId       string
	SessionId      string
	EnterpriseUrl  string
	PartnerUrl     string
	SObject        SObject `xml:"sObject"`
}

type SObject struct {
	TypeAttr      string `xml:"type,attr" json:"-"`
	Type          string
	SObjectFields []SObjectField `xml:",any" json:"-"`
	Fields        map[string]interface{}
}

func (sobj *SObject) findType() string {
	return strings.TrimPrefix(sobj.TypeAttr, "sf:")
}

func (sobj *SObject) populateFieldValues() {
	sobj.Type = sobj.findType()

	sobj.Fields = make(map[string]interface{})
	for _, field := range sobj.SObjectFields {
		fieldName := field.XMLName.Local
		fieldVal := field.Value
		sobj.Fields[fieldName] = fieldVal
	}
}

type SObjectField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func Unsoap(soapMessage []byte) (notifications []Notification, err error) {
	msg := &Message{}
	xml.Unmarshal([]byte(soapMessage), msg)

	if len(msg.Notifications.NotificationList) < 1 {
		return nil, fmt.Errorf("Message contains no notifications")
	}

	n := msg.Notifications
	for _, nt := range n.NotificationList {
		nt.OrganizationId, nt.ActionId, nt.SessionId = n.OrganizationId, n.ActionId, n.SessionId
		nt.EnterpriseUrl, nt.PartnerUrl = n.EnterpriseUrl, n.PartnerUrl
		nt.SObject.populateFieldValues()
		notifications = append(notifications, nt)
	}

	return notifications, nil
}
