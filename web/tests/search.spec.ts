import { test, expect } from '@playwright/test'

test.describe('Search Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display search page with title and search input', async ({ page }) => {
    // Title should be visible
    await expect(page.getByRole('heading', { name: /warframe item search/i })).toBeVisible()

    // Search input should be visible
    const searchInput = page.getByPlaceholder(/search for warframes/i)
    await expect(searchInput).toBeVisible()
  })

  test('should show placeholder text when no search query', async ({ page }) => {
    await expect(page.getByText(/start typing to search/i)).toBeVisible()
  })

  test('should show loading state while searching', async ({ page }) => {
    // Mock API to be slow
    await page.route('**/api/v1/items/search*', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 500))
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      })
    })

    const searchInput = page.getByPlaceholder(/search for warframes/i)
    await searchInput.fill('excalibur')

    // Should show loading spinner
    const spinner = page.locator('.animate-spin')
    await expect(spinner).toBeVisible()
  })

  test('should display search results', async ({ page }) => {
    // Mock API response
    await page.route('**/api/v1/items/search*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              uniqueName: '/Lotus/Powersuits/Excalibur/Excalibur',
              name: 'Excalibur',
              description: 'A balanced warframe',
              category: 'Warframes',
              imageName: 'excalibur.png',
            },
            {
              uniqueName: '/Lotus/Powersuits/Excalibur/ExcaliburPrime',
              name: 'Excalibur Prime',
              description: 'The prime variant',
              category: 'Warframes',
              imageName: 'excalibur-prime.png',
            },
          ],
        }),
      })
    })

    const searchInput = page.getByPlaceholder(/search for warframes/i)
    await searchInput.fill('excalibur')

    // Wait for results to appear
    await expect(page.getByText('Excalibur')).toBeVisible()
    await expect(page.getByText('Excalibur Prime')).toBeVisible()
  })

  test('should show no results message when search returns empty', async ({ page }) => {
    // Mock API response with empty results
    await page.route('**/api/v1/items/search*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      })
    })

    const searchInput = page.getByPlaceholder(/search for warframes/i)
    await searchInput.fill('nonexistentitem123')

    // Wait for debounce and API call
    await expect(page.getByText(/no items found/i)).toBeVisible()
  })

  test('should debounce search input', async ({ page }) => {
    let apiCallCount = 0

    await page.route('**/api/v1/items/search*', async (route) => {
      apiCallCount++
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      })
    })

    const searchInput = page.getByPlaceholder(/search for warframes/i)

    // Type quickly
    await searchInput.pressSequentially('test', { delay: 50 })

    // Wait for debounce (300ms) plus some buffer
    await page.waitForTimeout(500)

    // Should only have made one API call due to debouncing
    expect(apiCallCount).toBe(1)
  })

  test('should not show Add to Wishlist button when not authenticated', async ({ page }) => {
    // Mock API response
    await page.route('**/api/v1/items/search*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              uniqueName: '/Lotus/Powersuits/Excalibur/Excalibur',
              name: 'Excalibur',
              category: 'Warframes',
            },
          ],
        }),
      })
    })

    const searchInput = page.getByPlaceholder(/search for warframes/i)
    await searchInput.fill('excalibur')

    // Wait for results
    await expect(page.getByText('Excalibur')).toBeVisible()

    // Add to Wishlist button should not be visible when not authenticated
    await expect(page.getByRole('button', { name: /add to wishlist/i })).not.toBeVisible()
  })
})
