package store

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Visit represents a single proxied HTTP request.
type Visit struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	IP         string    `json:"ip"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	StatusCode int       `json:"status_code"`
	LatencyMs  float64   `json:"latency_ms"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	CreatedAt  time.Time `json:"created_at"`
}

// CountryStat holds a visit count grouped by country.
type CountryStat struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

// VisitBuffer is a channel-based batch writer for visits.
// It accumulates visits and flushes them to DB periodically.
// This ensures the proxy hot path never blocks on DB I/O.
type VisitBuffer struct {
	store         *Store
	visits        chan Visit
	batchSize     int
	flushInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewVisitBuffer creates a new VisitBuffer with the given batch config.
// The internal channel has a fixed capacity of 1000 (per design AD6).
func NewVisitBuffer(store *Store, batchSize int, flushInterval time.Duration) *VisitBuffer {
	return &VisitBuffer{
		store:         store,
		visits:        make(chan Visit, 1000),
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
}

// Start launches the background goroutine that drains the visit channel.
func (vb *VisitBuffer) Start() {
	vb.ctx, vb.cancel = context.WithCancel(context.Background())
	vb.wg.Add(1)
	go vb.run()
}

// Stop signals the background goroutine to flush remaining visits and exit.
func (vb *VisitBuffer) Stop() {
	if vb.cancel != nil {
		vb.cancel()
	}
	vb.wg.Wait()
}

// Log enqueues a visit for batch writing. Non-blocking — drops the visit
// if the buffer channel is full (logs a warning).
func (vb *VisitBuffer) Log(v Visit) {
	select {
	case vb.visits <- v:
	default:
		slog.Warn("visit buffer full, dropping event",
			"ip", v.IP,
			"path", v.Path,
		)
	}
}

func (vb *VisitBuffer) run() {
	defer vb.wg.Done()

	batch := make([]Visit, 0, vb.batchSize)
	ticker := time.NewTicker(vb.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-vb.ctx.Done():
			// Flush remaining on shutdown
			if len(batch) > 0 {
				vb.flush(batch)
			}
			return

		case v := <-vb.visits:
			batch = append(batch, v)
			if len(batch) >= vb.batchSize {
				vb.flush(batch)
				batch = make([]Visit, 0, vb.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				vb.flush(batch)
				batch = make([]Visit, 0, vb.batchSize)
			}
		}
	}
}

func (vb *VisitBuffer) flush(batch []Visit) {
	tx, err := vb.store.db.Begin()
	if err != nil {
		slog.Error("visit flush: begin tx", "error", err, "batch_size", len(batch))
		return
	}
	defer tx.Rollback() // no-op if committed

	stmt, err := tx.PrepareContext(vb.ctx,
		`INSERT INTO visits (site_id, ip, country, city, path, method, status_code, latency_ms, user_agent, referer)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		slog.Error("visit flush: prepare", "error", err)
		return
	}
	defer stmt.Close()

	for _, v := range batch {
		if _, err := stmt.ExecContext(vb.ctx,
			v.SiteID, v.IP, v.Country, v.City, v.Path, v.Method,
			v.StatusCode, v.LatencyMs, v.UserAgent, v.Referer,
		); err != nil {
			slog.Error("visit flush: exec", "error", err)
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("visit flush: commit", "error", err, "batch_size", len(batch))
		return
	}

	slog.Debug("visit flush: committed", "batch_size", len(batch))
}

// GetVisits returns a paginated list of visits, optionally filtered by site_id.
// Returns the visits slice, total count, and any error.
func (s *Store) GetVisits(page, limit int, siteID *int64) ([]*Visit, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	baseQuery := `FROM visits`
	var args []interface{}
	var whereClause string

	if siteID != nil {
		whereClause = " WHERE site_id = ?"
		args = append(args, *siteID)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count visits: %w", err)
	}

	// Get paginated data
	offset := (page - 1) * limit
	dataQuery := "SELECT id, site_id, ip, country, city, path, method, status_code, latency_ms, user_agent, referer, created_at" +
		baseQuery + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataArgs := append(args, limit, offset)

	rows, err := s.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query visits: %w", err)
	}
	defer rows.Close()

	var visits []*Visit
	for rows.Next() {
		v := &Visit{}
		if err := rows.Scan(
			&v.ID, &v.SiteID, &v.IP, &v.Country, &v.City,
			&v.Path, &v.Method, &v.StatusCode, &v.LatencyMs,
			&v.UserAgent, &v.Referer, &v.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan visit: %w", err)
		}
		visits = append(visits, v)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}
	if visits == nil {
		visits = []*Visit{}
	}

	return visits, total, nil
}

// GetTotalVisits returns the total number of visit records.
func (s *Store) GetTotalVisits() (int, error) {
	var total int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM visits`).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("total visits: %w", err)
	}
	return total, nil
}

// GetVisitsToday returns the number of visits recorded today (UTC).
func (s *Store) GetVisitsToday() (int, error) {
	var total int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM visits WHERE created_at >= date('now')`,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("visits today: %w", err)
	}
	return total, nil
}

// GetVisitsByCountry returns visit counts grouped by country, ordered by count descending.
func (s *Store) GetVisitsByCountry() ([]CountryStat, error) {
	rows, err := s.db.Query(
		`SELECT country, COUNT(*) as count
		 FROM visits
		 GROUP BY country
		 ORDER BY count DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("visits by country: %w", err)
	}
	defer rows.Close()

	var stats []CountryStat
	for rows.Next() {
		var cs CountryStat
		if err := rows.Scan(&cs.Country, &cs.Count); err != nil {
			return nil, fmt.Errorf("scan country stat: %w", err)
		}
		stats = append(stats, cs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	if stats == nil {
		stats = []CountryStat{}
	}
	return stats, nil
}
