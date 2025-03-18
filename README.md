# LinkHUB - 链接分享平台

## 项目简介

LinkHUB 是一个基于 Golang 开发的链接分享平台，用户可以在平台上分享、收藏和讨论有价值的链接资源。平台支持文章发布、评论互动、标签分类等功能，致力于为用户提供一个高效的网站分享和交流社区。

## 主要功能

- 用户系统
  - 用户注册和登录
  - 个人资料管理
  - 用户认证和授权

- 链接管理
  - 链接分享发布
  - 链接分类标签
  - 链接搜索筛选

- 文章系统
  - 文章发布和编辑
  - 文章分类管理
  - 文章搜索功能

- 互动功能
  - 评论系统
  - 点赞收藏

## 技术栈

- 后端框架：Gin
- 数据库：PostgreSQL
- ORM框架：GORM
- 配置管理：Viper
- 模板引擎：Go Templates
- 前端技术：HTML、CSS、JavaScript

## 项目结构

```
.
├── config/         # 配置文件和配置管理
├── database/       # 数据库连接和初始化
├── handlers/       # 请求处理器
├── middleware/     # 中间件
├── models/         # 数据模型
├── routes/         # 路由配置
├── static/         # 静态资源
├── templates/      # HTML模板
├── utils/          # 工具函数
└── main.go         # 程序入口
```

## 快速开始

### 环境要求

- Go 1.23.4 或更高版本
- PostgreSQL 数据库

### 安装步骤

1. 克隆项目
```bash
git clone https://github.com/your-username/LinkHUB.git
cd LinkHUB
```

2. 安装依赖
```bash
go mod download
```

3. 配置数据库
- 复制 `config/config_exp.yaml` 为 `config/config.yaml`
- 修改配置信息

4. 运行项目
```bash
go run main.go
```

访问 `http://localhost:5002` 即可看到项目运行效果

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进项目。在提交 PR 之前，请确保：

1. 代码符合项目规范
2. 更新相关注释

## 开源协议

本项目采用 MIT 协议开源。