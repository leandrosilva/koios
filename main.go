package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		log.Println("Environment variable ELASTICSEARCH_URL is not defined. Using default URL 'http://localhost:9200'.")
		esURL = "http://localhost:9200"
	}

	router := gin.Default()
	router.Static("/k/", "./public/")

	router.POST("/index/create", func(c *gin.Context) {
		content, err := createMovieIndex(c, esURL, "netflix_titles.csv")
		if err != nil {
			log.Printf("Got error: %s", err)
			c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]string{"message": content})
	})

	router.POST("/index/destroy", func(c *gin.Context) {
		content, err := destroyMovieIndex(c, esURL)
		if err != nil {
			log.Printf("Got error: %s", err)
			c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]string{"message": content})
	})

	router.GET("/search", func(c *gin.Context) {
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusOK, map[string]string{"count": "0"})
			return
		}
		content, err := searchMovieIndex(c, esURL, q)
		if err != nil {
			log.Printf("Got error: %s", err)
			c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	router.POST("/autocomplete", func(c *gin.Context) {
		q := c.PostForm("q")
		if q == "" {
			c.JSON(http.StatusOK, map[string]string{"count": "0"})
			return
		}
		content, err := searchMovieIndex(c, esURL, q)
		if err != nil {
			log.Printf("Got error: %s", err)
			c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, content)
	})

	router.Run(":6660")
}
