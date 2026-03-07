import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { InvoiceList } from './InvoiceList'
import { Invoice } from '@/types/invoice'

describe('InvoiceList', () => {
  const mockInvoices: Invoice[] = [
    {
      id: 1,
      user_id: 1,
      invoice_number: 'AB12345678',
      invoice_date: '2026-01-15T00:00:00Z',
      seller_name: '全家便利商店',
      seller_ban: '12345678',
      amount: 150,
      status: '已使用',
      items: null,
      is_duplicated: false,
      created_at: '2026-01-15T00:00:00Z',
    },
    {
      id: 2,
      user_id: 1,
      invoice_number: 'CD87654321',
      invoice_date: '2026-01-16T00:00:00Z',
      seller_name: '7-ELEVEN',
      seller_ban: '87654321',
      amount: 89,
      status: '已使用',
      items: null,
      is_duplicated: true,
      duplicated_transaction_id: 42,
      confidence_score: 0.95,
      created_at: '2026-01-16T00:00:00Z',
      duplicated_transaction: {
        id: 42,
        amount: 89,
        description: '7-ELEVEN',
        transaction_date: '2026-01-16T00:00:00Z',
        source: 'pdf',
      },
    },
  ]

  const defaultProps = {
    invoices: mockInvoices,
    total: 2,
    page: 1,
    pageSize: 20,
    isLoading: false,
    onViewDuplicate: vi.fn(),
    onDelete: vi.fn(),
    onPageChange: vi.fn(),
  }

  it('renders invoice list correctly', () => {
    render(<InvoiceList {...defaultProps} />)

    expect(screen.getByText('AB12345678')).toBeInTheDocument()
    expect(screen.getByText('CD87654321')).toBeInTheDocument()
    expect(screen.getByText('全家便利商店')).toBeInTheDocument()
    expect(screen.getByText('7-ELEVEN')).toBeInTheDocument()
  })

  it('displays loading state', () => {
    render(<InvoiceList {...defaultProps} isLoading={true} />)

    expect(screen.getByText('Loading invoices...')).toBeInTheDocument()
  })

  it('displays empty state', () => {
    render(<InvoiceList {...defaultProps} invoices={[]} />)

    expect(screen.getByText(/No invoices found/)).toBeInTheDocument()
  })

  it('shows duplicate badge with confidence score', () => {
    render(<InvoiceList {...defaultProps} />)

    expect(screen.getByText(/Duplicate/)).toBeInTheDocument()
    expect(screen.getByText(/(95%)/)).toBeInTheDocument()
  })

  it('shows unique badge for non-duplicated invoices', () => {
    render(<InvoiceList {...defaultProps} />)

    expect(screen.getByText('Unique')).toBeInTheDocument()
  })

  it('shows View Match button only for duplicated invoices', () => {
    render(<InvoiceList {...defaultProps} />)

    const viewMatchButtons = screen.getAllByText('View Match')
    expect(viewMatchButtons).toHaveLength(1)
  })

  it('calls onViewDuplicate when View Match is clicked', () => {
    render(<InvoiceList {...defaultProps} />)

    fireEvent.click(screen.getByText('View Match'))

    expect(defaultProps.onViewDuplicate).toHaveBeenCalledWith(mockInvoices[1])
  })

  it('calls onDelete when Delete is clicked', () => {
    render(<InvoiceList {...defaultProps} />)

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(defaultProps.onDelete).toHaveBeenCalledWith(1)
  })

  it('renders pagination when multiple pages', () => {
    render(<InvoiceList {...defaultProps} total={50} pageSize={20} />)

    expect(screen.getAllByText('Previous').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Next').length).toBeGreaterThan(0)
  })

  it('calls onPageChange when pagination button is clicked', () => {
    render(<InvoiceList {...defaultProps} total={50} pageSize={20} page={1} />)

    const nextButton = screen.getAllByText('Next')[0]
    fireEvent.click(nextButton)

    expect(defaultProps.onPageChange).toHaveBeenCalledWith(2)
  })
})
