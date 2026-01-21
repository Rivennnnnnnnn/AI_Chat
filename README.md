# AI Chat

基于 Go + React 的全栈智能聊天项目，支持人格对话、长期记忆、向量检索与现代化 UI 体验。

## 技术栈

- 后端：Go 1.21+, Gin, GORM, Redis, Eino
- 前端：React 18, Vite, Tailwind CSS, Zustand
- 数据：MySQL, Redis, Milvus
- 模型：DeepSeek Chat/Embedding

## 功能概览

- 用户系统：注册、登录、会话鉴权
- 人格会话：每个人格唯一会话，自动创建并持久化
- 上下文策略：仅取最近 30 轮对话
- 记忆检索：优先 Milvus，失败/无结果回退数据库
- 前端体验：`\n` 分段逐条展示 + “对方正在输入”动画

## 快速启动

### 1. 配置
- 复制示例配置并填写真实值：
  - `configs/chat.example.yaml` → `configs/chat.yaml`
  - `configs/mysql.example.yaml` → `configs/mysql.yaml`
  - `configs/redis.example.yaml` → `configs/redis.yaml`
  - `configs/milvus.example.yaml` → `configs/milvus.yaml`
- 准备依赖服务：MySQL、Redis、Milvus

### 2. 启动后端
```bash
go run main.go
```

### 3. 启动前端
```bash
cd frontend
npm install
npm run dev
```

## 文档

- API：`docs/API_Doc.md`
- 项目结构：`项目结构说明.md`
- 表结构：`表结构设计/`
