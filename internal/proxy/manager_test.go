package proxy

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	// Verify the internal map is initialized (Get on nonexistent returns nil)
	if proxy := m.Get("nonexistent"); proxy != nil {
		t.Fatal("expected nil for nonexistent domain")
	}
}

func TestAddOrUpdate(t *testing.T) {
	m := NewManager()
	err := m.AddOrUpdate("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	proxy := m.Get("example.com")
	if proxy == nil {
		t.Fatal("expected proxy for example.com")
	}
}

func TestAddOrUpdate_EmptyDomain(t *testing.T) {
	m := NewManager()
	err := m.AddOrUpdate("", "http://localhost:3000")
	if err == nil {
		t.Fatal("expected error for empty domain")
	}
}

func TestAddOrUpdate_InvalidUpstream(t *testing.T) {
	m := NewManager()
	err := m.AddOrUpdate("example.com", "://invalid")
	if err == nil {
		t.Fatal("expected error for invalid upstream")
	}
}

func TestAddOrUpdate_UpdateExisting(t *testing.T) {
	m := NewManager()
	err := m.AddOrUpdate("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	// Update the same domain with a different upstream
	err = m.AddOrUpdate("example.com", "http://localhost:4000")
	if err != nil {
		t.Fatal(err)
	}
	proxy := m.Get("example.com")
	if proxy == nil {
		t.Fatal("expected proxy after update")
	}
}

func TestRemove(t *testing.T) {
	m := NewManager()
	err := m.AddOrUpdate("example.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}

	m.Remove("example.com")
	if proxy := m.Get("example.com"); proxy != nil {
		t.Fatal("expected nil after removal")
	}
}

func TestRemove_Nonexistent(t *testing.T) {
	m := NewManager()
	// Should not panic
	m.Remove("nonexistent")
}

func TestGet_NotFound(t *testing.T) {
	m := NewManager()
	proxy := m.Get("nonexistent.com")
	if proxy != nil {
		t.Fatal("expected nil for nonexistent domain")
	}
}
