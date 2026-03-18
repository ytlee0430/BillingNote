import { test, expect } from '@playwright/test'
import path from 'path'

test.describe('PDF Parse Verification', () => {
  async function registerAndLogin(page: any) {
    const randomId = Math.random().toString(36).substring(2, 10)
    const testEmail = `verify${Date.now()}${randomId}@example.com`
    await page.goto('/register')
    await page.fill('input[placeholder="你的姓名"]', 'Verify User')
    await page.fill('input[placeholder="your@email.com"]', testEmail)
    await page.fill('input[placeholder="至少 6 個字元"]', 'password123')
    await page.fill('input[placeholder="再次輸入密碼"]', 'password123')
    await page.click('button[type="submit"]')
    await page.waitForURL(/\/(dashboard|transactions|upload|charts|settings)/, { timeout: 20000 })
  }

  test('set PDF password then upload encrypted PDF and verify amounts are correct', async ({ page }) => {
    await registerAndLogin(page)

    // Step 1: Go to Settings, set PDF password "mypass"
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Find the first password input in the PDF Password Management section
    const passwordSection = page.locator('text=PDF Password Management').locator('..')
    const pwdInputs = page.locator('input[type="password"]')
    await pwdInputs.first().fill('mypass')

    // Click Save Passwords
    await page.click('button:has-text("Save Passwords")')
    await expect(page.locator('text=success')).toBeVisible({ timeout: 5000 })

    // Step 2: Go to Upload page
    await page.goto('/upload')
    await page.waitForLoadState('networkidle')

    // Upload the encrypted test PDF
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles(path.resolve('/tmp/test_cathay_realistic_enc.pdf'))

    // Click Parse PDFs button
    await page.click('button:has-text("Parse PDFs")')

    // Wait for results
    await page.waitForTimeout(5000)

    // Step 3: Verify results
    // Bank should be identified as Cathay
    await expect(page.getByText('國泰世華')).toBeVisible({ timeout: 10000 })

    // Get all text content for verification
    const bodyText = await page.textContent('body')

    // Correct amounts should be present
    expect(bodyText).toContain('350')
    expect(bodyText).toContain('120')
    expect(bodyText).toContain('2,580')
    expect(bodyText).toContain('450')
    expect(bodyText).toContain('890')
    expect(bodyText).toContain('1,299')
    expect(bodyText).toContain('1,490')

    // Transaction descriptions should be present
    expect(bodyText).toContain('STARBUCKS')
    expect(bodyText).toContain('7-ELEVEN')
    expect(bodyText).toContain('COSTCO')
    expect(bodyText).toContain('UBER EATS')

    // Verify we got 7 transactions (not more - no header/footer mismatches)
    // Look for transaction rows
    const transactionRows = page.locator('tr').filter({ hasText: /\d{4}-\d{2}-\d{2}|12\/\d{2}/ })
    const rowCount = await transactionRows.count()
    console.log(`Found ${rowCount} transaction rows`)
    expect(rowCount).toBe(7)

    // Take screenshot for visual review
    await page.screenshot({ path: '/tmp/e2e-pdf-parse-result.png', fullPage: true })
    console.log('Screenshot saved to /tmp/e2e-pdf-parse-result.png')
  })

  test('upload unencrypted PDF without password should parse correctly', async ({ page }) => {
    await registerAndLogin(page)

    // Go directly to Upload (no password set)
    await page.goto('/upload')
    await page.waitForLoadState('networkidle')

    // Upload the unencrypted test PDF
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles(path.resolve('/tmp/test_cathay_realistic.pdf'))

    await page.click('button:has-text("Parse PDFs")')
    await page.waitForTimeout(5000)

    // Should identify as Cathay
    await expect(page.getByText('國泰世華')).toBeVisible({ timeout: 10000 })

    const bodyText = await page.textContent('body')
    expect(bodyText).toContain('350')
    expect(bodyText).toContain('STARBUCKS')

    await page.screenshot({ path: '/tmp/e2e-pdf-unencrypted-result.png', fullPage: true })
  })
})
