package game

import "testing"

func TestStrategicMapInitializesForEmperor(t *testing.T) {
	state := NewGame(2601)
	state.ForceCoronationForTest()

	if len(state.Strategy.Cities) < 12 {
		t.Fatalf("expected at least 12 strategic cities, got %d", len(state.Strategy.Cities))
	}
	if len(state.Strategy.Roads) < 14 {
		t.Fatalf("expected at least 14 strategic roads, got %d", len(state.Strategy.Roads))
	}
	if len(state.Strategy.Factions) < 5 {
		t.Fatalf("expected at least 5 strategic factions, got %d", len(state.Strategy.Factions))
	}
	if len(state.Strategy.Armies) < 4 {
		t.Fatalf("expected at least 4 strategic armies, got %d", len(state.Strategy.Armies))
	}

	for _, city := range state.Strategy.Cities {
		if city.ID == "" || city.Name == "" || city.OwnerID == "" || city.X <= 0 || city.Y <= 0 {
			t.Fatalf("city should be renderable and owned: %+v", city)
		}
		if len(state.Strategy.Neighbors(city.ID)) == 0 {
			t.Fatalf("city %s should have at least one road neighbor", city.ID)
		}
	}
}

func TestXuanshuoStrategicMapStartsWithNorthernWarFront(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2602)
	if err != nil {
		t.Fatalf("create xuanshuo game: %v", err)
	}
	state.ForceCoronationForTest()

	north, ok := state.Strategy.City("north")
	if !ok {
		t.Fatalf("expected north city in strategic map")
	}
	snow, ok := state.Strategy.City("snow-ridge")
	if !ok {
		t.Fatalf("expected snow-ridge city in strategic map")
	}
	beidi, ok := state.Strategy.Faction("beidi")
	if !ok {
		t.Fatalf("expected beidi faction in strategic map")
	}
	if !north.Front || !snow.Front {
		t.Fatalf("expected north and snow-ridge to be front cities, north=%+v snow=%+v", north, snow)
	}
	if snow.OwnerID != "beidi" {
		t.Fatalf("expected snow-ridge owned by beidi, got %+v", snow)
	}
	if beidi.Threat < 75 {
		t.Fatalf("expected xuanshuo beidi threat to start high, got %+v", beidi)
	}
	if !state.Strategy.AreAdjacent("north", "snow-ridge") {
		t.Fatalf("expected north and snow-ridge to be adjacent")
	}
}
