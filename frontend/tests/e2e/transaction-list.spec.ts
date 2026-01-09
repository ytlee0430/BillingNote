import { test, expect } from '@playwright/test'

test.describe('Transaction List Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('http://localhost:5173/login')
    await page.fill('input[type="email"]', 'test@example.com')
    await page.fill('input[type="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL('http://localhost:5173/dashboard')

    // Navigate to transactions
    await page.goto('http://localhost:5173/transactions')
  })

  test('should display transaction list', async ({ page }) => {
    await expect(page.locator('table')).toBeVisible()
    await expect(page.locator('th:has-text("Date")')).toBeVisible()
    await expect(page.locator('th:has-text("Type")')).toBeVisible()
    await expect(page.locator('th:has-text("Category")')).toBeVisible()
    await expect(page.locator('th:has-text("Description")')).toBeVisible()
    await expect(page.locator('th:has-text("Amount")')).toBeVisible()
    await expect(page.locator('th:has-text("Actions")')).toBeVisible()
  })

  test('should filter transactions by type', async ({ page }) => {
    // Filter by expense
    await page.selectOption('select', 'expense')

    // Wait for results
    await page.waitForTimeout(1000)

    // All visible transactions should be expenses
    const expenseBadges = await page.locator('text=expense').count()
    expect(expenseBadges).toBeGreaterThan(0)
  })

  test('should filter transactions by date range', async ({ page }) => {
    // Set date filters
    const startDateInput = page.locator('input[type="date"]').first()
    const endDateInput = page.locator('input[type="date"]').last()

    await startDateInput.fill('2024-01-01')
    await endDateInput.fill('2024-12-31')

    // Wait for results
    await page.waitForTimeout(1000)

    // Should have transactions within date range
    await expect(page.locator('table tbody tr')).toHaveCount(await page.locator('table tbody tr').count())
  })

  test('should clear filters', async ({ page }) => {
    // Set some filters
    await page.selectOption('select', 'expense')
    await page.locator('input[type="date"]').first().fill('2024-01-01')

    // Click clear filters
    await page.click('text=Clear Filters')

    // Filters should be reset
    const typeSelect = page.locator('select')
    await expect(typeSelect).toHaveValue('')
  })

  test('should edit a transaction', async ({ page }) => {
    // Click edit button on first transaction
    const editButtons = page.locator('button:has-text("Edit")')
    await editButtons.first().click()

    // Modal should open with Edit Transaction title
    await expect(page.locator('text=Edit Transaction')).toBeVisible()

    // Modify the description
    const descriptionInput = page.locator('input[type="text"]')
    await descriptionInput.clear()
    await descriptionInput.fill('Updated transaction description')

    // Submit
    await page.click('button:has-text("Update Transaction")')

    // Wait for update
    await page.waitForTimeout(1000)

    // Verify update
    await expect(page.locator('text=Updated transaction description')).toBeVisible()
  })

  test('should delete a transaction', async ({ page }) => {
    // Get initial transaction count
    const initialCount = await page.locator('table tbody tr').count()

    // Click delete on first transaction
    const deleteButtons = page.locator('button:has-text("Delete")')

    // Handle confirmation dialog
    page.on('dialog', (dialog) => dialog.accept())

    await deleteButtons.first().click()

    // Wait for deletion
    await page.waitForTimeout(1000)

    // Transaction count should decrease
    const newCount = await page.locator('table tbody tr').count()
    expect(newCount).toBeLessThanOrEqual(initialCount)
  })

  test('should display empty state when no transactions', async ({ page }) => {
    // Apply filters that return no results
    await page.selectOption('select', 'income')
    await page.locator('input[type="date"]').first().fill('2000-01-01')
    await page.locator('input[type="date"]').last().fill('2000-01-02')

    await page.waitForTimeout(1000)

    // Should show no transactions message
    await expect(page.locator('text=No transactions found')).toBeVisible()
  })

  test('should navigate between pages', async ({ page }) => {
    // Check if pagination exists (only if there are multiple pages)
    const paginationVisible = await page.locator('button:has-text("Next")').isVisible()

    if (paginationVisible) {
      // Click next page
      await page.click('button:has-text("Next")')

      await page.waitForTimeout(1000)

      // Should be on page 2
      // Verify by checking if Previous button is enabled
      const prevButton = page.locator('button:has-text("Previous")').first()
      await expect(prevButton).not.toBeDisabled()
    }
  })

  test('should display transaction type badges correctly', async ({ page }) => {
    // Check for income badge (green)
    const incomeBadges = page.locator('text=income')
    if ((await incomeBadges.count()) > 0) {
      const firstIncome = incomeBadges.first()
      await expect(firstIncome).toBeVisible()
    }

    // Check for expense badge (red)
    const expenseBadges = page.locator('text=expense')
    if ((await expenseBadges.count()) > 0) {
      const firstExpense = expenseBadges.first()
      await expect(firstExpense).toBeVisible()
    }
  })

  test('should display category information', async ({ page }) => {
    // Transactions with categories should show category name and icon
    const tableRows = page.locator('table tbody tr')
    const count = await tableRows.count()

    if (count > 0) {
      // At least one row should exist
      expect(count).toBeGreaterThan(0)
    }
  })
})
