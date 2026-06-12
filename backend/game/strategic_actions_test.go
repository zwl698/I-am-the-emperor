package game

import "testing"

func TestCityDevelopActionChangesStrategicCity(t *testing.T) {
	state := NewGame(2701)
	state.ForceCoronationForTest()
	beforeCommand := state.Command
	before, _ := state.Strategy.City("south")

	resolution, err := state.ApplyAction(ActionRequest{Kind: ActionCityDevelop, Target: "south", Mode: "farm"})
	if err != nil {
		t.Fatalf("apply city develop action: %v", err)
	}

	after, _ := state.Strategy.City("south")
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected city action to spend command, before %d after %d", beforeCommand, state.Command)
	}
	if after.Agriculture <= before.Agriculture || after.Grain <= before.Grain {
		t.Fatalf("expected farm to improve agriculture and grain, before %+v after %+v", before, after)
	}
	if resolution.Summary == "" {
		t.Fatalf("expected resolution summary")
	}
}

func TestArmyCommandMarchUsesRoadNetwork(t *testing.T) {
	state := NewGame(2702)
	state.ForceCoronationForTest()
	beforeCommand := state.Command

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionArmyCommand, Target: "imperial-guard:luoyang", Mode: "march"}); err != nil {
		t.Fatalf("march imperial guard: %v", err)
	}

	army, _ := state.Strategy.Army("imperial-guard")
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected march to spend command, before %d after %d", beforeCommand, state.Command)
	}
	if army.Location != "luoyang" || army.Status != "行军抵达" {
		t.Fatalf("expected army to move to luoyang, got %+v", army)
	}

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionArmyCommand, Target: "imperial-guard:east-sea", Mode: "march"}); err == nil {
		t.Fatalf("expected non-adjacent march to fail")
	}
}

func TestArmyAssaultCanCaptureForeignWarFront(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2703)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	armyIndex, ok := state.Strategy.armyIndex("northern-banner")
	if !ok {
		t.Fatalf("missing northern-banner")
	}
	state.Strategy.Armies[armyIndex].Troops = 42000
	state.Strategy.Armies[armyIndex].Grain = 90
	beforeThreat := state.Stats.BorderThreat

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionArmyCommand, Target: "northern-banner:snow-ridge", Mode: "assault"}); err != nil {
		t.Fatalf("assault snow ridge: %v", err)
	}

	snow, _ := state.Strategy.City("snow-ridge")
	army, _ := state.Strategy.Army("northern-banner")
	if snow.OwnerID != "court" {
		t.Fatalf("expected snow-ridge captured by court, got %+v", snow)
	}
	if army.Location != "snow-ridge" {
		t.Fatalf("expected army to enter captured city, got %+v", army)
	}
	if state.Stats.BorderThreat >= beforeThreat {
		t.Fatalf("expected captured front to reduce border threat, before %d after %d", beforeThreat, state.Stats.BorderThreat)
	}
}

func TestGovernorAssignActionSetsCityGovernor(t *testing.T) {
	state := NewGame(2704)
	state.ForceCoronationForTest()
	before, _ := state.Strategy.City("north")

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionGovernorAssign, Target: "north:gu", Mode: "appoint"}); err != nil {
		t.Fatalf("assign governor: %v", err)
	}

	after, _ := state.Strategy.City("north")
	if after.GovernorID != "gu" {
		t.Fatalf("expected governor gu, got %+v", after)
	}
	if after.Order <= before.Order {
		t.Fatalf("expected capable governor to improve order, before %+v after %+v", before, after)
	}
}
