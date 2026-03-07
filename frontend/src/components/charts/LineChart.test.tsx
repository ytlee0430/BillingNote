import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LineChart } from './LineChart'
import { TrendDataPoint } from '@/types/api'

describe('LineChart', () => {
  const mockData: TrendDataPoint[] = [
    { date: '2026-01', income: 50000, expense: 35000 },
    { date: '2026-02', income: 52000, expense: 38000 },
    { date: '2026-03', income: 48000, expense: 32000 },
  ]

  it('renders chart with data', () => {
    render(<LineChart data={mockData} />)
    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })

  it('shows empty state when no data', () => {
    render(<LineChart data={[]} />)
    expect(screen.getByText('No trend data available')).toBeInTheDocument()
  })

  it('shows empty state when data is null-like', () => {
    render(<LineChart data={undefined as any} />)
    expect(screen.getByText('No trend data available')).toBeInTheDocument()
  })
})
