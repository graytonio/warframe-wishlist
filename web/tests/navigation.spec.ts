import { test, expect } from '@playwright/test'

test.describe('Navigation', () => {
  test('should display header with logo and navigation links', async ({ page }) => {
    await page.goto('/')

    // Header should be visible
    const header = page.locator('header')
    await expect(header).toBeVisible()

    // Logo/brand should be visible
    await expect(page.getByText('Warframe Wishlist')).toBeVisible()

    // Search link should be visible
    await expect(page.getByRole('link', { name: /search/i })).toBeVisible()

    // Sign In button should be visible when not authenticated
    await expect(page.getByRole('link', { name: /sign in/i })).toBeVisible()
  })

  test('should navigate to login page when clicking Sign In', async ({ page }) => {
    await page.goto('/')

    await page.getByRole('link', { name: /sign in/i }).click()

    await expect(page).toHaveURL('/login')
  })

  test('should navigate to search page when clicking Search', async ({ page }) => {
    await page.goto('/login')

    await page.getByRole('link', { name: /search/i }).click()

    await expect(page).toHaveURL('/')
  })

  test('should navigate to search page when clicking logo', async ({ page }) => {
    await page.goto('/login')

    await page.getByText('Warframe Wishlist').click()

    await expect(page).toHaveURL('/')
  })
})
