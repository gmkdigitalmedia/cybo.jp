package tiler

import (
	"container/list"
	"sync"
	"time"
)

type TileCache struct {
	maxSize   int64
	currentSize int64
	items     map[string]*list.Element
	lru       *list.List
	mu        sync.RWMutex
	hits      uint64
	misses    uint64
}

type cacheEntry struct {
	key       string
	value     *TileResponse
	size      int64
	timestamp time.Time
}

func NewTileCache(maxSizeBytes int64) *TileCache {
	return &TileCache{
		maxSize: maxSizeBytes,
		items:   make(map[string]*list.Element),
		lru:     list.New(),
	}
}

func (c *TileCache) Get(key string) (*TileResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		c.misses++
		return nil, false
	}

	c.hits++
	c.lru.MoveToFront(elem)
	entry := elem.Value.(*cacheEntry)
	
	// Return a copy to prevent modification
	return &TileResponse{
		Data:        append([]byte(nil), entry.value.Data...),
		Width:       entry.value.Width,
		Height:      entry.value.Height,
		ContentType: entry.value.ContentType,
		CacheKey:    entry.value.CacheKey,
	}, true
}

func (c *TileCache) Set(key string, value *TileResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := int64(len(value.Data))

	// Check if key exists, update it
	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*cacheEntry)
		c.currentSize -= entry.size
		c.currentSize += size
		entry.value = value
		entry.size = size
		entry.timestamp = time.Now()
		c.lru.MoveToFront(elem)
		return
	}

	// Evict until we have space
	for c.currentSize+size > c.maxSize && c.lru.Len() > 0 {
		c.evictOldest()
	}

	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		size:      size,
		timestamp: time.Now(),
	}
	elem := c.lru.PushFront(entry)
	c.items[key] = elem
	c.currentSize += size
}

func (c *TileCache) evictOldest() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}

	entry := elem.Value.(*cacheEntry)
	c.lru.Remove(elem)
	delete(c.items, entry.key)
	c.currentSize -= entry.size
}

func (c *TileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lru = list.New()
	c.currentSize = 0
}

func (c *TileCache) Stats() (hits, misses uint64, size int64, count int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.hits, c.misses, c.currentSize, c.lru.Len()
}

func (c *TileCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}

// Prefetch tiles that are likely to be requested soon
func (c *TileCache) Prefetch(keys []string, fetcher func(string) (*TileResponse, error)) {
	for _, key := range keys {
		if _, ok := c.Get(key); ok {
			continue // Already cached
		}

		go func(k string) {
			resp, err := fetcher(k)
			if err == nil {
				c.Set(k, resp)
			}
		}(key)
	}
}
