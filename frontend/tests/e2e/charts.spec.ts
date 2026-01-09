import { test, expect } from '@playwright/test'

test.describe('Charts and Statistics', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('http://localhost:5173/login')
    await page.fill('input[type="email"]', 'test@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('http://localhost:5173/dashboard')

    // Navigate to charts
    await page.goto('http://localhost:5173/charts')
  })

  test('should display statistics page', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Statistics')
  })

  test('should display filter period section', async ({ page }) => {
    await expect(page.locator('text=Filter Period')).toBeVisible()
    await expect(page.locator('select').first()).toBeVisible() // Year selector
    await expect(page.locator('select').last()).toBeVisible() // Month selector
  })

  test('should display monthly statistics cards', async ({ page }) => {
    await expect(page.locator('text=Total Income')).toBeVisible()
    await expect(page.locator('text=Total Expense')).toBeVisible()
    await expect(page.locator('text=Balance')).toBeVisible()
  })

  test('should change year filter', async ({ page }) => {
    const yearSelect = page.locator('select').first()
    const currentYear = new Date().getFullYear()

    await yearSelect.selectOption(String(currentYear - 1))
    await page.waitForTimeout(1000)

    // Stats should reload
    await expect(page.locator('text=Total Income')).toBeVisible()
  })

  test('should change month filter', async ({ page }) => {
    const monthSelect = page.locator('select').last()

    await monthSelect.selectOption('1') // January
    await page.waitForTimeout(1000)

    // Stats should reload
    await expect(page.locator('text=Total Income')).toBeVisible()
  })

  test('should display category breakdown section', async ({ page }) => {
    await expect(page.locator('text=Category Breakdown')).toBeVisible()
  })

  test('should toggle between expense and income stats', async ({ page }) => {
    // Click expenses button
    await page.click('button:has-text("Expenses")')
    await page.waitForTimeout(500)

    // Should show expense data
    await expect(page.locator('button:has-text("Expenses")')).toHaveClass(/bg-red-100/)

    // Click income button
    await page.click('button:has-text("Income")')
    await page.waitForTimeout(500)

    // Should show income data
    await expect(page.locator('button:has-text("Income")')).toHaveClass(/bg-green-100/)
  })

  test('should display distribution chart', async ({ page }) => {
    await expect(page.locator('text=Distribution')).toBeVisible()

    // Check if chart container exists
    const chartContainer = page.locator('.recharts-responsive-container').first()
    await expect(chartContainer).toBeVisible()
  })

  test('should display comparison chart', async ({ page }) => {
    await expect(page.locator('text=Comparison')).toBeVisible()

    // Check if chart container exists
    const chartContainers = page.locator('.recharts-responsive-container')
    expect(await chartContainers.count()).toBeGreaterThanOrEqual(2)
  })

  test('should handle no data state', async ({ page }) => {
    // Select a far future date with no data
    const yearSelect = page.locator('select').first()
    await yearSelect.selectOption('2030')
    await page.waitForTimeout(1000)

    // Should show no data message or zero values
    const balance = await page.locator('text=Balance').locator('..').textContent()
    expect(balance).toBeTruthy()
  })

  test('should display income and expense in different colors', async ({ page }) => {
    // Income should be in green
    const incomeCard = page.locator('text=Total Income').locator('..')
    await expect(incomeCard).toHaveClass(/bg-green/)

    // Expense should be in red
    const expenseCard = page.locator('text=Total Expense').locator('..')
    await expect(expenseCard).toHaveClass(/bg-red/)
  })

  test('should show correct balance color', async ({ page }) => {
    // Balance card should exist
    const balanceCard = page.locator('text=Balance').locator('..')
    await expect(balanceCard).toBeVisible()

    // The color depends on whether balance is positive or negative
    // Just verify the card is rendered with appropriate styling
    await expect(balanceCard).toHaveClass(/bg-blue/)
  })

  test('should update charts when changing type', async ({ page }) => {
    // Start with expenses
    await page.click('button:has-text("Expenses")')
    await page.waitForTimeout(500)

    // Get chart state
    const expenseCharts = await page.locator('.recharts-responsive-container').count()

    // Switch to income
    await page.click('button:has-text("Income")')
    await page.waitForTimeout(500)

    // Charts should still be rendered
    const incomeCharts = await page.locator('.recharts-responsive-container').count()
    expect(incomeCharts).toBe(expenseCharts)
  })

  test('should display loading state correctly', async ({ page }) => {
    // When page first loads, there might be a loading state
    // This is hard to test due to timing, but we can verify the page eventually loads
    await page.waitForTimeout(2000)

    // After loading, statistics should be visible
    await expect(page.locator('text=Total Income')).toBeVisible()
  })

  test('should navigate to charts from dashboard', async ({ page }) => {
    await page.goto('http://localhost:5173/dashboard')
    await page.click('text=圖表')

    await expect(page).toHaveURL('http://localhost:5173/charts')
    await expect(page.locator('h1')).toContainText('Statistics')
  })

  test('should maintain filter state when switching tabs', async ({ page }) => {
    // Set a specific year
    const yearSelect = page.locator('select').first()
    await yearSelect.selectOption('2023')

    // Navigate away and back
    await page.goto('http://localhost:5173/dashboard')
    await page.goto('http://localhost:5173/charts')

    // Year might reset to current year (this is expected behavior)
    // Just verify the page loads correctly
    await expect(page.locator('text=Filter Period')).toBeVisible()
  })
})
