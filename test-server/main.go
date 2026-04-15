package main

import (
	"fmt"
	"log"

	"github.com/dpkrn/gotunnel/pkg/tunnel"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello world"})
	})

	router.POST("/name", func(c *gin.Context) {
		name := c.Query("name")
		c.JSON(200, gin.H{"message": "hello " + name})
	})

	url, stop, err := tunnel.StartTunnel("8080")
	if err != nil {
		log.Fatal(err)
	}
	defer stop()
	fmt.Println("Public URL:", url)
	router.Run(":8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
