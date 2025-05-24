package services

import (
    "context"

    "github.com/redis/go-redis/v9"
)

func getRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
}

func InvalidateClientCache() error {
    rdb := getRedisClient()
    defer rdb.Close()
    
    err := rdb.Publish(context.Background(), "cache_invalidation", "invalidate").Err()
    if err != nil {
        return err
    }
    
    return nil
}

// TODO: cache db requests
