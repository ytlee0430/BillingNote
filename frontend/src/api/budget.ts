import { apiClient } from './client'
import {
  Budget,
  CreateBudgetRequest,
  UpdateBudgetRequest,
  BudgetComparison,
} from '@/types/budget'

export const budgetApi = {
  list: async () => {
    const response = await apiClient.get<{ budgets: Budget[] }>('/api/budget')
    return response.data
  },

  create: async (data: CreateBudgetRequest) => {
    const response = await apiClient.post<Budget>('/api/budget', data)
    return response.data
  },

  update: async (id: number, data: UpdateBudgetRequest) => {
    const response = await apiClient.put<Budget>(`/api/budget/${id}`, data)
    return response.data
  },

  delete: async (id: number) => {
    const response = await apiClient.delete(`/api/budget/${id}`)
    return response.data
  },

  compare: async (year?: number, month?: number) => {
    const params: any = {}
    if (year) params.year = year
    if (month) params.month = month
    const response = await apiClient.get<{ comparisons: BudgetComparison[] }>('/api/budget/compare', params)
    return response.data
  },
}
