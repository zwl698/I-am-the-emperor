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

func TestWarTacticMobilizeFeedsStrategicFrontArmy(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 1503)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	armyIndex, _ := state.Strategy.armyIndex("northern-banner")
	state.Strategy.Armies[armyIndex].Grain = 12
	before := state.Strategy.Armies[armyIndex]

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionWarTactic, Target: "snow-ridge", Mode: "mobilize"}); err != nil {
		t.Fatalf("mobilize war tactic: %v", err)
	}

	after := state.Strategy.Armies[armyIndex]
	if after.Grain <= before.Grain || after.Morale <= before.Morale || after.Training <= before.Training {
		t.Fatalf("expected mobilize to resupply and drill strategic army, before %+v after %+v", before, after)
	}
	if len(state.Strategy.Logs) == 0 {
		t.Fatalf("expected strategic log after mobilize")
	}
}

func TestWarTacticCampaignResolvesStrategicAssault(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 1504)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	armyIndex, _ := state.Strategy.armyIndex("northern-banner")
	cityIndex, _ := state.Strategy.cityIndex("snow-ridge")
	state.Strategy.Armies[armyIndex].Troops = 52000
	state.Strategy.Armies[armyIndex].Grain = 100
	state.Strategy.Armies[armyIndex].Morale = 86
	state.Strategy.Armies[armyIndex].Training = 82
	state.Strategy.Cities[cityIndex].Troops = 3200
	state.Strategy.Cities[cityIndex].Defense = 18

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionWarTactic, Target: "snow-ridge", Mode: "campaign"}); err != nil {
		t.Fatalf("campaign war tactic: %v", err)
	}

	snow, _ := state.Strategy.City("snow-ridge")
	if snow.OwnerID != "court" {
		t.Fatalf("expected campaign to resolve on strategic city, got %+v", snow)
	}
	if len(state.Strategy.Battles) == 0 || state.Strategy.Battles[0].Outcome != "capture" {
		t.Fatalf("expected strategic battle report from campaign, got %+v", state.Strategy.Battles)
	}
}

func TestWarTacticFortifyAndTruceAffectStrategicFront(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 1505)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	beforeNorth, _ := state.Strategy.City("north")
	beforeBeidi, _ := state.Strategy.Faction("beidi")

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionWarTactic, Target: "snow-ridge", Mode: "fortify"}); err != nil {
		t.Fatalf("fortify war tactic: %v", err)
	}
	afterNorth, _ := state.Strategy.City("north")
	if afterNorth.Defense <= beforeNorth.Defense || afterNorth.Troops <= beforeNorth.Troops {
		t.Fatalf("expected fortify to reinforce strategic front, before %+v after %+v", beforeNorth, afterNorth)
	}

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionWarTactic, Target: "snow-ridge", Mode: "truce"}); err != nil {
		t.Fatalf("truce war tactic: %v", err)
	}
	afterBeidi, _ := state.Strategy.Faction("beidi")
	if afterBeidi.Threat >= beforeBeidi.Threat || afterBeidi.Relation <= beforeBeidi.Relation {
		t.Fatalf("expected truce to soften strategic faction, before %+v after %+v", beforeBeidi, afterBeidi)
	}
}

func TestWarTacticFortifyUsesCampaignFrontProvince(t *testing.T) {
	state, err := NewGameWithDynasty("dayin", 1506)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	westIndex, _ := state.findProvinceIndex("west")
	northIndex, _ := state.findProvinceIndex("north")
	beforeWest := state.Provinces[westIndex]
	beforeNorth := state.Provinces[northIndex]

	if _, err := state.ApplyAction(ActionRequest{Kind: ActionWarTactic, Target: "western-oath", Mode: "fortify"}); err != nil {
		t.Fatalf("fortify western war tactic: %v", err)
	}

	afterWest := state.Provinces[westIndex]
	afterNorth := state.Provinces[northIndex]
	if afterWest.Defense <= beforeWest.Defense {
		t.Fatalf("expected western fortify to reinforce west province, before %+v after %+v", beforeWest, afterWest)
	}
	if afterNorth.Defense != beforeNorth.Defense {
		t.Fatalf("expected western fortify not to reinforce north province, before %+v after %+v", beforeNorth, afterNorth)
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
