import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { TransactionList } from './TransactionList'
import { Transaction } from '@/types/transaction'

describe('TransactionList', () => {
  const mockTransactions: Transaction[] = [
    {
      id: 1,
      user_id: 1,
      category_id: 1,
      amount: 100.50,
      type: 'expense',
      description: 'Test expense',
      transaction_date: '2024-01-15T00:00:00Z',
      source: 'manual',
      created_at: '2024-01-15T00:00:00Z',
      updated_at: '2024-01-15T00:00:00Z',
      category: {
        id: 1,
        name: 'Food',
        type: 'expense',
        icon: '🍔',
        created_at: '2024-01-01T00:00:00Z',
      },
    },
    {
      id: 2,
      user_id: 1,
      amount: 500.00,
      type: 'income',
      description: 'Test income',
      transaction_date: '2024-01-16T00:00:00Z',
      source: 'manual',
      created_at: '2024-01-16T00:00:00Z',
      updated_at: '2024-01-16T00:00:00Z',
    },
  ]

  const defaultProps = {
    transactions: mockTransactions,
    total: 2,
    page: 1,
    pageSize: 10,
    isLoading: false,
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onPageChange: vi.fn(),
  }

  it('renders transaction list correctly', () => {
    render(<TransactionList {...defaultProps} />)

    expect(screen.getByText('Test expense')).toBeInTheDocument()
    expect(screen.getByText('Test income')).toBeInTheDocument()
    expect(screen.getByText('Food')).toBeInTheDocument()
  })

  it('displays loading state', () => {
    render(<TransactionList {...defaultProps} isLoading={true} />)

    expect(screen.getByText('Loading transactions...')).toBeInTheDocument()
  })

  it('displays empty state when no transactions', () => {
    render(<TransactionList {...defaultProps} transactions={[]} />)

    expect(screen.getByText('No transactions found')).toBeInTheDocument()
  })

  it('calls onEdit when edit button is clicked', () => {
    render(<TransactionList {...defaultProps} />)

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    expect(defaultProps.onEdit).toHaveBeenCalledWith(mockTransactions[0])
  })

  it('calls onDelete when delete button is clicked', () => {
    render(<TransactionList {...defaultProps} />)

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(defaultProps.onDelete).toHaveBeenCalledWith(1)
  })

  it('displays correct transaction type badges', () => {
    render(<TransactionList {...defaultProps} />)

    expect(screen.getByText('expense')).toBeInTheDocument()
    expect(screen.getByText('income')).toBeInTheDocument()
  })

  it('displays amount with correct sign and color', () => {
    render(<TransactionList {...defaultProps} />)

    const amounts = screen.getAllByText(/\$\d+\.\d{2}/)
    expect(amounts.length).toBeGreaterThan(0)
  })

  it('shows no category text when category is not provided', () => {
    render(<TransactionList {...defaultProps} />)

    expect(screen.getByText('No category')).toBeInTheDocument()
  })

  it('renders pagination when multiple pages', () => {
    render(<TransactionList {...defaultProps} total={25} pageSize={10} />)

    expect(screen.getByText('Previous')).toBeInTheDocument()
    expect(screen.getByText('Next')).toBeInTheDocument()
  })

  it('calls onPageChange when pagination button is clicked', () => {
    render(<TransactionList {...defaultProps} total={25} pageSize={10} page={1} />)

    const nextButton = screen.getAllByText('Next')[0]
    fireEvent.click(nextButton)

    expect(defaultProps.onPageChange).toHaveBeenCalledWith(2)
  })

  it('disables previous button on first page', () => {
    render(<TransactionList {...defaultProps} total={25} pageSize={10} page={1} />)

    const previousButtons = screen.getAllByText('Previous')
    expect(previousButtons[0]).toBeDisabled()
  })

  it('disables next button on last page', () => {
    render(<TransactionList {...defaultProps} total={25} pageSize={10} page={3} />)

    const nextButtons = screen.getAllByText('Next')
    expect(nextButtons[0]).toBeDisabled()
  })
})
