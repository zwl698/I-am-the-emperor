package game

import "testing"

func TestNewGameStartsAsInfantPrinceWithOpeningScene(t *testing.T) {
	state := NewGame(7)

	if state.ID == "" {
		t.Fatal("expected a game id")
	}
	if state.Phase != PhasePrince {
		t.Fatalf("expected prince phase, got %q", state.Phase)
	}
	if state.Age != 1 {
		t.Fatalf("expected age 1, got %d", state.Age)
	}
	if state.Scene == nil {
		t.Fatal("expected opening scene")
	}
	if len(state.Scene.Choices) < 2 {
		t.Fatalf("expected multiple opening choices, got %d", len(state.Scene.Choices))
	}
	if state.Stats.Legitimacy <= 0 || state.Stats.Health <= 0 {
		t.Fatalf("expected initialized stats, got %+v", state.Stats)
	}
}

func TestAvailableDynastiesExposeDistinctStartsAndAssets(t *testing.T) {
	dynasties := AvailableDynasties()

	if len(dynasties) < 4 {
		t.Fatalf("expected at least four playable dynasties, got %d", len(dynasties))
	}

	seen := map[string]bool{}
	for _, dynasty := range dynasties {
		if dynasty.ID == "" || dynasty.Name == "" || dynasty.Background == "" {
			t.Fatalf("dynasty should have identity and background: %+v", dynasty)
		}
		if dynasty.Asset == "" {
			t.Fatalf("dynasty should expose a generated asset: %+v", dynasty)
		}
		if len(dynasty.Features) < 2 {
			t.Fatalf("dynasty should have multiple features: %+v", dynasty)
		}
		seen[dynasty.ID] = true
	}

	for _, id := range []string{"dayin", "jingyao", "chengping", "xuanshuo"} {
		if !seen[id] {
			t.Fatalf("expected dynasty %q in list", id)
		}
	}
}

func TestNewGameWithDynastyChangesHistoricalPressure(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 11)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}

	if state.Dynasty.ID != "xuanshuo" {
		t.Fatalf("expected xuanshuo dynasty, got %+v", state.Dynasty)
	}
	if state.Assets.Hero == "" || state.Assets.Characters == "" {
		t.Fatalf("expected generated art assets in state, got %+v", state.Assets)
	}
	if state.Stats.BorderThreat < 55 {
		t.Fatalf("frontier dynasty should start under heavy border pressure, got %+v", state.Stats)
	}
	if len(state.Factions) < 4 {
		t.Fatalf("expected court factions, got %+v", state.Factions)
	}
	if len(state.Provinces) < 4 {
		t.Fatalf("expected provinces, got %+v", state.Provinces)
	}
	if state.Scene == nil || state.Scene.Art == "" {
		t.Fatalf("expected opening scene with generated art: %+v", state.Scene)
	}
}

func TestNewGameIncludesLongTermObjectives(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 17)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}

	if len(state.Objectives) < 4 {
		t.Fatalf("expected long-term objectives for a 30-minute play loop, got %+v", state.Objectives)
	}

	seen := map[string]bool{}
	for _, objective := range state.Objectives {
		if objective.ID == "" || objective.Title == "" || objective.Target <= 0 {
			t.Fatalf("objective should have id/title/target: %+v", objective)
		}
		seen[objective.ID] = true
	}
	for _, id := range []string{"secure_throne", "stabilize_realm", "reform_state", "pacify_borders"} {
		if !seen[id] {
			t.Fatalf("expected objective %q in %+v", id, state.Objectives)
		}
	}
}

func TestApplyChoiceChangesStatsAndAdvancesStory(t *testing.T) {
	state := NewGame(2)
	openingScene := state.Scene.ID
	choice := state.Scene.Choices[0].ID
	beforeLearning := state.Stats.Learning

	resolution, err := state.ApplyChoice(choice)
	if err != nil {
		t.Fatalf("apply choice: %v", err)
	}

	if resolution == nil || resolution.Summary == "" {
		t.Fatal("expected choice resolution summary")
	}
	if state.Turn != 1 {
		t.Fatalf("expected turn 1, got %d", state.Turn)
	}
	if state.Scene.ID == openingScene {
		t.Fatal("expected a new scene after applying a choice")
	}
	if state.Stats.Learning == beforeLearning {
		t.Fatal("expected opening choice to change learning")
	}
}

