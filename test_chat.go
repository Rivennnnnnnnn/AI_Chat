package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "http://localhost:8001/api/v1"

func main() {
	// 1. 模拟登录获取 SessionId (请确保用户已注册)
	// 如果你还没有用户，请先调用注册接口
	sessionId := login("xl18341", "Xl@1132511325")
	if sessionId == "" {
		fmt.Println("登录失败，请确保用户已存在且服务已启动")
		return
	}
	fmt.Printf("登录成功, SessionId: %s\n", sessionId)

	// 2. 创建对话
	convId := createConversation(sessionId, "测试对话")
	if convId == "" {
		fmt.Println("创建对话失败")
		return
	}
	fmt.Printf("对话创建成功, ConversationId: %s\n", convId)

	// 3. 发起聊天
	chat(sessionId, convId, "Riven的秘密是什么？", "你是一个幽默的编程助手")
}

func login(username, password string) string {
	data, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if data, ok := result["data"].(map[string]interface{}); ok && data != nil {
		return data["sessionId"].(string)
	}
	return ""
}

func createConversation(sessionId, title string) string {
	data, _ := json.Marshal(map[string]string{
		"title": title,
	})
	req, _ := http.NewRequest("POST", baseURL+"/ai/create-conversation", bytes.NewBuffer(data))
	req.Header.Set("SessionId", sessionId)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求错误: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取并打印返回内容
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("创建对话接口返回: %s\n", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if data, ok := result["data"].(map[string]interface{}); ok && data != nil {
		return data["conversationId"].(string)
	}
	return ""
}

func chat(sessionId, convId, query, systemPrompt string) {
	data, _ := json.Marshal(map[string]string{
		"query":          query,
		"conversationId": convId,
		"systemPrompt":   systemPrompt,
	})
	req, _ := http.NewRequest("POST", baseURL+"/ai/chat", bytes.NewBuffer(data))
	req.Header.Set("SessionId", sessionId)
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("正在等待 AI 回复...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("聊天请求出错: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("API 响应内容: %s\n", string(body))
}
