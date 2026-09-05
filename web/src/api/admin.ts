const BASE = '/api/admin'

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

// ── 用户列表 ──
export interface AdminUser {
  id: number
  username: string
  display_name: string
  role: number
  status: number
  email: string
  credits: number
  used_credits: number
  created_at: string
}

export interface UserListResponse {
  success: boolean
  data: { items: AdminUser[]; total: number }
  message?: string
}

export function getUsers(params: { p?: number; page_size?: number; keyword?: string }) {
  const query = new URLSearchParams()
  if (params.p) query.set('p', String(params.p))
  if (params.page_size) query.set('page_size', String(params.page_size))
  if (params.keyword) query.set('keyword', params.keyword)
  return request<UserListResponse>(`${BASE}/users?${query}`)
}

// ── 创建用户 ──
export function createUser(data: {
  username: string
  password: string
  display_name?: string
  role?: number
  credits?: number
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/users`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

// ── 更新用户 ──
export function updateUser(id: number, data: { username?: string; display_name?: string; password?: string }) {
  return request<{ success: boolean; message: string }>(`${BASE}/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

// ── 删除用户 ──
export function deleteUser(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/users/${id}`, {
    method: 'DELETE',
  })
}

// ── 更新用户状态 ──
export function updateUserStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/users/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

// ── 积分调整 ──
export function adjustUserCredits(id: number, mode: string, value: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/users/${id}/credits`, {
    method: 'PUT',
    body: JSON.stringify({ mode, value }),
  })
}

// ── 重置密码 ──
export function resetUserPassword(id: number, dataKey: string) {
  return request<{ success: boolean; message: string; data?: { encrypted_password: string } }>(`${BASE}/users/${id}/password`, {
    method: 'PUT',
    body: JSON.stringify({ data_key: dataKey }),
  })
}
