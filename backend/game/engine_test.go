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
