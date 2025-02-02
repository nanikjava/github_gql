package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"log"
	"net/url"
	"time"
)

const localEtcd = "127.0.0.1:7777"
const localEtcdScheme = "http"
const localEtcdDir = "etcd_data"

func InitEtcd() *embed.Etcd {
	cfg := embed.NewConfig()
	cfg.Dir = localEtcdDir
	cfg.ListenPeerUrls = []url.URL{}
	cfg.ListenClientUrls = []url.URL{{
		Scheme: localEtcdScheme,
		Host:   localEtcd,
	}}

	// Start embedded embedEtcd
	embedEtcd, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatal("Failed to start embedded embedEtcd:", err)
	}

	// Wait until embedEtcd is ready
	select {
	case <-embedEtcd.Server.ReadyNotify():
		log.Println("Embedded embedEtcd is ready!")
	case <-time.After(10 * time.Second):
		embedEtcd.Server.Stop() // Timeout safety
		log.Fatal("Embedded embedEtcd took too long to start")
	}

	return embedEtcd
}

func Get(client *clientv3.Client, key string) *clientv3.GetResponse {
	resp, err := client.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		log.Println("Failed to read data: ", err)
	}
	return resp
}

func Put(client *clientv3.Client, key string, value string) (context.Context, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := client.Put(ctx, key, value)

	if err != nil {
		log.Println("Failed to write data: ", err)
	}
	return ctx, err
}
