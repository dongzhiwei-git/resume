package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dongzhiwei-git/resume/config"
	"github.com/dongzhiwei-git/resume/metrics"
	"github.com/dongzhiwei-git/resume/models"

	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
	v, g := metrics.Snapshot()
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	canonical := scheme + "://" + c.Request.Host + c.Request.URL.Path
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":        "简单简历 - 在线简历制作",
		"ServerConfig": config.AppConfig,
		"Visits":       v,
		"Generates":    g,
		"Canonical":    canonical,
	})
}

func Editor(c *gin.Context) {
	selectedTemplate := c.Query("template")

	var initialResume models.Resume
	if selectedTemplate != "" {
		initialResume = models.GetDemoResume()
		initialResume.Config.Template = selectedTemplate
	} else {
		// Empty resume, but maybe set default config?
		initialResume = models.Resume{}
		initialResume.Config.Template = "" // Default
	}

	v, g := metrics.Snapshot()
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	canonical := scheme + "://" + c.Request.Host + c.Request.URL.Path
	c.HTML(http.StatusOK, "editor.html", gin.H{
		"title":     "编辑简历",
		"Resume":    initialResume,
		"Visits":    v,
		"Generates": g,
		"Canonical": canonical,
	})
}

func Preview(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.String(http.StatusBadRequest, "Invalid form")
		return
	}
	resume := parseResumeFromForm(c)

	if resume.Config.Color == "" {
		resume.Config.Color = "#333333"
	}
	if resume.Config.Template == "" {
		resume.Config.Template = "classic"
	}

	v, g := metrics.Snapshot()
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	canonical := scheme + "://" + c.Request.Host + "/view"
	c.HTML(http.StatusOK, "view.html", gin.H{
		"title":  "简历预览",
		"Resume": resume,
		"ResumeJSON": func() string {
			b, err := json.Marshal(resume)
			if err != nil {
				return "{}"
			}
			return string(b)
		}(),
		"Visits":    v,
		"Generates": g,
		"Canonical": canonical,
	})
}

func ApiPreview(c *gin.Context) {
	if _, err := c.MultipartForm(); err != nil {
		c.String(http.StatusBadRequest, "Invalid form")
		return
	}

	resume := parseResumeFromForm(c)
	fmt.Printf("Received Resume: %+v\n", resume)

	if resume.Config.Color == "" {
		resume.Config.Color = "#333333"
	}
	if resume.Config.Template == "" {
		resume.Config.Template = "classic"
	}

	c.HTML(http.StatusOK, "resume_content.html", gin.H{
		"Resume": resume,
	})
}

func GenerateEvent(c *gin.Context) {
	metrics.IncGenerate()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func SnapshotAPI(c *gin.Context) {
	v, g := metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{"visits": v, "generates": g})
}

func Health(c *gin.Context) {
	ready := metrics.Ready()
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": ready})
}

func DownloadPDF(c *gin.Context) {
	apiURL := os.Getenv("PDF_API_URL")
	apiKey := os.Getenv("PDF_API_KEY")
	if apiURL == "" || apiKey == "" {
		c.String(http.StatusBadRequest, "PDF service not configured")
		return
	}

	var resume models.Resume
	ct := c.GetHeader("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		if err := c.ShouldBindJSON(&resume); err != nil {
			c.String(http.StatusBadRequest, "Invalid JSON")
			return
		}
	} else {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.String(http.StatusBadRequest, "Invalid form")
			return
		}
		resume = parseResumeFromForm(c)
	}

	if resume.Config.Color == "" {
		resume.Config.Color = "#333333"
	}
	if resume.Config.Template == "" {
		resume.Config.Template = "classic"
	}
	if resume.Config.PaperSize == "" {
		resume.Config.PaperSize = "a4"
	}

	tpl, err := template.ParseFiles("templates/resume_content.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error")
		return
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, gin.H{"Resume": resume}); err != nil {
		c.String(http.StatusInternalServerError, "Render error")
		return
	}
	cssBytes, _ := os.ReadFile("static/css/style.css")
	html := "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><style>" + string(cssBytes) + "</style></head><body>" + buf.String() + "</body></html>"

	payload := map[string]any{
		"html": html,
		"options": map[string]any{
			"printBackground": true,
			"format":          strings.ToUpper(resume.Config.PaperSize),
			"margin":          map[string]string{"top": "0.5in", "bottom": "0.5in", "left": "0.5in", "right": "0.5in"},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.String(http.StatusBadGateway, "PDF service unavailable")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.String(http.StatusBadGateway, "PDF generation failed")
		return
	}
	metrics.IncGenerate()
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=resume.pdf")
	io.Copy(c.Writer, resp.Body)
}

