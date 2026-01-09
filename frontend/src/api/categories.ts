import { apiClient } from './client'
import { Category } from '@/types/transaction'

export const categoriesApi = {
  getAll: async () => {
    const response = await apiClient.get<Category[]>('/api/categories')
    return response.data
  },

  getByType: async (type: 'income' | 'expense') => {
    const response = await apiClient.get<Category[]>(`/api/categories/type/${type}`)
    return response.data
  },
}
