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

type commandRequest struct {
	CityID    string `json:"cityId"`
	GeneralID string `json:"generalId"`
	CommandID string `json:"commandId"`
}

type battleRequest struct {
	CityID       string `json:"cityId"`
	GeneralID    string `json:"generalId"`
	TargetCityID string `json:"targetCityId"`
}

type battleResponse struct {
	Outcome  *game.BattleOutcome `json:"outcome"`
	Snapshot *game.GameState     `json:"snapshot"`
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
		current:           newInitialGame(legacyArchivePath, "period-1", ""),
		legacyArchivePath: legacyArchivePath,
	}
	s.routes()
	return s
}

// newInitialGame builds the starting game from the real legacy archive when
// available, falling back to the authored seed if the archive cannot be read.
func newInitialGame(archivePath, scenarioID, playerID string) *game.GameState {
	if archivePath != "" {
		if state, err := game.NewGameFromArchive(archivePath, scenarioID, playerID); err == nil {
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
	s.mux.HandleFunc("GET /api/scenarios", s.handleScenarios)
	s.mux.HandleFunc("POST /api/games", s.handleCreateGame)
	s.mux.HandleFunc("GET /api/games/current", s.handleCurrentGame)
	s.mux.HandleFunc("POST /api/games/current/command", s.handleCommand)
	s.mux.HandleFunc("POST /api/games/current/battle", s.handleBattle)
	s.mux.HandleFunc("POST /api/games/current/advance-month", s.handleAdvanceMonth)
	s.mux.HandleFunc("GET /api/legacy/resources", s.handleLegacyResources)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "new_sanguoby",
		"status":  "ok",
	})
}

func (s *Server) handleScenarios(w http.ResponseWriter, r *http.Request) {
	options := make([]scenarioOption, 0, 4)
	for period := 1; period <= 4; period++ {
		scenarioID := scenarioIDForPeriod(period)
		state := newInitialGame(s.legacyArchivePath, scenarioID, "")
		rulers := make([]rulerOption, 0, len(state.Rulers))
		for _, ruler := range state.Rulers {
			if ruler.ID == "neutral" {
				continue
			}
			cityCount := 0
			for _, city := range state.Cities {
				if city.OwnerID == ruler.ID {
					cityCount++
				}
			}
			rulers = append(rulers, rulerOption{
				ID:        ruler.ID,
				Name:      ruler.Name,
				Character: ruler.Character,
				Color:     ruler.Color,
				CityCount: cityCount,
			})
		}
		options = append(options, scenarioOption{
			ID:      scenarioID,
			Period:  period,
			Name:    scenarioNameForPeriod(period),
			Year:    state.Date.Year,
			Rulers:  rulers,
			CityMax: len(state.Cities),
		})
	}
	writeJSON(w, http.StatusOK, scenarioListResponse{Scenarios: options})
}

func (s *Server) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	s.mu.Lock()
	s.current = newInitialGame(s.legacyArchivePath, req.ScenarioID, req.PlayerID)
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
	s.current.EndStrategyPhase()
	snapshot := s.current
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	var req commandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	s.mu.Lock()
	err := s.current.ApplyCommand(req.CityID, req.GeneralID, req.CommandID)
	snapshot := s.current
	s.mu.Unlock()

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) handleBattle(w http.ResponseWriter, r *http.Request) {
	var req battleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	s.mu.Lock()
	outcome, err := s.current.ApplyBattle(req.CityID, req.GeneralID, req.TargetCityID)
	snapshot := s.current
	s.mu.Unlock()

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, battleResponse{Outcome: outcome, Snapshot: snapshot})
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

type scenarioListResponse struct {
	Scenarios []scenarioOption `json:"scenarios"`
}

type scenarioOption struct {
	ID      string        `json:"id"`
	Period  int           `json:"period"`
	Name    string        `json:"name"`
	Year    int           `json:"year"`
	Rulers  []rulerOption `json:"rulers"`
	CityMax int           `json:"cityMax"`
}

type rulerOption struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Character string `json:"character"`
	Color     string `json:"color"`
	CityCount int    `json:"cityCount"`
}

func scenarioIDForPeriod(period int) string {
	return map[int]string{
		1: "period-1",
		2: "period-2",
		3: "period-3",
		4: "period-4",
	}[period]
}

func scenarioNameForPeriod(period int) string {
	return map[int]string{
		1: "董卓弄权",
		2: "曹操崛起",
		3: "赤壁之战",
		4: "三足鼎立",
	}[period]
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
