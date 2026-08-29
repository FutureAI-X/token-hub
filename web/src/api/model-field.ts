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

export interface ModelField {
  id: number
  model_id: number
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description: string
  sort_order: number
}

export function getModelFields(modelId: number) {
  return request<{ success: boolean; data: ModelField[] }>(`${BASE}/models/${modelId}/fields`)
}

export function createModelField(modelId: number, data: {
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description?: string
  sort_order?: number
}) {
  return request<{ success: boolean; message: string; data?: ModelField }>(`${BASE}/models/${modelId}/fields`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateModelField(modelId: number, fieldId: number, data: {
  field_key?: string
  field_name?: string
  field_type?: string
  required?: boolean
  description?: string
  sort_order?: number
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${modelId}/fields/${fieldId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deleteModelField(modelId: number, fieldId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/models/${modelId}/fields/${fieldId}`, {
    method: 'DELETE',
  })
}
