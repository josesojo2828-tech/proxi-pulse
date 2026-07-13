package geo

import (
	"fmt"
	"net"

	lru "github.com/hashicorp/golang-lru/v2"
)

// Cache wraps a MaxMind lookup with an LRU cache to avoid redundant lookups.
// Per design AD3: ~50k entries ≈ 25MB RAM, actively-used IPs stay cached.
type Cache struct {
	lru    *lru.Cache[string, *Record]
	lookup *Lookup
}

// NewCache creates a new geo cache backed by the given lookup.
func NewCache(lookup *Lookup, size int) (*Cache, error) {
	l, err := lru.New[string, *Record](size)
	if err != nil {
		return nil, fmt.Errorf("new lru cache: %w", err)
	}
	return &Cache{lru: l, lookup: lookup}, nil
}

// Get retrieves geolocation data for the given IP.
// Returns cached data if available, otherwise performs a MaxMind lookup
// and caches the result.
func (c *Cache) Get(ip net.IP) (*Record, error) {
	ipStr := ip.String()

	// Fast path: cache hit
	if rec, ok := c.lru.Get(ipStr); ok {
		return rec, nil
	}

	// Slow path: perform lookup
	rec, err := c.lookup.Lookup(ip)
	if err != nil {
		return nil, err
	}

	// Cache result (best-effort)
	c.lru.Add(ipStr, rec)
	return rec, nil
}

// Len returns the current number of cached entries.
func (c *Cache) Len() int {
	return c.lru.Len()
}
