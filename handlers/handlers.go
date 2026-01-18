package handlers

import (
	"bufio"
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
		"title":        "编辑简历",
		"Resume":       initialResume,
		"Visits":       v,
		"Generates":    g,
		"Canonical":    canonical,
		"ServerConfig": config.AppConfig,
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
		"Visits":       v,
		"Generates":    g,
		"Canonical":    canonical,
		"ServerConfig": config.AppConfig,
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

func AiPage(c *gin.Context) {
	if !config.AppConfig.EnableAIAssistant {
		c.String(http.StatusNotFound, "Not enabled")
		return
	}
	v, g := metrics.Snapshot()
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	canonical := scheme + "://" + c.Request.Host + c.Request.URL.Path
	c.HTML(http.StatusOK, "ai.html", gin.H{
		"title":        "AI 简历助手",
		"Visits":       v,
		"Generates":    g,
		"Canonical":    canonical,
		"ServerConfig": config.AppConfig,
	})
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiAskReq struct {
	Messages []chatMessage `json:"messages"`
}

func ApiAiAsk(c *gin.Context) {
	if !config.AppConfig.EnableAIAssistant {
		c.String(http.StatusForbidden, "Disabled")
		return
	}
	apiURL := os.Getenv("DEEPSEEK_API_URL")
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/v1/chat/completions"
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		c.String(http.StatusBadRequest, "Missing API key")
		return
	}
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}
	body := aiAskReq{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON")
		return
	}
	promptBytes, _ := os.ReadFile("docs/prompts/deepseek_resume_prompt.md")
	sys := chatMessage{Role: "system", Content: string(promptBytes)}
	msgs := append([]chatMessage{sys}, body.Messages...)
	payload := map[string]any{"model": model, "messages": msgs}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.String(http.StatusBadGateway, "AI unavailable")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.String(http.StatusBadGateway, "AI error")
		return
	}
	var out struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		c.String(http.StatusBadGateway, "Bad AI response")
		return
	}
	if len(out.Choices) == 0 {
		c.String(http.StatusBadGateway, "No answer")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": out.Choices[0].Message})
}

func ApiAiStream(c *gin.Context) {
	if !config.AppConfig.EnableAIAssistant {
		c.String(http.StatusForbidden, "Disabled")
		return
	}
	apiURL := os.Getenv("DEEPSEEK_API_URL")
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/v1/chat/completions"
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		c.String(http.StatusBadRequest, "Missing API key")
		return
	}
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}
	body := aiAskReq{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON")
		return
	}
	promptBytes, _ := os.ReadFile("docs/prompts/deepseek_resume_prompt.md")
	sys := chatMessage{Role: "system", Content: string(promptBytes)}
	msgs := append([]chatMessage{sys}, body.Messages...)
	payload := map[string]any{"model": model, "messages": msgs, "stream": true}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.String(http.StatusBadGateway, "AI unavailable")
		return
	}
	defer resp.Body.Close()
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	fl, ok := c.Writer.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "Streaming unsupported")
		return
	}
	r := bufio.NewReader(resp.Body)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			c.Writer.Write(line)
			fl.Flush()
		}
		if err != nil {
			break
		}
	}
}

type simpleGenReq struct {
	Input string `json:"input"`
}

