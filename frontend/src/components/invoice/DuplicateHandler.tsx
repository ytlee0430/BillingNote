import { Invoice } from '@/types/invoice'
import { formatCurrency, formatDate } from '@/utils/format'
import { Modal } from '@/components/common/Modal'
import { Button } from '@/components/common/Button'

interface DuplicateHandlerProps {
  invoice: Invoice | null
  isOpen: boolean
  onClose: () => void
  onConfirm: (invoiceId: number, transactionId: number) => void
  onDismiss: (invoiceId: number) => void
  isConfirming: boolean
}

export const DuplicateHandler = ({
  invoice,
  isOpen,
  onClose,
  onConfirm,
  onDismiss,
  isConfirming,
}: DuplicateHandlerProps) => {
  if (!invoice) return null

  const txn = invoice.duplicated_transaction

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Duplicate Match Detail" size="lg">
      <div className="space-y-6">
        {/* Invoice Info */}
        <div className="bg-blue-50 rounded-lg p-4">
          <h4 className="text-sm font-semibold text-blue-800 mb-2">Invoice</h4>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div>
              <span className="text-gray-500">Number:</span>{' '}
              <span className="font-mono">{invoice.invoice_number}</span>
            </div>
            <div>
              <span className="text-gray-500">Date:</span>{' '}
              {formatDate(invoice.invoice_date)}
            </div>
            <div>
              <span className="text-gray-500">Seller:</span>{' '}
              {invoice.seller_name}
            </div>
            <div>
              <span className="text-gray-500">Amount:</span>{' '}
              <span className="font-medium">{formatCurrency(invoice.amount)}</span>
            </div>
          </div>
        </div>

        {/* Matched Transaction */}
        {txn ? (
          <div className="bg-yellow-50 rounded-lg p-4">
            <h4 className="text-sm font-semibold text-yellow-800 mb-2">
              Matched Transaction
              {invoice.confidence_score !== undefined && (
                <span className="ml-2 text-xs font-normal bg-yellow-200 px-2 py-0.5 rounded">
                  {Math.round(invoice.confidence_score * 100)}% confidence
                </span>
              )}
            </h4>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <span className="text-gray-500">Description:</span>{' '}
                {txn.description}
              </div>
              <div>
                <span className="text-gray-500">Date:</span>{' '}
                {formatDate(txn.transaction_date)}
              </div>
              <div>
                <span className="text-gray-500">Amount:</span>{' '}
                <span className="font-medium">{formatCurrency(txn.amount)}</span>
              </div>
              <div>
                <span className="text-gray-500">Source:</span>{' '}
                {txn.source}
              </div>
            </div>
          </div>
        ) : (
          <div className="bg-gray-50 rounded-lg p-4 text-sm text-gray-500">
            No matched transaction details available.
          </div>
        )}

        {/* Actions */}
        <div className="flex justify-end space-x-3 pt-2">
          <Button
            variant="secondary"
            onClick={() => onDismiss(invoice.id)}
          >
            Not a Duplicate
          </Button>
          {txn && (
            <Button
              onClick={() => onConfirm(invoice.id, txn.id)}
              loading={isConfirming}
            >
              Confirm Duplicate
            </Button>
          )}
        </div>
      </div>
    </Modal>
  )
}
