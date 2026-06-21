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

func TestCommandEndpointAppliesPlayerOrder(t *testing.T) {
	handler := NewWithOptions(Options{LegacyArchivePath: ""})
	createBody := bytes.NewBufferString(`{"scenarioId":"dongzhuo","playerId":"caocao"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRec.Code, http.StatusCreated)
	}

	commandBody := bytes.NewBufferString(`{"cityId":"xuchang","generalId":"cao-cao","commandId":"assart"}`)
	commandReq := httptest.NewRequest(http.MethodPost, "/api/games/current/command", commandBody)
	commandRec := httptest.NewRecorder()
	handler.ServeHTTP(commandRec, commandReq)

	if commandRec.Code != http.StatusOK {
		t.Fatalf("command status = %d, want %d; body: %s", commandRec.Code, http.StatusOK, commandRec.Body.String())
	}
}

func TestCommandEndpointAppliesTargetedMove(t *testing.T) {
	handler := NewWithOptions(Options{LegacyArchivePath: ""})
	createBody := bytes.NewBufferString(`{"scenarioId":"dongzhuo","playerId":"caocao"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRec.Code, http.StatusCreated)
	}

	commandBody := bytes.NewBufferString(`{"cityId":"xuchang","generalId":"cao-cao","commandId":"move","targetCityId":"chenliu"}`)
	commandReq := httptest.NewRequest(http.MethodPost, "/api/games/current/command", commandBody)
	commandRec := httptest.NewRecorder()
	handler.ServeHTTP(commandRec, commandReq)

	if commandRec.Code != http.StatusOK {
		t.Fatalf("command status = %d, want %d; body: %s", commandRec.Code, http.StatusOK, commandRec.Body.String())
	}
	var body commandSnapshot
	if err := json.NewDecoder(commandRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode command response: %v", err)
	}
	for _, general := range body.Generals {
		if general.ID == "cao-cao" && general.CityID == "chenliu" {
			return
		}
	}
	t.Fatalf("cao-cao was not moved to chenliu: %#v", body.Generals)
}

