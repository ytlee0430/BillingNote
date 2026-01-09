import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BarChart } from './BarChart'
import { CategoryStats } from '@/types/api'

describe('BarChart', () => {
  const mockData: CategoryStats[] = [
    {
      category_id: 1,
      category_name: 'Food',
      amount: 500.0,
    },
    {
      category_id: 2,
      category_name: 'Transport',
      amount: 300.0,
    },
  ]

  it('renders chart with expense data', () => {
    render(<BarChart data={mockData} type="expense" />)

    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })

  it('renders chart with income data', () => {
    render(<BarChart data={mockData} type="income" />)

    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })

  it('displays no data message when data is empty', () => {
    render(<BarChart data={[]} type="expense" />)

    expect(screen.getByText('No data available')).toBeInTheDocument()
  })

  it('displays no data message when data is null', () => {
    render(<BarChart data={null as any} type="expense" />)

    expect(screen.getByText('No data available')).toBeInTheDocument()
  })

  it('renders bars for each category', () => {
    render(<BarChart data={mockData} type="expense" />)

    // Check that bars are rendered
    const bars = document.querySelectorAll('.recharts-bar-rectangle')
    expect(bars.length).toBeGreaterThan(0)
  })

  it('handles uncategorized items', () => {
    const dataWithUncategorized: CategoryStats[] = [
      {
        category_name: '',
        amount: 100.0,
      },
    ]

    render(<BarChart data={dataWithUncategorized} type="expense" />)

    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })
})
