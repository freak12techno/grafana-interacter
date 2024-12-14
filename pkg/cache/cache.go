package cache

import (
	"github.com/rs/zerolog"
)

type Cache struct {
	cache  map[string]string
	logger zerolog.Logger
}

func NewCache(logger *zerolog.Logger) *Cache {
	return &Cache{
		cache:  map[string]string{},
		logger: logger.With().Str("component", "cache").Logger(),
	}
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
	return key
}

func (c *Cache) Delete(key string) {
	delete(c.cache, key)
	c.logger.Trace().
		Str("key", key).
		Int("len", len(c.cache)).
		Msg("Cache delete item")
}