func TestBattleEndpointAppliesPlannedEmptyCityOccupation(t *testing.T) {
	handler := NewWithOptions(Options{LegacyArchivePath: ""})
	createBody := bytes.NewBufferString(`{"scenarioId":"dongzhuo","playerId":"caocao"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRec.Code, http.StatusCreated)
	}

	battleBody := bytes.NewBufferString(`{"cityId":"xuchang","generalIds":["cao-cao"],"targetCityId":"jingzhou","money":25,"food":100,"remainingFood":75,"fieldAdvantage":12}`)
	battleReq := httptest.NewRequest(http.MethodPost, "/api/games/current/battle", battleBody)
	battleRec := httptest.NewRecorder()
	handler.ServeHTTP(battleRec, battleReq)

	if battleRec.Code != http.StatusOK {
		t.Fatalf("battle status = %d, want %d; body: %s", battleRec.Code, http.StatusOK, battleRec.Body.String())
	}
	var body battleEndpointBody
	if err := json.NewDecoder(battleRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode battle response: %v", err)
	}
	if !body.Outcome.Won || !body.Outcome.Captured || body.Outcome.TargetCityID != "jingzhou" || body.Outcome.RemainingFood != 75 {
		t.Fatalf("planned empty-city outcome = %#v, want captured jingzhou with remainingFood 75", body.Outcome)
	}
	if body.Outcome.AttackerLosses != 0 || body.Outcome.DefenderLosses != 0 {
		t.Fatalf("empty-city losses = attacker %d defender %d, want 0/0", body.Outcome.AttackerLosses, body.Outcome.DefenderLosses)
	}
	if city := findBattleCity(body.Snapshot.Cities, "xuchang"); city.OwnerID != "caocao" || city.Money != 975 || city.Food != 1100 {
		t.Fatalf("origin city = %#v, want caocao money 975 food 1100", city)
	}
	if city := findBattleCity(body.Snapshot.Cities, "jingzhou"); city.OwnerID != "caocao" || city.Food != 1400 || city.PeopleDevotion != 72 {
		t.Fatalf("target city = %#v, want caocao food 1400 devotion 72", city)
	}
	if general := findBattleGeneral(body.Snapshot.Generals, "cao-cao"); general.CityID != "jingzhou" || general.Stamina != 96 || general.Soldiers != 1000 {
		t.Fatalf("attacker = %#v, want moved to jingzhou without loss/stamina cost", general)
	}
}

func TestScenariosEndpointListsLegacyPeriodsAndRulers(t *testing.T) {
	handler := NewWithOptions(Options{LegacyArchivePath: legacyArchivePath(t)})
	req := httptest.NewRequest(http.MethodGet, "/api/scenarios", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body scenarioListBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode scenarios response: %v", err)
	}
	if len(body.Scenarios) != 4 {
		t.Fatalf("scenarios = %d, want 4", len(body.Scenarios))
	}
	if body.Scenarios[0].ID != "period-1" || body.Scenarios[0].Year != 190 {
		t.Fatalf("first scenario = %#v, want period-1 year 190", body.Scenarios[0])
	}
	if len(body.Scenarios[0].Rulers) == 0 {
		t.Fatal("period-1 has no selectable rulers")
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

type commandSnapshot struct {
	Generals []struct {
		ID     string `json:"id"`
		CityID string `json:"cityId"`
	} `json:"generals"`
}

type battleEndpointBody struct {
	Outcome struct {
		Won            bool   `json:"won"`
		TargetCityID   string `json:"targetCityId"`
		Money          int    `json:"money"`
		Food           int    `json:"food"`
		RemainingFood  int    `json:"remainingFood"`
		FieldAdvantage int    `json:"fieldAdvantage"`
		AttackerLosses int    `json:"attackerLosses"`
		DefenderLosses int    `json:"defenderLosses"`
		Captured       bool   `json:"captured"`
	} `json:"outcome"`
	Snapshot struct {
		Cities []struct {
			ID             string `json:"id"`
			OwnerID        string `json:"ownerId"`
			Money          int    `json:"money"`
			Food           int    `json:"food"`
			PeopleDevotion int    `json:"peopleDevotion"`
		} `json:"cities"`
		Generals []struct {
			ID       string `json:"id"`
			CityID   string `json:"cityId"`
			Stamina  int    `json:"stamina"`
			Soldiers int    `json:"soldiers"`
		} `json:"generals"`
	} `json:"snapshot"`
}

type legacyResourcesBody struct {
	Count     int `json:"count"`
	Resources []struct {
		ID         uint16 `json:"id"`
		ItemCount  uint16 `json:"itemCount"`
		ItemLength uint16 `json:"itemLength"`
	} `json:"resources"`
}

type scenarioListBody struct {
	Scenarios []struct {
		ID     string `json:"id"`
		Year   int    `json:"year"`
		Rulers []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"rulers"`
	} `json:"scenarios"`
}

func (b legacyResourcesBody) hasResource(id, itemCount, itemLength uint16) bool {
	for _, resource := range b.Resources {
		if resource.ID == id && resource.ItemCount == itemCount && resource.ItemLength == itemLength {
			return true
		}
	}
	return false
}

func findBattleCity(cities []struct {
	ID             string `json:"id"`
	OwnerID        string `json:"ownerId"`
	Money          int    `json:"money"`
	Food           int    `json:"food"`
	PeopleDevotion int    `json:"peopleDevotion"`
}, id string) struct {
	ID             string `json:"id"`
	OwnerID        string `json:"ownerId"`
	Money          int    `json:"money"`
	Food           int    `json:"food"`
	PeopleDevotion int    `json:"peopleDevotion"`
} {
	for _, city := range cities {
		if city.ID == id {
			return city
		}
	}
	return struct {
		ID             string `json:"id"`
		OwnerID        string `json:"ownerId"`
		Money          int    `json:"money"`
		Food           int    `json:"food"`
		PeopleDevotion int    `json:"peopleDevotion"`
	}{}
}

func findBattleGeneral(generals []struct {
	ID       string `json:"id"`
	CityID   string `json:"cityId"`
	Stamina  int    `json:"stamina"`
	Soldiers int    `json:"soldiers"`
}, id string) struct {
	ID       string `json:"id"`
	CityID   string `json:"cityId"`
	Stamina  int    `json:"stamina"`
	Soldiers int    `json:"soldiers"`
} {
	for _, general := range generals {
		if general.ID == id {
			return general
		}
	}
	return struct {
		ID       string `json:"id"`
		CityID   string `json:"cityId"`
		Stamina  int    `json:"stamina"`
		Soldiers int    `json:"soldiers"`
	}{}
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
