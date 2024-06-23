package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

// NewCache creates new ready-to-use LRU cache.
func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

// Set adds value by its key into the cache.
// Returns true if key exists in the cache.
func (c *lruCache) Set(key Key, value interface{}) bool {
	if i, ok := c.get(key); ok {
		i.Value = value
		return true
	}

	i := c.queue.PushFront(value)
	if c.queue.Len() > c.capacity {
		c.queue.Remove(c.queue.Back())
		delete(c.items, key)
	}

	c.items[key] = i

	return false
}

// Get tries to get value from the cache by its key.
// Returns (nil, false) when key is not in the cache.
func (c *lruCache) Get(key Key) (interface{}, bool) {
	i, ok := c.get(key)
	if !ok {
		return nil, false
	}

	return i.Value, true
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
	c.queue = NewList()
	clear(c.items)
}
