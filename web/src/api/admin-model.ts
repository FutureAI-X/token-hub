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
  description: string
  tags: string
  status: number
  created_at: string
}

export function getModels() {
  return request<{ success: boolean; data: AdminModel[] }>(`${BASE}/models`)
}

export function createModel(data: {
  name: string
  owner: string
  description?: string
  tags?: string
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/models`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateModel(id: number, data: {
  name?: string
  owner?: string
  description?: string
  tags?: string
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

// ── 积分规则 ──

export interface QuotaRuleItem {
  id?: number
  rule_id?: number
  param_path: string
  param_value: string
  price: number
}

export interface QuotaRule {
  id: number
  model_id: number
  rule_type: string
  base_price: number
  description?: string
  status: number
  created_at: string
  updated_at: string
  items?: QuotaRuleItem[]
}

export function getQuotaRule(modelId: number) {
  return request<{ success: boolean; data: QuotaRule | null }>(`${BASE}/models/${modelId}/quota-rule`)
}

export function saveQuotaRule(modelId: number, data: {
  rule_type: string
  base_price: number
  description?: string
  items?: QuotaRuleItem[]
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${modelId}/quota-rule`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateQuotaRuleStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/quota-rules/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function deleteQuotaRule(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/quota-rules/${id}`, {
    method: 'DELETE',
  })
}

export function deleteModelQuotaRule(modelId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${modelId}/quota-rule`, {
    method: 'DELETE',
  })
}
