package game

import "testing"

func TestStrategicPressureConsumesArmySupply(t *testing.T) {
	state := NewGame(2801)
	state.ForceCoronationForTest()
	armyIndex, ok := state.Strategy.armyIndex("northern-banner")
	if !ok {
		t.Fatalf("missing northern-banner")
	}
	state.Strategy.Armies[armyIndex].Grain = 0
	before := state.Strategy.Armies[armyIndex]

	state.applyWorldPressure(DomainCourt)

	after := state.Strategy.Armies[armyIndex]
	if after.Troops >= before.Troops || after.Morale >= before.Morale {
		t.Fatalf("expected hungry army to lose troops and morale, before %+v after %+v", before, after)
	}
	if len(state.Strategy.Logs) == 0 {
		t.Fatalf("expected strategy log for supply pressure")
	}
}

func TestStrategicPressurePushesForeignWarFront(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2802)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	beforeNorth, _ := state.Strategy.City("north")
	beforeThreat := state.Stats.BorderThreat

	state.applyWorldPressure(DomainCourt)

	afterNorth, _ := state.Strategy.City("north")
	if afterNorth.Order >= beforeNorth.Order {
		t.Fatalf("expected Beidi pressure to lower northern order, before %+v after %+v", beforeNorth, afterNorth)
	}
	if state.Stats.BorderThreat <= beforeThreat {
		t.Fatalf("expected foreign front pressure to increase border threat, before %d after %d", beforeThreat, state.Stats.BorderThreat)
	}
	if len(state.Strategy.Logs) == 0 {
		t.Fatalf("expected strategy log for front pressure")
	}
}

func TestStrategicPressureSyncsForeignThreatFromMap(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2803)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	factionIndex, _ := state.Strategy.factionIndex("beidi")
	foreignIndex, _ := state.findForeignIndex("beidi")
	state.Strategy.Factions[factionIndex].Threat = 96
	state.Strategy.Factions[factionIndex].Relation = 18
	state.ForeignStates[foreignIndex].Threat = 22
	state.ForeignStates[foreignIndex].Relation = 60

	state.applyStrategicPressure(DomainCourt)

	foreign := state.ForeignStates[foreignIndex]
	if foreign.Threat < 70 || foreign.Relation >= 60 {
		t.Fatalf("expected strategic faction pressure to sync into foreign state, got %+v", foreign)
	}
}
