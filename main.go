package main

import (
	"embed"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/dongzhiwei-git/resume/handlers"
	"github.com/dongzhiwei-git/resume/metrics"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("recovered from panic: %v", r)
					time.Sleep(2 * time.Second)
				}
			}()

			router := gin.Default()
			router.Static("/static", "./static")
			tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
			router.SetHTMLTemplate(tmpl)

			router.GET("/", handlers.Home)
			router.GET("/editor", handlers.Editor)
			router.POST("/preview", handlers.Preview)
			router.POST("/api/preview", handlers.ApiPreview)
			router.POST("/import", handlers.Import)
			router.GET("/robots.txt", handlers.Robots)
			router.GET("/sitemap.xml", handlers.Sitemap)

			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}
			dsn := os.Getenv("MYSQL_DSN")
			if dsn != "" {
				ok := false
				for i := 0; i < 60; i++ {
					if db, err := metrics.SetupDB(dsn); err == nil && db != nil {
						metrics.Init(db)
						log.Printf("metrics persistence enabled")
						ok = true
						break
					} else if err != nil {
						log.Printf("metrics db setup retry %d: %v", i+1, err)
						time.Sleep(2 * time.Second)
					}
				}
				if !ok {
					panic("database not ready")
				}
			}
			if err := router.Run(":" + port); err != nil {
				log.Printf("server error: %v", err)
				time.Sleep(2 * time.Second)
			}
		}()
	}
}
