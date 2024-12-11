package cache

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

func (c *Cache) Set(key, value string) string {
	c.cache[key] = value
	return key
}

func (c *Cache) Delete(key string) {
	delete(c.cache, key)
}
