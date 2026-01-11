# ResumeToJob (Go 版本)

作者：ricardo

Language: 中文 | [English](README.en.md)

一个基于 Go + Gin + 原生 HTML/CSS 的简历生成器，支持中文界面、实时预览、模板样式、主题颜色、字体大小与文档尺寸等配置，并可打印/导出为 PDF。

在线预览：[https://ricardo-resume.zeabur.app/](https://ricardo-resume.zeabur.app/)

GitHub 仓库：[https://github.com/dongzhiwei-git/resume](https://github.com/dongzhiwei-git/resume)

## 功能特性
- 实时预览：编辑器左侧输入，右侧即时渲染
- 多条记录：工作经历、教育背景支持动态添加/删除
- 样式配置：模板风格（经典/现代/极简）、主题颜色、字体大小、文档尺寸（A4/Letter）
- 打印导出：预览页一键打印或保存为 PDF
- 纯前端模板：简单易改，易于自定义主题与布局

## 目录结构
- `main.go`：后端入口与路由、表单解析、模板渲染
- `templates/`：页面模板
  - `index.html`：首页
  - `editor.html`：编辑器（分栏 + 实时预览）
  - `view.html`：完整预览（用于打印）
  - `resume_content.html`：简历主体片段（供实时预览接口返回）
  - `header.html`、`footer.html`：公共头尾
- `static/css/style.css`：基础样式
- `go.mod`：依赖与版本

## 快速开始
- 环境：建议 Go 1.19 及以上
- 安装依赖：
  - `go mod tidy`
- 启动：
  - `go run main.go`
- 访问：
  - 打开浏览器访问 `http://localhost:8080`
  - 进入编辑器 `http://localhost:8080/editor`

## 使用说明
- 左侧表单填写个人信息、工作经历、教育背景，可通过“添加经历/添加教育”动态增加条目，右侧预览会自动更新
- 点击“生成完整预览 / 打印”进入全屏预览页，使用浏览器的打印对话框保存为 PDF

## API 接口
- `POST /api/preview`
  - 功能：根据表单数据返回简历主体 HTML 片段（`resume_content.html`）
  - 适用：编辑器右侧实时预览
- `POST /preview`
  - 功能：根据表单数据返回完整预览页面（用于打印）

## 表单字段约定
- 基本信息：
  - `name`、`email`、`phone`、`summary`
- 样式配置（点号形式）：
  - `config.template`、`config.color`、`config.font_size`、`config.paper_size`
- 工作经历（数组点号形式）：
  - `experience[0].title`、`experience[0].company`、`experience[0].date`、`experience[0].description`
  - 索引从 0 递增，删除或新增后会自动重排索引
- 教育背景（数组点号形式）：
  - `education[0].degree`、`education[0].school`、`education[0].date`

> 说明：为兼容旧版本 Go/Gin 的表单绑定差异，项目在后端实现了对 `PostForm` 的稳健解析，确保复杂数组字段能正确映射到数据结构。

## 自定义与扩展
- 模板扩展：
  - 在 `resume_content.html` 中调整结构与样式；可通过自定义 CSS 变量 `--theme-color`、模板类名 `template-xxx` 和字体大小类名 `font-size-xxx` 自由扩展
- 新模板：
  - 在编辑器增加新的 `config.template` 选项，并在 `resume_content.html` 中添加对应的样式块
- 字体与国际化：
  - 可在 `style.css` 中引入 Web Font；模板使用中文默认安全字体，避免乱码

## 常见问题
- 预览未显示工作/教育：
  - 确认字段名为点号形式（如 `experience[0].title`）；项目已内置索引重排与后端解析
- 依赖安装慢或失败：
  - 可配置 `GOPROXY`（例如：`go env -w GOPROXY=https://goproxy.cn,direct`）后再执行 `go mod tidy`

## 许可
- 本项目采用非商业许可，需署名为“ricardo”，禁止任何商业化使用。
