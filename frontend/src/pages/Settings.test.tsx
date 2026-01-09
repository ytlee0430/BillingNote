import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { Settings } from './Settings'
import { useAuthStore } from '@/store/authStore'

// Mock the auth store
vi.mock('@/store/authStore', () => ({
  useAuthStore: vi.fn(),
}))

// Mock useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('Settings', () => {
  const mockUser = {
    id: 1,
    email: 'test@example.com',
    name: 'Test User',
    created_at: '2024-01-01T00:00:00Z',
  }

  const mockLogout = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useAuthStore as any).mockReturnValue({
      user: mockUser,
      logout: mockLogout,
    })
  })

  const renderSettings = () => {
    return render(
      <BrowserRouter>
        <Settings />
      </BrowserRouter>
    )
  }

  it('renders settings page correctly', () => {
    renderSettings()

    expect(screen.getByText('Settings')).toBeInTheDocument()
    expect(screen.getByText('Profile Information')).toBeInTheDocument()
    expect(screen.getByText('Application Settings')).toBeInTheDocument()
    expect(screen.getByText('About')).toBeInTheDocument()
  })

  it('displays user information correctly', () => {
    renderSettings()

    expect(screen.getByText('test@example.com')).toBeInTheDocument()
    expect(screen.getByText('Test User')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('displays member since date', () => {
    renderSettings()

    expect(screen.getByText(/1\/1\/2024/)).toBeInTheDocument()
  })

  it('renders currency dropdown', () => {
    renderSettings()

    const currencySelect = screen.getByRole('combobox', { name: /currency/i })
    expect(currencySelect).toBeInTheDocument()
    expect(currencySelect).toHaveValue('USD')
  })

  it('renders date format dropdown', () => {
    renderSettings()

    const dateFormatSelect = screen.getByRole('combobox', { name: /date format/i })
    expect(dateFormatSelect).toBeInTheDocument()
  })

  it('renders theme dropdown', () => {
    renderSettings()

    const themeSelect = screen.getByRole('combobox', { name: /theme/i })
    expect(themeSelect).toBeInTheDocument()
  })

  it('displays coming soon messages for settings', () => {
    renderSettings()

    const comingSoonMessages = screen.getAllByText('This feature is coming soon')
    expect(comingSoonMessages.length).toBe(3)
  })

  it('displays about section with version info', () => {
    renderSettings()

    expect(screen.getByText(/Version:/)).toBeInTheDocument()
    expect(screen.getByText(/1.0.0/)).toBeInTheDocument()
    expect(screen.getByText(/Phase 1 - MVP/)).toBeInTheDocument()
  })

  it('displays danger zone section', () => {
    renderSettings()

    expect(screen.getByText('Danger Zone')).toBeInTheDocument()
  })

  it('calls logout when logout button is clicked', () => {
    renderSettings()

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    fireEvent.click(logoutButton)

    expect(mockLogout).toHaveBeenCalledTimes(1)
    expect(mockNavigate).toHaveBeenCalledWith('/login')
  })

  it('handles missing user data gracefully', () => {
    ;(useAuthStore as any).mockReturnValue({
      user: null,
      logout: mockLogout,
    })

    renderSettings()

    const naTexts = screen.getAllByText('N/A')
    expect(naTexts.length).toBeGreaterThan(0)
  })
})
