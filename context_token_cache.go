package weixinbot

import (
	"container/list"
	"sync"
)

const defaultContextTokenCacheMax = 10_000

// contextTokenCache is an LRU map of userID -> context_token with a fixed capacity.
type contextTokenCache struct {
	mu    sync.Mutex
	max   int
	items map[string]string
	order *list.List
	index map[string]*list.Element
}

func newContextTokenCache(max int) *contextTokenCache {
	if max <= 0 {
		max = defaultContextTokenCacheMax
	}
	return &contextTokenCache{
		max:   max,
		items: make(map[string]string),
		order: list.New(),
		index: make(map[string]*list.Element),
	}
}

func (c *contextTokenCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.items[key]
	if !ok {
		return "", false
	}
	if elem := c.index[key]; elem != nil {
		c.order.MoveToFront(elem)
	}
	return v, true
}

func (c *contextTokenCache) Set(key, val string) {
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if val == "" {
		c.deleteLocked(key)
		return
	}
	if elem, exists := c.index[key]; exists {
		c.items[key] = val
		c.order.MoveToFront(elem)
		return
	}
	for c.order.Len() >= c.max {
		c.evictOldestLocked()
	}
	elem := c.order.PushFront(key)
	c.index[key] = elem
	c.items[key] = val
}

func (c *contextTokenCache) deleteLocked(key string) {
	if elem := c.index[key]; elem != nil {
		c.order.Remove(elem)
		delete(c.index, key)
	}
	delete(c.items, key)
}

func (c *contextTokenCache) evictOldestLocked() {
	elem := c.order.Back()
	if elem == nil {
		return
	}
	key := elem.Value.(string)
	c.order.Remove(elem)
	delete(c.index, key)
	delete(c.items, key)
}

func (c *contextTokenCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]string)
	c.order = list.New()
	c.index = make(map[string]*list.Element)
}
