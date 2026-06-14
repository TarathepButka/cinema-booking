import { describe, it, expect, vi } from 'vitest'

import { mount } from '@vue/test-utils'
import App from '../App.vue'

vi.mock('@/components/NavBar.vue', () => ({
  default: {
    name: 'NavBar',
    template: '<nav />',
  },
}))

describe('App', () => {
  it('mounts the application shell', () => {
    const wrapper = mount(App, {
      global: {
        mocks: {
          $route: { meta: {} },
        },
        stubs: {
          NavBar: true,
          RouterView: true,
        },
      },
    })
    expect(wrapper.find('#app-root').exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'NavBar' }).exists()).toBe(true)
  })

  it('hides the navbar when the route requests it', () => {
    const wrapper = mount(App, {
      global: {
        mocks: {
          $route: { meta: { hideNavbar: true } },
        },
        stubs: {
          NavBar: true,
          RouterView: true,
        },
      },
    })

    expect(wrapper.findComponent({ name: 'NavBar' }).exists()).toBe(false)
  })
})
