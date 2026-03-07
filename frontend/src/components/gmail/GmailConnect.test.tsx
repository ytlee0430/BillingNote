import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { GmailConnect } from './GmailConnect'
import { gmailApi } from '@/api/gmail'

vi.mock('@/api/gmail', () => ({
  gmailApi: {
    getAuthURL: vi.fn(),
    getStatus: vi.fn(),
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    triggerScan: vi.fn(),
    disconnect: vi.fn(),
  },
}))

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('GmailConnect', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows connect button when not connected', async () => {
    vi.mocked(gmailApi.getStatus).mockResolvedValue({
      connected: false,
    })

    render(<GmailConnect />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('Connect Gmail')).toBeInTheDocument()
    })
  })

  it('shows connected status with email', async () => {
    vi.mocked(gmailApi.getStatus).mockResolvedValue({
      connected: true,
      email: 'test@gmail.com',
      last_scan_at: '2026-01-15T10:00:00Z',
    })
    vi.mocked(gmailApi.getSettings).mockResolvedValue({
      enabled: true,
      sender_keywords: ['credit', 'statement'],
      subject_keywords: ['帳單'],
      require_attachment: true,
    })

    render(<GmailConnect />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('Connected')).toBeInTheDocument()
      expect(screen.getByText('test@gmail.com')).toBeInTheDocument()
    })
  })

  it('shows disconnect button when connected', async () => {
    vi.mocked(gmailApi.getStatus).mockResolvedValue({
      connected: true,
      email: 'test@gmail.com',
    })
    vi.mocked(gmailApi.getSettings).mockResolvedValue({
      enabled: true,
      sender_keywords: [],
      subject_keywords: [],
      require_attachment: true,
    })

    render(<GmailConnect />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('Disconnect')).toBeInTheDocument()
    })
  })

  it('shows scan button when connected', async () => {
    vi.mocked(gmailApi.getStatus).mockResolvedValue({
      connected: true,
      email: 'test@gmail.com',
    })
    vi.mocked(gmailApi.getSettings).mockResolvedValue({
      enabled: true,
      sender_keywords: [],
      subject_keywords: [],
      require_attachment: true,
    })

    render(<GmailConnect />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('Scan Now')).toBeInTheDocument()
    })
  })

  it('populates settings fields from API', async () => {
    vi.mocked(gmailApi.getStatus).mockResolvedValue({
      connected: true,
      email: 'test@gmail.com',
    })
    vi.mocked(gmailApi.getSettings).mockResolvedValue({
      enabled: true,
      sender_keywords: ['credit', 'statement'],
      subject_keywords: ['帳單', '電子帳單'],
      require_attachment: true,
    })

    render(<GmailConnect />, { wrapper: createWrapper() })

    await waitFor(() => {
      const senderInput = screen.getByPlaceholderText(/credit, statement/) as HTMLInputElement
      expect(senderInput.value).toBe('credit, statement')
    })
  })

  it('shows loading state', () => {
    vi.mocked(gmailApi.getStatus).mockReturnValue(new Promise(() => {})) // never resolves

    render(<GmailConnect />, { wrapper: createWrapper() })

    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })
})
