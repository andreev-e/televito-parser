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
	// Если ключ уже есть, обновляем его значение
	if _, ok := l.cache[key]; ok {
		l.cache[key] = value
		l.updateKey(key)
		return
	}

	// Если достигли предела, удаляем самый старый ключ и его значение
	if len(l.cache) == l.capacity {
		oldestKey := l.keys[0]
		delete(l.cache, oldestKey)
		l.keys = l.keys[1:]
	}

	// Добавляем новый ключ и его значение
	l.cache[key] = value
	l.keys = append(l.keys, key)
}

func (l *LRUCache) updateKey(key string) {
	// Удаляем ключ из текущего положения
	for i, k := range l.keys {
		if k == key {
			l.keys = append(l.keys[:i], l.keys[i+1:]...)
			break
		}
	}
	// Помещаем ключ в конец списка
	l.keys = append(l.keys, key)
}

func main() {
	cache := Constructor(3)

	cache.Put("1", "a")
	cache.Put("2", "b")
	cache.Put("3", "c")

	fmt.Println(cache.Get("1")) // a
	fmt.Println(cache.Get("2")) // b
	fmt.Println(cache.Get("3")) // c

	cache.Put("4", "d")

	fmt.Println(cache.Get("1")) // ""
	fmt.Println(cache.Get("2")) // b
	fmt.Println(cache.Get("3")) // c
	fmt.Println(cache.Get("4")) // d
}
