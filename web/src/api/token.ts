const BASE = '/api/user'

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  // 注意顺序：options 先展开，headers 最后合并。
  // 若按原来的写法（headers 在前、...options 在后），任何传入 headers 的调用
  // 都会整体覆盖掉 Authorization，导致 401。
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
      ...((options?.headers as Record<string, string>) ?? {}),
    },
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
  // data_key 通过请求头传递，避免出现在 URL 查询串中（会进入访问日志/浏览器历史）
  return request<{ success: boolean; data: ApiKey[] }>(`${BASE}/tokens`, {
    headers: { 'X-Data-Key': dataKey },
  })
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
