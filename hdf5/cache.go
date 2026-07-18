package hdf5

import (
	"container/list"
	"sync"
)

type chunkCacheEntry struct {
	chunkIndex uint64
	data       []byte
	accessed   bool
	element    *list.Element
}

type ChunkCache struct {
	capacity int
	size     int
	cache    map[uint64]*chunkCacheEntry
	list     *list.List
	mutex    sync.RWMutex
	prefetch int
}

func NewChunkCache(capacity int, prefetch int) *ChunkCache {
	return &ChunkCache{
		capacity: capacity,
		cache:    make(map[uint64]*chunkCacheEntry),
		list:     list.New(),
		prefetch: prefetch,
	}
}

func (c *ChunkCache) Get(chunkIndex uint64) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if entry, ok := c.cache[chunkIndex]; ok {
		c.list.MoveToFront(entry.element)
		return entry.data, true
	}
	return nil, false
}

func (c *ChunkCache) Set(chunkIndex uint64, data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entry, ok := c.cache[chunkIndex]; ok {
		entry.data = data
		c.list.MoveToFront(entry.element)
		return
	}

	newSize := c.size + len(data)
	for newSize > c.capacity && c.list.Len() > 0 {
		oldest := c.list.Back()
		if oldest != nil {
			entry := oldest.Value.(*chunkCacheEntry)
			c.list.Remove(oldest)
			delete(c.cache, entry.chunkIndex)
			newSize -= len(entry.data)
		}
	}

	entry := &chunkCacheEntry{
		chunkIndex: chunkIndex,
		data:       data,
	}
	entry.element = c.list.PushFront(entry)
	c.cache[chunkIndex] = entry
	c.size = newSize
}

func (c *ChunkCache) Remove(chunkIndex uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entry, ok := c.cache[chunkIndex]; ok {
		c.list.Remove(entry.element)
		delete(c.cache, chunkIndex)
		c.size -= len(entry.data)
	}
}

func (c *ChunkCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[uint64]*chunkCacheEntry)
	c.list = list.New()
	c.size = 0
}

func (c *ChunkCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.size
}

func (c *ChunkCache) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

type FileCache struct {
	chunkCaches map[uint64]*ChunkCache
	mutex       sync.RWMutex
	defaultSize int
}

var globalFileCache = &FileCache{
	chunkCaches: make(map[uint64]*ChunkCache),
	defaultSize: 1024 * 1024 * 10,
}

func (fc *FileCache) GetOrCreate(datasetAddr uint64, capacity int) *ChunkCache {
	fc.mutex.RLock()
	if cache, ok := fc.chunkCaches[datasetAddr]; ok {
		fc.mutex.RUnlock()
		return cache
	}
	fc.mutex.RUnlock()

	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	if cache, ok := fc.chunkCaches[datasetAddr]; ok {
		return cache
	}

	if capacity <= 0 {
		capacity = fc.defaultSize
	}

	cache := NewChunkCache(capacity, 2)
	fc.chunkCaches[datasetAddr] = cache
	return cache
}

func (fc *FileCache) Remove(datasetAddr uint64) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	delete(fc.chunkCaches, datasetAddr)
}


