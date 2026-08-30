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
