# WordTask Web

Vue 3 + TypeScript + Tailwind CSS + DaisyUI 前端。

## 快速开始

```bash
npm install
npm run dev          # http://localhost:5173
```

> 后端启动在 `http://localhost:8080` 时，前端已通过 `vite.config.ts` 的 proxy
> 将 `/api` 转发给后端，无需手动处理 CORS。

## 脚本

| 命令 | 说明 |
| --- | --- |
| `npm run dev` | 启动本地开发服务器 |
| `npm run build` | 类型检查 + 产物构建（输出 `dist/`） |
| `npm run preview` | 本地预览 `build` 产物 |
| `npm run type-check` | 仅做 TS 类型检查 |
| `npm test` | 运行 Vitest 单元测试 |
| `npm run test:watch` | 监听模式测试 |

## 目录

```
src/
├── api/            Axios 与接口声明
├── assets/styles/  全局样式（Tailwind 入口）
├── components/
│   ├── common/     通用组件（Navbar 等）
│   └── review/     复习业务组件（WordCard/ActionBars/SessionSummary）
├── mock/           M1 阶段 Mock 数据
├── router/         Vue Router
├── store/          Pinia（review / setting / user）
├── types/          TS 类型声明
├── utils/          工具（Web Speech API 封装等）
├── views/          页面（Home / Review / Login）
├── App.vue
└── main.ts
```
