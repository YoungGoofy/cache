package cache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Clear()
	Add(key, value interface{})
	Get(key interface{}) (value interface{}, ok bool)
	Remove(key interface{})
	AddWithTTL(key, value interface{}, ttl time.Duration)
}

type Cache struct {
	sync.RWMutex
	items             map[interface{}]*list.Element
	capacity          int           // емкость кеша
	evictList         *list.List    // двусвязный список
	defaultExpiration time.Duration // дефолтное время жизни
	cleanupInterval   time.Duration // время очистки кеша
}

type Items struct {
	Key        interface{}
	Value      interface{}
	Created    time.Time
	Expiration int64
}

func New(defaultExpiration, cleanupInterval time.Duration, capacity int) *Cache {
	cache := Cache{
		items:             make(map[interface{}]*list.Element),
		capacity:          capacity,
		evictList:         list.New(),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	// Запускаю очистку просроченного кэша
	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

func (c *Cache) Add(key, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if _, exists := c.items[key]; exists {
		log.Printf("Key '%v' already exists", key)
		return
	}

	// Если емкость меньше длины списка, запись удаляется
	if c.evictList.Len() >= c.capacity {
		lastElem := c.evictList.Back()
		delete(c.items, lastElem.Value.(Items).Key)
		c.evictList.Remove(lastElem)
	}

	item := Items{
		Key:     key,
		Value:   value,
		Created: time.Now(),
	}

	// добавляется запись вперед
	elem := c.evictList.PushFront(item)
	c.items[key] = elem
	log.Printf("Save element '%v' with key '%v'", value, key)
}

func (c *Cache) AddWithTTL(key, value interface{}, ttl time.Duration) {
	c.Lock()
	defer c.Unlock()

	var expiration int64

	if _, exists := c.items[key]; exists {
		log.Printf("Key '%v' already exists", key)
		return
	}

	// если ttl не указан, выставляется дефолтное значение
	if ttl == 0 {
		ttl = c.defaultExpiration
	}

	// устанавливаем время истечения кэша
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	if c.evictList.Len() >= c.capacity {
		lastElem := c.evictList.Back()
		delete(c.items, lastElem.Value.(Items).Key)
		c.evictList.Remove(lastElem)
	}

	item := Items{
		Key:        key,
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}
	elem := c.evictList.PushFront(item)
	c.items[key] = elem
	log.Printf("Save element '%v' with key '%v'", value, key)
}

func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	c.RLock()
	defer c.RUnlock()

	if item, ok := c.items[key]; ok {
		// если данные вызывают, значит они еще актуальны, поэтому перекидываем данные вперед списка
		c.evictList.MoveToFront(item)
		return item.Value.(Items).Value, ok
	}

	return nil, false
}

func (c *Cache) Remove(key interface{}) {
	c.Lock()
	defer c.Unlock()

	element, found := c.items[key]
	if !found {
		log.Println("Key not found")
		return
	}
	c.evictList.Remove(element)
	delete(c.items, key)
	log.Printf("Removing item with key '%v'", key)
}

func (c *Cache) Cap() int {
	return c.capacity
}

func (c *Cache) Clear() {
	c.Lock()
	defer c.Unlock()

	c.evictList = c.evictList.Init()
	c.items = make(map[interface{}]*list.Element)
}

func (c *Cache) GetAll() map[interface{}]Items {
	c.RLock()
	defer c.RUnlock()

	items := make(map[interface{}]Items)
	for key, value := range c.items {
		items[key] = value.Value.(Items)
	}
	return items
}
