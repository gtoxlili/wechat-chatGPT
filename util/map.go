package util

import "sync"

type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Load(key K) (V, bool) {
	if value, ok := m.m.Load(key); ok {
		return value.(V), true
	}
	return *new(V), false
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) Len() int {
	count := 0
	m.Range(func(_ K, _ V) bool {
		count++
		return true
	})
	return count
}

func NewSyncMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: sync.Map{}}
}
