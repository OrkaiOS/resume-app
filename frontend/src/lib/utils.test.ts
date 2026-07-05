import { describe, it, expect } from 'vitest'
import { cn } from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional classes', () => {
    const hidden = false
    expect(cn('base', hidden && 'hidden', 'extra')).toBe('base extra')
  })

  it('deduplicates tailwind classes', () => {
    expect(cn('px-4', 'px-2')).toBe('px-2')
  })
})