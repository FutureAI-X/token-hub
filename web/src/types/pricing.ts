export interface PricingVendor {
  id: number
  name: string
  description?: string
  icon?: string
}

export interface PricingModel {
  id: number
  name: string
  description?: string
  icon?: string
  tags?: string
  vendor_id?: number
  vendor_name?: string
  quota_type: number // 0=按量计费, 1=按次计费
  model_ratio: number
  completion_ratio: number
  model_price: number
  context_length: number
  max_output_tokens: number
}

export interface PricingData {
  models: PricingModel[]
  vendors: PricingVendor[]
}
