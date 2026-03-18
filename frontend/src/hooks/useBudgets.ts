import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { budgetApi } from '@/api/budget'
import { CreateBudgetRequest, UpdateBudgetRequest } from '@/types/budget'

export const useBudgets = () => {
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery({
    queryKey: ['budgets'],
    queryFn: () => budgetApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateBudgetRequest) => budgetApi.create(data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['budgets'] })
      await queryClient.invalidateQueries({ queryKey: ['budget-compare'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateBudgetRequest }) =>
      budgetApi.update(id, data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['budgets'] })
      await queryClient.invalidateQueries({ queryKey: ['budget-compare'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => budgetApi.delete(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['budgets'] })
      await queryClient.invalidateQueries({ queryKey: ['budget-compare'] })
    },
  })

  return {
    budgets: data?.budgets || [],
    isLoading,
    error,
    createBudget: createMutation.mutate,
    createBudgetAsync: createMutation.mutateAsync,
    updateBudget: updateMutation.mutate,
    updateBudgetAsync: updateMutation.mutateAsync,
    deleteBudget: deleteMutation.mutate,
    isCreating: createMutation.isPending,
    isUpdating: updateMutation.isPending,
    isDeleting: deleteMutation.isPending,
  }
}

export const useBudgetComparison = (year?: number, month?: number) => {
  return useQuery({
    queryKey: ['budget-compare', year, month],
    queryFn: () => budgetApi.compare(year, month),
  })
}
