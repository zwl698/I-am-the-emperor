package legacyres

import (
	"testing"
)

func TestLoadScenarioAssemblesDongzhuo(t *testing.T) {
	archive, err := Open(decodeTestArchivePath(t))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	sc, err := archive.LoadScenario()
	if err != nil {
		t.Fatalf("LoadScenario error = %v", err)
	}

	if len(sc.Cities) != 38 {
		t.Fatalf("cities = %d, want 38", len(sc.Cities))
	}
	if len(sc.Persons) == 0 {
		t.Fatal("no persons loaded")
	}

	// City 0 should be 西凉 with a valid grid coordinate.
	if sc.Cities[0].Name != "西凉" {
		t.Errorf("city[0] name = %q, want 西凉", sc.Cities[0].Name)
	}
	if sc.Cities[0].X < 0 || sc.Cities[0].Y < 0 {
		t.Errorf("city[0] has no map coordinate: (%d,%d)", sc.Cities[0].X, sc.Cities[0].Y)
	}

	// All cities placed on the grid should have coordinates within bounds.
	placed := 0
	for _, c := range sc.Cities {
		if c.X >= 0 && c.Y >= 0 {
			placed++
			if c.X >= CityMapWidth || c.Y >= CityMapHeight {
				t.Errorf("city %q coordinate out of bounds: (%d,%d)", c.Name, c.X, c.Y)
			}
		}
	}
	if placed != 38 {
		t.Errorf("placed %d/38 cities on the grid", placed)
	}

	// Person 0 should be 董卓.
	if sc.Persons[0].Name != "董卓" {
		t.Errorf("person[0] name = %q, want 董卓", sc.Persons[0].Name)
	}

	// Log a summary of the assembled scenario.
	t.Logf("Scenario year=%d cities=%d persons=%d", sc.Year, len(sc.Cities), len(sc.Persons))
	for i := 0; i < 5; i++ {
		c := sc.Cities[i]
		t.Logf("  city[%d] %s (%d,%d) 归属%d 金%d 粮%d 人口%d", i, c.Name, c.X, c.Y, c.Belong, c.Money, c.Food, c.Population)
	}
	for i := 0; i < 5; i++ {
		p := sc.Persons[i]
		t.Logf("  person[%d] %s 武%d 智%d %s %s", i, p.Name, p.Force, p.IQ, p.Character, p.ArmsType)
	}
}

