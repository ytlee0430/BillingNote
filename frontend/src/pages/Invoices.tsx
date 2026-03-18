import { useState } from 'react'
import { InvoiceList } from '@/components/invoice/InvoiceList'
import { DuplicateHandler } from '@/components/invoice/DuplicateHandler'
import { Button } from '@/components/common/Button'
import { useInvoices } from '@/hooks/useInvoices'
import { Invoice, InvoiceFilter } from '@/types/invoice'

export const Invoices = () => {
  const [filter, setFilter] = useState<InvoiceFilter>({
    page: 1,
    page_size: 20,
  })
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null)
  const [syncDates, setSyncDates] = useState({
    start_date: '',
    end_date: '',
  })
  const [syncMessage, setSyncMessage] = useState<string | null>(null)

  const {
    invoices,
    total,
    page,
    pageSize,
    isLoading,
    syncInvoicesAsync,
    isSyncing,
    confirmDuplicateAsync,
    isConfirming,
    deleteInvoice,
  } = useInvoices(filter)

  const handleSync = async () => {
    if (!syncDates.start_date || !syncDates.end_date) {
      setSyncMessage('Please select both start and end dates.')
      return
    }

    setSyncMessage(null)
    try {
      const result = await syncInvoicesAsync({
        start_date: syncDates.start_date.replace(/-/g, '/'),
        end_date: syncDates.end_date.replace(/-/g, '/'),
      })
      setSyncMessage(`Synced ${result.synced} new invoices.`)
    } catch (error: any) {
      setSyncMessage(error?.response?.data?.error || 'Sync failed. Please check your carrier code in settings.')
    }
  }

  const handleConfirmDuplicate = async (invoiceId: number, transactionId: number) => {
    try {
      await confirmDuplicateAsync({ invoice_id: invoiceId, transaction_id: transactionId })
      setSelectedInvoice(null)
    } catch (error) {
      console.error('Failed to confirm duplicate:', error)
    }
  }

  const handleDismissDuplicate = (_invoiceId: number) => {
    setSelectedInvoice(null)
  }

  const handleDelete = (id: number) => {
    if (window.confirm('Are you sure you want to delete this invoice?')) {
      deleteInvoice(id)
    }
  }

  const handleFilterChange = (newFilter: Partial<InvoiceFilter>) => {
    setFilter((prev) => ({ ...prev, ...newFilter, page: 1 }))
  }

  const handlePageChange = (newPage: number) => {
    setFilter((prev) => ({ ...prev, page: newPage }))
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Invoices</h1>
      </div>

      {/* Sync Section */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">Sync Invoices from MOF</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Start Date
            </label>
            <input
              type="date"
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={syncDates.start_date}
              onChange={(e) =>
                setSyncDates((prev) => ({ ...prev, start_date: e.target.value }))
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
              value={syncDates.end_date}
              onChange={(e) =>
                setSyncDates((prev) => ({ ...prev, end_date: e.target.value }))
              }
            />
          </div>
          <div>
            <Button onClick={handleSync} loading={isSyncing}>
              {isSyncing ? 'Syncing...' : 'Sync Invoices'}
            </Button>
          </div>
        </div>
        {syncMessage && (
          <div
            className={`mt-3 p-3 rounded text-sm ${
              syncMessage.includes('failed') || syncMessage.includes('Please')
                ? 'bg-red-100 text-red-700'
                : 'bg-green-100 text-green-700'
            }`}
          >
            {syncMessage}
          </div>
        )}
      </div>

      {/* Filters */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">Filters</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
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
              onClick={() => setFilter({ page: 1, page_size: 20 })}
            >
              Clear Filters
            </Button>
          </div>
        </div>
      </div>

      {/* Invoice List */}
      <InvoiceList
        invoices={invoices}
        total={total}
        page={page}
        pageSize={pageSize}
        isLoading={isLoading}
        onViewDuplicate={setSelectedInvoice}
        onDelete={handleDelete}
        onPageChange={handlePageChange}
      />

      {/* Duplicate Handler Modal */}
      <DuplicateHandler
        invoice={selectedInvoice}
        isOpen={selectedInvoice !== null}
        onClose={() => setSelectedInvoice(null)}
        onConfirm={handleConfirmDuplicate}
        onDismiss={handleDismissDuplicate}
        isConfirming={isConfirming}
      />
    </div>
  )
}
