import { describe, it, expect } from 'vitest'
import { isValidEmail, isValidPassword, isPositiveNumber, isValidDate } from './validation'

describe('isValidEmail', () => {
  it('should return true for valid emails', () => {
    expect(isValidEmail('test@example.com')).toBe(true)
    expect(isValidEmail('user.name@domain.co.uk')).toBe(true)
  })

  it('should return false for invalid emails', () => {
    expect(isValidEmail('invalid')).toBe(false)
    expect(isValidEmail('test@')).toBe(false)
    expect(isValidEmail('@example.com')).toBe(false)
    expect(isValidEmail('test@example')).toBe(false)
  })
})

describe('isValidPassword', () => {
  it('should return true for passwords >= 6 characters', () => {
    expect(isValidPassword('123456')).toBe(true)
    expect(isValidPassword('longpassword')).toBe(true)
  })

  it('should return false for passwords < 6 characters', () => {
    expect(isValidPassword('12345')).toBe(false)
    expect(isValidPassword('abc')).toBe(false)
    expect(isValidPassword('')).toBe(false)
  })
})

describe('isPositiveNumber', () => {
  it('should return true for positive numbers', () => {
    expect(isPositiveNumber(1)).toBe(true)
    expect(isPositiveNumber(100.5)).toBe(true)
  })

  it('should return false for zero and negative numbers', () => {
    expect(isPositiveNumber(0)).toBe(false)
    expect(isPositiveNumber(-1)).toBe(false)
    expect(isPositiveNumber(-100.5)).toBe(false)
  })
})

describe('isValidDate', () => {
  it('should return true for valid date strings', () => {
    expect(isValidDate('2024-01-15')).toBe(true)
    expect(isValidDate('2024-12-31')).toBe(true)
  })

  it('should return false for invalid date strings', () => {
    expect(isValidDate('invalid')).toBe(false)
    expect(isValidDate('2024-13-01')).toBe(false)
    expect(isValidDate('not-a-date')).toBe(false)
  })
})
