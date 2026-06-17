package game

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func legacyArchivePathForGame(t *testing.T) string {
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

func TestNewGameFromArchiveBuildsRealScenario(t *testing.T) {
	state, err := NewGameFromArchive(legacyArchivePathForGame(t), "")
	if err != nil {
		t.Fatalf("NewGameFromArchive error = %v", err)
	}

	if len(state.Cities) != 38 {
		t.Fatalf("cities = %d, want 38", len(state.Cities))
	}
	if len(state.Rulers) < 2 {
		t.Fatalf("rulers = %d, want several factions", len(state.Rulers))
	}
	if len(state.Generals) == 0 {
		t.Fatal("no generals placed")
	}
	if len(state.Routes) == 0 {
		t.Fatal("no routes built")
	}

	// The first city should be a real legacy city with a name and coordinate.
	c0 := state.Cities[0]
	if c0.Name == "" {
		t.Error("city[0] has no name")
	}
	if c0.X < 0 || c0.Y < 0 {
		t.Errorf("city[0] %s has no coordinate (%d,%d)", c0.Name, c0.X, c0.Y)
	}

	// Every owned city must reference an existing ruler.
	rulerIDs := map[string]bool{}
	for _, r := range state.Rulers {
		rulerIDs[r.ID] = true
	}
	for _, c := range state.Cities {
		if !rulerIDs[c.OwnerID] {
			t.Errorf("city %s has unknown owner %q", c.Name, c.OwnerID)
		}
	}

	// Every general must reference an existing city and ruler.
	cityIDs := map[string]bool{}
	for _, c := range state.Cities {
		cityIDs[c.ID] = true
	}
	for _, g := range state.Generals {
		if !cityIDs[g.CityID] {
			t.Errorf("general %s in unknown city %q", g.Name, g.CityID)
		}
		if !rulerIDs[g.OwnerID] {
			t.Errorf("general %s has unknown owner %q", g.Name, g.OwnerID)
		}
		if g.Force <= 0 || g.Force > 100 {
			t.Errorf("general %s force=%d out of range", g.Name, g.Force)
		}
	}

	t.Logf("Real scenario: year=%d rulers=%d cities=%d generals=%d routes=%d player=%s",
		state.Date.Year, len(state.Rulers), len(state.Cities), len(state.Generals), len(state.Routes), state.PlayerID)
	for _, r := range state.Rulers {
		if r.ID == "neutral" {
			continue
		}
		t.Logf("  ruler %s (%s)", r.Name, r.Character)
	}
}

func TestNewGameFromArchiveAdvanceMonthWorks(t *testing.T) {
	state, err := NewGameFromArchive(legacyArchivePathForGame(t), "")
	if err != nil {
		t.Fatalf("NewGameFromArchive error = %v", err)
	}

	beforeMonth := state.Date.Month
	state.AdvanceMonth()
	if state.Date.Month == beforeMonth && state.Date.Year == 189 {
		t.Errorf("advance month did not change date: %d/%d", state.Date.Year, state.Date.Month)
	}
}

