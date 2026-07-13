package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrDuplicateDomain is returned when attempting to create a site with an existing domain.
var ErrDuplicateDomain = errors.New("duplicate domain")

// ErrSiteNotFound is returned when a site is not found by ID.
var ErrSiteNotFound = errors.New("site not found")

// Site represents a proxied upstream site configuration.
type Site struct {
	ID         int64     `json:"id"`
	Domain     string    `json:"domain"`
	Upstream   string    `json:"upstream"`
	SSHEnabled bool      `json:"ssl_enabled"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateSite inserts a new site and returns it with generated fields.
func (s *Store) CreateSite(domain, upstream string) (*Site, error) {
	res, err := s.db.Exec(
		`INSERT INTO sites (domain, upstream) VALUES (?, ?)`,
		domain, upstream,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateDomain, domain)
		}
		return nil, fmt.Errorf("create site: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.GetSite(id)
}

// GetSite retrieves a single site by its ID.
func (s *Store) GetSite(id int64) (*Site, error) {
	site := &Site{}
	err := s.db.QueryRow(
		`SELECT id, domain, upstream, ssl_enabled, active, created_at, updated_at
		 FROM sites WHERE id = ?`, id,
	).Scan(&site.ID, &site.Domain, &site.Upstream, &site.SSHEnabled, &site.Active, &site.CreatedAt, &site.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: id=%d", ErrSiteNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("get site %d: %w", id, err)
	}
	return site, nil
}

// ListSites returns all sites ordered by creation date descending.
func (s *Store) ListSites() ([]*Site, error) {
	rows, err := s.db.Query(
		`SELECT id, domain, upstream, ssl_enabled, active, created_at, updated_at
		 FROM sites ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sites: %w", err)
	}
	defer rows.Close()

	var sites []*Site
	for rows.Next() {
		site := &Site{}
		if err := rows.Scan(&site.ID, &site.Domain, &site.Upstream, &site.SSHEnabled, &site.Active, &site.CreatedAt, &site.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan site: %w", err)
		}
		sites = append(sites, site)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return sites, nil
}

// UpdateSite modifies domain and upstream for an existing site.
func (s *Store) UpdateSite(id int64, domain, upstream string) (*Site, error) {
	res, err := s.db.Exec(
		`UPDATE sites SET domain = ?, upstream = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		domain, upstream, id,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateDomain, domain)
		}
		return nil, fmt.Errorf("update site %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("%w: id=%d", ErrSiteNotFound, id)
	}
	return s.GetSite(id)
}

// DeleteSite performs a soft delete by setting active = 0.
func (s *Store) DeleteSite(id int64) error {
	res, err := s.db.Exec(
		`UPDATE sites SET active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete site %d: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: id=%d", ErrSiteNotFound, id)
	}
	return nil
}

// GetActiveSitesMap returns a map of domain -> upstream for all active sites.
// Used by the proxy manager to route requests based on Host header.
func (s *Store) GetActiveSitesMap() (map[string]string, error) {
	rows, err := s.db.Query(
		`SELECT domain, upstream FROM sites WHERE active = 1`,
	)
	if err != nil {
		return nil, fmt.Errorf("get active sites: %w", err)
	}
	defer rows.Close()

	sites := make(map[string]string)
	for rows.Next() {
		var domain, upstream string
		if err := rows.Scan(&domain, &upstream); err != nil {
			return nil, fmt.Errorf("scan active site: %w", err)
		}
		sites[domain] = upstream
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return sites, nil
}

// isUniqueConstraintError checks if the error is a SQLite UNIQUE constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite returns errors containing "UNIQUE constraint failed"
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
