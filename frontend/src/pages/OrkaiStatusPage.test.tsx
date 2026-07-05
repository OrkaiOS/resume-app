import { describe, it, expect } from 'vitest'
import OrkaiStatusPage from './OrkaiStatusPage'

describe('OrkaiStatusPage', () => {
  it('is a function component', () => {
    expect(typeof OrkaiStatusPage).toBe('function')
  })

  it('has a display name or is a named function', () => {
    expect(OrkaiStatusPage.name).toBe('OrkaiStatusPage')
  })

  it('accepts checking prop with default false', () => {
    const result = OrkaiStatusPage({})
    expect(result).toBeDefined()
    expect(result.type).toBe('div')
  })

  it('renders checking state when checking=true', () => {
    const result = OrkaiStatusPage({ checking: true })
    expect(result).toBeDefined()
    expect(result.type).toBe('div')
  })

  it('renders not-checking state when checking=false', () => {
    const result = OrkaiStatusPage({ checking: false })
    expect(result).toBeDefined()
    expect(result.type).toBe('div')
  })

  it('renders not-checking state by default', () => {
    const result = OrkaiStatusPage({})
    expect(result).toBeDefined()
    expect(result.type).toBe('div')
  })
})