import { Modal } from '@/components/common/Modal'
import { TransactionForm } from './TransactionForm'
import { CreateTransactionRequest, Transaction } from '@/types/transaction'

interface TransactionModalProps {
  isOpen: boolean
  onClose: () => void
  transaction?: Transaction | null
  onSave: (data: CreateTransactionRequest) => void | Promise<void>
  isSaving?: boolean
}

export const TransactionModal = ({
  isOpen,
  onClose,
  transaction,
  onSave,
  isSaving = false,
}: TransactionModalProps) => {
  return (
    <Modal isOpen={isOpen} onClose={onClose} title={transaction ? 'Edit Transaction' : 'Add Transaction'}>
      <TransactionForm
        transaction={transaction}
        onSubmit={onSave}
        onCancel={onClose}
        isSubmitting={isSaving}
      />
    </Modal>
  )
}
