import { apiClient } from './client'

export interface CategoryKeyword {
  id: number
  user_id: number
  category_id: number
  keyword: string
  created_at: string
  category?: {
    id: number
    name: string
    type: string
    icon: string
    color: string
  }
}

export const categoryKeywordsApi = {
  list: async () => {
    const response = await apiClient.get<CategoryKeyword[]>('/api/category-keywords')
    return response.data
  },

  add: async (categoryId: number, keyword: string) => {
    const response = await apiClient.post<CategoryKeyword>('/api/category-keywords', {
      category_id: categoryId,
      keyword,
    })
    return response.data
  },

  remove: async (id: number) => {
    await apiClient.delete(`/api/category-keywords/${id}`)
  },

  batchSet: async (categoryId: number, keywords: string[]) => {
    await apiClient.put('/api/category-keywords/batch', {
      category_id: categoryId,
      keywords,
    })
  },

  initDefaults: async () => {
    await apiClient.post('/api/category-keywords/init-defaults')
  },

  reclassify: async () => {
    const response = await apiClient.post<{ updated: number }>('/api/category-keywords/reclassify')
    return response.data
  },
}