func simpleFromInput(s string) models.Resume {
	r := models.Resume{Summary: strings.TrimSpace(s)}
	// try extract email
	if m := regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`).FindString(s); m != "" {
		r.Email = m
	}
	// try extract phone (11 digits)
	if m := regexp.MustCompile(`\b1[3-9][0-9]{9}\b`).FindString(s); m != "" {
		r.Phone = m
	}
	// set one experience if keywords present
	title := "工程师"
	if strings.Contains(s, "后端") {
		title = "后端工程师"
	}
	if strings.Contains(s, "前端") {
		title = "前端工程师"
	}
	if strings.Contains(s, "产品") {
		title = "产品经理"
	}
	r.Experience = []models.Exp{{Title: title, Company: "", Date: time.Now().Format("2006-01"), Description: s}}
	r.Education = []models.Edu{}
	r.Config = models.ThemeConfig{Template: "classic", Color: "#333333", PaperSize: "a4"}
	return r
}

func ApiAiGenerateSimple(c *gin.Context) {
	if !config.AppConfig.EnableAIAssistant {
		c.String(http.StatusForbidden, "Disabled")
		return
	}
	apiURL := os.Getenv("DEEPSEEK_API_URL")
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/v1/chat/completions"
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		c.String(http.StatusBadRequest, "Missing API key")
		return
	}
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	reqBody := simpleGenReq{}
	if err := c.ShouldBindJSON(&reqBody); err != nil || strings.TrimSpace(reqBody.Input) == "" {
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	schema := `仅输出一个严格的 JSON 对象，键名与结构如下（全部小写）：
{"name":"","email":"","phone":"","summary":"","avatar":"","config":{"template":"classic","color":"#333333","font":"","font_size":"","paper_size":"a4"},"experience":[{"title":"","company":"","date":"YYYY-MM","description":""}],"education":[{"degree":"","school":"","date":"YYYY-MM"}]}`
	sys := chatMessage{Role: "system", Content: "你是简历生成助手。请在不臆造个人信息的前提下，尽量饱满地填充内容：总结要简洁全面，经验描述采用 3–5 条要点以中文分号分隔，包含动作、方法、数据结果。严格生成符合站点 schema 的 JSON，键名小写，只返回 JSON。"}
	user := chatMessage{Role: "user", Content: "输入：" + reqBody.Input + "\n要求：" + schema + "\n只返回 JSON，不要任何解释。未知值留空字符串或空数组。"}
	payload := map[string]any{"model": model, "messages": []chatMessage{sys, user}}
	b, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(b))
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		r := simpleFromInput(reqBody.Input)
		c.JSON(http.StatusOK, r)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		r := simpleFromInput(reqBody.Input)
		c.JSON(http.StatusOK, r)
		return
	}
	var out struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		r := simpleFromInput(reqBody.Input)
		c.JSON(http.StatusOK, r)
		return
	}
	if len(out.Choices) == 0 {
		r := simpleFromInput(reqBody.Input)
		c.JSON(http.StatusOK, r)
		return
	}
	var r models.Resume
	content := out.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), &r); err != nil {
		i := strings.Index(content, "{")
		j := strings.LastIndex(content, "}")
		if i >= 0 && j > i {
			if err2 := json.Unmarshal([]byte(content[i:j+1]), &r); err2 != nil {
				r = simpleFromInput(reqBody.Input)
			}
		} else {
			r = simpleFromInput(reqBody.Input)
		}
	}
	if r.Config.Color == "" {
		r.Config.Color = "#333333"
	}
	if r.Config.Template == "" {
		r.Config.Template = "classic"
	}
	c.JSON(http.StatusOK, r)
}

type reviseReq struct {
	Instruction string        `json:"instruction"`
	Resume      models.Resume `json:"resume"`
}

func ApiAiRevise(c *gin.Context) {
	if !config.AppConfig.EnableAIAssistant {
		c.String(http.StatusForbidden, "Disabled")
		return
	}
	apiURL := os.Getenv("DEEPSEEK_API_URL")
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/v1/chat/completions"
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		c.String(http.StatusBadRequest, "Missing API key")
		return
	}
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	var reqBody reviseReq
	if err := c.ShouldBindJSON(&reqBody); err != nil || strings.TrimSpace(reqBody.Instruction) == "" {
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	schema := `仅输出一个严格的 JSON 对象，键名与结构如下（全部小写）：
{"name":"","email":"","phone":"","summary":"","avatar":"","config":{"template":"classic","color":"#333333","font":"","font_size":"","paper_size":"a4"},"experience":[{"title":"","company":"","date":"YYYY-MM","description":""}],"education":[{"degree":"","school":"","date":"YYYY-MM"}]}`
	sys := chatMessage{Role: "system", Content: "你是简历生成助手。根据现有 JSON 简历与用户修改要求，更新并优化简历：保持事实，不臆造；经验描述以 3–5 条要点的中文分号分隔补充动作、方法、数据。严格返回符合站点 schema 的 JSON，键名小写，只返回 JSON。"}
	oldJSON, _ := json.Marshal(reqBody.Resume)
	user := chatMessage{Role: "user", Content: "现有简历：" + string(oldJSON) + "\n修改要求：" + reqBody.Instruction + "\n输出：" + schema + "\n只返回 JSON，不要解释。"}
	payload := map[string]any{"model": model, "messages": []chatMessage{sys, user}}
	b, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(b))
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		r := reqBody.Resume
		c.JSON(http.StatusOK, r)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		r := reqBody.Resume
		c.JSON(http.StatusOK, r)
		return
	}
	var out struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		c.JSON(http.StatusOK, reqBody.Resume)
		return
	}
	if len(out.Choices) == 0 {
		c.JSON(http.StatusOK, reqBody.Resume)
		return
	}
	content := out.Choices[0].Message.Content
	var r models.Resume
	if err := json.Unmarshal([]byte(content), &r); err != nil {
		i := strings.Index(content, "{")
		j := strings.LastIndex(content, "}")
		if i >= 0 && j > i {
			if err2 := json.Unmarshal([]byte(content[i:j+1]), &r); err2 != nil {
				r = reqBody.Resume
			}
		} else {
			r = reqBody.Resume
		}
	}
	if r.Config.Color == "" {
		r.Config.Color = "#333333"
	}
	if r.Config.Template == "" {
		r.Config.Template = "classic"
	}
	if r.Config.PaperSize == "" {
		r.Config.PaperSize = "a4"
	}
	c.JSON(http.StatusOK, r)
}

func ApiPreviewJSON(c *gin.Context) {
	var resume models.Resume
	if err := c.ShouldBindJSON(&resume); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON")
		return
	}
	if resume.Config.Color == "" {
		resume.Config.Color = "#333333"
	}
	if resume.Config.Template == "" {
		resume.Config.Template = "classic"
	}
	c.HTML(http.StatusOK, "resume_content.html", gin.H{"Resume": resume})
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
