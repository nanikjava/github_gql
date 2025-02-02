package github

import (
	"context"
	"github.com/machinebox/graphql"
	"github_gql/pkg/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

// GithubGQL to query Github using GraphqQL
func GithubGQL(client *clientv3.Client, token string) {
	graphClient := graphql.NewClient(GithubURL)

	req := graphql.NewRequest(GithubQuery)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	ctx, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	var resp *GithubResponse

	if err := graphClient.Run(ctx, req, &resp); err != nil {
		log.Fatal(err)
	}

	ctrPut := 0
	if resp != nil {
		//loop through the result (Edges)
		for _, e := range resp.Search.Edges {
			r := etcd.Get(client, e.Node.Url)

			//store the repo information if does not exist
			if r.Count <= 0 {
				var desc = ""

				//...sometimes Description is empty check first
				if e.Node.Description != nil {
					desc = *e.Node.Description
				}
				etcd.Put(client, e.Node.Url, desc)
				ctrPut++
			}
		}
	}
	log.Println("New repo stored: ", ctrPut)
}

// GetGithubData extract data from etcd using nextRevision as parameter
// for the next sets of data
func GetGithubData(nextRevision int64) *Data {
	d := &Data{}
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{EtcdURL},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to connect to embedded embedEtcd:", err)
	}
	defer c.Close()

	//Extracting data from etcd. The way it works is to get data and limit it
	//by the dataLimit parameter. Data are sorted by key and in ascending order.
	//WithMinModRev is to indicate which revision number we want to get the data starting
	//from, this is used as a pagination method
	resp, err := c.Get(context.TODO(), "",
		clientv3.WithLimit(dataLimit),
		clientv3.WithSort(clientv3.SortByCreateRevision|clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithMinModRev(nextRevision),
		clientv3.WithPrefix())
	if err != nil {
		log.Fatal("Failed to read data:", err)
	}

	for _, kv := range resp.Kvs {
		d.GithubData = append(d.GithubData, GithubData{string(kv.Key), string(kv.Value)})
		nextRevision = kv.CreateRevision
	}

	//revision need to be added to the next number
	//to avoid duplication of the same data being queried again
	nextRevision++

	d.NextRevision = nextRevision

	//More gives indication whether there will be more data available to
	//be read
	d.Next = resp.More
	return d
}
