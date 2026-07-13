package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/josesojo2828/proxi-pulse/internal/proxy"
	"github.com/josesojo2828/proxi-pulse/internal/store"
)

func setupTestAPI(t *testing.T) (*store.Store, chi.Router) {
	t.Helper()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	pm := proxy.NewManager()
	router := NewRouter(st, pm)
	return st, router
}

func TestCreateSite_Handler(t *testing.T) {
	_, router := setupTestAPI(t)
	body := `{"domain":"test.com","upstream":"http://localhost:3000"}`
	req := httptest.NewRequest("POST", "/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got %+v", resp)
	}
}

func TestCreateSite_MissingDomain(t *testing.T) {
	_, router := setupTestAPI(t)
	body := `{"upstream":"http://localhost:3000"}`
	req := httptest.NewRequest("POST", "/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSite_MissingUpstream(t *testing.T) {
	_, router := setupTestAPI(t)
	body := `{"domain":"test.com"}`
	req := httptest.NewRequest("POST", "/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSite_InvalidJSON(t *testing.T) {
	_, router := setupTestAPI(t)
	body := `{invalid json`
	req := httptest.NewRequest("POST", "/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListSites_Handler(t *testing.T) {
	st, router := setupTestAPI(t)

	// Create a couple of sites
	_, err := st.CreateSite("alpha.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.CreateSite("beta.com", "http://localhost:3001")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/sites", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got %+v", resp)
	}
}

func TestGetSite_Handler(t *testing.T) {
	st, router := setupTestAPI(t)

	site, err := st.CreateSite("test.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/sites/%d", site.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got %+v", resp)
	}
}

func TestGetSite_NotFound(t *testing.T) {
	_, router := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/sites/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSite_Handler(t *testing.T) {
	st, router := setupTestAPI(t)

	site, err := st.CreateSite("test.com", "http://localhost:3000")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/sites/%d", site.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got %+v", resp)
	}
}

func TestDeleteSite_NotFound(t *testing.T) {
	_, router := setupTestAPI(t)

	req := httptest.NewRequest("DELETE", "/sites/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStatsCountries_Handler(t *testing.T) {
	_, router := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/stats/countries", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got %+v", resp)
	}
}
