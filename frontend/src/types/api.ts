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

export interface TrendDataPoint {
  date: string
  income: number
  expense: number
}

export interface TrendStatsResponse {
  data: TrendDataPoint[]
}
