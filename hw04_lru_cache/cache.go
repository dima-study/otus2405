package hw04lrucache

import "sync"

// Key represents Cache key.
type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

// lruItemValue represents items value for lruCache.
// Duplicate key to remove from lruCache.items map due cache overflow.
// value holds cached value.
type lruItemValue struct {
	key   Key
	value interface{}
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem

	mx sync.RWMutex
}

// NewCache creates new ready-to-use LRU cache.
func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		mx:       sync.RWMutex{},
	}
}

// Set adds value by its key into the cache.
// Returns true if key exists in the cache.
func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mx.Lock()
	defer c.mx.Unlock()

	iv := lruItemValue{
		key:   key,
		value: value,
	}
	if i, ok := c.get(key); ok {
		i.Value = iv
		return true
	}

	i := c.queue.PushFront(iv)
	if c.queue.Len() > c.capacity {
		back := c.queue.Back()
		c.queue.Remove(back)

		backIv := back.Value.(lruItemValue)
		delete(c.items, backIv.key)
	}

	c.items[key] = i

	return false
}

// Get tries to get value from the cache by its key.
// Returns (nil, false) when key is not in the cache.
func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	i, ok := c.get(key)
	if !ok {
		return nil, false
	}

	iv := i.Value.(lruItemValue)
	return iv.value, true
}

// get tries o get value from the cache by its key, and moves it to front when found.
func (c *lruCache) get(key Key) (*ListItem, bool) {
	i, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.queue.MoveToFront(i)

	return i, true
}

// Clear resets the cache.
func (c *lruCache) Clear() {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.queue = NewList()
	clear(c.items)
}
