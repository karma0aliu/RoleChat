import { authState, setAuth, clearAuth } from './auth';
import router from './router';

const API_BASE = 'http://127.0.0.1:8080/api';

let isRefreshing = false;
let failedQueue: { resolve: (value: unknown) => void; reject: (reason?: any) => void; }[] = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

function notifySessionExpired() {
  if (typeof window !== 'undefined') {
    if (!sessionStorage.getItem('sessionExpiredShown')) {
      try { alert('登录已过期，请重新登录。'); } catch {}
      sessionStorage.setItem('sessionExpiredShown', '1');
    }
  }
}

async function refreshToken() {
  isRefreshing = true;
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) {
    clearAuth();
    notifySessionExpired();
    router.push('/login');
    return Promise.reject(new Error('No refresh token available'));
  }

  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    let data: any = {};
    try { data = await res.json(); } catch {}
    if (!res.ok) {
      throw new Error(data?.detail || data?.message || 'Failed to refresh token');
    }
    const newAccessToken = data.access_token || data.token;
  const newRefreshToken = data.refresh_token;
    if (!newAccessToken) {
      throw new Error('New access token not provided');
    }
    if (newRefreshToken) {
      try { localStorage.setItem('refreshToken', newRefreshToken); } catch {}
    }
    setAuth(newAccessToken, authState.user || undefined);
    processQueue(null, newAccessToken);
    return newAccessToken;
  } catch (error) {
    processQueue(error, null);
    clearAuth();
    notifySessionExpired();
    router.push('/login');
    return Promise.reject(error);
  } finally {
    isRefreshing = false;
  }
}

export async function fetchWithAuth(url: string, options: RequestInit = {}) {
  let token = authState.token;

  const finalUrl = url.startsWith('http') ? url : `${API_BASE}${url}`;

  const headers = new Headers(options.headers || {});
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  options.headers = headers;

  let response = await fetch(finalUrl, options);

  if (response.status === 401) {
    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      })
      .then(newToken => {
        headers.set('Authorization', `Bearer ${newToken}`);
        options.headers = headers;
        return fetch(finalUrl, options);
      });
    }

    try {
      const newToken = await refreshToken();
      headers.set('Authorization', `Bearer ${newToken}`);
      options.headers = headers;
      response = await fetch(finalUrl, options);
      if (response.status === 401) {
        notifySessionExpired();
        clearAuth();
        router.push('/login');
        return Promise.reject(new Error('Session expired'));
      }
    } catch (error) {
      return Promise.reject(error);
    }
  }

  return response;
}

// API functions
export async function getTopicsWithLimit(limit: number) {
  const response = await fetchWithAuth(`/chat/topics/limit?n=${limit}`);
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'Failed to fetch topics');
  }
  return response.json();
}

export async function getTopics() {
  const response = await fetchWithAuth('/chat/topics');
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'Failed to fetch topics');
  }
  return response.json();
}
