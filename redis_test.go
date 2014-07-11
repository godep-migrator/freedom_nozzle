package main

import "testing"

func Test_NewNotificationId(t *testing.T) {
	redisPool = newRedisPool("localhost:6379")

	clearNotificationId("testing_freedom_nozzle")

	//first time checking, should be new
	isNew, err := newNotificationId("testing_freedom_nozzle")
	if err != nil {
		t.Fatal("Could not test notification id handling")
	}

	if !isNew {
		t.Error("Expected notification id to be new")
	}

	//since it's already been checked should be false now
	isNew, err = newNotificationId("testing_freedom_nozzle")
	if err != nil {
		t.Fatal("Could not test notification id handling")
	}

	if isNew {
		t.Error("Expected notification id to not be new")
	}

	clearNotificationId("testing_freedom_nozzle")
}
