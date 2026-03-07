import { useState, useEffect } from 'react'
import { Button } from '@/components/common/Button'
import { Modal } from '@/components/common/Modal'
import { useBudgets, useBudgetComparison } from '@/hooks/useBudgets'
import { formatCurrency } from '@/utils/format'
import { categoriesApi } from '@/api/categories'
import { Category } from '@/types/transaction'

export const Budget = () => {
  const currentDate = new Date()
  const [selectedYear, setSelectedYear] = useState(currentDate.getFullYear())
  const [selectedMonth, setSelectedMonth] = useState(currentDate.getMonth() + 1)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [categories, setCategories] = useState<Category[]>([])
  const [newCategoryId, setNewCategoryId] = useState<number>(0)
  const [newAmount, setNewAmount] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editAmount, setEditAmount] = useState('')

  const {
    budgets,
    isLoading,
    createBudgetAsync,
    updateBudgetAsync,
    deleteBudget,
    isCreating,
    isUpdating,
  } = useBudgets()

  const { data: compareData, isLoading: isCompareLoading } = useBudgetComparison(
    selectedYear,
    selectedMonth
  )

  useEffect(() => {
    categoriesApi.getByType('expense').then(setCategories).catch(console.error)
  }, [])

  const handleCreate = async () => {
    if (!newCategoryId || !newAmount) return
    try {
      await createBudgetAsync({
        category_id: newCategoryId,
        monthly_amount: parseFloat(newAmount),
      })
      setIsModalOpen(false)
      setNewCategoryId(0)
      setNewAmount('')
    } catch (error) {
      console.error('Failed to create budget:', error)
    }
  }

  const handleUpdate = async (id: number) => {
    if (!editAmount) return
    try {
      await updateBudgetAsync({ id, data: { monthly_amount: parseFloat(editAmount) } })
      setEditingId(null)
      setEditAmount('')
    } catch (error) {
      console.error('Failed to update budget:', error)
    }
  }

  const handleDelete = (id: number) => {
    if (window.confirm('Are you sure you want to delete this budget?')) {
      deleteBudget(id)
    }
  }

  const usedCategoryIds = new Set(budgets.map((b) => b.category_id))
  const availableCategories = categories.filter((c) => !usedCategoryIds.has(c.id))

  const years = Array.from({ length: 5 }, (_, i) => currentDate.getFullYear() - i)
  const months = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December',
  ]

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Budget</h1>
        <Button onClick={() => setIsModalOpen(true)} disabled={availableCategories.length === 0}>
          Add Budget
        </Button>
      </div>

      {/* Period Selector */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">Period</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <select
            className="w-full border border-gray-300 rounded-md px-3 py-2"
            value={selectedYear}
            onChange={(e) => setSelectedYear(parseInt(e.target.value))}
          >
            {years.map((year) => (
              <option key={year} value={year}>{year}</option>
            ))}
          </select>
          <select
            className="w-full border border-gray-300 rounded-md px-3 py-2"
            value={selectedMonth}
            onChange={(e) => setSelectedMonth(parseInt(e.target.value))}
          >
            {months.map((month, i) => (
              <option key={i} value={i + 1}>{month}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Budget Comparison */}
      {isCompareLoading || isLoading ? (
        <div className="bg-white shadow rounded-lg p-8 text-center">
          <div className="text-gray-500">Loading budgets...</div>
        </div>
      ) : compareData?.comparisons && compareData.comparisons.length > 0 ? (
        <div className="space-y-4">
          {compareData.comparisons.map((comp) => (
            <div key={comp.budget.id} className="bg-white shadow rounded-lg p-6">
              <div className="flex justify-between items-start mb-3">
                <div>
                  <h3 className="font-semibold text-gray-900">
                    {comp.budget.category?.icon} {comp.budget.category?.name || 'Unknown'}
                  </h3>
                  <p className="text-sm text-gray-500">
                    Budget: {formatCurrency(comp.budget.monthly_amount)}
                  </p>
                </div>
                <div className="text-right">
                  {editingId === comp.budget.id ? (
                    <div className="flex gap-2">
                      <input
                        type="number"
                        value={editAmount}
                        onChange={(e) => setEditAmount(e.target.value)}
                        className="w-32 border border-gray-300 rounded-md px-2 py-1 text-sm"
                      />
                      <Button
                        size="sm"
                        onClick={() => handleUpdate(comp.budget.id)}
                        loading={isUpdating}
                      >
                        Save
                      </Button>
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => setEditingId(null)}
                      >
                        Cancel
                      </Button>
                    </div>
                  ) : (
                    <div className="flex gap-2">
                      <button
                        onClick={() => {
                          setEditingId(comp.budget.id)
                          setEditAmount(String(comp.budget.monthly_amount))
                        }}
                        className="text-blue-600 hover:text-blue-900 text-sm"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(comp.budget.id)}
                        className="text-red-600 hover:text-red-900 text-sm"
                      >
                        Delete
                      </button>
                    </div>
                  )}
                </div>
              </div>

              {/* Progress Bar */}
              <div className="w-full bg-gray-200 rounded-full h-4 mb-2">
                <div
                  className={`h-4 rounded-full transition-all ${
                    comp.is_over_budget ? 'bg-red-500' : comp.percentage > 80 ? 'bg-yellow-500' : 'bg-green-500'
                  }`}
                  style={{ width: `${Math.min(comp.percentage, 100)}%` }}
                />
              </div>

              <div className="flex justify-between text-sm">
                <span className={comp.is_over_budget ? 'text-red-600 font-semibold' : 'text-gray-600'}>
                  Spent: {formatCurrency(comp.actual_amount)} ({comp.percentage}%)
                </span>
                <span className={comp.remaining < 0 ? 'text-red-600' : 'text-green-600'}>
                  {comp.remaining >= 0 ? `Remaining: ${formatCurrency(comp.remaining)}` : `Over: ${formatCurrency(Math.abs(comp.remaining))}`}
                </span>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="bg-white shadow rounded-lg p-8 text-center">
          <div className="text-gray-500">
            {budgets.length === 0
              ? 'No budgets set. Click "Add Budget" to get started.'
              : 'No budget data for this period.'}
          </div>
        </div>
      )}

      {/* Add Budget Modal */}
      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} title="Add Budget">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
            <select
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={newCategoryId}
              onChange={(e) => setNewCategoryId(parseInt(e.target.value))}
            >
              <option value={0}>Select category</option>
              {availableCategories.map((cat) => (
                <option key={cat.id} value={cat.id}>
                  {cat.icon} {cat.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Monthly Amount</label>
            <input
              type="number"
              value={newAmount}
              onChange={(e) => setNewAmount(e.target.value)}
              placeholder="Enter amount"
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              loading={isCreating}
              disabled={!newCategoryId || !newAmount}
            >
              Create
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
