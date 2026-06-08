package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGameAPI(t *testing.T) {
	app := New()
	req := httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBufferString(`{"seed":42,"dynastyId":"jingyao"}`))
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["id"] == "" {
		t.Fatalf("expected id in response: %+v", payload)
	}
	if payload["scene"] == nil {
		t.Fatalf("expected scene in response: %+v", payload)
	}
	dynasty := payload["dynasty"].(map[string]any)
	if dynasty["id"] != "jingyao" {
		t.Fatalf("expected selected dynasty, got %+v", dynasty)
	}
}

func TestDynastiesAPI(t *testing.T) {
	app := New()
	req := httptest.NewRequest(http.MethodGet, "/api/dynasties", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload) < 4 {
		t.Fatalf("expected dynasties, got %+v", payload)
	}
}

func TestApplyChoiceAPI(t *testing.T) {
	app := New()
	create := httptest.NewRecorder()
	app.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBufferString(`{"seed":3}`)))

	var state struct {
		ID    string `json:"id"`
		Scene struct {
			Choices []struct {
				ID string `json:"id"`
			} `json:"choices"`
		} `json:"scene"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	body := []byte(`{"choiceId":"` + state.Scene.Choices[0].ID + `"}`)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/games/"+state.ID+"/choices", bytes.NewBuffer(body)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["resolution"] == nil || payload["state"] == nil {
		t.Fatalf("expected resolution and state: %+v", payload)
	}
}
