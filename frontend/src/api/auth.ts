import { apiClient } from './client'
import { LoginRequest, RegisterRequest, AuthResponse } from '@/types/auth'

export const authApi = {
  login: async (data: LoginRequest) => {
    const response = await apiClient.post<AuthResponse>('/api/auth/login', data)
    return response.data
  },

  register: async (data: RegisterRequest) => {
    const response = await apiClient.post<AuthResponse>('/api/auth/register', data)
    return response.data
  },

  me: async () => {
    const response = await apiClient.get('/api/auth/me')
    return response.data
  },
}
