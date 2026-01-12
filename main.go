package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dongzhiwei-git/resume/config"
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
			router.Use(func(c *gin.Context) {
				if c.Request.Method == "GET" {
					p := c.Request.URL.Path
					if !strings.HasPrefix(p, "/static") && !strings.HasPrefix(p, "/.well-known") && p != "/robots.txt" && p != "/sitemap.xml" && p != "/favicon.ico" && p != "/metrics/snapshot" {
						metrics.IncVisit()
					}
				}
			})
			router.Static("/static", "./static")
			tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
			router.SetHTMLTemplate(tmpl)

			router.GET("/", handlers.Home)
			router.GET("/editor", handlers.Editor)
			router.POST("/preview", handlers.Preview)
			router.POST("/api/preview", handlers.ApiPreview)
			router.GET("/ai", handlers.AiPage)
			router.POST("/api/ai/ask", handlers.ApiAiAsk)
			router.POST("/import", handlers.Import)
			router.GET("/robots.txt", handlers.Robots)
			router.GET("/sitemap.xml", handlers.Sitemap)
			router.POST("/metrics/generate", handlers.GenerateEvent)
			router.GET("/metrics/snapshot", handlers.SnapshotAPI)
			router.GET("/healthz", handlers.Health)

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
			srv := &http.Server{Addr: ":" + port, Handler: router}
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("server error: %v", err)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("server shutdown error: %v", err)
			}
			if v := os.Getenv("ENABLE_AI_ASSISTANT"); v != "" {
				config.AppConfig.EnableAIAssistant = v == "1" || strings.ToLower(v) == "true"
			}
		}()
	}
}
