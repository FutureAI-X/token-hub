export interface PricingModel {
  id: number
  name: string
  description?: string
  tags?: string
  owner: string
  status: number
}

export interface PricingData {
  models: PricingModel[]
}
