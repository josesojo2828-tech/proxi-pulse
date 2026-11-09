package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Manager maintains a map of domain → ReverseProxy.
// Updated when sites are created, edited, or deleted.
type Manager struct {
	mu      sync.RWMutex
	proxies map[string]*httputil.ReverseProxy // domain -> proxy
}

// NewManager creates an empty proxy manager.
func NewManager() *Manager {
	return &Manager{
		proxies: make(map[string]*httputil.ReverseProxy),
	}
}

// AddOrUpdate creates or updates a reverse proxy for a domain.
func (m *Manager) AddOrUpdate(domain, upstream string) error {
	if domain == "" {
		return fmt.Errorf("domain must not be empty")
	}
	u, err := url.Parse(upstream)
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}
	m.mu.Lock()
	m.proxies[domain] = proxy
	m.mu.Unlock()
	return nil
}

// Remove deletes a proxy for a domain.
func (m *Manager) Remove(domain string) {
	m.mu.Lock()
	delete(m.proxies, domain)
	m.mu.Unlock()
}

// Get returns the proxy for a domain, or nil if not found.
func (m *Manager) Get(domain string) *httputil.ReverseProxy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.proxies[domain]
}

// ReloadFromStore replaces all proxies with active sites from the store.
func (m *Manager) ReloadFromStore(siteStore interface {
	GetActiveSitesMap() (map[string]string, error)
}) error {
	sites, err := siteStore.GetActiveSitesMap()
	if err != nil {
		return err
	}
	newProxies := make(map[string]*httputil.ReverseProxy, len(sites))
	for domain, upstream := range sites {
		u, err := url.Parse(upstream)
		if err != nil {
			continue // skip invalid upstreams
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		newProxies[domain] = proxy
	}
	m.mu.Lock()
	m.proxies = newProxies
	m.mu.Unlock()
	return nil
}
