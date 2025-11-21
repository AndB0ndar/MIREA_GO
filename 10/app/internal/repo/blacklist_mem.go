package repo

import (
	"sync"
	"time"
)

type BlacklistMem struct {
	mu    sync.RWMutex
	items map[string]time.Time
}

func NewBlacklistMem() *BlacklistMem {
	return &BlacklistMem{
		items: make(map[string]time.Time),
	}
}

func (b *BlacklistMem) Add(token string, expiresAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[token] = expiresAt
}

func (b *BlacklistMem) IsRevoked(token string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	expires, exists := b.items[token]
	if !exists {
		return false
	}

	// Удаляем просроченные токены при проверке
	if time.Now().After(expires) {
		b.mu.RUnlock()
		b.mu.Lock()
		delete(b.items, token)
		b.mu.Unlock()
		b.mu.RLock()
		return false
	}

	return true
}

func (b *BlacklistMem) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for token, expires := range b.items {
		if now.After(expires) {
			delete(b.items, token)
		}
	}
}
