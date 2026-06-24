import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ShelfLocation } from './ShelfLocation'

describe('ShelfLocation', () => {
  it('shows stored zero-based shelf and slot as one-based for the user', () => {
    render(<ShelfLocation bookcase="Lounge" shelf={0} slot={0} />)

    expect(screen.getByText('Lounge · shelf 1 · slot 1')).toBeInTheDocument()
  })
})
