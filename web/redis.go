package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func clearCache() error {
	rdb := getRedisClient()
	defer rdb.Close()

	err := rdb.FlushDB(ctx).Err()
	if err != nil {
		return err
	}

	return nil
}

func listenForCacheInvalidation() {
	rdb := getRedisClient()
	defer rdb.Close()

	pubsub := rdb.Subscribe(ctx, "cache_invalidation")
	defer pubsub.Close()

	channel := pubsub.Channel()

	log.Println("Listening for cache invalidation")

	for msg := range channel {
		if msg.Payload != "invalidate" {
			continue
		}
		log.Println("Clearing cache")
		err := clearCache()
		if err != nil {
			log.Printf("Error clearing cache: %v", err)
		}
	}
}

func sendPageIfCached(w http.ResponseWriter, url string) bool {
	rdb := getRedisClient()
	defer rdb.Close()

	cachedPage, err := rdb.Get(ctx, "PAGE:"+url).Result()
	if err == nil { // cache hit
		w.Write([]byte(cachedPage))
		return true
	}
	return false
}

func writePageToCache(pageBuf bytes.Buffer, url string) {
	rdb := getRedisClient()
	defer rdb.Close()

	rdb.Set(ctx, "PAGE:"+url, pageBuf.String(), time.Hour*24)
}
