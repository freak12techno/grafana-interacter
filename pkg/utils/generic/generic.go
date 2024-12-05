package generic

import "math"

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

func Paginate[T any](input []T, page, perPage int) ([]T, int) {
	lowerBound := page * perPage
	upperBound := (page + 1) * perPage
	totalPages := math.Ceil(float64(len(input)) / float64(perPage))

	if len(input) < lowerBound {
		lowerBound = len(input)
	}

	if len(input) < upperBound {
		upperBound = len(input)
	}

	return input[lowerBound:upperBound], int(totalPages)
}
