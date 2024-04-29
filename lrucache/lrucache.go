package Lrucache

import "fmt"

var CachedLocations LRUCache

type LRUCache struct {
	capacity int
	cache    map[string]string
	keys     []string // для отслеживания порядка использования
}

func Constructor(capacity int) LRUCache {
	return LRUCache{
		capacity: capacity,
		cache:    make(map[string]string),
		keys:     make([]string, 0),
	}
}

func (l *LRUCache) Get(key string) (string, error) {
	if val, ok := l.cache[key]; ok {
		l.updateKey(key)
		return val, nil
	}
	return "", fmt.Errorf("Key not found")
}

func (l *LRUCache) Put(key, value string) {
	if _, ok := l.cache[key]; ok {
		l.cache[key] = value
		l.updateKey(key)
		return
	}

	if len(l.cache) == l.capacity {
		oldestKey := l.keys[0]
		delete(l.cache, oldestKey)
		l.keys = l.keys[1:]
	}

	l.cache[key] = value
	l.keys = append(l.keys, key)
}

func (l *LRUCache) updateKey(key string) {
	for i, k := range l.keys {
		if k == key {
			l.keys = append(l.keys[:i], l.keys[i+1:]...)
			break
		}
	}

	l.keys = append(l.keys, key)
}
