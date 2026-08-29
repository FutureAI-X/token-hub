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

export interface Vendor {
  id: number
  name: string
  description: string
  base_url: string
  api_key: string
  protocol_type: string
  status: number
  created_at: string
}

export function getVendors() {
  return request<{ success: boolean; data: Vendor[] }>(`${BASE}/vendors`)
}

export function createVendor(data: {
  name: string
  description?: string
  base_url: string
  api_key: string
  protocol_type: string
  data_key: string
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendors`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateVendor(id: number, data: {
  name?: string
  description?: string
  base_url?: string
  api_key?: string
  protocol_type?: string
  data_key?: string
}) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendors/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function updateVendorStatus(id: number, status: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendors/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function deleteVendor(id: number) {
  return request<{ success: boolean; message: string }>(`${BASE}/vendors/${id}`, {
    method: 'DELETE',
  })
}
