package github

import "time"

const dataLimit = 25
const GithubURL = "https://api.github.com/graphql"
const EtcdURL = "127.0.0.1:7777"

// GithubQuery contains the GraphQL to query
// "All Go repository that have stars more than 100 and sorted by the updated date
const GithubQuery = `
	{
	 search(
		query: "language:Go stars:>50  sort:updated"
		type: REPOSITORY
		first: 100
	 ) {
		edges {
		  node {
			... on Repository {
			  nameWithOwner
			  description
			  url
			  updatedAt
			}
		  }
		}
		pageInfo {
		  endCursor
		  hasNextPage
		}
	 }
	}`

type GithubResponse struct {
	Search Search `json:"search"`
}

type Search struct {
	Edges    []Edge `json:"edges"`
	PageInfo `json:"pageInfo"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Node struct {
	Description   *string   `json:"description"`
	NameWithOwner string    `json:"nameWithOwner"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Url           string    `json:"url"`
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type GithubData struct {
	Key   string
	Value string
}

type Data struct {
	GithubData   []GithubData
	NextRevision int64
	Start        int64
	End          int64
	Next         bool
}

const PageDisplayed = 10

type Pagination struct {
	GithubData []GithubData
	PageNumber int64
	Start      int64
	End        int64
	Refetch    bool // to fetch the next X number of pages
}
