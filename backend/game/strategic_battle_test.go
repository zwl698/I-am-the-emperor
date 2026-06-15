package game

import "testing"

func TestAssaultUsesFriendlySupportAndRecordsBattleReport(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2901)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	mainIndex, _ := state.Strategy.armyIndex("northern-banner")
	supportIndex, _ := state.Strategy.armyIndex("imperial-guard")
	state.Strategy.Armies[mainIndex].Troops = 9000
	state.Strategy.Armies[mainIndex].Morale = 58
	state.Strategy.Armies[mainIndex].Training = 54
	state.Strategy.Armies[supportIndex].Location = "north"
	state.Strategy.Armies[supportIndex].Troops = 38000
	state.Strategy.Armies[supportIndex].Grain = 90
	beforeSupportTroops := state.Strategy.Armies[supportIndex].Troops

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionArmyCommand, Target: "northern-banner:snow-ridge", Mode: "assault"}); err != nil {
		t.Fatalf("assault with support: %v", err)
	}

	snow, _ := state.Strategy.City("snow-ridge")
	support, _ := state.Strategy.Army("imperial-guard")
	if snow.OwnerID != "court" {
		t.Fatalf("expected support to help capture snow-ridge, got %+v", snow)
	}
	if support.Troops >= beforeSupportTroops {
		t.Fatalf("expected supporting army to share losses, before %d after %d", beforeSupportTroops, support.Troops)
	}
	if len(state.Strategy.Battles) == 0 {
		t.Fatalf("expected battle report")
	}
	report := state.Strategy.Battles[0]
	if report.CityID != "snow-ridge" || report.Outcome != "capture" {
		t.Fatalf("expected capture report for snow-ridge, got %+v", report)
	}
	if report.AttackerLoss <= 0 || report.DefenderLoss <= 0 {
		t.Fatalf("expected non-zero losses in report: %+v", report)
	}
	if !containsString(report.Participants, "imperial-guard") {
		t.Fatalf("expected support participant in report: %+v", report)
	}
	if len(report.Factors) < 3 || !containsText(report.Factors, "攻势") || !containsText(report.Factors, "守势") {
		t.Fatalf("expected battle report to explain combat factors, got %+v", report.Factors)
	}
}

func TestBesiegeCanForceSurrenderAfterMultipleRounds(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2902)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	state.Command = 5
	armyIndex, _ := state.Strategy.armyIndex("northern-banner")
	cityIndex, _ := state.Strategy.cityIndex("snow-ridge")
	state.Strategy.Armies[armyIndex].Troops = 26000
	state.Strategy.Armies[armyIndex].Grain = 90
	state.Strategy.Cities[cityIndex].Grain = 10
	state.Strategy.Cities[cityIndex].Order = 18

	for i := 0; i < 3; i++ {
		if _, err := state.ApplyAction(ActionRequest{Kind: ActionSiegeCommand, Target: "northern-banner:snow-ridge", Mode: "besiege"}); err != nil {
			t.Fatalf("besiege round %d: %v", i+1, err)
		}
	}

	snow, _ := state.Strategy.City("snow-ridge")
	army, _ := state.Strategy.Army("northern-banner")
	if snow.OwnerID != "court" {
		t.Fatalf("expected snow-ridge to surrender after siege, got %+v", snow)
	}
	if army.Status != "围城迫降" {
		t.Fatalf("expected army status to record surrender, got %+v", army)
	}
	if len(state.Strategy.Battles) == 0 || state.Strategy.Battles[0].Outcome != "surrender" {
		t.Fatalf("expected surrender battle report, got %+v", state.Strategy.Battles)
	}
	if !containsText(state.Strategy.Battles[0].Factors, "围城") || !containsText(state.Strategy.Battles[0].Factors, "城粮") {
		t.Fatalf("expected surrender report to expose siege factors, got %+v", state.Strategy.Battles[0].Factors)
	}
}

func TestEnemyStrategicAIActivelyCapturesWeakFrontCity(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 2903)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	cityIndex, _ := state.Strategy.cityIndex("north")
	armyIndex, _ := state.Strategy.armyIndex("beidi-vanguard")
	state.Strategy.Cities[cityIndex].Troops = 900
	state.Strategy.Cities[cityIndex].Defense = 6
	state.Strategy.Cities[cityIndex].Order = 12
	state.Strategy.Armies[armyIndex].Troops = 36000
	state.Strategy.Armies[armyIndex].Morale = 85
	beforeThreat := state.Stats.BorderThreat

	state.applyWorldPressure(DomainCourt)

	north, _ := state.Strategy.City("north")
	enemy, _ := state.Strategy.Army("beidi-vanguard")
	if north.OwnerID != "beidi" {
		t.Fatalf("expected beidi to capture weak north, got %+v", north)
	}
	if enemy.Location != "north" {
		t.Fatalf("expected enemy army to enter north, got %+v", enemy)
	}
	if state.Stats.BorderThreat <= beforeThreat {
		t.Fatalf("expected border threat to rise, before %d after %d", beforeThreat, state.Stats.BorderThreat)
	}
	if len(state.Strategy.Battles) == 0 || state.Strategy.Battles[0].Outcome != "enemy_capture" {
		t.Fatalf("expected enemy capture report, got %+v", state.Strategy.Battles)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsText(items []string, text string) bool {
	for _, item := range items {
		if len(item) >= len(text) {
			for i := 0; i <= len(item)-len(text); i++ {
				if item[i:i+len(text)] == text {
					return true
				}
			}
		}
	}
	return false
}
