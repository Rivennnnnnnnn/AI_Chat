import React, { useState, useEffect, useRef } from 'react';
import { LogOut, Send, Bot, User, Settings, TestTube, PlusCircle, Loader2, MessageSquare } from 'lucide-react';
import { useAuthStore } from '../store/authStore';
import { useNavigate } from 'react-router-dom';
import api, { chatApi } from '../services/api';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
}

interface Conversation {
  id: string;
  title: string;
  createdAt: string;
}

const ChatPage: React.FC = () => {
  const { username, clearSession } = useAuthStore();
  const navigate = useNavigate();
  const [testResult, setTestResult] = useState<string | null>(null);
  const [loadingTest, setLoadingTest] = useState(false);
  
  // Chat state
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [isSending, setIsSending] = useState(false);
  const [isCreatingConv, setIsCreatingConv] = useState(false);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Load conversations on mount
  useEffect(() => {
    fetchConversations();
  }, []);

  const fetchConversations = async () => {
    try {
      const res = await chatApi.getConversations();
      if (res.data.code === 0) {
        setConversations(res.data.data.conversations || []);
      }
    } catch (err) {
      console.error('Failed to fetch conversations', err);
    }
  };

  const loadConversationHistory = async (id: string) => {
    if (id === conversationId) return;
    
    setIsLoadingHistory(true);
    setConversationId(id);
    try {
      const res = await chatApi.getConversationMessages(id);
      if (res.data.code === 0) {
        setMessages(res.data.data.messages || []);
      } else {
        setTestResult(`加载历史失败: ${res.data.message}`);
      }
    } catch (err: any) {
      setTestResult(`请求历史失败: ${err.message}`);
    } finally {
      setIsLoadingHistory(false);
    }
  };

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout');
    } catch (err) {
      console.error('Logout failed', err);
    } finally {
      clearSession();
      navigate('/login');
    }
  };

  const createNewConversation = async () => {
    setIsCreatingConv(true);
    try {
      const res = await chatApi.createConversation(`与 ${username} 的新对话`);
      if (res.data.code === 0) {
        const newConv = {
          id: res.data.data.conversationId,
          title: `与 ${username} 的新对话`,
          createdAt: new Date().toISOString()
        };
        setConversations(prev => [newConv, ...prev]);
        setConversationId(newConv.id);
        setMessages([]);
        setTestResult('会话创建成功，可以开始聊天了！');
      } else {
        setTestResult(`创建失败: ${res.data.message}`);
      }
    } catch (err: any) {
      setTestResult(`创建会话失败: ${err.message}`);
    } finally {
      setIsCreatingConv(false);
      setTimeout(() => setTestResult(null), 3000);
    }
  };

  const handleSendMessage = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!input.trim() || isSending) return;

    // If no conversation yet, create one first
    let currentConvId = conversationId;
    if (!currentConvId) {
      setIsCreatingConv(true);
      try {
        const res = await chatApi.createConversation(`与 ${username} 的新对话`);
        if (res.data.code === 0) {
          currentConvId = res.data.data.conversationId;
          setConversationId(currentConvId);
          // Refresh list to show new conversation
          fetchConversations();
        } else {
          setTestResult(`创建会话失败: ${res.data.message}`);
          setIsCreatingConv(false);
          return;
        }
      } catch (err: any) {
        setTestResult(`请求失败: ${err.message}`);
        setIsCreatingConv(false);
        return;
      } finally {
        setIsCreatingConv(false);
      }
    }

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsSending(true);

    try {
      const res = await chatApi.sendMessage(input, currentConvId!);
      if (res.data.code === 0) {
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: res.data.data.message,
        };
        setMessages(prev => [...prev, assistantMessage]);
      } else {
        setTestResult(`聊天失败: ${res.data.message}`);
      }
    } catch (err: any) {
      setTestResult(`聊天请求失败: ${err.message}`);
    } finally {
      setIsSending(false);
    }
  };

  const runTest = async () => {
    setLoadingTest(true);
    try {
      const res = await api.post('/test');
      if (res.data.code === 0) {
        setTestResult('API 连通性测试成功！');
      } else {
        setTestResult(`测试失败: ${res.data.message}`);
      }
    } catch (err: any) {
      setTestResult(`请求失败: ${err.message}`);
    } finally {
      setLoadingTest(false);
      setTimeout(() => setTestResult(null), 3000);
    }
  };

  return (
    <div className="flex h-screen w-full bg-zinc-950 text-zinc-100 overflow-hidden">
      {/* Sidebar */}
      <div className="w-64 bg-zinc-900 border-r border-zinc-800 flex flex-col">
        <div className="p-6 border-b border-zinc-800 flex items-center gap-3">
          <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <Bot className="h-5 w-5 text-white" />
          </div>
          <span className="font-bold text-xl tracking-tight">AI Chat</span>
        </div>
        
        <nav className="flex-1 p-4 space-y-2 overflow-y-auto custom-scrollbar">
          <div className="text-xs font-semibold text-zinc-500 uppercase tracking-wider px-2 mb-2">会话历史</div>
          <button 
            onClick={createNewConversation}
            disabled={isCreatingConv}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white transition-colors disabled:opacity-50 mb-4"
          >
            {isCreatingConv ? <Loader2 className="h-4 w-4 animate-spin" /> : <PlusCircle className="h-4 w-4" />}
            <span>新对话</span>
          </button>
          
          <div className="space-y-1">
            {conversations.map((conv) => (
              <button
                key={conv.id}
                onClick={() => loadConversationHistory(conv.id)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors group ${
                  conversationId === conv.id 
                    ? 'bg-zinc-800 text-white' 
                    : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'
                }`}
              >
                <MessageSquare className={`h-4 w-4 flex-shrink-0 ${conversationId === conv.id ? 'text-blue-400' : 'text-zinc-500'}`} />
                <span className="text-sm truncate text-left">{conv.title}</span>
              </button>
            ))}
          </div>

          <div className="mt-8 pt-4 border-t border-zinc-800/50">
            <div className="text-xs font-semibold text-zinc-500 uppercase tracking-wider px-2 mb-2">系统</div>
            <button onClick={runTest} disabled={loadingTest} className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-zinc-800 text-zinc-400 hover:text-white transition-colors disabled:opacity-50">
              <TestTube className={`h-4 w-4 ${loadingTest ? 'animate-pulse' : ''}`} />
              <span>连通性测试</span>
            </button>
          </div>
        </nav>

        <div className="p-4 border-t border-zinc-800 bg-zinc-900/50">
          <div className="flex items-center gap-3 px-2 py-3 mb-2">
            <div className="h-10 w-10 bg-zinc-800 rounded-full flex items-center justify-center border border-zinc-700">
              <User className="h-6 w-6 text-zinc-400" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{username}</p>
              <p className="text-xs text-zinc-500 truncate">在线</p>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-red-500/10 text-zinc-400 hover:text-red-400 transition-colors"
          >
            <LogOut className="h-4 w-4" />
            <span>退出登录</span>
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col relative">
        {/* Header */}
        <header className="h-16 border-b border-zinc-800 flex items-center justify-between px-6 bg-zinc-900/50 backdrop-blur-sm">
          <div className="flex items-center gap-2">
            <h1 className="text-lg font-semibold truncate max-w-[400px]">
              {conversationId 
                ? conversations.find(c => c.id === conversationId)?.title || `会话: ${conversationId.slice(0, 8)}...`
                : '开始新的对话'}
            </h1>
            {(isSending || isLoadingHistory) && <Loader2 className="h-3 w-3 animate-spin text-blue-500" />}
          </div>
          <button className="p-2 text-zinc-400 hover:text-white transition-colors">
            <Settings className="h-5 w-5" />
          </button>
        </header>

        {/* Messages area */}
        <main className="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar">
          {isLoadingHistory ? (
            <div className="h-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
            </div>
          ) : messages.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center space-y-6">
              <div className="h-20 w-20 bg-zinc-900 border border-zinc-800 rounded-2xl flex items-center justify-center shadow-xl mb-4">
                <Bot className="h-10 w-10 text-blue-500" />
              </div>
              <div className="max-w-md">
                <h2 className="text-2xl font-bold mb-2">你好, {username}!</h2>
                <p className="text-zinc-500">
                  我是你的 AI 助手。你可以问我任何问题，或者从侧边栏选择一个历史对话。
                </p>
              </div>
            </div>
          ) : (
            <div className="max-w-4xl mx-auto space-y-6">
              {messages.map((msg, index) => (
                <div key={msg.id || index} className={`flex gap-4 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}>
                  <div className={`h-8 w-8 rounded-lg flex items-center justify-center flex-shrink-0 ${
                    msg.role === 'user' ? 'bg-blue-600' : 'bg-zinc-800 border border-zinc-700'
                  }`}>
                    {msg.role === 'user' ? <User className="h-5 w-5 text-white" /> : <Bot className="h-5 w-5 text-blue-400" />}
                  </div>
                  <div className={`max-w-[85%] rounded-2xl px-4 py-2 text-sm leading-relaxed ${
                    msg.role === 'user' 
                      ? 'bg-blue-600 text-white rounded-tr-none' 
                      : 'bg-zinc-900 border border-zinc-800 text-zinc-200 rounded-tl-none'
                  }`}>
                    {msg.role === 'assistant' ? (
                      <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        components={{
                          code({ node, inline, className, children, ...props }: any) {
                            const match = /language-(\w+)/.exec(className || '');
                            return !inline && match ? (
                              <SyntaxHighlighter
                                style={vscDarkPlus}
                                language={match[1]}
                                PreTag="div"
                                {...props}
                              >
                                {String(children).replace(/\n$/, '')}
                              </SyntaxHighlighter>
                            ) : (
                              <code className={className} {...props}>
                                {children}
                              </code>
                            );
                          },
                          p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
                          ul: ({ children }) => <ul className="list-disc ml-4 mb-2">{children}</ul>,
                          ol: ({ children }) => <ol className="list-decimal ml-4 mb-2">{children}</ol>,
                          li: ({ children }) => <li className="mb-1">{children}</li>,
                          a: ({ children, href }) => <a href={href} className="text-blue-400 underline" target="_blank" rel="noreferrer">{children}</a>,
                          blockquote: ({ children }) => <blockquote className="border-l-4 border-zinc-700 pl-4 italic my-2">{children}</blockquote>,
                          table: ({ children }) => (
                            <div className="overflow-x-auto my-4 rounded-lg border border-zinc-800">
                              <table className="w-full border-collapse text-left text-sm">
                                {children}
                              </table>
                            </div>
                          ),
                          thead: ({ children }) => <thead className="bg-zinc-800/50 text-zinc-300">{children}</thead>,
                          th: ({ children }) => <th className="px-4 py-2 border-b border-zinc-700 font-semibold">{children}</th>,
                          td: ({ children }) => <td className="px-4 py-2 border-b border-zinc-800 text-zinc-400">{children}</td>,
                          tr: ({ children }) => <tr className="even:bg-zinc-900/50 hover:bg-zinc-800/30 transition-colors last:border-0">{children}</tr>,
                        }}
                      >
                        {msg.content}
                      </ReactMarkdown>
                    ) : (
                      msg.content
                    )}
                  </div>
                </div>
              ))}
              {isSending && (
                <div className="flex gap-4">
                  <div className="h-8 w-8 rounded-lg bg-zinc-800 border border-zinc-700 flex items-center justify-center">
                    <Bot className="h-5 w-5 text-blue-400" />
                  </div>
                  <div className="bg-zinc-900 border border-zinc-800 text-zinc-400 rounded-2xl rounded-tl-none px-4 py-2 text-sm flex items-center gap-2">
                    <Loader2 className="h-3 w-3 animate-spin" />
                    AI 正在思考...
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>
          )}

          {testResult && (
            <div className="fixed bottom-24 left-1/2 -translate-x-1/2 z-50">
              <div className={`px-6 py-3 rounded-xl border shadow-2xl animate-in fade-in slide-in-from-bottom-4 duration-300 ${
                testResult.includes('成功') ? 'bg-green-500/10 border-green-500/20 text-green-400' : 'bg-red-500/10 border-red-500/20 text-red-400'
              }`}>
                {testResult}
              </div>
            </div>
          )}
        </main>

        {/* Input area */}
        <footer className="p-6 border-t border-zinc-800 bg-zinc-950">
          <form onSubmit={handleSendMessage} className="max-w-4xl mx-auto relative">
            <textarea
              rows={1}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSendMessage();
                }
              }}
              disabled={isSending || isCreatingConv}
              placeholder={isCreatingConv ? "正在创建会话..." : "输入消息... (Shift+Enter 换行)"}
              className="w-full bg-zinc-900 border border-zinc-800 rounded-2xl pl-6 pr-14 py-4 focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all text-zinc-300 disabled:opacity-50 resize-none overflow-hidden"
              style={{ minHeight: '56px' }}
            />
            <button
              type="submit"
              disabled={!input.trim() || isSending || isCreatingConv}
              className="absolute right-3 bottom-3 h-10 w-10 bg-blue-600 hover:bg-blue-500 rounded-xl flex items-center justify-center text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSending ? <Loader2 className="h-5 w-5 animate-spin" /> : <Send className="h-5 w-5" />}
            </button>
          </form>
          <p className="text-center text-[10px] text-zinc-600 mt-4 uppercase tracking-widest font-semibold">
            Powered by DeepSeek & Gemini
          </p>
        </footer>
      </div>
    </div>
  );
};

export default ChatPage;
