package main

import (
	//"fmt"
	"log" // log.Fatal()
	// "pacb.com/seq/paws/pkg/stuff"
	// "pacb.com/seq/paws/pkg/stiff"
	//"github.com/gofiber/fiber/v2"
	//_ "github.com/gofiber/fiber/v2/middleware/recover" // to trap panics
	//"github.com/gofiber/fiber/v2/utils"
	//"github.com/gofiber/template/html"
	"github.com/gin-gonic/gin"
	"pacb.com/seq/paws/pkg/web"
	"runtime" // only for GOOS
)

func main() {
	//router := gin.Default()
	// Or explicitly:
	router := gin.New()
	router.Use(
		//gin.Logger(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/pathsNotToLog/"), // useful!
		//gin.Recovery(),
	)

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})

	router.GET("/os", func(c *gin.Context) {
		c.String(200, runtime.GOOS)
	})

	web.AddRoutes(router)

	log.Fatal(router.Run(":5000")) // logger maybe not needed, but does not seem to hurt
}
