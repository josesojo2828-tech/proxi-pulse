package geo

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
)

// Record holds geolocation data for a single IP lookup.
type Record struct {
	Country string  `json:"country"`
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// Lookup wraps the MaxMind GeoLite2 reader (v2 API).
type Lookup struct {
	db *geoip2.Reader
}

// NewLookup opens the MaxMind .mmdb database at the given path.
func NewLookup(mmdbPath string) (*Lookup, error) {
	db, err := geoip2.Open(mmdbPath)
	if err != nil {
		return nil, fmt.Errorf("open mmdb %q: %w", mmdbPath, err)
	}
	return &Lookup{db: db}, nil
}

// Close releases the underlying database reader.
func (l *Lookup) Close() error {
	return l.db.Close()
}

// Lookup performs a geolocation lookup for the given IP address.
// Returns the country ISO code, city name, and coordinates.
func (l *Lookup) Lookup(ip net.IP) (*Record, error) {
	// geoip2-golang/v2 uses netip.Addr instead of net.IP
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		// Try IPv4-mapped IPv6 (16-byte representation)
		ip4 := ip.To4()
		if ip4 != nil {
			addr, ok = netip.AddrFromSlice(ip4)
		}
		if !ok {
			return nil, fmt.Errorf("invalid IP address: %s", ip)
		}
	}

	record, err := l.db.City(addr)
	if err != nil {
		return nil, fmt.Errorf("geo lookup %s: %w", ip, err)
	}

	r := &Record{
		Country: record.Country.ISOCode,
		City:    record.City.Names.English,
	}
	if record.Location.HasCoordinates() {
		r.Lat = *record.Location.Latitude
		r.Lon = *record.Location.Longitude
	}
	return r, nil
}

// IsPrivate checks whether the given IP is in a private or reserved range.
// This covers RFC1918 (10/8, 172.16/12, 192.168/16), loopback (127/8),
// link-local (169.254/16), IPv6 loopback (::1), and IPv6 unique-local (fc00::/7).
func IsPrivate(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()
}
