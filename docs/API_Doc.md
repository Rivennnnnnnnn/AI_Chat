# AI Chat 项目接口文档

本文档描述了 AI Chat 项目目前已完成的 API 接口。

## 基础信息

- **基础路径**: `/api/v1`
- **响应格式**: `JSON`
- **认证方式**: 私有接口需在 Header 中携带 `SessionId`

### 通用响应结构

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| code | int | 状态码（0 表示成功，非 0 表示失败） |
| message | string | 错误描述或成功提示 |
| data | object | 业务数据 |

---

## 认证接口 (Auth) [已对接]

### 1. 用户注册 [已对接]

用于新用户创建账号。

- **接口地址**: `/auth/register`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 限制 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| username | string | 是 | 3-20字符 | 用户名 |
| password | string | 是 | 8-20字符 | 密码 |
| email | string | 是 | 邮箱格式 | 电子邮箱 |

- **响应示例 (成功)**:

```json
{
    "code": 0,
    "message": "success",
    "data": null
}
```

---

## 测试接口 (Test) [已对接]

### 1. 连通性测试 [已对接]

用于测试 API 连通性及登录状态。

- **接口地址**: `/test`
- **请求方法**: `POST`
- **请求头**: 
    - `SessionId`: 用户会话 ID (必填)

- **响应示例 (成功)**:

```json
{
    "code": 0,
    "message": "success",
    "data": null
}
```

---

## 人格管理接口 (Persona) [新增加]

### 1. 创建人格 [已完成]
用户可以自定义或根据模板创建 AI 人格。

- **接口地址**: `/persona/create`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| name | string | 是 | 人格名称 |
| description | string | 是 | 原始描述（用户输入的背景/性格） |
| systemPrompt | string | 是 | 最终的系统提示词（用于指导 LLM） |
| mode | int | 是 | 模式（1: 自定义, 2: 模拟） |
| avatar | string | 否 | 头像 URL |

### 2. 获取人格列表 [已完成]
获取当前用户创建的所有人格。

- **接口地址**: `/persona/list`
- **请求方法**: `GET`

- **响应示例 (成功)**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "personas": [
            {
                "id": "per:xxxx",
                "name": "傲娇学姐",
                "description": "...",
                "systemPrompt": "...",
                "mode": 1,
                "avatar": "..."
            }
        ]
    }
}
```

### 3. 获取人格记忆列表 [已完成]
获取指定人格的长期记忆（仅当前用户）。

- **接口地址**: `/persona/{personaId}/memory/list`
- **请求方法**: `GET`

- **响应示例 (成功)**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "memories": [
            {
                "id": "mem:xxxx",
                "persona_id": "per:xxxx",
                "user_id": 1,
                "type": "preference",
                "content": "用户喜欢深色主题",
                "keywords": "深色主题,偏好",
                "source": "auto",
                "status": "active",
                "hit_count": 0,
                "last_hit_at": null,
                "created_at": "2026-01-20T10:00:00Z",
                "updated_at": "2026-01-20T10:00:00Z"
            }
        ]
    }
}
```

### 4. 手动创建记忆 [已完成]
在指定人格下手动创建一条记忆。

- **接口地址**: `/persona/{personaId}/memory/create`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| type | string | 是 | 类型：fact/preference/event/emotion/relationship |
| content | string | 是 | 记忆内容 |

### 5. 更新记忆内容 [已完成]
更新指定记忆的内容。

- **接口地址**: `/persona/{personaId}/memory/{memoryId}`
- **请求方法**: `PUT`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| content | string | 是 | 记忆内容 |

### 6. 删除记忆 [已完成]
软删除指定记忆。

- **接口地址**: `/persona/{personaId}/memory/{memoryId}`
- **请求方法**: `DELETE`

---

## AI 聊天接口 (AI Chat) [已对接]

所有 AI 接口均需要登录后调用，并在 Header 中携带 `SessionId`。

### 1. 创建对话 [已对接]
用于初始化一个新的聊天会话。

- **接口地址**: `/ai/create-conversation`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| title | string | 是 | 对话标题 |
| personaId | string | 是 | AI 人格 ID（同一人格仅对应一个会话） |

> 说明：如果该 `personaId` 已存在会话，会直接返回该会话 ID，不再新建。

- **响应示例 (成功)**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "conversationId": "con:xxxx-xxxx-xxxx"
    }
}
```

### 2. 发送聊天消息 (人格版) [已更新]
在指定的对话中发送消息，并指定 AI 人格进行回复。

- **接口地址**: `/ai/chat-with-persona`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| query | string | 是 | 用户提问内容 |
| conversationId | string | 是 | 对话 ID |
| personaId | string | 是 | AI 人格 ID |

- **上下文策略**: 仅使用最近 30 轮对话作为 LLM 上下文。
- **回复格式**: 回复内容可能包含 `\n` 作为分段符号，用于前端模拟逐条消息显示。

- **响应示例 (成功)**:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "message": "（根据选定人格生成的回复内容）"
    }
}
```

### 3. 获取对话列表 [已对接]
获取当前用户的所有聊天对话。

- **接口地址**: `/ai/conversations`
- **请求方法**: `GET`

### 4. 获取对话消息历史 [已对接]
获取指定对话的消息历史记录。

- **接口地址**: `/ai/conversation-messages`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| conversationId | string | 是 | 对话 ID |

---

## 状态码定义

| 状态码 | 描述 |
| :--- | :--- |
| 0 | 成功 (SuccessCode) |
| 1 | 参数格式错误 (FailedCode) |
| 1001 | 用户名或密码错误 (LoginFailedCode) |
| 1002 | 注册失败 (RegisterFailedCode) |
| 1003 | 数据库操作失败 (DataBaseFailedCode) |
| 1004 | Redis 连接或操作失败 (RedisFailedCode) |
| 1005 | 会话已过期 (SessionExpiredCode) |
| 1006 | 聊天失败 (ChatFailedCode) |
