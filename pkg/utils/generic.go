package utils

func Filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

func Map[T, V any](slice []T, f func(T) V) []V {
	n := make([]V, len(slice))
	for index, e := range slice {
		n[index] = f(e)
	}
	return n
}
