package api

// mapSlice converts a slice by applying convert to each element. It captures the
// make-and-loop boilerplate the list handlers would otherwise repeat, and always
// returns a non-nil slice so an empty result serialises as [] rather than null.
func mapSlice[T, U any](in []T, convert func(T) U) []U {
	out := make([]U, len(in))
	for index, value := range in {
		out[index] = convert(value)
	}

	return out
}
