package store

import (
	"testing"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateSite(t *testing.T) {
	s := setupTestStore(t)
	site, err := s.CreateSite("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	if site.Domain != "example.com" {
		t.Errorf("expected example.com, got %s", site.Domain)
	}
	if !site.Active {
		t.Error("expected active")
	}
}

func TestCreateDuplicateDomain(t *testing.T) {
	s := setupTestStore(t)
	_, err := s.CreateSite("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateSite("example.com", "http://localhost:3001")
	if err == nil {
		t.Fatal("expected error for duplicate domain")
	}
}

func TestGetSite(t *testing.T) {
	s := setupTestStore(t)
	created, err := s.CreateSite("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSite(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Domain != created.Domain {
		t.Errorf("expected %s, got %s", created.Domain, got.Domain)
	}
}

func TestGetSite_NotFound(t *testing.T) {
	s := setupTestStore(t)
	_, err := s.GetSite(999)
	if err == nil {
		t.Fatal("expected error for non-existent site")
	}
}

func TestListSites(t *testing.T) {
	s := setupTestStore(t)

	_, err := s.CreateSite("alpha.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateSite("beta.com", "http://localhost:3001")
	if err != nil {
		t.Fatal(err)
	}

	sites, err := s.ListSites()
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) != 2 {
		t.Errorf("expected 2 sites, got %d", len(sites))
	}
}

func TestDeleteSite(t *testing.T) {
	s := setupTestStore(t)
	site, err := s.CreateSite("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}

	err = s.DeleteSite(site.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Site should be soft-deleted (not active)
	got, err := s.GetSite(site.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Active {
		t.Error("expected site to be inactive after delete")
	}
}

func TestGetActiveSitesMap(t *testing.T) {
	s := setupTestStore(t)

	_, err := s.CreateSite("active.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	site2, err := s.CreateSite("todelete.com", "http://localhost:3001")
	if err != nil {
		t.Fatal(err)
	}

	// Soft-delete the second site
	err = s.DeleteSite(site2.ID)
	if err != nil {
		t.Fatal(err)
	}

	activeSites, err := s.GetActiveSitesMap()
	if err != nil {
		t.Fatal(err)
	}

	if len(activeSites) != 1 {
		t.Errorf("expected 1 active site, got %d", len(activeSites))
	}
	if upstream, ok := activeSites["active.com"]; !ok {
		t.Error("expected active.com in active sites map")
	} else if upstream != "http://localhost:3000" {
		t.Errorf("expected http://localhost:3000, got %s", upstream)
	}
}

func TestUpdateSite(t *testing.T) {
	s := setupTestStore(t)
	site, err := s.CreateSite("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := s.UpdateSite(site.ID, "example.org", "http://localhost:4000")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Domain != "example.org" {
		t.Errorf("expected example.org, got %s", updated.Domain)
	}
	if updated.Upstream != "http://localhost:4000" {
		t.Errorf("expected http://localhost:4000, got %s", updated.Upstream)
	}
}

func TestUpdateSite_NotFound(t *testing.T) {
	s := setupTestStore(t)
	_, err := s.UpdateSite(999, "nowhere.com", "http://localhost:3000")
	if err == nil {
		t.Fatal("expected error for non-existent site")
	}
}