func Robots(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	host := c.Request.Host
	body := "User-agent: *\nAllow: /\nSitemap: " + scheme + "://" + host + "/sitemap.xml\n"
	c.String(http.StatusOK, body)
}

func Sitemap(c *gin.Context) {
	c.Header("Content-Type", "application/xml; charset=utf-8")
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	host := c.Request.Host
	today := time.Now().Format("2006-01-02")
	b := strings.Builder{}
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	for _, p := range []string{"/", "/editor"} {
		b.WriteString("  <url>\n")
		b.WriteString("    <loc>" + scheme + "://" + host + p + "</loc>\n")
		b.WriteString("    <lastmod>" + today + "</lastmod>\n")
		b.WriteString("    <changefreq>weekly</changefreq>\n")
		b.WriteString("    <priority>0.8</priority>\n")
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>")
	c.String(http.StatusOK, b.String())
}

func Import(c *gin.Context) {
	if !config.AppConfig.EnableImport {
		c.String(http.StatusForbidden, "Import feature is disabled")
		return
	}

	file, err := c.FormFile("resume_json")
	if err != nil {
		c.String(http.StatusBadRequest, "Upload failed")
		return
	}

	// Read file content
	f, err := file.Open()
	if err != nil {
		c.String(http.StatusBadRequest, "Open file failed")
		return
	}
	defer f.Close()

	// Decode JSON
	var resume models.Resume
	if err := json.NewDecoder(f).Decode(&resume); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON: %v", err)
		return
	}

	// Render editor with data
	c.HTML(http.StatusOK, "editor.html", gin.H{
		"title":  "编辑简历",
		"Resume": resume,
	})
}

func parseResumeFromForm(c *gin.Context) models.Resume {
	r := models.Resume{}
	r.Name = c.PostForm("name")
	r.Email = c.PostForm("email")
	r.Phone = c.PostForm("phone")
	r.Summary = c.PostForm("summary")

	r.Config.Template = c.PostForm("config.template")
	r.Config.Color = c.PostForm("config.color")
	r.Config.Font = c.PostForm("config.font")
	r.Config.FontSize = c.PostForm("config.font_size")
	r.Config.PaperSize = c.PostForm("config.paper_size")

	// Handle File Upload
	file, err := c.FormFile("avatar")
	if err == nil {
		// Generate a unique filename or keep original if simple
		dst := "static/uploads/" + file.Filename
		// Ensure unique name to prevent collisions in real app, but for now:
		if err := c.SaveUploadedFile(file, dst); err == nil {
			r.Avatar = "/" + dst
		} else {
			fmt.Println("Save file error:", err)
		}
	} else if c.PostForm("avatar_existing") != "" {
		// Keep existing avatar if not re-uploaded
		r.Avatar = c.PostForm("avatar_existing")
	}

	// Parse Experience and Education using regex
	expMap := map[int]*models.Exp{}
	eduMap := map[int]*models.Edu{}

	expRe := regexp.MustCompile(`^experience\[(\d+)\]\.(title|company|date|description)$`)
	eduRe := regexp.MustCompile(`^education\[(\d+)\]\.(degree|school|date)$`)

	for key, vals := range c.Request.PostForm {
		if len(vals) == 0 {
			continue
		}
		val := vals[0]

		if m := expRe.FindStringSubmatch(key); len(m) == 3 {
			idx, _ := strconv.Atoi(m[1])
			field := m[2]
			e := expMap[idx]
			if e == nil {
				e = &models.Exp{}
				expMap[idx] = e
			}
			switch field {
			case "title":
				e.Title = val
			case "company":
				e.Company = val
			case "date":
				e.Date = val
			case "description":
				e.Description = val
			}
		} else if m := eduRe.FindStringSubmatch(key); len(m) == 3 {
			idx, _ := strconv.Atoi(m[1])
			field := m[2]
			e := eduMap[idx]
			if e == nil {
				e = &models.Edu{}
				eduMap[idx] = e
			}
			switch field {
			case "degree":
				e.Degree = val
			case "school":
				e.School = val
			case "date":
				e.Date = val
			}
		}
	}

	// Convert maps to slices, sorted by index
	if len(expMap) > 0 {
		idxs := make([]int, 0, len(expMap))
		for i := range expMap {
			idxs = append(idxs, i)
		}
		sort.Ints(idxs)
		for _, i := range idxs {
			r.Experience = append(r.Experience, *expMap[i])
		}
	}

	if len(eduMap) > 0 {
		idxs := make([]int, 0, len(eduMap))
		for i := range eduMap {
			idxs = append(idxs, i)
		}
		sort.Ints(idxs)
		for _, i := range idxs {
			r.Education = append(r.Education, *eduMap[i])
		}
	}

	return r
}
