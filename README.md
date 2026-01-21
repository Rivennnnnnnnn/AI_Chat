# AI Chat - 全栈智能聊天应用

基于 **Go (Gin)** + **React (TypeScript)** + **DeepSeek** 构建的全栈 AI 聊天项目。支持用户注册登录、多轮对话管理、流式响应（开发中）等功能。

## 🚀 技术栈

- **后端**: Go 1.21+, Gin, GORM, Go-Redis, Eino (AI SDK)
- **前端**: React 18, Vite, Tailwind CSS, Lucide React, Zustand
- **数据库**: MySQL 8.0, Redis 7.0, Milvus (向量检索)
- **AI 模型**: DeepSeek-Chat (通过 API 接入)

## 📦 功能特性

- [x] **用户系统**: 注册、登录、Session 鉴权、登出。
- [x] **对话管理**: 每个人格唯一会话，自动创建并保存历史消息。
- [x] **AI 聊天**: 接入 DeepSeek 模型，支持系统提示词与最近 30 轮上下文。
- [x] **记忆检索**: 优先 Milvus 向量检索，失败或无结果回退数据库。
- [x] **现代 UI**: 基于 Tailwind 设计的深色模式界面，自适应滚动和响应式布局。
- [x] **调试工具**: 接口连通性测试按钮、后端独立测试脚本。

## 🛠️ 快速启动

### 1. 准备工作
- 安装 MySQL 并创建数据库 `ai_chat`。
- 安装 Redis。
- 获取 DeepSeek API Key。

### 2. 后端启动
1. 进入项目根目录。
2. 在 `configs/` 下配置 `mysql.yaml`, `redis.yaml`, `chat.yaml`, `milvus.yaml`。
   - 可参考 `configs/*.example.yaml` 示例文件，复制后再填写实际值。
3. 运行服务：
   ```bash
   go run main.go
   ```

### 3. 前端启动
1. 进入 `frontend/` 目录。
2. 安装依赖：
   ```bash
   npm install
   ```
3. 启动开发服务器：
   ```bash
   npm run dev
   ```

## 📖 文档指南

- **接口定义**: 参见 `docs/API_Doc.md`
- **项目结构**: 参见 `项目结构说明.md`
- **数据库设计**: 参见 `表结构设计/`


