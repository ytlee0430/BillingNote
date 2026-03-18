import { apiClient } from './client'
import { PairingCodeResponse, ConnectionsResponse } from '@/types/sharing'

export const sharingApi = {
  getMyCode: async () => {
    const response = await apiClient.get<PairingCodeResponse>('/api/shared/my-code')
    return response.data
  },

  regenerateCode: async () => {
    const response = await apiClient.post<PairingCodeResponse>('/api/shared/regenerate-code')
    return response.data
  },

  pair: async (code: string) => {
    const response = await apiClient.post('/api/shared/pair', { code })
    return response.data
  },

  getConnections: async () => {
    const response = await apiClient.get<ConnectionsResponse>('/api/shared/connections')
    return response.data
  },

  revokeAccess: async (uid: number) => {
    const response = await apiClient.delete(`/api/shared/connections/${uid}`)
    return response.data
  },
}
