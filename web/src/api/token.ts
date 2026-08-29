const BASE = '/api/user'

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    ...options,
  })
  return res.json()
}

export interface ApiKey {
  id: number
  name: string
  key: string
  status: number
  created_at: string
}

export function getTokens(dataKey: string) {
  return request<{ success: boolean; data: ApiKey[] }>(`${BASE}/tokens?data_key=${encodeURIComponent(dataKey)}`)
}

export function createToken(name: string, dataKey: string) {
  return request<{ success: boolean; message: string; data?: ApiKey }>(`${BASE}/tokens`, {
    method: 'POST',
    body: JSON.stringify({ name, data_key: dataKey }),
  })
}

export function updateToken(id: number, name: string) {
  return request<{ success: boolean; message: string }>(`${BASE}/tokens/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  })
}

export function deleteToken(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/tokens/${id}`, {
    method: 'DELETE',
  })
}
