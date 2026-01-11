# 免费部署到中国可访问区域

推荐使用 Zeabur（免费计划，支持亚洲区域，Docker 部署）或 Railway（免费试用，部分地区访问较快）。本项目已提供 `Dockerfile`，可直接导入部署。

## Zeabur 部署步骤
- 注册并登录 Zeabur，选择 “Create Project”
- 选择 “Import from GitHub”，关联仓库 `github.com/dongzhiwei-git/resume`
- 服务类型选择 “Service”，构建方式选择 “Dockerfile”
- 环境变量：
  - `GIN_MODE=release`
  - `PORT=8080`
- 区域选择亚洲（如 Hong Kong 或 Singapore），提升中国大陆访问速度
- 保存并部署，完成后访问分配的域名

## Railway 部署步骤
- 登录 Railway，创建新项目
- 选择从 GitHub 导入仓库 `github.com/dongzhiwei-git/resume`
- 使用 Docker 部署或 Go Build（推荐 Docker）
- 设置环境变量：`GIN_MODE=release`、`PORT=8080`
- 运行并查看分配的域名

## 注意事项
- 平台可能默认使用端口 `PORT` 变量，本项目已在 `main.go` 中读取 `PORT`，默认 `8080`
- 中国大陆访问速度与平台区域有关，建议选择香港/新加坡区域
- 如需自定义域名，平台支持绑定并提供基础证书
