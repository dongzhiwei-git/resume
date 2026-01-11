package main

import (
	"embed"
	"html/template"
	"os"

	"github.com/dongzhiwei-git/resume/handlers"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	r := gin.Default()

	r.Static("/static", "./static")
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", handlers.Home)
	r.GET("/editor", handlers.Editor)
	r.POST("/preview", handlers.Preview)
	r.POST("/api/preview", handlers.ApiPreview)
	r.POST("/import", handlers.Import)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}
