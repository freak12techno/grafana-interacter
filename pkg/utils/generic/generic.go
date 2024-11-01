package generic

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

func Find[T any](slice []T, f func(T) bool) (*T, bool) {
	for _, item := range slice {
		if f(item) {
			return &item, true
		}
	}

	return nil, false
}

func MergeMaps(first, second map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range first {
		result[key] = value
	}

	for key, value := range second {
		result[key] = value
	}

	return result
}

func SplitArrayIntoChunks[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T

	var currentChunk []T
	for i := range slice {
		currentChunk = append(currentChunk, slice[i])
		if len(currentChunk) == chunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = []T{}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
