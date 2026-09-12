const BASE = '/api/admin'

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  // options 先展开、headers 最后合并：若反过来，任何传入 headers 的调用
  // 都会整体覆盖掉 Authorization，导致 401
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
