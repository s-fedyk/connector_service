package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func init() {
	log.Print("Initializing Redis...")
	ctx := context.Background()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is not set")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
    Password: "",
		DB:   0,
    DialTimeout: 5*time.Second,
    ReadTimeout: 3*time.Second,
    WriteTimeout: 3*time.Second,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis successfully!")
	}
}

func store(buf []byte, filename string) (bool, error) {
	ctx := context.Background()

	redisClient.Set(ctx, filename, buf, 10*time.Minute)

	return true, nil
}
