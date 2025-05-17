package main

import (
    "log"

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

func listenCacheInvalidation() {
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
            log.Print("Error clearing cache: %v", err)
        }
    }
}
