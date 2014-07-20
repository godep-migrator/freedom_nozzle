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
	OrganizationID   string         `xml:"OrganizationId"`
	ActionID         string         `xml:"ActionId"`
	SessionID        string         `xml:"SessionId"`
	EnterpriseURL    string         `xml:"EnterpriseUrl"`
	PartnerURL       string         `xml:"PartnerUrl"`
	NotificationList []notification `xml:"Notification"`
}

type notification struct {
	ID             string  `xml:"Id" json:"Id"`
	OrganizationID string  `json:"OrganizationId"`
	ActionID       string  `json:"ActionId"`
	SessionID      string  `json:"SessionId"`
	EnterpriseURL  string  `json:"EnterpriseUrl"`
	PartnerURL     string  `json:"PartnerUrl"`
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
	xml.Unmarshal(soap, msg)

	if len(msg.Notifications.NotificationList) < 1 {
		return nil, fmt.Errorf("message contains no notifications")
	}

	n := msg.Notifications
	for _, nt := range n.NotificationList {
		nt.OrganizationID, nt.ActionID, nt.SessionID = n.OrganizationID, n.ActionID, n.SessionID
		nt.EnterpriseURL, nt.PartnerURL = n.EnterpriseURL, n.PartnerURL
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
