import { describe, it, expect } from 'vitest'
import BackendHealthProvider from './BackendHealthProvider'
import { useBackendHealth } from './useBackendHealth'

describe('BackendHealthProvider', () => {
  it('is a function component', () => {
    expect(typeof BackendHealthProvider).toBe('function')
  })

  it('has a display name or is a named function', () => {
    expect(BackendHealthProvider.name).toBe('BackendHealthProvider')
  })
})

describe('useBackendHealth', () => {
  it('is a function', () => {
    expect(typeof useBackendHealth).toBe('function')
  })
})
