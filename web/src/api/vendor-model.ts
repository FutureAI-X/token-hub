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

export interface VendorModel {
  id: number
  vendor_id: number
  model_id: number
  vendor_model_id: string
  request_path: string
  is_async: boolean
  status: number
  created_at: string
  vendor_name: string
  model_name: string
}

export function getVendorModels() {
  return request<{ success: boolean; data: VendorModel[] }>(`${BASE}/vendor-models`)
}

export function createVendorModel(data: {
  vendor_id: number
  model_id: number
  vendor_model_id: string
  request_path?: string
  is_async?: boolean
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateVendorModel(id: number, data: {
  vendor_id?: number
  model_id?: number
  vendor_model_id?: string
  request_path?: string
  is_async?: boolean
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function updateVendorModelStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function deleteVendorModel(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models/${id}`, {
    method: 'DELETE',
  })
}

// ── 供应商模型字段 ──

export interface VendorModelField {
  id: number
  vendor_model_id: number
  model_field_id: number | null
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description: string
  sort_order: number
}

export function getVendorModelFields(vmId: number) {
  return request<{ success: boolean; data: VendorModelField[] }>(`${BASE}/vendor-models/${vmId}/fields`)
}

export function syncFieldsFromModel(vmId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models/${vmId}/fields/sync`, {
    method: 'POST',
  })
}

export function createVMField(vmId: number, data: {
  field_key: string
  field_name: string
  field_type: string
  required: boolean
  description?: string
  model_field_id?: number | null
}) {
  return request<{ success: boolean; message: string; data?: VendorModelField }>(`${BASE}/vendor-models/${vmId}/fields`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateVMField(vmId: number, fieldId: number, data: {
  field_key?: string
  field_name?: string
  field_type?: string
  required?: boolean
  description?: string
  model_field_id?: number | null
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models/${vmId}/fields/${fieldId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deleteVMField(vmId: number, fieldId: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendor-models/${vmId}/fields/${fieldId}`, {
    method: 'DELETE',
  })
}
