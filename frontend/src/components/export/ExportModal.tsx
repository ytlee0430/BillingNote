import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { transactionsApi } from '@/api/transactions'

interface ExportModalProps {
  isOpen: boolean
  onClose: () => void
}

export const ExportModal = ({ isOpen, onClose }: ExportModalProps) => {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
  const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0)

  const [startDate, setStartDate] = useState(firstDay.toISOString().split('T')[0])
  const [endDate, setEndDate] = useState(lastDay.toISOString().split('T')[0])
  const [isExporting, setIsExporting] = useState(false)
  const [error, setError] = useState('')

  if (!isOpen) return null

  const handleExport = async () => {
    if (!startDate || !endDate) {
      setError('Please select both start and end dates')
      return
    }
    if (startDate > endDate) {
      setError('Start date must be before end date')
      return
    }

    setError('')
    setIsExporting(true)
    try {
      const { blob, filename } = await transactionsApi.exportCSV(startDate, endDate)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      onClose()
    } catch {
      setError('Failed to export CSV. Please try again.')
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen px-4">
        <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />

        <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full p-6">
          <h2 className="text-xl font-semibold mb-4">Export Transactions</h2>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <input
                type="date"
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                End Date
              </label>
              <input
                type="date"
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>

            {error && (
              <p className="text-red-600 text-sm">{error}</p>
            )}
          </div>

          <div className="flex justify-end gap-3 mt-6">
            <Button variant="secondary" onClick={onClose} disabled={isExporting}>
              Cancel
            </Button>
            <Button onClick={handleExport} loading={isExporting}>
              Export CSV
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
