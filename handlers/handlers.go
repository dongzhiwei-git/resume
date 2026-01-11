package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/dongzhiwei-git/resume/config"
	"github.com/dongzhiwei-git/resume/models"

	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":        "简单简历 - 在线简历制作",
		"ServerConfig": config.AppConfig,
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

	c.HTML(http.StatusOK, "editor.html", gin.H{
		"title":  "编辑简历",
		"Resume": initialResume,
	})
}

func Preview(c *gin.Context) {
	c.Request.ParseMultipartForm(32 << 20) // Parse form data
	resume := parseResumeFromForm(c)

	if resume.Config.Color == "" {
		resume.Config.Color = "#333333"
	}
	if resume.Config.Template == "" {
		resume.Config.Template = "classic"
	}

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
	})
}

func ApiPreview(c *gin.Context) {
	c.MultipartForm() // Ensure form is parsed

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
