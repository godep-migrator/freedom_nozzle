package main

import (
	"log"

	"github.com/garyburd/redigo/redis"
)

var redisPool *redis.Pool

func newRedisPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   20,
		MaxActive: 200,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				log.Fatalf("Could not connect to redis: %s", err)
			}
			return c, err
		},
	}
}

func testRedisPool(pool *redis.Pool) {
	c := pool.Get()
	defer c.Close()
	_, _ = c.Do("PING")
}

func newNotificationID(id string) (isNew bool, err error) {
	conn := redisPool.Get()
	defer conn.Close()

	key := "notification:" + id
	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}

	if exists {
		return false, nil
	}

	//store this notification id for 25 hours, salesforce may try to resend for up to 24 hours
	conn.Do("SET", key, 1, "EX", 90000)
	return true, nil
}

func clearNotificationID(id string) (err error) {
	conn := redisPool.Get()
	defer conn.Close()

	key := "notification:" + id
	_, err = conn.Do("DEL", key)
	return
}
