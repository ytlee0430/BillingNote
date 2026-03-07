import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { invoicesApi } from '@/api/invoices'
import {
  InvoiceFilter,
  InvoiceSyncRequest,
  ConfirmDuplicateRequest,
  InvoiceSettingsInput,
} from '@/types/invoice'

export const useInvoices = (filter: InvoiceFilter = {}) => {
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery({
    queryKey: ['invoices', filter],
    queryFn: () => invoicesApi.list(filter),
  })

  const syncMutation = useMutation({
    mutationFn: (data: InvoiceSyncRequest) => invoicesApi.sync(data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['invoices'] })
    },
  })

  const confirmDuplicateMutation = useMutation({
    mutationFn: (data: ConfirmDuplicateRequest) => invoicesApi.confirmDuplicate(data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['invoices'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => invoicesApi.delete(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['invoices'] })
    },
  })

  const updateSettingsMutation = useMutation({
    mutationFn: (data: InvoiceSettingsInput) => invoicesApi.updateSettings(data),
  })

  return {
    invoices: data?.invoices || [],
    total: data?.total || 0,
    page: data?.page || 1,
    pageSize: data?.page_size || 20,
    isLoading,
    error,
    syncInvoices: syncMutation.mutate,
    syncInvoicesAsync: syncMutation.mutateAsync,
    isSyncing: syncMutation.isPending,
    syncError: syncMutation.error,
    confirmDuplicate: confirmDuplicateMutation.mutate,
    confirmDuplicateAsync: confirmDuplicateMutation.mutateAsync,
    isConfirming: confirmDuplicateMutation.isPending,
    deleteInvoice: deleteMutation.mutate,
    isDeleting: deleteMutation.isPending,
    updateSettings: updateSettingsMutation.mutate,
    updateSettingsAsync: updateSettingsMutation.mutateAsync,
    isUpdatingSettings: updateSettingsMutation.isPending,
  }
}
