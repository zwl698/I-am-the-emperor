package game

import "testing"

func TestApplyActionWarTacticUsesUnifiedActionLayer(t *testing.T) {
	state := NewGame(1501)
	state.ForceCoronationForTest()
	war := state.Wars[0]
	beforeCommand := state.Command

	resolution, err := state.ApplyAction(ActionRequest{Kind: ActionWarTactic, Target: war.ID, Mode: "campaign"})
	if err != nil {
		t.Fatalf("apply war tactic action: %v", err)
	}

	after := state.Wars[0]
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected action to spend command, before %d after %d", beforeCommand, state.Command)
	}
	if after.Progress <= war.Progress || after.Threat >= war.Threat {
		t.Fatalf("expected campaign action to advance war, before %+v after %+v", war, after)
	}
	if resolution.Summary == "" {
		t.Fatalf("expected action resolution summary")
	}
}

func TestApplyActionCoversMapTrialOfficeHeirAndDiplomacy(t *testing.T) {
	state := NewGame(1502)
	state.ForceCoronationForTest()
	state.Command = 6
	province := state.Provinces[0]
	legalCase := state.LegalCases[0]
	office := state.Offices[0]
	minister := state.Court[0]
	heir := state.Heirs[0]
	foreign := state.ForeignStates[0]

	actions := []ActionRequest{
		{Kind: ActionMapAllocation, Target: province.ID, Mode: "relief"},
		{Kind: ActionTrialMove, Target: legalCase.ID, Mode: "open_trial"},
		{Kind: ActionOfficeAssign, Target: office.ID + ":" + minister.ID},
		{Kind: ActionHeirLesson, Target: heir.ID, Mode: "study"},
		{Kind: ActionEnvoyMission, Target: foreign.ID, Mode: "embassy"},
	}
	for _, action := range actions {
		if _, err := state.ApplyAction(action); err != nil {
			t.Fatalf("apply action %+v: %v", action, err)
		}
	}

	if state.Command != 1 {
		t.Fatalf("expected five actions to spend command points, got %d", state.Command)
	}
	if state.Provinces[0].Disaster >= province.Disaster {
		t.Fatalf("expected map allocation to reduce disaster, before %+v after %+v", province, state.Provinces[0])
	}
	if !state.LegalCases[0].Resolved {
		t.Fatalf("expected trial move to resolve case: %+v", state.LegalCases[0])
	}
	if state.Offices[0].HolderID != minister.ID {
		t.Fatalf("expected office assignment to set holder, got %+v", state.Offices[0])
	}
	if state.Heirs[0].Talent <= heir.Talent {
		t.Fatalf("expected heir lesson to improve talent, before %+v after %+v", heir, state.Heirs[0])
	}
	if state.ForeignStates[0].Relation <= foreign.Relation {
		t.Fatalf("expected envoy mission to improve relation, before %+v after %+v", foreign, state.ForeignStates[0])
	}
}

func TestActionCatalogExposesPlayableCategories(t *testing.T) {
	catalog := ActionCatalog()
	if len(catalog) < 18 {
		t.Fatalf("expected rich action catalog, got %d", len(catalog))
	}
	seen := map[ActionKind]bool{}
	for _, action := range catalog {
		if action.Kind == "" || action.Mode == "" || action.Label == "" || action.Panel == "" {
			t.Fatalf("action catalog entry should be playable: %+v", action)
		}
		seen[action.Kind] = true
	}
	for _, kind := range []ActionKind{ActionMapAllocation, ActionWarTactic, ActionTrialMove, ActionOfficeAssign, ActionHeirLesson, ActionEnvoyMission} {
		if !seen[kind] {
			t.Fatalf("expected action kind %q in catalog: %+v", kind, catalog)
		}
	}
}
