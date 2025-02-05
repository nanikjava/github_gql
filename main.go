package main

import (
	"github_gql/pkg/etcd"
	"github_gql/pkg/github"
	"github_gql/pkg/web"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var GithubToken string

func init() {
	t, exist := os.LookupEnv("GITHUB_TOKEN")

	if !exist {
		log.Fatalln("Error getting GITHUB_TOKEN")
		os.Exit(1)
	}

	GithubToken = t
}

func main() {
	embedEtcd := etcd.InitEtcd()

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:7777"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to connect to embedded embedEtcd:", err)
	}
	defer client.Close()

	startWebServer()
	startGithubProcess(client)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGKILL)

	for {
		<-sig
		exit(1, embedEtcd)
	}
}

func startGithubProcess(client *clientv3.Client) {
	go func() {
		for {
			log.Println("Processing Github.....")
			github.GithubGQL(client, GithubToken)
			log.Println("Waking up in 2mins.....")
			time.Sleep(2 * time.Minute)
		}
	}()
}

func startWebServer() {
	go func() {
		web.StartServer()
	}()
}

func exit(code int, embedEtcd *embed.Etcd) {
	log.Println("Finishing.....")

	if embedEtcd != nil {
		log.Println("Closing etcd.....")
		embedEtcd.Server.Stop()
	}
	os.Exit(code)
}
