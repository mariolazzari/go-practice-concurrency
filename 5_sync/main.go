package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Entry struct {
	value      any
	expiration time.Time
}

type Cache struct {
	size int
	ttl  time.Duration

	mu sync.Mutex
	m  map[string]Entry
	// maintain insertion order to evict oldest when full
	keys []string
}

func New(size int, ttl time.Duration) (*Cache, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be positive")
	}
	return &Cache{
		size: size,
		ttl:  ttl,
		m:    make(map[string]Entry),
	}, nil
}

func (c *Cache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = nil
	c.keys = nil
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, found := c.m[key]
	if !found {
		return nil, false
	}

	// expired?
	if time.Since(entry.expiration) > 0 {
		delete(c.m, key)
		c.removeKey(key)
		return nil, false
	}
	return entry.value, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// if exists, update value and expiration, no need to reorder
	if _, found := c.m[key]; found {
		c.m[key] = Entry{
			value:      value,
			expiration: time.Now().Add(c.ttl),
		}
		return
	}

	// if full, evict oldest
	if len(c.m) >= c.size {
		oldest := c.keys[0]
		delete(c.m, oldest)
		c.keys = c.keys[1:]
	}

	c.m[key] = Entry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
	c.keys = append(c.keys, key)
}

func (c *Cache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.m))
	for _, k := range c.keys {
		if _, found := c.m[k]; found {
			keys = append(keys, k)
		}
	}
	return keys
}

func (c *Cache) removeKey(key string) {
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			return
		}
	}
}

func main() {
	keyFmt := "key-%02d"
	keyName := func(i int) string { return fmt.Sprintf(keyFmt, i) }

	size := 5
	ttl := 10 * time.Millisecond
	log.Printf("info: creating cache: size=%d, ttl=%v", size, ttl)
	c, err := New(size, ttl)
	if err != nil {
		log.Printf("error: can't create - %s", err)
		return
	}
	log.Printf("info: OK")

	log.Printf("info: checking TTL")
	key, val := keyName(1), 3
	c.Set(key, val)
	v, ok := c.Get(key)
	if !ok || v != val {
		log.Printf("error: %q: got %v (ok=%v)", key, v, ok)
		return
	}

	// Let key expire
	time.Sleep(2 * ttl)
	_, ok = c.Get(key)
	if ok {
		log.Printf("error: %q: got value after TTL", key)
		return
	}
	log.Printf("info: OK")

	log.Printf("info: checking overflow")
	n := size * 2
	for i := 0; i < n; i++ {
		c.Set(keyName(i), i)
	}
	_, ok = c.Get(keyName(1))
	if ok {
		log.Printf("error: %q: got value after overflow", key)
		return
	}
	_, ok = c.Get(keyName(n - 1))
	if !ok {
		log.Printf("error: %q: not found", keyName(n-1))
		return
	}
	log.Printf("info: OK")

	numGr := size * 3
	count := 1000
	log.Printf("info: checking concurrency (%d goroutines, %d loops each)", numGr, count)

	var wg sync.WaitGroup
	wg.Add(numGr)
	for i := 0; i < numGr; i++ {
		key := keyName(i)
		go func() {
			defer wg.Done()
			for i := 0; i < count; i++ {
				time.Sleep(time.Microsecond)
				c.Set(key, i)
			}
		}()
	}
	wg.Wait()
	log.Printf("info: OK")
}
