package main

import (
	"github.com/dongzhiwei-git/resume/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", handlers.Home)
	r.GET("/editor", handlers.Editor)
	r.POST("/preview", handlers.Preview)
	r.POST("/api/preview", handlers.ApiPreview)
	r.POST("/import", handlers.Import)

	r.Run(":8080")
}
