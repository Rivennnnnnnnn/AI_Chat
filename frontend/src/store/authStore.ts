import { create } from 'zustand';

interface AuthState {
  sessionId: string | null;
  username: string | null;
  setSession: (sessionId: string, username: string) => void;
  clearSession: () => void;
  isAuthenticated: boolean;
}

export const useAuthStore = create<AuthState>((set) => ({
  sessionId: localStorage.getItem('sessionId'),
  username: localStorage.getItem('username'),
  isAuthenticated: !!localStorage.getItem('sessionId'),
  setSession: (sessionId, username) => {
    localStorage.setItem('sessionId', sessionId);
    localStorage.setItem('username', username);
    set({ sessionId, username, isAuthenticated: true });
  },
  clearSession: () => {
    localStorage.removeItem('sessionId');
    localStorage.removeItem('username');
    set({ sessionId: null, username: null, isAuthenticated: false });
  },
}));
