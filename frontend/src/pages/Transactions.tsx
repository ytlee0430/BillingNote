import { useState } from 'react'
import { TransactionList } from '@/components/transaction/TransactionList'
import { TransactionModal } from '@/components/transaction/TransactionModal'
import { Button } from '@/components/common/Button'
import { useTransactions } from '@/hooks/useTransactions'
import { TransactionFilter, Transaction } from '@/types/transaction'

export const Transactions = () => {
  const [filter, setFilter] = useState<TransactionFilter>({
    page: 1,
    page_size: 10,
  })
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingTransaction, setEditingTransaction] = useState<Transaction | null>(null)

  const {
    transactions,
    total,
    page,
    pageSize,
    isLoading,
    createTransactionAsync,
    updateTransactionAsync,
    deleteTransaction,
    isCreating,
    isUpdating,
  } = useTransactions(filter)

  const handleCreate = () => {
    setEditingTransaction(null)
    setIsModalOpen(true)
  }

  const handleEdit = (transaction: Transaction) => {
    setEditingTransaction(transaction)
    setIsModalOpen(true)
  }

  const handleDelete = (id: number) => {
    if (window.confirm('Are you sure you want to delete this transaction?')) {
      deleteTransaction(id)
    }
  }

  const handleFilterChange = (newFilter: Partial<TransactionFilter>) => {
    setFilter((prev) => ({ ...prev, ...newFilter, page: 1 }))
  }

  const handlePageChange = (newPage: number) => {
    setFilter((prev) => ({ ...prev, page: newPage }))
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Transactions</h1>
        <Button onClick={handleCreate}>Add Transaction</Button>
      </div>

      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">Filters</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Type
            </label>
            <select
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={filter.type || ''}
              onChange={(e) =>
                handleFilterChange({
                  type: e.target.value as 'income' | 'expense' | undefined || undefined,
                })
              }
            >
              <option value="">All</option>
              <option value="income">Income</option>
              <option value="expense">Expense</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Start Date
            </label>
            <input
              type="date"
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={filter.start_date || ''}
              onChange={(e) =>
                handleFilterChange({ start_date: e.target.value || undefined })
              }
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              End Date
            </label>
            <input
              type="date"
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={filter.end_date || ''}
              onChange={(e) =>
                handleFilterChange({ end_date: e.target.value || undefined })
              }
            />
          </div>

          <div className="flex items-end">
            <Button
              variant="secondary"
              onClick={() => setFilter({ page: 1, page_size: 10 })}
            >
              Clear Filters
            </Button>
          </div>
        </div>
      </div>

      <TransactionList
        transactions={transactions}
        total={total}
        page={page}
        pageSize={pageSize}
        isLoading={isLoading}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onPageChange={handlePageChange}
      />

      <TransactionModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        transaction={editingTransaction}
        isSaving={isCreating || isUpdating}
        onSave={async (data) => {
          try {
            if (editingTransaction) {
              await updateTransactionAsync({ id: editingTransaction.id, data })
            } else {
              await createTransactionAsync(data)
            }
            setIsModalOpen(false)
          } catch (error) {
            console.error('Failed to save transaction:', error)
          }
        }}
      />
    </div>
  )
}
