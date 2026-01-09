import { describe, it, expect } from 'vitest'
import { formatCurrency, formatDate, formatNumber } from './format'

describe('formatCurrency', () => {
  it('should format positive numbers correctly', () => {
    expect(formatCurrency(1000)).toContain('1,000')
  })

  it('should format zero correctly', () => {
    expect(formatCurrency(0)).toContain('0')
  })

  it('should format negative numbers correctly', () => {
    expect(formatCurrency(-500)).toContain('500')
  })
})

describe('formatDate', () => {
  it('should format date string correctly', () => {
    const result = formatDate('2024-01-15')
    expect(result).toBe('2024-01-15')
  })

  it('should format Date object correctly', () => {
    const date = new Date('2024-01-15')
    const result = formatDate(date)
    expect(result).toBe('2024-01-15')
  })

  it('should support custom format', () => {
    const result = formatDate('2024-01-15', 'yyyy/MM/dd')
    expect(result).toBe('2024/01/15')
  })
})

describe('formatNumber', () => {
  it('should format numbers with thousand separators', () => {
    expect(formatNumber(1000)).toBe('1,000')
  })

  it('should format large numbers correctly', () => {
    expect(formatNumber(1234567)).toBe('1,234,567')
  })

  it('should handle zero', () => {
    expect(formatNumber(0)).toBe('0')
  })
})
