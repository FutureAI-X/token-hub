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

// ── 供应商端点 ──

export interface VendorEndpoint {
  id: number
  vendor_id: number
  endpoint_id: number
  path: string
  name: string
  description: string
  is_async: boolean
  status: number
  created_at: string
  vendor_name: string
  endpoint_path: string
  endpoint_name: string
}

export function getVendorEndpoints() {
  return request<{ success: boolean; data: VendorEndpoint[] }>(`${BASE}/vendor-endpoints`)
}

export function createVendorEndpoint(data: {
  vendor_id: number
  endpoint_id: number
  path?: string
  name?: string
  description?: string
  is_async?: boolean
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateVendorEndpoint(id: number, data: {
  vendor_id?: number
  endpoint_id?: number
  path?: string
  name?: string
  description?: string
  is_async?: boolean
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function updateVendorEndpointStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function deleteVendorEndpoint(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints/${id}`, {
    method: 'DELETE',
  })
}

// ── 供应商端点字段 ──

export interface VendorEndpointField {
  id: number
  vendor_endpoint_id: number
  endpoint_field_id: number | null
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description: string
  sort_order: number
}

export function getVendorEndpointFields(veId: number) {
  return request<{ success: boolean; data: VendorEndpointField[] }>(`${BASE}/vendor-endpoints/${veId}/fields`)
}

export function syncVEFieldsFromEndpoint(veId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints/${veId}/fields/sync`, {
    method: 'POST',
  })
}

export function createVEField(veId: number, data: {
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description?: string
  endpoint_field_id?: number | null
}) {
  return request<{ success: boolean; message: string; data?: VendorEndpointField }>(`${BASE}/vendor-endpoints/${veId}/fields`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateVEField(veId: number, fieldId: number, data: {
  field_key?: string
  field_name?: string
  field_type?: string
  required?: boolean
  description?: string
  endpoint_field_id?: number | null
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints/${veId}/fields/${fieldId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deleteVEField(veId: number, fieldId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-endpoints/${veId}/fields/${fieldId}`, {
    method: 'DELETE',
  })
}
