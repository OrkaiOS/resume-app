import { describe, it, expect } from 'vitest'
import { OrkaiHealthProvider, useOrkaiHealth } from './useOrkaiHealth'

describe('OrkaiHealthProvider', () => {
  it('is a function component', () => {
    expect(typeof OrkaiHealthProvider).toBe('function')
  })

  it('has a display name or is a named function', () => {
    expect(OrkaiHealthProvider.name).toBe('OrkaiHealthProvider')
  })
})

describe('useOrkaiHealth', () => {
  it('is a function', () => {
    expect(typeof useOrkaiHealth).toBe('function')
  })
})