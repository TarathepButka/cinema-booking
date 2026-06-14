import process from 'node:process'
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? [['line'], ['html', { open: 'never' }]] : 'line',
  use: {
    baseURL: process.env.CI ? 'http://localhost:4173' : 'http://localhost:5173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    headless: process.env.HEADED !== 'true',
  },
  projects: [
    {
      name: 'chromium',
      use: devices['Desktop Chrome'],
    },
    {
      name: 'mobile-chromium',
      use: devices['Pixel 5'],
    },
  ],
  webServer: {
    command: process.env.CI ? 'npm run preview' : 'npm run dev',
    port: process.env.CI ? 4173 : 5173,
    reuseExistingServer: !process.env.CI,
  },
})
