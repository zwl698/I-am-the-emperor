package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"new_sanguoby/backend/internal/game"
	"new_sanguoby/backend/internal/legacyres"
)

type Server struct {
	mux               *http.ServeMux
	mu                sync.Mutex
	current           *game.GameState
	legacyArchivePath string
}

type Options struct {
	LegacyArchivePath string
}

type createGameRequest struct {
	ScenarioID string `json:"scenarioId"`
	PlayerID   string `json:"playerId"`
}

func New() http.Handler {
	return NewWithOptions(Options{})
}

func NewWithOptions(options Options) http.Handler {
	legacyArchivePath := options.LegacyArchivePath
	if legacyArchivePath == "" {
		legacyArchivePath = defaultLegacyArchivePath()
	}
	s := &Server{
		mux:               http.NewServeMux(),
		current:           newInitialGame(legacyArchivePath, ""),
		legacyArchivePath: legacyArchivePath,
	}
	s.routes()
	return s
}

// newInitialGame builds the starting game from the real legacy archive when
// available, falling back to the authored seed if the archive cannot be read.
func newInitialGame(archivePath, playerID string) *game.GameState {
	if archivePath != "" {
		if state, err := game.NewGameFromArchive(archivePath, playerID); err == nil {
			return state
		}
	}
	return game.NewGame("dongzhuo", playerID)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("POST /api/games", s.handleCreateGame)
	s.mux.HandleFunc("GET /api/games/current", s.handleCurrentGame)
	s.mux.HandleFunc("POST /api/games/current/advance-month", s.handleAdvanceMonth)
	s.mux.HandleFunc("GET /api/legacy/resources", s.handleLegacyResources)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "new_sanguoby",
		"status":  "ok",
	})
}

func (s *Server) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	s.mu.Lock()
	s.current = newInitialGame(s.legacyArchivePath, req.PlayerID)
	snapshot := s.current
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, snapshot)
}

func (s *Server) handleCurrentGame(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	snapshot := s.current
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) handleAdvanceMonth(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.current.AdvanceMonth()
	snapshot := s.current
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) handleLegacyResources(w http.ResponseWriter, r *http.Request) {
	archive, err := legacyres.Open(s.legacyArchivePath)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "legacy archive unavailable")
		return
	}

	resources := archive.List(120)
	writeJSON(w, http.StatusOK, legacyResourcesResponse{
		Source:    archive.Path(),
		Count:     len(resources),
		Resources: resources,
	})
}

type legacyResourcesResponse struct {
	Source    string             `json:"source"`
	Count     int                `json:"count"`
	Resources []legacyres.Header `json:"resources"`
}

func defaultLegacyArchivePath() string {
	if path := os.Getenv("LEGACY_ARCHIVE_PATH"); path != "" {
		return path
	}
	return filepath.Clean("../sanguobaye_c-master/src/dat.lib.orig")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
