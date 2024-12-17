package cache

import (
	"encoding/json"
	"main/pkg/fs"

	"github.com/rs/zerolog"
)

type Cache struct {
	filesystem fs.FS
	cachePath  string
	cache      map[string]string
	logger     zerolog.Logger
}

func NewCache(logger *zerolog.Logger, filesystem fs.FS, cachePath string) *Cache {
	return &Cache{
		cache:      map[string]string{},
		filesystem: filesystem,
		cachePath:  cachePath,
		logger:     logger.With().Str("component", "cache").Logger(),
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
		c.logger.Trace().Msg("Not using persistent cache, not loading cache from file")
		return
	}

	bytes, err := c.filesystem.ReadFile(c.cachePath)
	if err != nil {
		c.logger.Warn().Err(err).Msg("Error loading cache")
		return
	}

	if unmarshalErr := json.Unmarshal(bytes, &c.cache); unmarshalErr != nil {
		c.logger.Warn().Err(err).Msg("Error parsing cache")
	}

	c.logger.Trace().Int("len", len(c.cache)).Msg("Cache loaded")
}

func (c *Cache) Save() {
	if c.cachePath == "" {
		return
	}

	bytes, _ := json.Marshal(c.cache) //nolint:errchkjson

	if err := c.filesystem.WriteFile(c.cachePath, bytes, 0o755); err != nil {
		c.logger.Warn().Err(err).Msg("Error writing cache")
	}
}
