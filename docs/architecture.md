# 架构文档

本项目采用简洁的 MVC 式结构：Gin 负责路由与模板渲染，模板负责前端展示与交互，实时预览通过轻量接口返回 HTML 片段。

## 总览
- Web 框架：`Gin`
- 模板引擎：`html/template`（Go 模板）
- 静态资源：`static/`（CSS/JS）
- 模板页面：`templates/`（HTML 模板）

## 关键模块
- 路由与控制器：`main.go`
  - `GET /` 首页
  - `GET /editor` 编辑器（分栏 + 实时预览）
  - `POST /api/preview` 返回简历主体片段，用于右侧实时预览
  - `POST /preview` 返回完整预览页，用于打印/导出

- 表单解析：`parseResumeFromForm(c *gin.Context)`
  - 直接解析 `PostForm` 字段，支持数组字段：`experience[0].title`、`education[0].school` 等
  - 解决旧版本环境下自动绑定不稳定的问题

- 模板分层：
  - `editor.html`：编辑表单 + 预览容器 + 动态添加/删除逻辑
  - `resume_content.html`：简历主体内容（作为片段可复用在完整预览与实时预览）
  - `view.html`：完整预览页（隐藏导航/页脚、优化打印样式）
  - `header.html` / `footer.html`：公共头尾

## 数据模型
- `Resume`
  - `Name`、`Email`、`Phone`、`Summary`
  - `Config`：`Template`、`Color`、`FontSize`、`PaperSize`
  - `Experience`：`[]Exp{Title, Company, Date, Description}`
  - `Education`：`[]Edu{Degree, School, Date}`

## 交互流程
1. 用户在编辑器填写表单（含动态增删条目）
2. 前端将表单通过 `POST /api/preview` 提交，后端解析并返回 `resume_content.html` 渲染片段
3. 右侧预览区域替换片段，实现实时更新
4. 用户点击“生成完整预览/打印”，通过 `POST /preview` 进入可打印的全屏页面

## 设计原则
- 简单可维护：模板与样式尽量直观，易于二次开发
- 兼容性优先：后端主动解析复杂表单结构，避免绑定差异
- 无状态服务：不持久化用户数据，隐私友好（可按需扩展存储）

## 扩展方向
- 增加主题模板与打印版式（如双栏、时间轴）
- 引入多语言与字体管理（Web Font）
- 增加导出为 DOCX/Markdown 的格式转换（可在后端使用转换库或在前端生成）
- 增加持久化与分享功能（数据库 + 链接分享）
