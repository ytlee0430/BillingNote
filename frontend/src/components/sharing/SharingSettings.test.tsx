import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { SharingSettings } from './SharingSettings'

vi.mock('@/api/sharing', () => ({
  sharingApi: {
    getMyCode: vi.fn(),
    regenerateCode: vi.fn(),
    pair: vi.fn(),
    getConnections: vi.fn(),
    revokeAccess: vi.fn(),
  },
}))

import { sharingApi } from '@/api/sharing'

describe('SharingSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(sharingApi.getMyCode).mockResolvedValue({ code: 'AB12-CD34' })
    vi.mocked(sharingApi.getConnections).mockResolvedValue({ viewers: [], owners: [] })
  })

  it('should display pairing code after loading', async () => {
    render(<SharingSettings />)

    await waitFor(() => {
      expect(screen.getByText('AB12-CD34')).toBeInTheDocument()
    })
  })

  it('should regenerate code', async () => {
    vi.mocked(sharingApi.regenerateCode).mockResolvedValue({ code: 'XY56-ZW78' })

    render(<SharingSettings />)
    await waitFor(() => {
      expect(screen.getByText('AB12-CD34')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Regenerate'))

    await waitFor(() => {
      expect(screen.getByText('XY56-ZW78')).toBeInTheDocument()
    })
  })

  it('should pair with a code', async () => {
    vi.mocked(sharingApi.pair).mockResolvedValue({ message: 'paired' })

    render(<SharingSettings />)
    await waitFor(() => {
      expect(screen.getByText('AB12-CD34')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText('AB12-CD34')
    fireEvent.change(input, { target: { value: 'QR99-ST00' } })
    fireEvent.click(screen.getByText('Pair'))

    await waitFor(() => {
      expect(sharingApi.pair).toHaveBeenCalledWith('QR99-ST00')
    })
  })

  it('should display viewers list', async () => {
    vi.mocked(sharingApi.getConnections).mockResolvedValue({
      viewers: [
        { id: 1, owner_id: 1, viewer_id: 2, created_at: '', viewer: { id: 2, email: 'viewer@test.com', name: 'Viewer' } },
      ],
      owners: [],
    })

    render(<SharingSettings />)

    await waitFor(() => {
      expect(screen.getByText('viewer@test.com')).toBeInTheDocument()
      expect(screen.getByText('Revoke')).toBeInTheDocument()
    })
  })

  it('should display owners list', async () => {
    vi.mocked(sharingApi.getConnections).mockResolvedValue({
      viewers: [],
      owners: [
        { id: 1, owner_id: 3, viewer_id: 1, created_at: '', owner: { id: 3, email: 'owner@test.com', name: 'Owner' } },
      ],
    })

    render(<SharingSettings />)

    await waitFor(() => {
      expect(screen.getByText('owner@test.com')).toBeInTheDocument()
    })
  })
})
