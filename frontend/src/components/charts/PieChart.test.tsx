import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PieChart } from './PieChart'
import { CategoryStats } from '@/types/api'

describe('PieChart', () => {
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
    {
      category_id: 3,
      category_name: 'Entertainment',
      amount: 200.0,
    },
  ]

  it('renders chart with data', () => {
    render(<PieChart data={mockData} type="expense" />)

    // Check if Recharts ResponsiveContainer is rendered
    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })

  it('displays no data message when data is empty', () => {
    render(<PieChart data={[]} type="expense" />)

    expect(screen.getByText('No data available')).toBeInTheDocument()
  })

  it('displays no data message when data is null', () => {
    render(<PieChart data={null as any} type="expense" />)

    expect(screen.getByText('No data available')).toBeInTheDocument()
  })

  it('renders with income type', () => {
    render(<PieChart data={mockData} type="income" />)

    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })

  it('handles uncategorized items', () => {
    const dataWithUncategorized: CategoryStats[] = [
      {
        category_name: '',
        amount: 100.0,
      },
    ]

    render(<PieChart data={dataWithUncategorized} type="expense" />)

    const container = document.querySelector('.recharts-responsive-container')
    expect(container).toBeTruthy()
  })

  it('renders multiple categories correctly', () => {
    render(<PieChart data={mockData} type="expense" />)

    // Check that the chart has the correct number of cells
    const cells = document.querySelectorAll('.recharts-pie-sector')
    expect(cells.length).toBe(mockData.length)
  })
})
