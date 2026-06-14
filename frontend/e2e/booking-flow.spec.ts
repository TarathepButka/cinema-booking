import { expect, test, type Page } from '@playwright/test'

const movie = {
  id: 'movie-1',
  title: 'Test Movie',
  description: 'A movie used to verify the booking journey.',
  genre: ['Action'],
  duration: 120,
  rating: 8.5,
  language: 'EN',
  director: 'Test Director',
  cast: ['Actor One'],
  posterUrl: '',
  releaseDate: '2026-01-01T00:00:00Z',
  isActive: true,
}

async function mockAuthenticatedUser(page: Page) {
  await page.route('**/api/auth/me', (route) =>
    route.fulfill({
      json: {
        userId: 'user-1',
        email: 'user@example.com',
        name: 'Test User',
        role: 'USER',
        dbRole: 'USER',
      },
    }),
  )
}

test('opens a movie and displays an available showtime', async ({ page }) => {
  await mockAuthenticatedUser(page)
  await page.route('**/api/movies', (route) =>
    route.fulfill({ json: { data: [movie], total: 1 } }),
  )
  await page.route('**/api/movies/movie-1', (route) =>
    route.fulfill({ json: { data: movie } }),
  )
  await page.route('**/api/showtimes?*', (route) =>
    route.fulfill({
      json: {
        data: [{
          id: 'showtime-1',
          movieId: movie.id,
          theaterId: 'theater-1',
          theaterName: 'Hall A',
          startTime: '2099-06-14T10:00:00+07:00',
          endTime: '2099-06-14T12:00:00+07:00',
          slotLabel: 'Morning',
          priceByZone: { FRONT: 180 },
        }],
        total: 1,
      },
    }),
  )

  await page.goto('/')
  await expect(page.getByText('Test Movie', { exact: true })).toBeVisible()
  await page.locator('.movie-card').click()

  await expect(page).toHaveURL(/\/movies\/movie-1\/showtime$/)
  await expect(page.getByRole('heading', { name: 'Test Movie' })).toBeVisible()
  await expect(page.locator('.section-title-accent', { hasText: 'Select Showtime' })).toBeVisible()
  await expect(page.locator('.slot-card').filter({ hasText: 'Morning' })).toBeVisible()
})

