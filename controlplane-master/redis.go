package main

import (
	"os"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
)

func MustConnectToRedis() *redis.Client {
	addr := os.Getenv("REDIS_HOST")
	if addr == "" {
		glog.Exit("no REDIS_HOST")
	}

	passwd := os.Getenv("REDIS_PASSWORD")
	if passwd != "" {
		glog.Infof("connecting to Redis server %q with password", addr)
	} else {
		glog.Infof("connecting to Redis server %q without password", addr)
	}

	glog.Infof("connecting to Redis at %q", addr)
	client := redis.NewClient(&redis.Options{
		Addr:       addr,
		Password:   passwd,
		MaxRetries: 3,
	})

	for i := 1; i <= 60; i++ {
		glog.Infof("checking connectivity to Redis, attempt %d", i)
		if _, err := client.Ping().Result(); err != nil {
			glog.Infof("failed to get connect to Redis, will retry shortly: %v", err)
			time.Sleep(time.Second)
		} else {
			glog.Info("successfully connected to Redis!")
			return client
		}
	}

	glog.Exit("failed to connect to Redis")
	return nil
}
