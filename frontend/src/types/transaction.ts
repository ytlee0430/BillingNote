export interface Transaction {
  id: number
  user_id: number
  category_id?: number
  amount: number
  type: 'income' | 'expense'
  description: string
  transaction_date: string
  source: string
  tags: string[]
  created_at: string
  updated_at: string
  category?: Category
}

export interface Category {
  id: number
  name: string
  type: 'income' | 'expense'
  icon?: string
  color?: string
  created_at: string
}

export interface CreateTransactionRequest {
  category_id?: number
  amount: number
  type: 'income' | 'expense'
  description: string
  transaction_date: string
  source?: string
  tags?: string[]
}

export interface UpdateTransactionRequest {
  category_id?: number
  amount?: number
  type?: 'income' | 'expense'
  description?: string
  transaction_date?: string
  tags?: string[]
}

export interface TransactionFilter {
  type?: 'income' | 'expense'
  start_date?: string
  end_date?: string
  category_id?: number
  q?: string
  tags?: string
  min_amount?: number
  max_amount?: number
  page?: number
  page_size?: number
}

export interface TransactionListResponse {
  data: Transaction[]
  total: number
  page: number
  page_size: number
}
