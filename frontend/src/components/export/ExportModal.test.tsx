import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ExportModal } from './ExportModal'

vi.mock('@/api/transactions', () => ({
  transactionsApi: {
    exportCSV: vi.fn(),
  },
}))

import { transactionsApi } from '@/api/transactions'

describe('ExportModal', () => {
  const mockOnClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should not render when closed', () => {
    render(<ExportModal isOpen={false} onClose={mockOnClose} />)
    expect(screen.queryByText('Export Transactions')).not.toBeInTheDocument()
  })

  it('should render when open', () => {
    render(<ExportModal isOpen={true} onClose={mockOnClose} />)
    expect(screen.getByText('Export Transactions')).toBeInTheDocument()
    expect(screen.getByText('Start Date')).toBeInTheDocument()
    expect(screen.getByText('End Date')).toBeInTheDocument()
    expect(screen.getByText('Export CSV')).toBeInTheDocument()
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('should call onClose when Cancel is clicked', () => {
    render(<ExportModal isOpen={true} onClose={mockOnClose} />)
    fireEvent.click(screen.getByText('Cancel'))
    expect(mockOnClose).toHaveBeenCalled()
  })

  it('should show error when start date is after end date', async () => {
    render(<ExportModal isOpen={true} onClose={mockOnClose} />)

    const inputs = screen.getAllByDisplayValue(/.+/)
    fireEvent.change(inputs[0], { target: { value: '2026-02-01' } })
    fireEvent.change(inputs[1], { target: { value: '2026-01-01' } })

    fireEvent.click(screen.getByText('Export CSV'))

    expect(screen.getByText('Start date must be before end date')).toBeInTheDocument()
  })

  it('should trigger download on successful export', async () => {
    const mockBlob = new Blob(['csv data'], { type: 'text/csv' })
    vi.mocked(transactionsApi.exportCSV).mockResolvedValue({
      blob: mockBlob,
      filename: 'transactions.csv',
    })

    const createObjectURL = vi.fn(() => 'blob:test')
    const revokeObjectURL = vi.fn()
    global.URL.createObjectURL = createObjectURL
    global.URL.revokeObjectURL = revokeObjectURL

    render(<ExportModal isOpen={true} onClose={mockOnClose} />)
    fireEvent.click(screen.getByText('Export CSV'))

    await waitFor(() => {
      expect(transactionsApi.exportCSV).toHaveBeenCalled()
      expect(mockOnClose).toHaveBeenCalled()
    })
  })

  it('should show error on export failure', async () => {
    vi.mocked(transactionsApi.exportCSV).mockRejectedValue(new Error('Failed'))

    render(<ExportModal isOpen={true} onClose={mockOnClose} />)
    fireEvent.click(screen.getByText('Export CSV'))

    await waitFor(() => {
      expect(screen.getByText('Failed to export CSV. Please try again.')).toBeInTheDocument()
    })
  })
})
