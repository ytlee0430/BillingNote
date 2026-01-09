import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { CreateTransactionRequest, Transaction } from '@/types/transaction'
import { categoriesApi } from '@/api/categories'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/Input'

interface TransactionFormProps {
  transaction?: Transaction | null
  onSubmit: (data: CreateTransactionRequest) => void | Promise<void>
  onCancel: () => void
  isSubmitting?: boolean
}

export const TransactionForm = ({
  transaction,
  onSubmit,
  onCancel,
  isSubmitting = false,
}: TransactionFormProps) => {
  const [formData, setFormData] = useState<CreateTransactionRequest>({
    amount: 0,
    type: 'expense',
    description: '',
    transaction_date: new Date().toISOString().split('T')[0],
    category_id: undefined,
    source: 'manual',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.getAll(),
  })

  useEffect(() => {
    if (transaction) {
      setFormData({
        amount: transaction.amount,
        type: transaction.type,
        description: transaction.description,
        transaction_date: transaction.transaction_date.split('T')[0],
        category_id: transaction.category_id,
        source: transaction.source,
      })
    }
  }, [transaction])

  const filteredCategories = categories?.filter(
    (cat) => cat.type === formData.type
  )

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.amount || formData.amount <= 0) {
      newErrors.amount = 'Amount must be greater than 0'
    }

    if (!formData.transaction_date) {
      newErrors.transaction_date = 'Transaction date is required'
    }

    if (!formData.description.trim()) {
      newErrors.description = 'Description is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    // Convert date to ISO format for backend
    const submitData = {
      ...formData,
      transaction_date: new Date(formData.transaction_date).toISOString(),
    }

    onSubmit(submitData)
  }

  const handleChange = (
    field: keyof CreateTransactionRequest,
    value: any
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }))

    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev }
        delete newErrors[field]
        return newErrors
      })
    }

    // Reset category when type changes
    if (field === 'type') {
      setFormData((prev) => ({
        ...prev,
        category_id: undefined,
      }))
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Type <span className="text-red-500">*</span>
        </label>
        <div className="flex space-x-4">
          <label className="flex items-center">
            <input
              type="radio"
              value="expense"
              checked={formData.type === 'expense'}
              onChange={(e) => handleChange('type', e.target.value)}
              className="mr-2"
            />
            <span>Expense</span>
          </label>
          <label className="flex items-center">
            <input
              type="radio"
              value="income"
              checked={formData.type === 'income'}
              onChange={(e) => handleChange('type', e.target.value)}
              className="mr-2"
            />
            <span>Income</span>
          </label>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Amount <span className="text-red-500">*</span>
        </label>
        <Input
          type="number"
          step="0.01"
          value={formData.amount}
          onChange={(e) => handleChange('amount', parseFloat(e.target.value))}
          error={errors.amount}
          placeholder="0.00"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Category
        </label>
        <select
          className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          value={formData.category_id || ''}
          onChange={(e) =>
            handleChange(
              'category_id',
              e.target.value ? parseInt(e.target.value) : undefined
            )
          }
        >
          <option value="">No category</option>
          {filteredCategories?.map((category) => (
            <option key={category.id} value={category.id}>
              {category.icon} {category.name}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Description <span className="text-red-500">*</span>
        </label>
        <Input
          type="text"
          value={formData.description}
          onChange={(e) => handleChange('description', e.target.value)}
          error={errors.description}
          placeholder="Enter description"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Date <span className="text-red-500">*</span>
        </label>
        <Input
          type="date"
          value={formData.transaction_date}
          onChange={(e) => handleChange('transaction_date', e.target.value)}
          error={errors.transaction_date}
        />
      </div>

      <div className="flex justify-end space-x-2 pt-4">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={isSubmitting}>
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Saving...' : `${transaction ? 'Update' : 'Create'} Transaction`}
        </Button>
      </div>
    </form>
  )
}
