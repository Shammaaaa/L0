package cache

import (
	"context"
	"sync"
	"time"

	"order/domain"
)

// InMemory простая и дефолтная реализация кеша в памяти с мапой и мютексом под капотом
type InMemory struct {
	mutex *sync.RWMutex
	items map[string]domain.Order
}

func NewInMemory() *InMemory {
	return &InMemory{
		mutex: &sync.RWMutex{},
		items: map[string]domain.Order{},
	}
}

func (i *InMemory) Has(_ context.Context, key string) bool {
	i.mutex.RLock()
	_, ok := i.items[key]
	i.mutex.RUnlock()
	return ok
}

func (i *InMemory) Set(_ context.Context, key string, value domain.Order, _ time.Duration) error {
	i.mutex.Lock()
	i.items[key] = value
	i.mutex.Unlock()
	return nil
}

func (i *InMemory) Get(_ context.Context, key string) (domain.Order, bool, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	item, ok := i.items[key]
	return item, ok, nil
}
