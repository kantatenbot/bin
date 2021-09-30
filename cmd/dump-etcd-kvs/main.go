package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func main() {
	host := os.Args[1]
	kvs, err := Run(host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting keys for %s, %s\n", host, err)
		os.Exit(1)
	}

	b, _ := json.Marshal(kvs)
	os.Stdout.Write(b)
}

// Run lists keys. host should be an IP:port pair
func Run(host string) ([]*mvccpb.KeyValue, error) {
	connectTimeout := 3
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{host},
		// this dial timeout doesn't actually cause an error to be returned if
		// the host is unreachable, so we hack around it with a status call
		// afterwards
		DialTimeout: time.Duration(connectTimeout) * time.Second,
		Logger:      zap.New(nil), // client noisy as hell stfu
	})
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// hack to timeout on unreachable hosts
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectTimeout)*time.Second)
	defer cancel()
	_, err = client.Status(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("unabled to get status (host is probably down. the client lib doesnt surface network errors), %s", err.Error())
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.KV.Get(ctx, "", clientv3.WithFromKey())
	// https://github.com/etcd-io/etcd/blob/58fb625d1226d7dc47f4fe0355ac5b169e7d3ef6/client/v3/op.go#L411-L419
	// https://github.com/etcd-io/etcd/blob/2c834459e1aab78a5d5219c7dfe42335fc4b617a/etcdserver/etcdserverpb/rpc.proto#L382

	if err != nil {
		return nil, err
	}
	return resp.Kvs, nil
}
