package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	handler := New()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body["status"] != "ok" || body["service"] != "new_sanguoby" {
		t.Fatalf("body = %#v, want service new_sanguoby with status ok", body)
	}
}

func TestGameEndpointsCreateReadAndAdvance(t *testing.T) {
	handler := New()

	createBody := bytes.NewBufferString(`{"scenarioId":"dongzhuo","playerId":"caocao"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRec.Code, http.StatusCreated)
	}
	created := decodeSnapshot(t, createRec)
	if created.PlayerID != "caocao" || created.Date.Month != 1 {
		t.Fatalf("created snapshot = %#v, want caocao at month 1", created)
	}

	currentReq := httptest.NewRequest(http.MethodGet, "/api/games/current", nil)
	currentRec := httptest.NewRecorder()
	handler.ServeHTTP(currentRec, currentReq)
	current := decodeSnapshot(t, currentRec)
	if current.Date.Month != 1 {
		t.Fatalf("current month = %d, want 1", current.Date.Month)
	}

	advanceReq := httptest.NewRequest(http.MethodPost, "/api/games/current/advance-month", nil)
	advanceRec := httptest.NewRecorder()
	handler.ServeHTTP(advanceRec, advanceReq)
	advanced := decodeSnapshot(t, advanceRec)
	if advanced.Date.Month != 2 {
		t.Fatalf("advanced month = %d, want 2", advanced.Date.Month)
	}
}

func TestLegacyResourcesEndpointListsArchive(t *testing.T) {
	handler := NewWithOptions(Options{LegacyArchivePath: legacyArchivePath(t)})
	req := httptest.NewRequest(http.MethodGet, "/api/legacy/resources", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body legacyResourcesBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode legacy resources response: %v", err)
	}
	if body.Count != 87 {
		t.Fatalf("count = %d, want 87", body.Count)
	}
	if !body.hasResource(58, 43, 10) {
		t.Fatalf("resources missing city names header: %#v", body.Resources)
	}
	if !body.hasResource(64, 163, 0) {
		t.Fatalf("resources missing variable string constants header: %#v", body.Resources)
	}
}

func decodeSnapshot(t *testing.T, rec *httptest.ResponseRecorder) snapshot {
	t.Helper()
	if rec.Code < 200 || rec.Code > 299 {
		t.Fatalf("status = %d, want success; body: %s", rec.Code, rec.Body.String())
	}
	var body snapshot
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

type snapshot struct {
	PlayerID string `json:"playerId"`
	Date     struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	} `json:"date"`
}

type legacyResourcesBody struct {
	Count     int `json:"count"`
	Resources []struct {
		ID         uint16 `json:"id"`
		ItemCount  uint16 `json:"itemCount"`
		ItemLength uint16 `json:"itemLength"`
	} `json:"resources"`
}

func (b legacyResourcesBody) hasResource(id, itemCount, itemLength uint16) bool {
	for _, resource := range b.Resources {
		if resource.ID == id && resource.ItemCount == itemCount && resource.ItemLength == itemLength {
			return true
		}
	}
	return false
}

func legacyArchivePath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	path := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "sanguobaye_c-master", "src", "dat.lib.orig"))
	if _, err := os.Stat(path); err != nil {
		t.Skipf("legacy archive not found: %v", err)
	}
	return path
}
