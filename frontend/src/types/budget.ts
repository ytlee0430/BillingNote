import { Category } from './transaction'

export interface Budget {
  id: number
  user_id: number
  category_id: number
  monthly_amount: number
  created_at: string
  updated_at: string
  category?: Category
}

export interface CreateBudgetRequest {
  category_id: number
  monthly_amount: number
}

export interface UpdateBudgetRequest {
  monthly_amount: number
}

export interface BudgetComparison {
  budget: Budget
  actual_amount: number
  remaining: number
  percentage: number
  is_over_budget: boolean
}
