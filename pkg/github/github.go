package github

import (
	"context"
	"github_gql/pkg/etcd"
	"log"
	"strings"
	"time"

	"github.com/machinebox/graphql"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func x(s string) {
	log.Println(s)
}

// GithubGQL to query Github using GraphqQL
func GithubGQL(client *clientv3.Client, token string) {
	graphClient := graphql.NewClient(GithubURL)

	graphClient.Log = x

	req := graphql.NewRequest(GithubQuery)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	ctx, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	var resp *GithubResponse

	if err := graphClient.Run(ctx, req, &resp); err != nil {
		log.Println(err)
		return
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

func GetPaginatedGithubData(pageNumber int64, start int64, end int64) []*Pagination {
	d := []*Pagination{}

	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{EtcdURL},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to connect to embedded embedEtcd:", err)
	}
	defer c.Close()
	var resp *clientv3.GetResponse

	resp, err = c.Get(context.TODO(), "",
		//clientv3.WithLimit(dataLimit*PageDisplayed),
		clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
		clientv3.WithPrefix())

	if err != nil {
		log.Fatal("Failed to read data:", err)
	}
	defer c.Close()

	// got a lot of data so we need to iterate in batches
	totalPage := int64(len(resp.Kvs)) / int64(dataLimit)
	var startIdx int64 = 0
	//loop through the totalPage calculated
	var l int64 = 0
	for ; l < totalPage; l++ {
		start := resp.Kvs[startIdx].CreateRevision

		var endIdx int64 = 0
		if int64(dataLimit)*(l+1) >= int64(len(resp.Kvs)) {
			endIdx = int64(dataLimit)*(l+1) - 1
		} else {
			endIdx = int64(dataLimit) * (l + 1)
		}
		end := resp.Kvs[endIdx].CreateRevision
		elem := &Pagination{
			GithubData: nil,
			PageNumber: l,
			Start:      start,
			End:        end,
			Refetch:    false,
		}
		d = append(d, elem)

		if l == pageNumber {
			max := int64(int64(dataLimit) * (l + 1))
			for r := startIdx; r < max; r++ {
				elem.GithubData = append(elem.GithubData, GithubData{
					Key:   string(resp.Kvs[r].Key),
					Value: string(resp.Kvs[r].Value),
				})
			}
		}
		startIdx = dataLimit * (l + 1)
	}

	return d
}

// GetGithubData extract data from etcd using nextRevision as parameter
// for the next sets of data
func GetGithubData(origRevision int64, step string, start int64, end int64) *Data {
	d := &Data{}
	var revision = origRevision
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
	var resp *clientv3.GetResponse

	if strings.ToLower(step) == "next" || strings.ToLower(step) == "reset" {
		resp, err = c.Get(context.TODO(), "",
			clientv3.WithLimit(dataLimit),
			clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
			clientv3.WithMinCreateRev(revision),
			clientv3.WithPrefix())
		if err != nil {
			log.Fatal("Failed to read data:", err)
		}
	} else if strings.ToLower(step) == "prev" {
		resp, err = c.Get(context.TODO(), "",
			clientv3.WithLimit(dataLimit),
			clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
			clientv3.WithMinCreateRev(start),
			clientv3.WithMaxCreateRev(end),
			//clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortDescend),
			//clientv3.WithMaxCreateRev(revision),
			clientv3.WithPrefix())
		if err != nil {
			log.Fatal("Failed to read data:", err)
		}
	}

	for _, kv := range resp.Kvs {
		d.GithubData = append(d.GithubData, GithubData{string(kv.Key), string(kv.Value)})
		revision = kv.CreateRevision
	}

	//revision need to be added to the next number
	//to avoid duplication of the same data being queried again
	d.NextRevision = revision

	//More gives indication whether there will be more data available to
	//be read
	d.Next = true //resp.More

	if strings.ToLower(step) == "reset" {
		d.Start = resp.Kvs[0].CreateRevision
		d.End = resp.Kvs[len(resp.Kvs)-1].CreateRevision
	} else if strings.ToLower(step) == "next" && (origRevision == end) {
		d.Start = start
		d.End = end
	} else if strings.ToLower(step) == "prev" {
		d.End = start
	} else {
		d.Start = end
		d.End = origRevision
	}

	return d
}
