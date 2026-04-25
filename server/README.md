# WordTask Server

基于 Gin 的后端服务，为 WordTask 前端提供单词复习相关接口。

## 运行环境

- Go 1.25+

## 快速开始

```bash
cd server
go mod tidy
go run .
```

服务启动后监听 `http://localhost:8080`。

## 接口列表（M2 Mock 版本）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/ping` | 健康检查 |
| GET  | `/api/words/today` | 获取今日待复习单词列表 |
| POST | `/api/words/review` | 提交对单词的复习评价 |

### 统一响应格式

```json
{
  "code": 200,
  "msg": "success",
  "data": { }
}
```

### POST /api/words/review

请求体：

```json
{
  "word_id": 1,
  "grade": "green"   // red | yellow | green
}
```

## 目录结构

```
server/
├── main.go           # 入口：路由、CORS
├── api/
│   ├── response.go   # 统一响应
│   └── handlers.go   # 接口处理函数
├── mock/
│   └── words.go      # Mock 数据（M3 后由 SQLite 替代）
└── go.mod
```

## 里程碑

- **M2（当前）**：Gin 基础服务 + 内存 Mock 数据，打通前后端联调。
- **M3**：接入 SQLite，导入 `data/wordbooks/*.json` 词库，建立 `account + user_id` 多用户映射。
- **M4**：集成 FSRS 记忆算法（`go-fsrs`），由后端计算下一次复习时间。
