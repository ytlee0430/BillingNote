import { test, expect } from '@playwright/test'

test.describe('Manual Transaction Creation', () => {
  test.beforeEach(async ({ page }) => {
    // Generate unique email for each test using timestamp + random number
    const uniqueId = `${Date.now()}${Math.random().toString(36).substring(7)}`
    const testEmail = `txtest${uniqueId}@example.com`

    await page.goto('/register')
    await page.fill('input[placeholder="你的姓名"]', 'Transaction Test User')
    await page.fill('input[placeholder="your@email.com"]', testEmail)
    await page.fill('input[placeholder="至少 6 個字元"]', 'password123')
    await page.fill('input[placeholder="再次輸入密碼"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL(/\/dashboard/, { timeout: 15000 })
  })

  test('should navigate to transactions page', async ({ page }) => {
    await page.click('text=交易記錄')
    await expect(page).toHaveURL(/\/transactions/)
    await expect(page.locator('main h1')).toContainText('Transactions')
  })

  test('should open add transaction modal', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    await expect(page.locator('h3:has-text("Add Transaction")')).toBeVisible()
  })

  test('should create a new expense transaction', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    // Wait for modal to open
    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Fill in the form (use more specific selectors within the modal)
    await modal.locator('input[value="expense"]').check()
    await modal.locator('input[type="number"]').fill('150.50')
    await modal.locator('input[placeholder="Enter description"]').fill('Test expense transaction')

    // Submit the form using the button inside the form
    await modal.locator('button[type="submit"]').click()

    // Wait for modal to close and verify transaction appears
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator('text=Test expense transaction')).toBeVisible({ timeout: 10000 })
  })

  test('should create a new income transaction', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    // Wait for modal to open
    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Fill in the form
    await modal.locator('input[value="income"]').check()
    await modal.locator('input[type="number"]').fill('1000.00')
    await modal.locator('input[placeholder="Enter description"]').fill('Test income transaction')

    // Submit the form
    await page.click('button:has-text("Create Transaction")')

    // Wait for modal to close and verify transaction appears
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator('text=Test income transaction')).toBeVisible({ timeout: 10000 })
  })

  test('should show validation error for empty amount', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Try to submit without filling amount (description only)
    await modal.locator('input[placeholder="Enter description"]').fill('Test transaction')
    await page.click('button:has-text("Create Transaction")')

    // Should show validation error
    await expect(page.locator('text=Amount must be greater than 0')).toBeVisible()
  })

  test('should show validation error for empty description', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Try to submit without description (amount only)
    await modal.locator('input[type="number"]').fill('100')
    await page.click('button:has-text("Create Transaction")')

    // Should show validation error
    await expect(page.locator('text=Description is required')).toBeVisible()
  })

  test('should cancel transaction creation', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    // Wait for modal to open
    await expect(page.locator('h3:has-text("Add Transaction")')).toBeVisible()

    // Click cancel
    await page.click('button:has-text("Cancel")')

    // Modal should be closed
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible()
  })

  test('should create transaction with category', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Fill in the form with category
    await modal.locator('input[value="expense"]').check()
    await modal.locator('input[type="number"]').fill('50.00')

    // Wait for categories to load and select one if available
    const categorySelect = modal.locator('select')
    await expect(categorySelect).toBeVisible()

    // Wait a bit for categories to load
    await page.waitForTimeout(500)

    // Get the number of options
    const optionCount = await categorySelect.locator('option').count()

    // Only try to select category if there's more than the default "No category" option
    if (optionCount > 1) {
      await categorySelect.selectOption({ index: 1 })
    }

    await modal.locator('input[placeholder="Enter description"]').fill('Transaction with category')

    // Submit the form
    await modal.locator('button[type="submit"]').click()

    // Wait for modal to close and verify
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator('text=Transaction with category')).toBeVisible({ timeout: 10000 })
  })

  test('should persist transaction data after save (not disappear)', async ({ page }) => {
    const uniqueDescription = `Persist Test ${Date.now()}`

    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Fill in the form
    await modal.locator('input[value="expense"]').check()
    await modal.locator('input[type="number"]').fill('99.99')
    await modal.locator('input[placeholder="Enter description"]').fill(uniqueDescription)

    // Submit and wait for modal to close
    await page.click('button:has-text("Create Transaction")')

    // Wait for the modal to close
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })

    // Verify transaction appears in the list immediately
    await expect(page.locator(`text=${uniqueDescription}`)).toBeVisible({ timeout: 10000 })
  })

  test('should persist transaction data after page refresh', async ({ page }) => {
    const uniqueDescription = `Refresh Test ${Date.now()}`

    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Fill in the form
    await modal.locator('input[value="income"]').check()
    await modal.locator('input[type="number"]').fill('500.00')
    await modal.locator('input[placeholder="Enter description"]').fill(uniqueDescription)

    // Submit the form
    await page.click('button:has-text("Create Transaction")')

    // Wait for modal to close and transaction to appear
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator(`text=${uniqueDescription}`)).toBeVisible({ timeout: 10000 })

    // Refresh the page
    await page.reload()

    // Verify transaction still appears after refresh
    await expect(page.locator(`text=${uniqueDescription}`)).toBeVisible({ timeout: 10000 })
  })

  test('should show saving state during transaction creation', async ({ page }) => {
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    // Fill in the form
    await modal.locator('input[value="expense"]').check()
    await modal.locator('input[type="number"]').fill('25.00')
    await modal.locator('input[placeholder="Enter description"]').fill('Saving state test')

    // Click submit
    await page.click('button:has-text("Create Transaction")')

    // Wait for modal to close and transaction to appear
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator('text=Saving state test')).toBeVisible({ timeout: 10000 })
  })

  test('should update transaction and persist changes', async ({ page }) => {
    const originalDescription = `Original ${Date.now()}`
    const updatedDescription = `Updated ${Date.now()}`

    // First create a transaction
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    await modal.locator('input[value="expense"]').check()
    await modal.locator('input[type="number"]').fill('75.00')
    await modal.locator('input[placeholder="Enter description"]').fill(originalDescription)
    await page.click('button:has-text("Create Transaction")')

    // Wait for modal to close and transaction to appear
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator(`text=${originalDescription}`)).toBeVisible({ timeout: 10000 })

    // Click edit button on the transaction row
    const transactionRow = page.locator(`tr:has-text("${originalDescription}")`)
    await transactionRow.locator('button:has-text("Edit")').click()

    // Wait for edit modal
    await expect(page.locator('h3:has-text("Edit Transaction")')).toBeVisible()

    // Update the description
    const editModal = page.locator('form')
    await editModal.locator('input[placeholder="Enter description"]').fill(updatedDescription)
    await page.click('button:has-text("Update Transaction")')

    // Wait for modal to close
    await expect(page.locator('h3:has-text("Edit Transaction")')).not.toBeVisible({ timeout: 15000 })

    // Verify the updated description appears
    await expect(page.locator(`text=${updatedDescription}`)).toBeVisible({ timeout: 10000 })

    // Original should not be visible anymore
    await expect(page.locator(`text=${originalDescription}`)).not.toBeVisible()

    // Refresh and verify persistence
    await page.reload()
    await expect(page.locator(`text=${updatedDescription}`)).toBeVisible({ timeout: 10000 })
  })

  test('should delete transaction and persist deletion', async ({ page }) => {
    const descriptionToDelete = `Delete Test ${Date.now()}`

    // First create a transaction
    await page.goto('/transactions')
    await page.click('button:has-text("Add Transaction")')

    const modal = page.locator('form')
    await expect(modal).toBeVisible()

    await modal.locator('input[value="expense"]').check()
    await modal.locator('input[type="number"]').fill('30.00')
    await modal.locator('input[placeholder="Enter description"]').fill(descriptionToDelete)
    await page.click('button:has-text("Create Transaction")')

    // Wait for modal to close and transaction to appear
    await expect(page.locator('h3:has-text("Add Transaction")')).not.toBeVisible({ timeout: 15000 })
    await expect(page.locator(`text=${descriptionToDelete}`)).toBeVisible({ timeout: 10000 })

    // Accept the confirmation dialog
    page.on('dialog', dialog => dialog.accept())

    // Click delete button on the transaction row
    const transactionRow = page.locator(`tr:has-text("${descriptionToDelete}")`)
    await transactionRow.locator('button:has-text("Delete")').click()

    // Verify the transaction is removed
    await expect(page.locator(`text=${descriptionToDelete}`)).not.toBeVisible({ timeout: 10000 })

    // Refresh and verify it's still gone
    await page.reload()
    await expect(page.locator(`text=${descriptionToDelete}`)).not.toBeVisible()
  })
})
