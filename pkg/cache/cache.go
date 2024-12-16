package cache

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog"
)

type Cache struct {
	cachePath string
	cache     map[string]string
	logger    zerolog.Logger
}

func NewCache(logger *zerolog.Logger, cachePath string) *Cache {
	cache := &Cache{
		cache:     map[string]string{},
		cachePath: cachePath,
		logger:    logger.With().Str("component", "cache").Logger(),
	}

	cache.Load()
	return cache
}

func (c *Cache) Get(key string) (string, bool) {
	value, found := c.cache[key]
	return value, found
}

func (c *Cache) Set(key, value string) string {
	c.cache[key] = value
	c.logger.Trace().
		Str("key", key).
		Str("value", value).
		Int("len", len(c.cache)).
		Msg("Cache set item")

	c.Save()

	return key
}

func (c *Cache) Delete(key string) {
	delete(c.cache, key)
	c.logger.Trace().
		Str("key", key).
		Int("len", len(c.cache)).
		Msg("Cache delete item")

	c.Save()
}

func (c *Cache) Length() int {
	return len(c.cache)
}

func (c *Cache) Load() {
	if c.cachePath == "" {
		return
	}

	bytes, err := os.ReadFile(c.cachePath)
	if err != nil {
		c.logger.Warn().Err(err).Msg("Error loading cache")
		return
	}

	if unmarshalErr := json.Unmarshal(bytes, &c.cache); unmarshalErr != nil {
		c.logger.Warn().Err(err).Msg("Error parsing cache")
	}
}

func (c *Cache) Save() {
	if c.cachePath == "" {
		return
	}

	bytes, _ := json.Marshal(c.cache) //nolint:errchkjson

	if err := os.WriteFile(c.cachePath, bytes, 0o755); err != nil {
		c.logger.Warn().Err(err).Msg("Error writing cache")
	}
}
