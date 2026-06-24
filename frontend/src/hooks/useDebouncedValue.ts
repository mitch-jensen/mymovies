import { useEffect, useState } from 'react'

// Returns `value` delayed by `delayMs`, so search-as-you-type doesn't fire a
// Request on every keystroke. The debounce is a UI concern, so it lives here
// Rather than in the generated client.
export function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delayMs)

    return () => clearTimeout(id)
  }, [value, delayMs])

  return debounced
}
