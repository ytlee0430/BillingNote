import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { transactionsApi } from '@/api/transactions'
import {
  TransactionFilter,
  CreateTransactionRequest,
  UpdateTransactionRequest,
} from '@/types/transaction'

export const useTransactions = (filter: TransactionFilter = {}) => {
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery({
    queryKey: ['transactions', filter],
    queryFn: () => transactionsApi.list(filter),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateTransactionRequest) => transactionsApi.create(data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions'] })
      await queryClient.invalidateQueries({ queryKey: ['monthly-stats'] })
      await queryClient.invalidateQueries({ queryKey: ['category-stats'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateTransactionRequest }) =>
      transactionsApi.update(id, data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions'] })
      await queryClient.invalidateQueries({ queryKey: ['monthly-stats'] })
      await queryClient.invalidateQueries({ queryKey: ['category-stats'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => transactionsApi.delete(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions'] })
      await queryClient.invalidateQueries({ queryKey: ['monthly-stats'] })
      await queryClient.invalidateQueries({ queryKey: ['category-stats'] })
    },
  })

  return {
    transactions: data?.data || [],
    total: data?.total || 0,
    page: data?.page || 1,
    pageSize: data?.page_size || 10,
    isLoading,
    error,
    createTransaction: createMutation.mutate,
    createTransactionAsync: createMutation.mutateAsync,
    updateTransaction: updateMutation.mutate,
    updateTransactionAsync: updateMutation.mutateAsync,
    deleteTransaction: deleteMutation.mutate,
    deleteTransactionAsync: deleteMutation.mutateAsync,
    isCreating: createMutation.isPending,
    isUpdating: updateMutation.isPending,
    isDeleting: deleteMutation.isPending,
  }
}

export const useMonthlyStats = (year?: number, month?: number) => {
  return useQuery({
    queryKey: ['monthly-stats', year, month],
    queryFn: () => transactionsApi.getMonthlyStats(year, month),
  })
}

export const useCategoryStats = (
  startDate: string,
  endDate: string,
  type?: 'income' | 'expense'
) => {
  return useQuery({
    queryKey: ['category-stats', startDate, endDate, type],
    queryFn: () => transactionsApi.getCategoryStats(startDate, endDate, type),
    enabled: !!startDate && !!endDate,
  })
}