func TestPrinceCanAscendToEmperorThroughStoryChoices(t *testing.T) {
	state := NewGame(13)

	for state.Phase != PhaseEmperor && state.Ending == nil {
		if _, err := state.ApplyChoice(state.Scene.Choices[0].ID); err != nil {
			t.Fatalf("apply choice: %v", err)
		}
	}

	if state.Ending != nil {
		t.Fatalf("expected coronation, got ending: %+v", state.Ending)
	}
	if state.Phase != PhaseEmperor {
		t.Fatalf("expected emperor phase, got %q", state.Phase)
	}
	if state.Age < 18 {
		t.Fatalf("expected adult emperor age, got %d", state.Age)
	}
	if state.Stats.Treasury <= 0 || state.Stats.Army <= 0 || state.Stats.Diplomacy <= 0 {
		t.Fatalf("expected imperial stats to be initialized, got %+v", state.Stats)
	}
}

func TestEmperorChoicesAffectStrategicDomains(t *testing.T) {
	state := NewGame(21)
	state.ForceCoronationForTest()
	before := state.Stats

	var militaryChoice string
	for _, choice := range state.Scene.Choices {
		if choice.Domain == DomainMilitary {
			militaryChoice = choice.ID
			break
		}
	}
	if militaryChoice == "" {
		t.Fatalf("expected a military choice in emperor scene: %+v", state.Scene.Choices)
	}

	if _, err := state.ApplyChoice(militaryChoice); err != nil {
		t.Fatalf("apply choice: %v", err)
	}

	if state.Stats.Army <= before.Army {
		t.Fatalf("expected army to increase, before %+v after %+v", before, state.Stats)
	}
	if state.Stats.Treasury >= before.Treasury {
		t.Fatalf("expected military investment to cost treasury, before %+v after %+v", before, state.Stats)
	}
}

func TestEmperorSceneOffersDeepStrategicChoices(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 31)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()

	if len(state.Scene.Choices) < 6 {
		t.Fatalf("expected a richer emperor choice set, got %+v", state.Scene.Choices)
	}

	domains := map[Domain]bool{}
	for _, choice := range state.Scene.Choices {
		domains[choice.Domain] = true
	}
	for _, domain := range []Domain{DomainDomestic, DomainEconomy, DomainMilitary, DomainDiplomacy, DomainReform, DomainIntrigue} {
		if !domains[domain] {
			t.Fatalf("expected domain %q in emperor choices: %+v", domain, state.Scene.Choices)
		}
	}
}

func TestStrategicChoiceMovesFactionAndProvinceState(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 41)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	beforeProvince := state.Provinces[0]
	beforeFaction := state.Factions[0]

	var reformChoice string
	for _, choice := range state.Scene.Choices {
		if choice.Domain == DomainReform {
			reformChoice = choice.ID
			break
		}
	}
	if reformChoice == "" {
		t.Fatal("expected reform choice")
	}

	if _, err := state.ApplyChoice(reformChoice); err != nil {
		t.Fatalf("apply choice: %v", err)
	}

	if state.Provinces[0] == beforeProvince {
		t.Fatalf("expected province state to change, before %+v after %+v", beforeProvince, state.Provinces[0])
	}
	if state.Factions[0] == beforeFaction {
		t.Fatalf("expected faction state to change, before %+v after %+v", beforeFaction, state.Factions[0])
	}
	if state.Season == "" || state.ReignYear < 1 {
		t.Fatalf("expected calendar to advance, got season %q year %d", state.Season, state.ReignYear)
	}
}

func TestStrategicChoiceAdvancesObjectiveProgress(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 51)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	before := objectiveProgress(t, state, "reform_state")

	var reformChoice string
	for _, choice := range state.Scene.Choices {
		if choice.Domain == DomainReform {
			reformChoice = choice.ID
			break
		}
	}
	if reformChoice == "" {
		t.Fatal("expected reform choice")
	}

	if _, err := state.ApplyChoice(reformChoice); err != nil {
		t.Fatalf("apply choice: %v", err)
	}

	after := objectiveProgress(t, state, "reform_state")
	if after <= before {
		t.Fatalf("expected reform objective progress to advance, before %d after %d", before, after)
	}
}

func TestBadRuleCanEndTheDynasty(t *testing.T) {
	state := NewGame(99)
	state.ForceCoronationForTest()
	state.Stats.Populace = 1
	state.Stats.Stability = 1
	state.Stats.BorderThreat = 95

	for i := 0; i < 6 && state.Ending == nil; i++ {
		if _, err := state.ApplyChoice(state.Scene.Choices[len(state.Scene.Choices)-1].ID); err != nil {
			t.Fatalf("apply choice: %v", err)
		}
	}

	if state.Ending == nil {
		t.Fatal("expected an ending after catastrophic rule")
	}
	if state.Ending.Kind != EndingCollapse {
		t.Fatalf("expected collapse ending, got %+v", state.Ending)
	}
}

func objectiveProgress(t *testing.T, state *GameState, id string) int {
	t.Helper()
	for _, objective := range state.Objectives {
		if objective.ID == id {
			return objective.Progress
		}
	}
	t.Fatalf("objective %q not found in %+v", id, state.Objectives)
	return 0
}
