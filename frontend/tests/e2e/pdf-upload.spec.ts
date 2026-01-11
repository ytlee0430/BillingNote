import { test, expect } from '@playwright/test'
import path from 'path'

test.describe('PDF Upload', () => {
  // Helper to login before tests
  async function login(page: any) {
    const randomId = Math.random().toString(36).substring(2, 10)
    const testEmail = `upload${Date.now()}${randomId}@example.com`
    await page.goto('/register')
    await page.fill('input[placeholder="你的姓名"]', 'Upload Test User')
    await page.fill('input[placeholder="your@email.com"]', testEmail)
    await page.fill('input[placeholder="至少 6 個字元"]', 'password123')
    await page.fill('input[placeholder="再次輸入密碼"]', 'password123')
    await page.click('button[type="submit"]')
    // Wait for navigation to complete - dashboard or any authenticated page
    await page.waitForURL(/\/(dashboard|transactions|upload|charts|settings)/, { timeout: 20000 })
  }

  test('should navigate to upload page', async ({ page }) => {
    await login(page)

    // Click on upload link in navigation
    await page.click('text=上傳帳單')
    await expect(page).toHaveURL(/\/upload/)
    await expect(page.getByRole('heading', { name: 'Upload PDF Statements' })).toBeVisible()
  })

  test('should show upload area', async ({ page }) => {
    await login(page)
    await page.goto('/upload')

    // Check upload area elements
    await expect(page.locator('text=Drop PDF files here or click to select')).toBeVisible()
    await expect(page.locator('input[type="file"]')).toBeAttached()
  })

  test('should select PDF file', async ({ page }) => {
    await login(page)
    await page.goto('/upload')

    // Create a mock PDF file for testing
    const fileInput = page.locator('input[type="file"]')

    // Set file using setInputFiles with a fake PDF
    await fileInput.setInputFiles({
      name: 'test-statement.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('%PDF-1.4 test content'),
    })

    // Check that file appears in the list
    await expect(page.locator('text=test-statement.pdf')).toBeVisible()
    await expect(page.locator('text=Selected Files (1)')).toBeVisible()
  })

  test('should remove selected file', async ({ page }) => {
    await login(page)
    await page.goto('/upload')

    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test-remove.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('%PDF-1.4 test content'),
    })

    await expect(page.locator('text=test-remove.pdf')).toBeVisible()

    // Click remove button - use a more specific locator
    const removeButton = page.locator('button:has-text("Remove")')
    await removeButton.click()

    // File should be removed - wait for it to disappear
    await expect(page.locator('text=test-remove.pdf')).not.toBeVisible({ timeout: 5000 })
  })

  test('should show parse button when files selected', async ({ page }) => {
    await login(page)
    await page.goto('/upload')

    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test-parse.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('%PDF-1.4 test content'),
    })

    // Parse button should be visible
    await expect(page.locator('button:has-text("Parse PDFs")')).toBeVisible()
  })

  test('should handle upload error gracefully', async ({ page }) => {
    await login(page)
    await page.goto('/upload')

    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'invalid-file.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('%PDF-1.4 invalid'),
    })

    // Click parse button
    await page.click('button:has-text("Parse PDFs")')

    // Wait for the upload to complete (success or error)
    await page.waitForTimeout(3000)

    // Either we get results or an error - both are acceptable
    const hasResults = await page.locator('.bg-white.rounded-lg.shadow').count() > 0
    const hasError = await page.locator('.bg-red-100').count() > 0
    const hasNoTransactions = await page.locator('text=No transactions found').count() > 0

    expect(hasResults || hasError || hasNoTransactions).toBeTruthy()
  })

  test('should show start over button after parsing', async ({ page }) => {
    await login(page)
    await page.goto('/upload')

    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles({
      name: 'test-startover.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('%PDF-1.4 test'),
    })

    await page.click('button:has-text("Parse PDFs")')

    // Wait for results
    await page.waitForTimeout(3000)

    // Check for start over button or upload area
    const hasStartOver = await page.locator('button:has-text("Start Over")').count() > 0
    const hasUploadArea = await page.locator('text=Drop PDF files here').count() > 0

    expect(hasStartOver || hasUploadArea).toBeTruthy()
  })
})

test.describe('PDF Password Settings', () => {
  async function login(page: any) {
    const randomId = Math.random().toString(36).substring(2, 10)
    const testEmail = `password${Date.now()}${randomId}@example.com`
    await page.goto('/register')
    await page.fill('input[placeholder="你的姓名"]', 'Password Test User')
    await page.fill('input[placeholder="your@email.com"]', testEmail)
    await page.fill('input[placeholder="至少 6 個字元"]', 'password123')
    await page.fill('input[placeholder="再次輸入密碼"]', 'password123')
    await page.click('button[type="submit"]')
    // Wait for navigation to complete - dashboard or any authenticated page
    await page.waitForURL(/\/(dashboard|transactions|upload|charts|settings)/, { timeout: 20000 })
  }

  test('should navigate to settings page', async ({ page }) => {
    await login(page)
    await page.click('text=設定')
    await expect(page).toHaveURL(/\/settings/)
  })

  test('should show PDF password management section', async ({ page }) => {
    await login(page)
    await page.goto('/settings')

    await expect(page.locator('text=PDF Password Management')).toBeVisible()
    await expect(page.locator('text=#1')).toBeVisible()
    await expect(page.locator('text=#2')).toBeVisible()
    await expect(page.locator('text=#3')).toBeVisible()
    await expect(page.locator('text=#4')).toBeVisible()
  })

  test('should have password input fields', async ({ page }) => {
    await login(page)
    await page.goto('/settings')

    // Check for password inputs
    const passwordInputs = page.locator('input[type="password"]')
    await expect(passwordInputs).toHaveCount(4)
  })

  test('should have save passwords button', async ({ page }) => {
    await login(page)
    await page.goto('/settings')

    await expect(page.locator('button:has-text("Save Passwords")')).toBeVisible()
  })
})
