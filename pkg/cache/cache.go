package cache

import "github.com/google/uuid"

type Cache struct {
	cache map[string]string
}

func NewCache() *Cache {
	return &Cache{
		cache: map[string]string{},
	}
}

func (c *Cache) Get(key string) (string, bool) {
	value, found := c.cache[key]
	return value, found
}

func (c *Cache) Set(value string) string {
	key := uuid.New().String()[0:8]
	c.cache[key] = value
	return key
}
