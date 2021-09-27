package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()
	host := os.Args[1]
	keys, err := Run(ctx, host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting keys for %s, %s\n", host, err)
		os.Exit(1)
	}
	b, _ := json.Marshal(keys)
	os.Stdout.Write(b)
}

// Run lists keys. host should be an IP:port pair
func Run(ctx context.Context, host string) ([]string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	keys := rdb.Keys(ctx, "*")
	err := keys.Err()
	if err != nil {
		return nil, err
	}

	return keys.Val(), nil
}
