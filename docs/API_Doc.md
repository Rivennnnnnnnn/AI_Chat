# AI Chat 项目接口文档

本文档描述了 AI Chat 项目目前已完成的 API 接口。

## 基础信息

- **基础路径**: `/api/v1`
- **响应格式**: `JSON`

### 通用响应结构

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| code | int | 状态码（0 表示成功，非 0 表示失败） |
| message | string | 错误描述或成功提示 |
| data | object | 业务数据 |

---

## 认证接口 (Auth)

### 1. 用户注册

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

- **响应示例 (参数错误)**:

```json
{
    "code": 1,
    "message": "参数格式错误",
    "data": null
}
```

- **响应示例 (注册失败)**:

```json
{
    "code": 1002,
    "message": "注册失败，请检查格式或稍后重试",
    "data": null
}
```

### 2. 用户登录

用户通过用户名和密码登录，获取会话 ID。

- **接口地址**: `/auth/login`
- **请求方法**: `POST`
- **请求参数 (JSON)**:

| 参数名 | 类型 | 必填 | 限制 | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| username | string | 是 | 3-20字符 | 用户名 |
| password | string | 是 | 8-20字符 | 密码 |

- **响应示例 (成功)**:

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "sessionId": "452778945612b78a9..."
    }
}
```

- **响应示例 (登录失败)**:

```json
{
    "code": 1001,
    "message": "用户名或密码错误",
    "data": null
}
```

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
