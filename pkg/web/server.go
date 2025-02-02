package web

import (
	"github.com/gin-gonic/gin"
	"github_gql/pkg/github"
	"log"
	"net/http"
	"strconv"
)

func StartServer() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	// Route to serve the main template
	r.GET("/", func(c *gin.Context) {
		nextRevisionStr := c.Query("NextRevision")

		nextRevision, err := strconv.ParseInt(nextRevisionStr, 10, 64)
		if err != nil {
			log.Println("Error parsing NextRevision")
			nextRevision = 0
		}

		data := github.GetGithubData(nextRevision)

		c.HTML(http.StatusOK, "main.tmpl", gin.H{
			"data": data,
		})
	})

	log.Println("Listening on port 8080")
	r.Run(":8080")
}
