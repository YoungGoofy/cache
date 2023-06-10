package cache

import "time"


func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {
	for {
		// ждем интервал
		<- time.After(c.cleanupInterval)
		if c.items == nil {
			return
		}

		// поиск ключей, которые можно удалить
		if keys := c.expiredKeys(); len(keys) != 0 {
			// удаление ключей
			c.cleanItems(keys)
		}
	}
}

func (c *Cache) expiredKeys() (keys []interface{}) {
	c.RLock()
	defer c.RUnlock()

	for key, value := range c.items {
		if time.Now().UnixNano() > value.Value.(Items).Expiration && value.Value.(Items).Expiration > 0 {
			keys = append(keys, key)
		}
	}
	return
}

func (c *Cache) cleanItems(keys []interface{}) {
	c.Lock()
	defer c.Unlock()

	for _, key := range keys {
		element := c.items[key]
		c.evictList.Remove(element)
		delete(c.items, key)
	}
}