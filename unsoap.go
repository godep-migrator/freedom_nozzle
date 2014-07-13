package main

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type message struct {
	Notifications notifications `xml:"Body>notifications"`
}

type notifications struct {
	OrganizationId   string
	ActionId         string
	SessionId        string
	EnterpriseUrl    string
	PartnerUrl       string
	NotificationList []notification `xml:"Notification"`
}

type notification struct {
	Id             string
	OrganizationId string
	ActionId       string
	SessionId      string
	EnterpriseUrl  string
	PartnerUrl     string
	SObject        sObject `xml:"sObject"`
}

func (n *notification) routingKey() (key string, err error) {
	key = strings.ToLower(n.SObject.Type)
	if len(key) == 0 {
		return "", fmt.Errorf("could not create routing key")
	}

	action := n.actionType()
	if len(action) > 0 {
		key = key + "." + action
	}

	return key, nil
}

func (n *notification) actionType() (action string) {
	created, err := parseSfTime(n.SObject.Fields["CreatedDate"])
	if err != nil {
		return
	}

	modified, err := parseSfTime(n.SObject.Fields["LastModifiedDate"])
	if err != nil {
		return
	}

	switch {
	case modified.Equal(created):
		action = "create"
	case modified.After(created):
		action = "update"
	}
	return
}

type sObject struct {
	TypeAttr      string `xml:"type,attr" json:"-"`
	Type          string
	SObjectFields []sObjectField `xml:",any" json:"-"`
	Fields        map[string]interface{}
}

func (s *sObject) findType() string {
	return strings.TrimPrefix(s.TypeAttr, "sf:")
}

func (s *sObject) populateFieldValues() {
	s.Type = s.findType()

	s.Fields = make(map[string]interface{})
	for _, field := range s.SObjectFields {
		fieldName := field.XMLName.Local
		fieldVal := field.Value
		s.Fields[fieldName] = fieldVal
	}
}

type sObjectField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func unsoap(soap []byte) (notifications []notification, err error) {
	msg := &message{}
	xml.Unmarshal([]byte(soap), msg)

	if len(msg.Notifications.NotificationList) < 1 {
		return nil, fmt.Errorf("message contains no notifications")
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

func parseSfTime(timeField interface{}) (t time.Time, err error) {
	if timeField == nil {
		return time.Now(), fmt.Errorf("no time found")
	}

	t, err = time.Parse(time.RFC3339, timeField.(string))
	return
}
