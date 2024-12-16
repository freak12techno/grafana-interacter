package cache

import (
	"errors"
	"main/pkg/fs"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheLoadFailedToLoad(t *testing.T) {
	t.Parallel()

	cache := NewCache(loggerPkg.GetNopLogger(), &fs.TestFS{}, "not-found.json")
	cache.Load()
	require.Equal(t, 0, cache.Length())
}

func TestCacheLoadFailedToParse(t *testing.T) {
	t.Parallel()

	cache := NewCache(loggerPkg.GetNopLogger(), &fs.TestFS{}, "invalid.yml")
	cache.Load()
	require.Equal(t, 0, cache.Length())
}

func TestCacheLoadOk(t *testing.T) {
	t.Parallel()

	cache := NewCache(loggerPkg.GetNopLogger(), &fs.TestFS{}, "cache.json")
	cache.Load()
	require.Equal(t, 2, cache.Length())
}

func TestCacheSaveFailed(t *testing.T) {
	t.Parallel()

	cache := NewCache(loggerPkg.GetNopLogger(), &fs.TestFS{
		WriteError: errors.New("custom error"),
	}, "cache.json")
	cache.Save()
}
