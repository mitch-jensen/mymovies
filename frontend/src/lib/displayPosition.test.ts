import { describe, expect, it } from 'vitest'
import { displayPosition } from './displayPosition'

describe('displayPosition', () => {
  it('shows a zero-based index as a one-based number for humans', () => {
    expect(displayPosition(0)).toBe(1)
  })

  it('keeps later positions sequential', () => {
    expect(displayPosition(3)).toBe(4)
  })
})
