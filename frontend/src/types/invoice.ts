export interface Invoice {
  id: number
  user_id: number
  invoice_number: string
  invoice_date: string
  seller_name: string
  seller_ban: string
  amount: number
  status: string
  items: InvoiceItem[] | null
  is_duplicated: boolean
  duplicated_transaction_id?: number
  confidence_score?: number
  created_at: string
  duplicated_transaction?: {
    id: number
    amount: number
    description: string
    transaction_date: string
    source: string
  }
}

export interface InvoiceItem {
  description: string
  quantity: string
  unit_price: string
  amount: string
}

export interface InvoiceFilter {
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
}

export interface InvoiceListResponse {
  invoices: Invoice[]
  total: number
  page: number
  page_size: number
}

export interface InvoiceSyncRequest {
  start_date: string
  end_date: string
}

export interface InvoiceSyncResponse {
  synced: number
  message: string
}

export interface ConfirmDuplicateRequest {
  invoice_id: number
  transaction_id: number
}

export interface InvoiceSettingsInput {
  invoice_carrier: string
}
