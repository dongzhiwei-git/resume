package models

type Exp struct {
	Title       string `form:"title" json:"title"`
	Company     string `form:"company" json:"company"`
	Date        string `form:"date" json:"date"`
	Description string `form:"description" json:"description"`
}

type Edu struct {
	Degree string `form:"degree" json:"degree"`
	School string `form:"school" json:"school"`
	Date   string `form:"date" json:"date"`
}

type ThemeConfig struct {
	Template  string `form:"template" json:"template"`
	Color     string `form:"color" json:"color"`
	Font      string `form:"font" json:"font"`
	FontSize  string `form:"font_size" json:"font_size"`
	PaperSize string `form:"paper_size" json:"paper_size"`
}

type Resume struct {
	Name       string      `form:"name" json:"name"`
	Email      string      `form:"email" json:"email"`
	Phone      string      `form:"phone" json:"phone"`
	Avatar     string      `form:"-" json:"avatar"` // File path
	Summary    string      `form:"summary" json:"summary"`
	Experience []Exp       `form:"experience" json:"experience"`
	Education  []Edu       `form:"education" json:"education"`
	Config     ThemeConfig `form:"config" json:"config"`
}

func GetDemoResume() Resume {
	return Resume{
		Name:    "张三",
		Email:   "zhangsan@example.com",
		Phone:   "13800138000",
		Summary: "资深软件工程师，拥有5年全栈开发经验。擅长Go, Python, React等技术栈。致力于构建高性能、可扩展的Web应用。",
		Experience: []Exp{
			{
				Title:       "高级后端开发工程师",
				Company:     "科技创新有限公司",
				Date:        "2021 - 至今",
				Description: "负责核心业务系统的架构设计与开发，带领团队完成微服务改造，提升系统吞吐量300%。",
			},
			{
				Title:       "全栈开发工程师",
				Company:     "未来网络科技有限公司",
				Date:        "2018 - 2021",
				Description: "参与电商平台的前后端开发，主导支付模块的重构，实现了99.99%的系统可用性。",
			},
		},
		Education: []Edu{
			{
				Degree: "计算机科学与技术 学士",
				School: "某某大学",
				Date:   "2014 - 2018",
			},
		},
	}
}
