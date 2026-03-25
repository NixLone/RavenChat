const API = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token')
  const headers: HeadersInit = { 'Content-Type': 'application/json', ...(opts.headers || {}) }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${API}${path}`, { ...opts, headers })
  if (!res.ok) throw new Error(await res.text())
  if (res.status === 204) return {} as T
  return res.json()
}

export const wsUrl = (chatId: string) => `${API.replace('http', 'ws')}/api/v1/chats/${chatId}/ws`
