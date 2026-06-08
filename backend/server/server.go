package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"i-am-the-emperor/backend/game"
)

type App struct {
	mu    sync.RWMutex
	games map[string]*game.GameState
	mux   *http.ServeMux
}

type createGameRequest struct {
	Seed int64 `json:"seed"`
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
	a.mux.HandleFunc("/api/games", a.handleGames)
	a.mux.HandleFunc("/api/games/", a.handleGameAction)
	a.mux.Handle("/", http.FileServer(http.Dir(filepath.Clean("web"))))
}

func (a *App) handleGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "只支持 POST 创建新游戏")
		return
	}

	var req createGameRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	state := game.NewGame(req.Seed)

	a.mu.Lock()
	a.games[state.ID] = state
	a.mu.Unlock()

	writeJSON(w, http.StatusOK, state)
}

func (a *App) handleGameAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/games/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "未找到游戏")
		return
	}
	id := parts[0]

	if len(parts) == 1 && r.Method == http.MethodGet {
		a.mu.RLock()
		state := a.games[id]
		a.mu.RUnlock()
		if state == nil {
			writeError(w, http.StatusNotFound, "存档不存在")
			return
		}
		writeJSON(w, http.StatusOK, state)
		return
	}

	if len(parts) == 2 && parts[1] == "choices" && r.Method == http.MethodPost {
		a.applyChoice(w, r, id)
		return
	}

	writeError(w, http.StatusNotFound, "未知的游戏接口")
}

func (a *App) applyChoice(w http.ResponseWriter, r *http.Request, id string) {
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
