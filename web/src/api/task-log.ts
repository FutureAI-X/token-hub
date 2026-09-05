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

export interface TaskLog {
  id: number
  task_id: string
  vendor_id: number
  model_id: number
  endpoint_id: number
  status: string
  vendor_response?: string
  query_response?: string
  created_at: string
  updated_at: string
}

export interface TaskLogListResponse {
  items: TaskLog[]
  total: number
  page: number
  page_size: number
}

export function getTaskLogs(params: {
  page?: number
  page_size?: number
  status?: string
}) {
  const searchParams = new URLSearchParams()
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.page_size) searchParams.set('page_size', params.page_size.toString())
  if (params.status) searchParams.set('status', params.status)

  return request<{ success: boolean; data: TaskLogListResponse }>(
    `${BASE}/task-logs?${searchParams.toString()}`
  )
}

export function getTaskLogDetail(id: number) {
  return request<{ success: boolean; data: TaskLog }>(`${BASE}/task-logs/${id}`)
}
