# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

本项目是一个基于 Go (Gin) 和 TypeScript 的前后端分离单词记忆应用，旨在提供流畅的移动端背诵体验与高效的词库管理系统。

## Tech Stack

**后端 (Backend)**
- 语言: Go 1.2x+
- 框架: Gin Gonic (RESTful API)
- 数据库: SQLite (开发便捷)
- ORM: GORM
- 记忆算法: go-fsrs 或 SM-2（M4 集成）
- 认证: JWT (M3 预留)

**前端 (Frontend)**
- 框架: Vue 3 + TypeScript
- 构建: Vite 5
- 样式: Tailwind CSS + DaisyUI
- 状态: Pinia
- 路由: Vue Router 4
- 通信: Axios
- 发音: 浏览器原生 Web Speech API
- 测试: Vitest + @vue/test-utils

**前后端通信**
- 完全分离，JSON 格式，统一响应结构 `{ code, data, msg }`
- 开发期前端通过 Vite proxy (`/api` → `http://localhost:8080`) 转发，免 CORS 配置
- 线上部署由后端开启 CORS 允许前端域名

## Repository Structure

```
WordTask/
├── server/                  # Go 后端（M2 启动搭建）
│   ├── cmd/                 # 程序入口
│   ├── internal/            # 业务模块（handler/service/repo/model）
│   ├── pkg/                 # 可复用公共包（fsrs 封装等）
│   ├── migrations/          # SQL 迁移或 GORM AutoMigrate 脚本
│   ├── data/                # 运行期 SQLite 文件（被 .gitignore）
│   ├── go.mod
│   └── main.go
│
├── web/                     # Vue 3 前端
│   ├── public/
│   ├── src/
│   │   ├── api/             # Axios 实例 + 接口声明
│   │   ├── assets/styles/   # 全局样式（Tailwind 入口）
│   │   ├── components/
│   │   │   ├── common/      # Navbar 等通用组件
│   │   │   └── review/      # WordCard / ActionBars / SessionSummary
│   │   ├── mock/            # M1 阶段 Mock 数据
│   │   ├── router/          # Vue Router
│   │   ├── store/           # Pinia（review / setting / user）
│   │   ├── types/           # TS 类型声明
│   │   ├── utils/           # Web Speech API 封装等
│   │   ├── views/           # Home / Review / Login
│   │   ├── App.vue
│   │   └── main.ts
│   ├── index.html
│   ├── package.json
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── 排期.md                   # Milestone 规划
├── CLAUDE.md
└── .gitignore
```

## Common Commands

### 前端（web/）

```bash
cd web

# 安装依赖
npm install

# 本地开发（默认 http://localhost:5173）
npm run dev

# 构建生产产物（先做 vue-tsc 类型检查，再 vite build，输出 web/dist/）
npm run build

# 本地预览构建产物
npm run preview

# 仅类型检查（CI 友好）
npm run type-check

# 单元测试（Vitest，一次性跑完）
npm test

# 单元测试（监听模式）
npm run test:watch

# Lint / 自动修复（eslint 配置按需补充）
npm run lint
```

### 后端（server/，M2 启用）

```bash
cd server

# 安装依赖
go mod tidy

# 本地运行（默认 :8080）
go run ./cmd/server

# 构建
go build -o bin/server ./cmd/server

# 单元测试
go test ./...

# 代码格式化 / 静态检查
gofmt -w .
go vet ./...
```

### 一键联调

1. 终端 A：`cd server && go run ./cmd/server`（监听 8080）
2. 终端 B：`cd web && npm run dev`（监听 5173）
3. 浏览器访问 `http://localhost:5173`；前端 `/api/*` 请求会被 Vite proxy 转到后端。

## Architecture & Key Concepts

### UI/UX 规范 (Design Guidelines)
- **移动端优先**：所有按钮点击区域高度不低于 **44px**（Tailwind 中通过 `min-h-touch` / `min-w-touch` 实现）。
- **卡片式布局**：柔和阴影 + 圆角（`rounded-3xl` + `shadow-xl`）。
- **单词卡片**：默认只显示单词与音标；点击翻转显示释义与例句（3D `rotateY` 动画）。
- **快速反馈**：底部红（忘记）/ 黄（模糊）/ 绿（认识）三个按钮。
- **进度展示**：Navbar 同时展示 **当前批次进度（如 2/5）** 与 **今日整体进度（如 12/50）**。

### 背诵闯关模式
- 以**批次（Session）**为单位背诵，每组数量由 `settingStore.batchSize` 控制（默认 5）。
- `reviewStore` 状态分层：`dailyQueue` / `sessionQueue` / `sessionIndex`。
- 每批次结束展示 `SessionSummary`，提示“再背一组”或“休息一下”。
- 评分**单条异步上传**（`api/word.ts#submitReview`），避免批次中途丢失进度。

### 开发规范 (Development Rules)
- **代码风格**
  - Go：`gofmt` + `go vet`，命名遵循官方约定。
  - TS：驼峰命名，严格类型检查（`strict: true`）。
- **错误处理**：所有 API 返回格式统一 `{"code": 200, "data": {}, "msg": "success"}`，非 200 视为业务错误。
- **记忆算法（M4）**：后端 `next` 接口结合 Progress 表，计算 `NextReview < now` 的单词，并按熟悉度升序返回。

## Notes for Claude / CodeBuddy

- 优先使用中文回答。
- 修改前请先阅读相关模块，避免破坏既有约定。
- **不要删除 `.codebuddy/` 目录**（其中包含项目持久化数据）。
- 数据库文件（`*.db` / `*.sqlite`）已在 `.gitignore` 中忽略。
- TODO:
  - [ ] M2：搭建 `server/` 目录与 `/api/words/today`、`/api/words/review` 接口。
  - [ ] M3：接入 JWT + 多账号映射。
  - [ ] M4：集成 go-fsrs，落地 Progress 表与下次复习时间计算。
