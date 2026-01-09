export interface ApiError {
  error: string
}

export interface MonthlyStats {
  income: number
  expense: number
  balance: number
}

export interface CategoryStats {
  category_id?: number
  category_name: string
  amount: number
}
