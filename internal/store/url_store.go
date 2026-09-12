package store

import (
	"errors"
	"sync"
	"time"

	"github.com/k11ngp1ng/url-shortener-go/internal/domain"
)

var ErrNotFound = errors.New("short URL not found")
var ErrCodeAlreadyExists = errors.New("short code already exists")

type URLStore interface {
	Create(url domain.URL) (domain.URL, error)
	Get(code string) (domain.URL, error)
	IncrementClicks(code string) (domain.URL, error)
}

type MemoryURLStore struct {
	mu   sync.RWMutex
	urls map[string]domain.URL
}

func NewMemoryURLStore() *MemoryURLStore {
	return &MemoryURLStore{urls: make(map[string]domain.URL)}
}

func (s *MemoryURLStore) Create(url domain.URL) (domain.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.urls[url.Code]; exists {
		return domain.URL{}, ErrCodeAlreadyExists
	}
	if url.CreatedAt.IsZero() {
		url.CreatedAt = time.Now().UTC()
	}
	s.urls[url.Code] = url
	return url, nil
}

func (s *MemoryURLStore) Get(code string) (domain.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.urls[code]
	if !exists {
		return domain.URL{}, ErrNotFound
	}
	return url, nil
}

func (s *MemoryURLStore) IncrementClicks(code string) (domain.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, exists := s.urls[code]
	if !exists {
		return domain.URL{}, ErrNotFound
	}
	url.Clicks++
	s.urls[code] = url
	return url, nil
}
