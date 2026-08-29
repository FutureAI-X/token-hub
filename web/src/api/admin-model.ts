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

export interface AdminModel {
  id: number
  name: string
  owner: string
  model_type: string
  request_path: string
  description: string
  quota_type: number
  model_ratio: number
  completion_ratio: number
  context_length: number
  max_output_tokens: number
  status: number
  created_at: string
}

export function getModels() {
  return request<{ success: boolean; data: AdminModel[] }>(`${BASE}/models`)
}

export function createModel(data: {
  name: string
  owner: string
  model_type: string
  request_path?: string
  description?: string
  context_length?: number
  max_output_tokens?: number
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/models`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateModel(id: number, data: {
  name?: string
  owner?: string
  model_type?: string
  request_path?: string
  description?: string
  context_length?: number
  max_output_tokens?: number
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function updateModelStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function deleteModel(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${id}`, {
    method: 'DELETE',
  })
}

// ── 模型端点关联 ──

export interface ModelEndpoint {
  id: number
  model_id: number
  endpoint_id: number
  priority: number
  endpoint_path: string
  endpoint_name: string
}

export function getModelEndpoints(modelId: number) {
  return request<{ success: boolean; data: ModelEndpoint[] }>(`${BASE}/models/${modelId}/endpoints`)
}

export function syncModelEndpoints(modelId: number, endpointIds: number[]) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${modelId}/endpoints`, {
    method: 'PUT',
    body: JSON.stringify({ endpoint_ids: endpointIds }),
  })
}
