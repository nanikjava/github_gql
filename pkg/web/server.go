package web

import (
	"github.com/gin-gonic/gin"
	"github_gql/pkg/github"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func StartServer() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	// Route to serve the main template
	r.GET("/", func(c *gin.Context) {
		nextRevisionStr := c.Query("NextRevision")
		startStr := c.Query("Start")
		endStr := c.Query("End")

		step := c.Query("Step")

		if strings.ToLower(step) == "" {
			step = "reset"
		}

		nextRevision, err := strconv.ParseInt(nextRevisionStr, 10, 64)
		if err != nil {
			log.Println("Error parsing NextRevision")
			nextRevision = 0
		}
		start, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			log.Println("Error parsing start")
		}
		end, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			log.Println("Error parsing end")
		}

		data := github.GetGithubData(nextRevision, step, start, end)

		c.HTML(http.StatusOK, "main.tmpl", gin.H{
			"data": data,
		})
	})

	r.GET("/p/", func(c *gin.Context) {
		pageNumberStr := c.Query("pn")
		startStr := c.Query("start")
		endStr := c.Query("end")

		pageNumber, err := strconv.ParseInt(pageNumberStr, 10, 64)
		if err != nil {
			log.Println("Error parsing pageNumber")
		}
		start, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			log.Println("Error parsing start")
		}
		end, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			log.Println("Error parsing end")
		}
		data := github.GetPaginatedGithubData(pageNumber, start, end)

		c.HTML(http.StatusOK, "pagination.tmpl", gin.H{
			"data": data,
		})
	})

	log.Println("Listening on port 8080")
	r.Run(":8080")
}
