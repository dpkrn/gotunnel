package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dpkrn/gotunnel/pkg/tunnel"
	"github.com/gin-gonic/gin"
)

func main() {
	// Quiet Gin startup logs; avoid “trust all proxies” default in debug mode.
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello world"})
	})

	router.POST("/name", func(c *gin.Context) {
		name := c.Query("name")
		c.JSON(200, gin.H{"message": "hello " + name})
	})

	// POST JSON body: {"num1":1,"num2":2} → {"sum":3}
	router.POST("/sum", func(c *gin.Context) {
		var body struct {
			Num1 int `json:"num1"`
			Num2 int `json:"num2"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"sum": body.Num1 + body.Num2})
	})

	tunnelOptions := tunnel.TunnelOptions{
		Inspector:    true,
		InspectorAdd: "4040",
	}

	url, stop, err := tunnel.StartTunnel("8080", tunnelOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer stop()
	fmt.Println("Public URL:", url)
	router.Run(":8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
