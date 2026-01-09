import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || ''

export interface ParsedTransaction {
  date: string
  description: string
  amount: number
  currency: string
  category: string
  card_last4: string
  is_duplicate: boolean
}

export interface UploadResult {
  filename: string
  bank: string
  transactions: ParsedTransaction[]
  total_amount: number
  error?: string
}

export interface UploadResponse {
  results: UploadResult[]
}

export interface ImportResponse {
  imported: number
  message: string
}

export interface PDFPassword {
  id: number
  priority: number
  label: string
  has_value: boolean
  created_at: string
  updated_at: string
}

export interface PDFPasswordInput {
  password: string
  priority: number
  label?: string
}

// Upload PDF files for parsing
export async function uploadPDFs(files: File[]): Promise<UploadResponse> {
  const formData = new FormData()
  files.forEach(file => {
    formData.append('files', file)
  })

  const token = localStorage.getItem('token')
  const response = await axios.post<UploadResponse>(
    `${API_BASE_URL}/api/upload/pdf`,
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
        Authorization: `Bearer ${token}`,
      },
    }
  )

  return response.data
}

// Import parsed transactions
export async function importTransactions(
  transactions: ParsedTransaction[]
): Promise<ImportResponse> {
  const token = localStorage.getItem('token')
  const response = await axios.post<ImportResponse>(
    `${API_BASE_URL}/api/transactions/import`,
    { transactions },
    {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    }
  )

  return response.data
}

// Get PDF passwords
export async function getPDFPasswords(): Promise<{ passwords: PDFPassword[] }> {
  const token = localStorage.getItem('token')
  const response = await axios.get<{ passwords: PDFPassword[] }>(
    `${API_BASE_URL}/api/settings/pdf-passwords`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  )

  return response.data
}

// Set a PDF password
export async function setPDFPassword(input: PDFPasswordInput): Promise<void> {
  const token = localStorage.getItem('token')
  await axios.post(
    `${API_BASE_URL}/api/settings/pdf-passwords`,
    input,
    {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    }
  )
}

// Set multiple PDF passwords
export async function setMultiplePDFPasswords(
  passwords: PDFPasswordInput[]
): Promise<void> {
  const token = localStorage.getItem('token')
  await axios.put(
    `${API_BASE_URL}/api/settings/pdf-passwords`,
    { passwords },
    {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    }
  )
}

// Delete a PDF password
export async function deletePDFPassword(priority: number): Promise<void> {
  const token = localStorage.getItem('token')
  await axios.delete(`${API_BASE_URL}/api/settings/pdf-passwords/${priority}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}
