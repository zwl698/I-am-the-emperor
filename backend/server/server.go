package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"

	"i-am-the-emperor/backend/game"
)

type App struct {
	mu    sync.RWMutex
	games map[string]*game.GameState
	mux   *http.ServeMux
}

type createGameRequest struct {
	Seed      int64  `json:"seed"`
	DynastyID string `json:"dynastyId"`
}

type choiceRequest struct {
	ChoiceID string `json:"choiceId"`
}

type choiceResponse struct {
	Resolution *game.Resolution `json:"resolution"`
	State      *game.GameState  `json:"state"`
}

func New() *App {
	app := &App{
		games: make(map[string]*game.GameState),
		mux:   http.NewServeMux(),
	}
	app.routes()
	return app
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) routes() {
	a.mux.HandleFunc("GET /api/dynasties", a.handleDynasties)
	a.mux.HandleFunc("POST /api/games", a.handleGames)
	a.mux.HandleFunc("GET /api/games/{id}", a.handleGetGame)
	a.mux.HandleFunc("POST /api/games/{id}/choices", a.handleApplyChoice)
	a.mux.Handle("/", http.FileServer(http.Dir(filepath.Clean("web"))))
}

func (a *App) handleDynasties(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, game.AvailableDynasties())
}

func (a *App) handleGames(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.DynastyID == "" {
		req.DynastyID = "dayin"
	}
	state, err := game.NewGameWithDynasty(req.DynastyID, req.Seed)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.mu.Lock()
	a.games[state.ID] = state
	a.mu.Unlock()

	writeJSON(w, http.StatusOK, state)
}

func (a *App) handleGetGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a.mu.RLock()
	state := a.games[id]
	a.mu.RUnlock()
	if state == nil {
		writeError(w, http.StatusNotFound, "存档不存在")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (a *App) handleApplyChoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req choiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "选择请求格式错误")
		return
	}
	if req.ChoiceID == "" {
		writeError(w, http.StatusBadRequest, "缺少 choiceId")
		return
	}

	a.mu.Lock()
	state := a.games[id]
	if state == nil {
		a.mu.Unlock()
		writeError(w, http.StatusNotFound, "存档不存在")
		return
	}
	resolution, err := state.ApplyChoice(req.ChoiceID)
	a.mu.Unlock()

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, choiceResponse{Resolution: resolution, State: state})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error":  message,
		"status": strconv.Itoa(status),
	})
}
