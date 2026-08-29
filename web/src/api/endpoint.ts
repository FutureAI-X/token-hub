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

// ── 端点 ──

export interface Endpoint {
  id: number
  path: string
  name: string
  description: string
  status: number
  created_at: string
}

export function getEndpoints() {
  return request<{ success: boolean; data: Endpoint[] }>(`${BASE}/endpoints`)
}

export function createEndpoint(data: { path: string; name: string; description?: string }) {
  return request<{ success: boolean; message: string }>(`${BASE}/endpoints`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateEndpoint(id: number, data: { path?: string; name?: string; description?: string }) {
  return request<{ success: boolean; message: string }>(`${BASE}/endpoints/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function updateEndpointStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/endpoints/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function deleteEndpoint(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/endpoints/${id}`, {
    method: 'DELETE',
  })
}

// ── 端点字段 ──

export interface EndpointField {
  id: number
  endpoint_id: number
  section: string
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description: string
  sort_order: number
}

export function getEndpointFields(endpointId: number, section?: string) {
  const query = section ? `?section=${section}` : ''
  return request<{ success: boolean; data: EndpointField[] }>(`${BASE}/endpoints/${endpointId}/fields${query}`)
}

export function createEndpointField(endpointId: number, data: {
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description?: string
  sort_order?: number
  section?: string
}) {
  return request<{ success: boolean; message: string; data?: EndpointField }>(`${BASE}/endpoints/${endpointId}/fields`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateEndpointField(endpointId: number, fieldId: number, data: {
  field_key?: string
  field_name?: string
  field_type?: string
  required?: boolean
  description?: string
  sort_order?: number
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/endpoints/${endpointId}/fields/${fieldId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deleteEndpointField(endpointId: number, fieldId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/endpoints/${endpointId}/fields/${fieldId}`, {
    method: 'DELETE',
  })
}
