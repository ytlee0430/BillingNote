import { apiClient } from './client'
import {
  Transaction,
  CreateTransactionRequest,
  UpdateTransactionRequest,
  TransactionFilter,
  TransactionListResponse,
} from '@/types/transaction'
import { MonthlyStats, CategoryStats } from '@/types/api'

export const transactionsApi = {
  list: async (filter: TransactionFilter = {}) => {
    const response = await apiClient.get<TransactionListResponse>('/api/transactions', filter)
    return response.data
  },

  get: async (id: number) => {
    const response = await apiClient.get<Transaction>(`/api/transactions/${id}`)
    return response.data
  },

  create: async (data: CreateTransactionRequest) => {
    const response = await apiClient.post<Transaction>('/api/transactions', data)
    return response.data
  },

  update: async (id: number, data: UpdateTransactionRequest) => {
    const response = await apiClient.put<Transaction>(`/api/transactions/${id}`, data)
    return response.data
  },

  delete: async (id: number) => {
    const response = await apiClient.delete(`/api/transactions/${id}`)
    return response.data
  },

  getMonthlyStats: async (year?: number, month?: number) => {
    const params: any = {}
    if (year) params.year = year
    if (month) params.month = month
    const response = await apiClient.get<MonthlyStats>('/api/stats/monthly', params)
    return response.data
  },

  getCategoryStats: async (startDate: string, endDate: string, type?: 'income' | 'expense') => {
    const params: any = { start_date: startDate, end_date: endDate }
    if (type) params.type = type
    const response = await apiClient.get<CategoryStats[]>('/api/stats/category', params)
    return response.data
  },
}
