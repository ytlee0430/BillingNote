import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { DuplicateHandler } from './DuplicateHandler'
import { Invoice } from '@/types/invoice'

describe('DuplicateHandler', () => {
  const mockInvoice: Invoice = {
    id: 1,
    user_id: 1,
    invoice_number: 'AB12345678',
    invoice_date: '2026-01-15T00:00:00Z',
    seller_name: '全家便利商店',
    seller_ban: '12345678',
    amount: 150,
    status: '已使用',
    items: null,
    is_duplicated: true,
    duplicated_transaction_id: 42,
    confidence_score: 0.92,
    created_at: '2026-01-15T00:00:00Z',
    duplicated_transaction: {
      id: 42,
      amount: 150,
      description: '全家便利商店',
      transaction_date: '2026-01-15T00:00:00Z',
      source: 'pdf',
    },
  }

  const defaultProps = {
    invoice: mockInvoice,
    isOpen: true,
    onClose: vi.fn(),
    onConfirm: vi.fn(),
    onDismiss: vi.fn(),
    isConfirming: false,
  }

  it('renders invoice details', () => {
    render(<DuplicateHandler {...defaultProps} />)

    expect(screen.getByText('AB12345678')).toBeInTheDocument()
    expect(screen.getAllByText('全家便利商店')).toHaveLength(2) // invoice seller + transaction description
  })

  it('renders matched transaction details', () => {
    render(<DuplicateHandler {...defaultProps} />)

    expect(screen.getByText('Matched Transaction')).toBeInTheDocument()
    expect(screen.getByText(/92% confidence/)).toBeInTheDocument()
  })

  it('shows confirm and dismiss buttons', () => {
    render(<DuplicateHandler {...defaultProps} />)

    expect(screen.getByText('Confirm Duplicate')).toBeInTheDocument()
    expect(screen.getByText('Not a Duplicate')).toBeInTheDocument()
  })

  it('calls onConfirm with correct IDs', () => {
    render(<DuplicateHandler {...defaultProps} />)

    fireEvent.click(screen.getByText('Confirm Duplicate'))

    expect(defaultProps.onConfirm).toHaveBeenCalledWith(1, 42)
  })

  it('calls onDismiss when Not a Duplicate is clicked', () => {
    render(<DuplicateHandler {...defaultProps} />)

    fireEvent.click(screen.getByText('Not a Duplicate'))

    expect(defaultProps.onDismiss).toHaveBeenCalledWith(1)
  })

  it('does not render when invoice is null', () => {
    const { container } = render(
      <DuplicateHandler {...defaultProps} invoice={null} />
    )

    expect(container.innerHTML).toBe('')
  })

  it('does not render when isOpen is false', () => {
    render(<DuplicateHandler {...defaultProps} isOpen={false} />)

    expect(screen.queryByText('Duplicate Match Detail')).not.toBeInTheDocument()
  })

  it('shows no match message when transaction is missing', () => {
    const invoiceNoTxn = { ...mockInvoice, duplicated_transaction: undefined }
    render(<DuplicateHandler {...defaultProps} invoice={invoiceNoTxn} />)

    expect(screen.getByText(/No matched transaction/)).toBeInTheDocument()
    expect(screen.queryByText('Confirm Duplicate')).not.toBeInTheDocument()
  })
})
