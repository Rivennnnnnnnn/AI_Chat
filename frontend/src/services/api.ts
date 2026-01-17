import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器：注入 SessionId
api.interceptors.request.use((config) => {
  const sessionId = localStorage.getItem('sessionId');
  if (sessionId) {
    config.headers['SessionId'] = sessionId;
  }
  return config;
});

// 响应拦截器：处理未授权情况
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // 处理过期的 session
      localStorage.removeItem('sessionId');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const chatApi = {
  // 创建对话
  createConversation: (title: string) => 
    api.post('/ai/create-conversation', { title }),
  
  // 发送消息
  sendMessage: (query: string, conversationId: string, systemPrompt: string = '你是一个专业的 AI 助手。') => 
    api.post('/ai/chat', { query, conversationId, systemPrompt }),

  // 获取对话列表
  getConversations: () => 
    api.get('/ai/conversations'),

  // 获取对话消息历史
  getConversationMessages: (conversationId: string) => 
    api.post('/ai/conversation-messages', { conversationId }),
};

export default api;
