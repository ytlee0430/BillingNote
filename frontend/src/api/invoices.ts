import { apiClient } from './client'
import {
  InvoiceFilter,
  InvoiceListResponse,
  InvoiceSyncRequest,
  InvoiceSyncResponse,
  ConfirmDuplicateRequest,
  InvoiceSettingsInput,
} from '@/types/invoice'

export const invoicesApi = {
  list: async (filter: InvoiceFilter = {}) => {
    const response = await apiClient.get<InvoiceListResponse>('/api/invoice/list', filter)
    return response.data
  },

  sync: async (data: InvoiceSyncRequest) => {
    const response = await apiClient.post<InvoiceSyncResponse>('/api/invoice/sync', data)
    return response.data
  },

  confirmDuplicate: async (data: ConfirmDuplicateRequest) => {
    const response = await apiClient.post<{ message: string }>('/api/invoice/confirm-duplicate', data)
    return response.data
  },

  delete: async (id: number) => {
    const response = await apiClient.delete(`/api/invoice/${id}`)
    return response.data
  },

  updateSettings: async (data: InvoiceSettingsInput) => {
    const response = await apiClient.put<{ message: string }>('/api/invoice/settings', data)
    return response.data
  },
}
