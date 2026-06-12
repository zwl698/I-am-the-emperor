package game

import "testing"

func TestEnsurePlayableStateRepairsLegacyEmperorState(t *testing.T) {
	state := NewGame(1301)
	state.Phase = PhaseEmperor
	state.Turn = 19
	state.ReignYear = 0
	state.Season = ""
	state.Command = 0
	state.Harem = nil
	state.Heirs = nil
	state.Offices = nil
	state.Projects = nil
	state.Policies = nil
	state.Relations = nil
	state.ForeignStates = nil
	state.Plots = nil
	state.LegalCases = nil
	state.PublicOpinion = PublicOpinion{}
	state.Provinces = nil
	state.Wars = nil
	state.Scene = nil

	state.EnsurePlayableState()

	if state.ReignYear < 1 || state.Season == "" {
		t.Fatalf("expected repaired calendar, got year=%d season=%q", state.ReignYear, state.Season)
	}
	if len(state.Harem) == 0 || len(state.Offices) == 0 || len(state.Projects) == 0 || len(state.ForeignStates) == 0 || len(state.LegalCases) == 0 {
		t.Fatalf("expected systems to be repaired: harem=%d offices=%d projects=%d foreign=%d cases=%d", len(state.Harem), len(state.Offices), len(state.Projects), len(state.ForeignStates), len(state.LegalCases))
	}
	if len(state.Provinces) == 0 || len(state.Wars) == 0 {
		t.Fatalf("expected world systems to be repaired: provinces=%d wars=%d", len(state.Provinces), len(state.Wars))
	}
	if state.Scene == nil || len(state.Scene.Choices) == 0 {
		t.Fatalf("expected emperor scene after repair, got %+v", state.Scene)
	}
	if state.Command != 0 {
		t.Fatalf("repair should not refill spent command points, got %d", state.Command)
	}
}
