package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type apiGameState struct {
	ID      string `json:"id"`
	Phase   string `json:"phase"`
	Command int    `json:"command"`
	Wars    []struct {
		ID       string `json:"id"`
		Threat   int    `json:"threat"`
		Progress int    `json:"progress"`
	} `json:"wars"`
	Scene struct {
		Choices []struct {
			ID string `json:"id"`
		} `json:"choices"`
	} `json:"scene"`
}

type apiActionResponse struct {
	Resolution map[string]any `json:"resolution"`
	State      apiGameState   `json:"state"`
}

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

func TestCreateGameRejectsUnknownDynasty(t *testing.T) {
	app := New()
	req := httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBufferString(`{"seed":42,"dynastyId":"missing"}`))
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
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

	var state apiGameState
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

func TestApplyOrderAPI(t *testing.T) {
	app := New()
	create := httptest.NewRecorder()
	app.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBufferString(`{"seed":9,"dynastyId":"chengping"}`)))

	var state apiGameState
	if err := json.Unmarshal(create.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	for i := 0; i < 5; i++ {
		body := []byte(`{"choiceId":"` + state.Scene.Choices[0].ID + `"}`)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/games/"+state.ID+"/choices", bytes.NewBuffer(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("advance to emperor status %d: %s", rec.Code, rec.Body.String())
		}
		var payload apiActionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode advance response: %v", err)
		}
		state = payload.State
	}
	if state.Phase != "emperor" {
		t.Fatalf("expected emperor phase before issuing order, got %+v", state)
	}
	beforeCommand := state.Command
	if beforeCommand <= 0 {
		t.Fatalf("expected command points before order, got %+v", state)
	}

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/games/"+state.ID+"/orders", bytes.NewBufferString(`{"kind":"relief","target":"capital"}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload apiActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Resolution == nil || payload.State.ID == "" {
		t.Fatalf("expected resolution and state: %+v", payload)
	}
	if payload.State.Command != beforeCommand-1 {
		t.Fatalf("expected order to spend one command point, before %d after %d", beforeCommand, payload.State.Command)
	}
}

func TestApplyWarOrderAPI(t *testing.T) {
	app := New()
	create := httptest.NewRecorder()
	app.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBufferString(`{"seed":19,"dynastyId":"xuanshuo"}`)))

	var state apiGameState
	if err := json.Unmarshal(create.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	for i := 0; i < 5; i++ {
		body := []byte(`{"choiceId":"` + state.Scene.Choices[0].ID + `"}`)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/games/"+state.ID+"/choices", bytes.NewBuffer(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("advance to emperor status %d: %s", rec.Code, rec.Body.String())
		}
		var payload apiActionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode advance response: %v", err)
		}
		state = payload.State
	}
	if len(state.Wars) == 0 {
		t.Fatalf("expected war campaigns after coronation, got %+v", state)
	}

	beforeCommand := state.Command
	beforeWar := state.Wars[0]
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/games/"+state.ID+"/orders", bytes.NewBufferString(`{"kind":"campaign","target":"`+beforeWar.ID+`"}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload apiActionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	afterWar := payload.State.Wars[0]
	if payload.State.Command != beforeCommand-1 {
		t.Fatalf("expected campaign order to spend one command point, before %d after %d", beforeCommand, payload.State.Command)
	}
	if afterWar.Progress <= beforeWar.Progress || afterWar.Threat >= beforeWar.Threat {
		t.Fatalf("expected campaign to advance and reduce threat, before %+v after %+v", beforeWar, afterWar)
	}
}
