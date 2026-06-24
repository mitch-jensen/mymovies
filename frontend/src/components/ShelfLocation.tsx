import { displayPosition } from '../lib/displayPosition'

interface ShelfLocationProps {
  bookcase: string
  shelf: number
  slot: number
}

// Renders a placed release's physical location. Shelf and slot are stored
// Zero-based but shown to the user one-based.
export function ShelfLocation({ bookcase, shelf, slot }: ShelfLocationProps) {
  return (
    <span>
      {bookcase} · shelf {displayPosition(shelf)} · slot {displayPosition(slot)}
    </span>
  )
}
