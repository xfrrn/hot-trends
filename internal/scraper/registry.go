package scraper

import (
	"sync"
)

// Registry manages all registered scrapers
type Registry struct {
	scrapers map[string]Scraper
	mu       sync.RWMutex
}

// NewRegistry creates a new scraper registry
func NewRegistry() *Registry {
	return &Registry{
		scrapers: make(map[string]Scraper),
	}
}

// Register adds a scraper to the registry
func (r *Registry) Register(s Scraper) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scrapers[s.Name()] = s
}

// Get retrieves a scraper by name
func (r *Registry) Get(name string) (Scraper, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.scrapers[name]
	return s, ok
}

// List returns all registered scrapers
func (r *Registry) List() []Scraper {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scrapers := make([]Scraper, 0, len(r.scrapers))
	for _, s := range r.scrapers {
		scrapers = append(scrapers, s)
	}
	return scrapers
}

// Count returns the number of registered scrapers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.scrapers)
}
