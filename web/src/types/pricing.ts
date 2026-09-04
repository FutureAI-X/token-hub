export interface QuotaRuleItem {
  id: number
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
  items?: QuotaRuleItem[]
}

export interface PricingModel {
  id: number
  name: string
  description?: string
  tags?: string
  owner: string
  status: number
  quota_rule?: QuotaRule
}

export interface PricingData {
  models: PricingModel[]
}