test('successfully completes the full booking flow (select seat and confirm payment)', async ({ page }) => {
  await mockAuthenticatedUser(page)

  let seatStatus = 'AVAILABLE'
  let lockedBy: string | undefined = undefined
  let lockedUntil: string | undefined = undefined

  // Mock initial pages load
  await page.route('**/api/movies', (route) =>
    route.fulfill({ json: { data: [movie], total: 1 } }),
  )
  await page.route('**/api/movies/movie-1', (route) =>
    route.fulfill({ json: { data: movie } }),
  )
  await page.route('**/api/showtimes?*', (route) =>
    route.fulfill({
      json: {
        data: [{
          id: 'showtime-1',
          movieId: movie.id,
          theaterId: 'theater-1',
          theaterName: 'Hall A',
          startTime: '2099-06-14T10:00:00+07:00',
          endTime: '2099-06-14T12:00:00+07:00',
          slotLabel: 'Morning',
          priceByZone: { FRONT: 180 },
        }],
        total: 1,
      },
    }),
  )

  // Mock seat map details load (uses seatStatus state)
  await page.route('**/api/showtimes/showtime-1', (route) =>
    route.fulfill({
      json: {
        data: {
          id: 'showtime-1',
          movieId: movie.id,
          theaterId: 'theater-1',
          theaterName: 'Hall A',
          startTime: '2099-06-14T10:00:00+07:00',
          endTime: '2099-06-14T12:00:00+07:00',
          slotLabel: 'Morning',
          priceByZone: { FRONT: 180 },
          movie: {
            id: 'movie-1',
            title: 'Test Movie',
          },
          seats: [
            {
              row: 'A',
              number: 1,
              seatLabel: 'A1',
              zone: 'FRONT',
              price: 180,
              status: seatStatus,
              lockedBy: lockedBy,
              lockedUntil: lockedUntil,
            },
          ],
        },
      },
    }),
  )

  // Mock lock seat request (transitions state to LOCKED)
  await page.route('**/api/bookings/locks', (route) => {
    if (route.request().method() === 'POST') {
      seatStatus = 'LOCKED'
      lockedBy = 'user-1'
      lockedUntil = new Date(Date.now() + 300000).toISOString()
      return route.fulfill({
        json: {
          data: {
            showtimeId: 'showtime-1',
            seatLabel: 'A1',
            expiresAt: lockedUntil,
          },
        },
      })
    }
    return route.continue()
  })

  // Mock confirm booking request
  await page.route('**/api/bookings', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: {
          data: {
            id: 'booking-12345678',
            userId: 'user-1',
            showtimeId: 'showtime-1',
            seats: [
              {
                seatLabel: 'A1',
                row: 'A',
                number: 1,
                zone: 'FRONT',
                price: 180,
              },
            ],
            status: 'CONFIRMED',
            totalPrice: 180,
            userEmail: 'user@example.com',
            userName: 'Test User',
            movieTitle: 'Test Movie',
            theaterName: 'Hall A',
            startTime: '2099-06-14T10:00:00+07:00',
            createdAt: new Date().toISOString(),
          },
        },
      })
    }
    return route.continue()
  })

  // Start navigation
  await page.goto('/')

  // 1. Home Page -> click movie
  await expect(page.getByText('Test Movie', { exact: true })).toBeVisible()
  await page.locator('.movie-card').click({ force: true })

  // 2. Movie Detail -> click showtime slot card
  await expect(page).toHaveURL(/\/movies\/movie-1\/showtime$/)
  await page.locator('.slot-card').filter({ hasText: 'Morning' }).click({ force: true })

  // 3. Seat Map Page -> select seat A1 -> click Proceed
  await expect(page).toHaveURL(/\/movies\/movie-1\/showtime\/showtime-1\/select-seat$/)

  // Click seat button labeled "A1"
  const seatBtn = page.getByRole('button', { name: 'Seat A1', exact: false })
  await expect(seatBtn).toBeVisible()
  await seatBtn.click({ force: true })

  // Click proceed button
  const proceedBtn = page.getByRole('button', { name: 'Proceed' })
  await expect(proceedBtn).toBeVisible()
  await proceedBtn.click({ force: true })

  // 4. Confirm/Payment Page -> click pay
  await expect(page).toHaveURL(/\/payment$/)
  await expect(page.getByRole('heading', { name: 'Payment & Confirmation' })).toBeVisible()
  await expect(page.getByText('A1', { exact: true })).toBeVisible()

  const payBtn = page.getByRole('button', { name: /Confirm & Pay/i })
  await expect(payBtn).toBeVisible()
  await payBtn.click({ force: true })

  // 5. Booking Success Page -> check booking code
  await expect(page.getByRole('heading', { name: 'Booking Successful!' })).toBeVisible()
  await expect(page.locator('.booking-code')).toHaveText('12345678')
  await expect(page.locator('.ticket-movie')).toHaveText('Test Movie')
})

test('releases seat when user deselects it', async ({ page }) => {
  await mockAuthenticatedUser(page)

  // Mock seat as already locked by user-1 initially
  await page.route('**/api/showtimes/showtime-1', (route) =>
    route.fulfill({
      json: {
        data: {
          id: 'showtime-1',
          movieId: movie.id,
          theaterId: 'theater-1',
          theaterName: 'Hall A',
          startTime: '2099-06-14T10:00:00+07:00',
          endTime: '2099-06-14T12:00:00+07:00',
          slotLabel: 'Morning',
          priceByZone: { FRONT: 180 },
          movie: { id: 'movie-1', title: 'Test Movie' },
          seats: [
            {
              row: 'A',
              number: 1,
              seatLabel: 'A1',
              zone: 'FRONT',
              price: 180,
              status: 'LOCKED',
              lockedBy: 'user-1',
              lockedUntil: new Date(Date.now() + 300000).toISOString(),
            },
          ],
        },
      },
    }),
  )

  await page.route('**/api/bookings/locks', (route) => {
    if (route.request().method() === 'DELETE') {
      return route.fulfill({ status: 204 })
    }
    return route.continue()
  })

  await page.goto('/movies/movie-1/showtime/showtime-1/select-seat')

  // The seat button labeled "A1" should be visible and selected initially
  const seatBtn = page.getByRole('button', { name: 'Seat A1', exact: false })
  await expect(seatBtn).toBeVisible()

  // Click the seat to deselect it and wait for DELETE to locks API call
  const [releaseResponse] = await Promise.all([
    page.waitForResponse(resp => resp.url().includes('/api/bookings/locks') && resp.request().method() === 'DELETE'),
    seatBtn.click({ force: true })
  ])

  expect(releaseResponse.status()).toBe(204)
  expect(JSON.parse(releaseResponse.request().postData() || '{}')).toEqual({
    showtimeId: 'showtime-1',
    seatLabel: 'A1',
  })
})

