// People count shelves and slots from 1 when organising a collection; the API
// Stores them from 0. Use this wherever a stored position is shown to the user.
export function displayPosition(zeroBasedIndex: number): number {
  return zeroBasedIndex + 1
}
