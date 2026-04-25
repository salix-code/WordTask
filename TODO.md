# 任务清单

## 1. 前台与后台对接
完成从前台向后台请求单词的逻辑。要配合：
1. 登录
2. 选择book。
暂时先不用考虑登录，仅仅在支持功能2的情况下完成。
### 1.1 前台任务清单

- [x] `settingStore.ts`：新增 `wordbook` 字段，默认值 `'KET'`
- [x] `word.ts`：`fetchTodayWords` 接收 `wordbook` 参数，请求时带上 `userId: '0000-0000-0000-0000'` 和 `wordbook`
- [x] `reviewStore.ts`：`initDaily()` 改为 `async`，删除 mockWords，改为调用 `fetchTodayWords(setting.wordbook)`
- [x] `Review.vue`：`onMounted` 改为 `async/await`，等待 `initDaily()` 完成后再渲染

### 1.2 后台任务清单

**数据层**
- [x] `internal/model/word.go`：Word 结构体，含 wordbook / source_order 等字段
- [x] `internal/model/progress.go`：UserProgress 结构体，主键 (user_id, wordbook)，字段 completed_count
- [x] `internal/db/seed.go`：启动时检查 words 表为空则从 JSON 文件 seed（批量 200 条）
- [x] `internal/db/db.go`：GORM 初始化 + AutoMigrate

**业务层**
- [x] `internal/service/word.go`：GetTodayWords(userId, wordbook) → 查 offset，取 words，检查是否学完；响应含 dailyGoal / completed / words（含 sortKey 字段供 M4 FSRS 替换）
- [x] `internal/service/progress.go`：AdvanceProgress(userId, wordbook, count) → completed_count += count

**接口层**
- [x] `api/handlers.go`：替换 mock 版 GetTodayWords，新增 AdvanceProgress handler；userId 完全从请求参数读取
- [x] `main.go`：初始化 DB 和 seed，新增 POST /api/progress/advance 路由；DB_PATH / WORDBOOK_DIR 支持环境变量覆盖

**接口说明**
- `GET  /api/words/today?userId=xxx&wordbook=KET` → 返回 `{ dailyGoal, completed, words[] }`
- `POST /api/progress/advance` body: `{ userId, wordbook, count }` → 推进 offset