test('shows timeout modal and redirects to seat map when reservation expires', async ({ page }) => {
  await mockAuthenticatedUser(page)

  let seatStatus = 'AVAILABLE'
  let lockedBy: string | undefined = undefined
  let lockedUntil: string | undefined = undefined

  // Mock initial pages load
  await page.route('**/api/movies/movie-1', (route) =>
    route.fulfill({ json: { data: movie } }),
  )

  // Mock seat map details load
  await page.route('**/api/showtimes/showtime-1', (route) =>
    route.fulfill({
      json: {
        data: {
          id: 'showtime-1',
          movieId: movie.id,
          theaterId: 'theater-1',
          theaterName: 'Hall A',
          startTime: '2099-06-14T10:00:00+07:00',
          endTime: '2099-06-14T12:00:00+07:00',
          slotLabel: 'Morning',
          priceByZone: { FRONT: 180 },
          movie: { id: 'movie-1', title: 'Test Movie' },
          seats: [
            {
              row: 'A',
              number: 1,
              seatLabel: 'A1',
              zone: 'FRONT',
              price: 180,
              status: seatStatus,
              lockedBy: lockedBy,
              lockedUntil: lockedUntil,
            },
          ],
        },
      },
    }),
  )

  // Mock lock seat request with short expiry (1.5 seconds)
  let shortExpiry: string
  await page.route('**/api/bookings/locks', (route) => {
    if (route.request().method() === 'POST') {
      shortExpiry = new Date(Date.now() + 1500).toISOString()
      seatStatus = 'LOCKED'
      lockedBy = 'user-1'
      lockedUntil = shortExpiry
      return route.fulfill({
        json: {
          data: {
            showtimeId: 'showtime-1',
            seatLabel: 'A1',
            expiresAt: shortExpiry,
          },
        },
      })
    }
    return route.continue()
  })

  // Start at seat map page
  await page.goto('/movies/movie-1/showtime/showtime-1/select-seat')

  // Click seat button labeled "A1"
  const seatBtn = page.getByRole('button', { name: 'Seat A1', exact: false })
  await expect(seatBtn).toBeVisible()
  await seatBtn.click({ force: true })

  // Click proceed button
  const proceedBtn = page.getByRole('button', { name: 'Proceed' })
  await expect(proceedBtn).toBeVisible()
  await proceedBtn.click({ force: true })

  // 4. Confirm/Payment Page
  await expect(page).toHaveURL(/\/payment$/)
  await expect(page.getByRole('heading', { name: 'Payment & Confirmation' })).toBeVisible()

  // Wait 2 seconds for reservation to expire
  await page.waitForTimeout(2000)

  // The Session Expired modal should show up
  const expiredModal = page.locator('.modal-overlay', { hasText: 'Session Expired' })
  await expect(expiredModal).toBeVisible()

  // Intercept the unlock delete call when they click OK
  let unlockCalled = false
  await page.route('**/api/bookings/locks', (route) => {
    if (route.request().method() === 'DELETE') {
      unlockCalled = true
      return route.fulfill({ status: 204 })
    }
    return route.continue()
  })

  // Click OK on the modal
  await page.getByRole('button', { name: 'OK' }).click({ force: true })

  // Verify redirect back to select-seat page and that the unlock API was invoked
  await expect(page).toHaveURL(/\/select-seat$/)
  expect(unlockCalled).toBe(true)
})


