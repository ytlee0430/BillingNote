import { apiClient } from './client'
import {
  GmailStatus,
  GmailSettings,
  GmailSettingsInput,
  GmailScanResult,
  GmailCallbackRequest,
} from '@/types/gmail'

export const gmailApi = {
  getAuthURL: async () => {
    const response = await apiClient.get<{ url: string }>('/api/gmail/auth')
    return response.data
  },

  handleCallback: async (data: GmailCallbackRequest) => {
    const response = await apiClient.post<{ message: string }>('/api/gmail/callback', data)
    return response.data
  },

  getStatus: async () => {
    const response = await apiClient.get<GmailStatus>('/api/gmail/status')
    return response.data
  },

  getSettings: async () => {
    const response = await apiClient.get<GmailSettings>('/api/gmail/settings')
    return response.data
  },

  updateSettings: async (data: GmailSettingsInput) => {
    const response = await apiClient.put<{ message: string }>('/api/gmail/settings', data)
    return response.data
  },

  triggerScan: async () => {
    const response = await apiClient.post<GmailScanResult>('/api/gmail/scan')
    return response.data
  },

  disconnect: async () => {
    const response = await apiClient.delete('/api/gmail/disconnect')
    return response.data
  },
}
