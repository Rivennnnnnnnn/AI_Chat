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
  
  // 发送消息 (与人格对话)
  sendMessage: (query: string, conversationId: string, personaId: string) => 
    api.post('/ai/chat-with-persona', { query, conversationId, personaId }),

  // 获取对话列表
  getConversations: () => 
    api.get('/ai/conversations'),

  // 获取对话消息历史
  getConversationMessages: (conversationId: string) => 
    api.post('/ai/conversation-messages', { conversationId }),
};

export const personaApi = {
  // 创建人格
  createPersona: (data: { name: string, description: string, systemPrompt: string, mode: number, avatar: string }) =>
    api.post('/persona/create', data),

  // 获取人格列表
  getPersonas: () =>
    api.get('/persona/list'),
};

export default api;
