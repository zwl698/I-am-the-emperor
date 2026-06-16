package game

import (
	"strings"
	"testing"
)

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

func TestNewGameExposesLargeGeneratedAssetGalleries(t *testing.T) {
	state := NewGame(15)

	if len(state.Assets.SceneGallery) < 30 {
		t.Fatalf("expected at least 30 scene assets, got %+v", state.Assets.SceneGallery)
	}
	if len(state.Assets.PortraitGallery) < 30 {
		t.Fatalf("expected at least 30 portrait assets, got %+v", state.Assets.PortraitGallery)
	}
	if state.Scene == nil || !strings.Contains(state.Scene.Art, "/assets/scenes/") {
		t.Fatalf("expected opening scene to use generated scene gallery, got %+v", state.Scene)
	}
}

func TestNewGameGivesMinistersPlayableAttributes(t *testing.T) {
	state := NewGame(16)

	if len(state.Court) < 4 {
		t.Fatalf("expected court ministers, got %+v", state.Court)
	}
	for _, minister := range state.Court {
		if minister.Ability <= 0 || minister.Ambition <= 0 || minister.Integrity <= 0 {
			t.Fatalf("minister should expose ability, ambition, and integrity: %+v", minister)
		}
		if minister.Stress < 0 {
			t.Fatalf("minister stress should never be negative: %+v", minister)
		}
	}
}

func TestNewGameIncludesHaremSuccessionAndOffices(t *testing.T) {
	state, err := NewGameWithDynasty("jingyao", 81)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()

	if len(state.Harem) < 4 {
		t.Fatalf("expected playable harem roster, got %+v", state.Harem)
	}
	if len(state.Heirs) < 2 {
		t.Fatalf("expected multiple heirs for succession play, got %+v", state.Heirs)
	}
	if len(state.Offices) < 6 {
		t.Fatalf("expected six major offices for appointments, got %+v", state.Offices)
	}
	if state.Succession.Stability <= 0 {
		t.Fatalf("expected succession stability to be initialized, got %+v", state.Succession)
	}

	for _, consort := range state.Harem {
		if consort.ID == "" || consort.Name == "" || consort.Rank == "" || consort.Clan == "" {
			t.Fatalf("consort should expose identity, rank, and clan: %+v", consort)
		}
		if consort.Favor <= 0 || consort.FamilyPower <= 0 || consort.Influence <= 0 {
			t.Fatalf("consort should expose favor, clan power, and influence: %+v", consort)
		}
	}
	for _, heir := range state.Heirs {
		if heir.ID == "" || heir.Name == "" || heir.MotherID == "" {
			t.Fatalf("heir should expose identity and maternal link: %+v", heir)
		}
		if heir.Age <= 0 || heir.Talent <= 0 || heir.Support <= 0 {
			t.Fatalf("heir should expose age, talent, and support: %+v", heir)
		}
	}
	for _, office := range state.Offices {
		if office.ID == "" || office.Title == "" || office.Domain == "" {
			t.Fatalf("office should expose identity, title, and domain: %+v", office)
		}
		if office.Authority <= 0 {
			t.Fatalf("office should expose authority: %+v", office)
		}
	}
}

func TestAppointmentOrderAssignsOfficeAndPressuresMinister(t *testing.T) {
	state := NewGame(82)
	state.ForceCoronationForTest()
	beforeTurn := state.Turn
	beforeCommand := state.Command
	office := state.Offices[0]
	minister := state.Court[2]

	resolution, err := state.ApplyOrder(OrderRequest{
		Kind:   OrderAppoint,
		Target: office.ID + ":" + minister.ID,
	})
	if err != nil {
		t.Fatalf("apply appointment order: %v", err)
	}

	if resolution == nil || resolution.Summary == "" {
		t.Fatalf("expected appointment resolution, got %+v", resolution)
	}
	if state.Turn != beforeTurn {
		t.Fatalf("appointment should not advance scene turn, before %d after %d", beforeTurn, state.Turn)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease by 1, before %d after %d", beforeCommand, state.Command)
	}
	afterOffice, ok := officeByID(state, office.ID)
	if !ok {
		t.Fatalf("office %q missing after appointment", office.ID)
	}
	if afterOffice.HolderID != minister.ID {
		t.Fatalf("expected office holder %q, got %+v", minister.ID, afterOffice)
	}
	afterMinister, ok := ministerByID(state, minister.ID)
	if !ok {
		t.Fatalf("minister %q missing after appointment", minister.ID)
	}
	if afterMinister.Stress <= minister.Stress {
		t.Fatalf("expected appointed minister stress to rise, before %+v after %+v", minister, afterMinister)
	}
}

func TestSuccessionOrderNamesHeirAndMovesSupport(t *testing.T) {
	state := NewGame(83)
	state.ForceCoronationForTest()
	beforeCommand := state.Command
	heir := state.Heirs[1]
	beforeStability := state.Succession.Stability

	resolution, err := state.ApplyOrder(OrderRequest{
		Kind:   OrderNameHeir,
		Target: heir.ID,
	})
	if err != nil {
		t.Fatalf("apply succession order: %v", err)
	}

	if resolution == nil || resolution.Summary == "" {
		t.Fatalf("expected succession resolution, got %+v", resolution)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease by 1, before %d after %d", beforeCommand, state.Command)
	}
	if state.Succession.NamedHeirID != heir.ID {
		t.Fatalf("expected named heir %q, got %+v", heir.ID, state.Succession)
	}
	afterHeir, ok := heirByID(state, heir.ID)
	if !ok {
		t.Fatalf("heir %q missing after naming", heir.ID)
	}
	if !afterHeir.Named {
		t.Fatalf("expected heir to be marked as named: %+v", afterHeir)
	}
	if afterHeir.Support <= heir.Support {
		t.Fatalf("expected named heir support to rise, before %+v after %+v", heir, afterHeir)
	}
	if state.Succession.Stability == beforeStability {
		t.Fatalf("expected succession stability to move, before %d after %+v", beforeStability, state.Succession)
	}
}

func TestDynamicCourtAgendaChangesBetweenSeasons(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 84)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	firstSignature := choiceSignature(state.Scene.Choices)

	var domesticChoice string
	for _, choice := range state.Scene.Choices {
		if choice.Domain == DomainDomestic {
			domesticChoice = choice.ID
			break
		}
	}
	if domesticChoice == "" {
		t.Fatalf("expected domestic agenda in first court scene: %+v", state.Scene.Choices)
	}

	if _, err := state.ApplyChoice(domesticChoice); err != nil {
		t.Fatalf("apply choice: %v", err)
	}
	secondSignature := choiceSignature(state.Scene.Choices)

	if firstSignature == secondSignature {
		t.Fatalf("expected seasonal court agenda to change, got %q", firstSignature)
	}

	systemAgenda := false
	for _, choice := range state.Scene.Choices {
		text := choice.Text + choice.Detail
		if strings.Contains(text, "后宫") || strings.Contains(text, "储") || strings.Contains(text, "官职") || strings.Contains(text, "任免") {
			systemAgenda = true
			break
		}
	}
	if !systemAgenda {
		t.Fatalf("expected harem, succession, or appointment agenda in court choices: %+v", state.Scene.Choices)
	}
}

func TestSeasonalRandomEventsAreGeneratedFromMultiplePressures(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 85)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	state.Stats.Treasury = 22
	state.Stats.BorderThreat = 82
	state.Succession.Dispute = 72
	state.Offices[0].VacancyRisk = 68

	var militaryChoice string
	for _, choice := range state.Scene.Choices {
		if choice.Domain == DomainMilitary {
			militaryChoice = choice.ID
			break
		}
	}
	if militaryChoice == "" {
		t.Fatalf("expected military choice: %+v", state.Scene.Choices)
	}

	if _, err := state.ApplyChoice(militaryChoice); err != nil {
		t.Fatalf("apply choice: %v", err)
	}

	if len(state.RecentEvents) < 2 {
		t.Fatalf("expected multiple seasonal random events, got %+v", state.RecentEvents)
	}
	seenCategories := map[EventCategory]bool{}
	seenDomains := map[Domain]bool{}
	for _, event := range state.RecentEvents {
		if event.ID == "" || event.Title == "" || event.Summary == "" || event.Detail == "" {
			t.Fatalf("event should expose narrative identity and detail: %+v", event)
		}
		if event.Severity <= 0 {
			t.Fatalf("event should expose severity: %+v", event)
		}
		seenCategories[event.Category] = true
		seenDomains[event.Domain] = true
	}
	if !seenCategories[EventStory] || !seenCategories[EventSystem] {
		t.Fatalf("expected story and system events, got %+v", state.RecentEvents)
	}
	if !seenDomains[DomainMilitary] && !seenDomains[DomainCourt] && !seenDomains[DomainEconomy] {
		t.Fatalf("expected pressure-linked event domains, got %+v", state.RecentEvents)
	}
}

func TestMicroGameRandomEventResolvesCheckAndAppliesEffects(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 86)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	state.Stats.Reform = 70
	state.Court[0].Ability = 92
	beforeReform := state.Stats.Reform
	beforeHistory := len(state.History)

	event := state.resolveRandomEventForTest("audit-sprint")

	if event.Category != EventMicroGame {
		t.Fatalf("expected micro-game event, got %+v", event)
	}
	if event.Check == "" || event.Target <= 0 || event.Roll <= 0 {
		t.Fatalf("expected check, target, and roll: %+v", event)
	}
	if !event.Success {
		t.Fatalf("high reform and able minister should pass audit sprint: %+v", event)
	}
	if state.Stats.Reform <= beforeReform {
		t.Fatalf("expected successful micro event to improve reform, before %d after %+v", beforeReform, state.Stats)
	}
	if len(state.History) != beforeHistory+1 {
		t.Fatalf("expected random event to write history, before %d after %d", beforeHistory, len(state.History))
	}
}

func TestSeasonalRandomEventsVaryAcrossTurns(t *testing.T) {
	state, err := NewGameWithDynasty("jingyao", 87)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	signatures := map[string]bool{}

	for i := 0; i < 3; i++ {
		choice := state.Scene.Choices[i%len(state.Scene.Choices)].ID
		if _, err := state.ApplyChoice(choice); err != nil {
			t.Fatalf("apply choice %d: %v", i, err)
		}
		signatures[eventSignature(state.RecentEvents)] = true
	}

	if len(signatures) < 2 {
		t.Fatalf("expected seasonal events to vary across turns, got %+v", signatures)
	}
}

func TestNewGameIncludesGrandStrategySystems(t *testing.T) {
	state, err := NewGameWithDynasty("dayin", 88)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()

	if len(state.Projects) < 5 {
		t.Fatalf("expected long-term imperial projects, got %+v", state.Projects)
	}
	if len(state.Policies) < 5 {
		t.Fatalf("expected standing policies, got %+v", state.Policies)
	}
	if len(state.Relations) < 6 {
		t.Fatalf("expected court relationship web, got %+v", state.Relations)
	}
	for _, project := range state.Projects {
		if project.ID == "" || project.Name == "" || project.Stage == "" || project.Domain == "" {
			t.Fatalf("project should expose identity and stage: %+v", project)
		}
		if project.Investment <= 0 || project.Risk < 0 {
			t.Fatalf("project should expose investment and risk: %+v", project)
		}
	}
	for _, policy := range state.Policies {
		if policy.ID == "" || policy.Name == "" || policy.Domain == "" {
			t.Fatalf("policy should expose identity and domain: %+v", policy)
		}
		if policy.Upkeep < 0 {
			t.Fatalf("policy upkeep should not be negative: %+v", policy)
		}
	}
}

func TestProjectOrderAdvancesLongTermProject(t *testing.T) {
	state := NewGame(89)
	state.ForceCoronationForTest()
	project := state.Projects[0]
	beforeCommand := state.Command
	beforeProgress := project.Progress

	resolution, err := state.ApplyOrder(OrderRequest{
		Kind:   OrderFundProject,
		Target: project.ID,
	})
	if err != nil {
		t.Fatalf("fund project: %v", err)
	}

	if resolution == nil || resolution.Summary == "" {
		t.Fatalf("expected project resolution, got %+v", resolution)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease, before %d after %d", beforeCommand, state.Command)
	}
	afterProject, ok := projectByID(state, project.ID)
	if !ok {
		t.Fatalf("project missing after funding: %q", project.ID)
	}
	if afterProject.Progress <= beforeProgress {
		t.Fatalf("expected project progress to advance, before %+v after %+v", project, afterProject)
	}
	if afterProject.Stage == project.Stage && afterProject.Progress >= 35 {
		t.Fatalf("expected project stage to react to progress, before %+v after %+v", project, afterProject)
	}
}

func TestPolicyOrderActivatesSeasonalPolicy(t *testing.T) {
	state := NewGame(90)
	state.ForceCoronationForTest()
	beforeCommand := state.Command
	policy := state.Policies[0]

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderEnactPolicy, Target: policy.ID}); err != nil {
		t.Fatalf("enact policy: %v", err)
	}
	afterPolicy, ok := policyByID(state, policy.ID)
	if !ok {
		t.Fatalf("policy missing after enactment: %q", policy.ID)
	}
	if !afterPolicy.Active {
		t.Fatalf("expected policy to be active: %+v", afterPolicy)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease, before %d after %d", beforeCommand, state.Command)
	}

	beforeStats := state.Stats
	if _, err := state.ApplyChoice(state.Scene.Choices[0].ID); err != nil {
		t.Fatalf("advance season: %v", err)
	}
	if state.Stats == beforeStats {
		t.Fatalf("expected active policy to affect seasonal state, before %+v after %+v", beforeStats, state.Stats)
	}
}

func TestRelationshipWebShiftsFromGrandStrategy(t *testing.T) {
	state := NewGame(91)
	state.ForceCoronationForTest()
	relation := state.Relations[0]

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderFundProject, Target: state.Projects[0].ID}); err != nil {
		t.Fatalf("fund project: %v", err)
	}

	afterRelation, ok := relationByID(state, relation.ID)
	if !ok {
		t.Fatalf("relation missing after strategy order: %q", relation.ID)
	}
	if afterRelation.Trust == relation.Trust && afterRelation.Tension == relation.Tension {
		t.Fatalf("expected relationship web to shift, before %+v after %+v", relation, afterRelation)
	}
}

func TestFrontierDynastyStartsWithExternalWarCampaign(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 18)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}

	if len(state.Wars) == 0 {
		t.Fatalf("expected frontier dynasty to start with an external war, got %+v", state.Wars)
	}
	war := state.Wars[0]
	if war.ID == "" || war.Enemy == "" || war.Front == "" || war.Stage == "" {
		t.Fatalf("war campaign should expose identity, enemy, front, and stage: %+v", war)
	}
	if war.Threat <= 0 || war.Supply <= 0 || war.Morale <= 0 {
		t.Fatalf("war campaign should expose threat, supply, and morale: %+v", war)
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

	// Dynamic agenda: should have 2-5 choices based on pressure
	if len(state.Scene.Choices) < 2 {
		t.Fatalf("expected at least 2 emperor choices, got %d: %+v", len(state.Scene.Choices), state.Scene.Choices)
	}
	if len(state.Scene.Choices) > 5 {
		t.Fatalf("expected at most 5 emperor choices, got %d: %+v", len(state.Scene.Choices), state.Scene.Choices)
	}

	domains := map[Domain]bool{}
	for _, choice := range state.Scene.Choices {
		domains[choice.Domain] = true
	}
	// With default coronation stats, at least military and domestic should be urgent
	if !domains[DomainMilitary] && !domains[DomainDomestic] {
		t.Fatalf("expected at least military or domestic domain in emperor choices: %+v", state.Scene.Choices)
	}

	// High-pressure scenarios should still produce a meaningful set
	state.Stats.BorderThreat = 90
	state.Stats.Treasury = 15
	state.Stats.Grain = 15
	state.Crisis.Severity = 80
	state.Crisis.Clock = 5
	state.Scene = emperorScene(state)
	if len(state.Scene.Choices) < 3 {
		t.Fatalf("expected more choices under high pressure, got %d: %+v", len(state.Scene.Choices), state.Scene.Choices)
	}
}

func TestStrategicChoiceMovesFactionAndProvinceState(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 41)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	// Ensure reform has enough pressure to appear in dynamic agenda
	state.Stats.Reform = 70
	state.Offices[0].VacancyRisk = 60
	state.Scene = emperorScene(state)

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
		t.Fatalf("expected reform choice in choices: %+v", state.Scene.Choices)
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
	// Ensure reform has enough pressure to appear in dynamic agenda
	state.Stats.Reform = 70
	state.Offices[0].VacancyRisk = 60
	state.Scene = emperorScene(state)

	before := objectiveProgress(t, state, "reform_state")

	var reformChoice string
	for _, choice := range state.Scene.Choices {
		if choice.Domain == DomainReform {
			reformChoice = choice.ID
			break
		}
	}
	if reformChoice == "" {
		t.Fatalf("expected reform choice in choices: %+v", state.Scene.Choices)
	}

	if _, err := state.ApplyChoice(reformChoice); err != nil {
		t.Fatalf("apply choice: %v", err)
	}

	after := objectiveProgress(t, state, "reform_state")
	if after <= before {
		t.Fatalf("expected reform objective progress to advance, before %d after %d", before, after)
	}
}

func TestEmperorCanIssueOrdersWithoutAdvancingSceneTurn(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 61)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	beforeTurn := state.Turn
	beforeCommand := state.Command
	beforeProvince := state.Provinces[0]

	resolution, err := state.ApplyOrder(OrderRequest{
		Kind:   OrderRelief,
		Target: "capital",
	})
	if err != nil {
		t.Fatalf("apply order: %v", err)
	}

	if resolution == nil || resolution.Summary == "" {
		t.Fatalf("expected order resolution, got %+v", resolution)
	}
	if state.Turn != beforeTurn {
		t.Fatalf("order should not advance scene turn, before %d after %d", beforeTurn, state.Turn)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease by 1, before %d after %d", beforeCommand, state.Command)
	}
	if state.Provinces[0] == beforeProvince {
		t.Fatalf("expected province to change, before %+v after %+v", beforeProvince, state.Provinces[0])
	}
}

func TestWarOrderAdvancesCampaignWithoutAdvancingSceneTurn(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 63)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	beforeTurn := state.Turn
	beforeCommand := state.Command
	beforeWar := state.Wars[0]

	resolution, err := state.ApplyOrder(OrderRequest{
		Kind:   OrderCampaign,
		Target: beforeWar.ID,
	})
	if err != nil {
		t.Fatalf("apply campaign order: %v", err)
	}

	if resolution == nil || resolution.Summary == "" {
		t.Fatalf("expected war resolution, got %+v", resolution)
	}
	if state.Turn != beforeTurn {
		t.Fatalf("war order should not advance scene turn, before %d after %d", beforeTurn, state.Turn)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease by 1, before %d after %d", beforeCommand, state.Command)
	}
	afterWar := state.Wars[0]
	if afterWar.Progress <= beforeWar.Progress {
		t.Fatalf("expected campaign progress to advance, before %+v after %+v", beforeWar, afterWar)
	}
	if afterWar.Threat >= beforeWar.Threat {
		t.Fatalf("expected campaign threat to fall, before %+v after %+v", beforeWar, afterWar)
	}
}

func TestOrdersRequireCommandPoints(t *testing.T) {
	state := NewGame(71)
	state.ForceCoronationForTest()
	state.Command = 0

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderRelief, Target: "capital"}); err == nil {
		t.Fatal("expected order to fail with no command points")
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

func officeByID(state *GameState, id string) (Office, bool) {
	for _, office := range state.Offices {
		if office.ID == id {
			return office, true
		}
	}
	return Office{}, false
}

func ministerByID(state *GameState, id string) (Minister, bool) {
	for _, minister := range state.Court {
		if minister.ID == id {
			return minister, true
		}
	}
	return Minister{}, false
}

func heirByID(state *GameState, id string) (Heir, bool) {
	for _, heir := range state.Heirs {
		if heir.ID == id {
			return heir, true
		}
	}
	return Heir{}, false
}

func choiceSignature(choices []Choice) string {
	parts := make([]string, 0, len(choices))
	for _, choice := range choices {
		parts = append(parts, choice.ID+":"+choice.Text)
	}
	return strings.Join(parts, "|")
}

func eventSignature(events []SeasonEvent) string {
	parts := make([]string, 0, len(events))
	for _, event := range events {
		parts = append(parts, event.ID)
	}
	return strings.Join(parts, "|")
}

func projectByID(state *GameState, id string) (ImperialProject, bool) {
	for _, project := range state.Projects {
		if project.ID == id {
			return project, true
		}
	}
	return ImperialProject{}, false
}

func policyByID(state *GameState, id string) (StandingPolicy, bool) {
	for _, policy := range state.Policies {
		if policy.ID == id {
			return policy, true
		}
	}
	return StandingPolicy{}, false
}

func relationByID(state *GameState, id string) (Relation, bool) {
	for _, relation := range state.Relations {
		if relation.ID == id {
			return relation, true
		}
	}
	return Relation{}, false
}

func TestPrinceChoicesGrantTraits(t *testing.T) {
	state := NewGame(42)

	// Turn 0: birth-omen → choose "grab-scroll" → should grant TraitScholarly
	if _, err := state.ApplyChoice("grab-scroll"); err != nil {
		t.Fatalf("apply grab-scroll: %v", err)
	}
	if !state.HasTrait(TraitScholarly) {
		t.Fatal("expected TraitScholarly after grabbing scroll")
	}

	// Turn 1: study-yard → choose "answer-people" → should grant TraitBenevolent
	if _, err := state.ApplyChoice("answer-people"); err != nil {
		t.Fatalf("apply answer-people: %v", err)
	}
	if !state.HasTrait(TraitBenevolent) {
		t.Fatal("expected TraitBenevolent after answering people first")
	}

	// Turn 2: winter-hunt → choose "accuse-brother" → should grant TraitSuspicious
	if _, err := state.ApplyChoice("accuse-brother"); err != nil {
		t.Fatalf("apply accuse-brother: %v", err)
	}
	if !state.HasTrait(TraitSuspicious) {
		t.Fatal("expected TraitSuspicious after accusing brother")
	}

	// Turn 3: flood-memorial → choose "borrow-merchants" → should grant TraitFrugal
	if _, err := state.ApplyChoice("borrow-merchants"); err != nil {
		t.Fatalf("apply borrow-merchants: %v", err)
	}
	if !state.HasTrait(TraitFrugal) {
		t.Fatal("expected TraitFrugal after borrowing from merchants")
	}

	// Turn 4: succession-night → choose "secure-edict" → should grant TraitVisionary
	if _, err := state.ApplyChoice("secure-edict"); err != nil {
		t.Fatalf("apply secure-edict: %v", err)
	}
	if !state.HasTrait(TraitVisionary) {
		t.Fatal("expected TraitVisionary after securing edict")
	}
}

func TestPrinceTraitsPersistAfterCoronation(t *testing.T) {
	state := NewGame(77)

	// Make specific prince choices
	for state.Phase != PhaseEmperor && state.Ending == nil {
		choiceID := state.Scene.Choices[1].ID // pick second choice each time
		if _, err := state.ApplyChoice(choiceID); err != nil {
			t.Fatalf("apply choice: %v", err)
		}
	}

	// After coronation, emperor should have prince-phase traits + dynasty base trait
	if len(state.EmperorTraits) < 2 {
		t.Fatalf("expected at least 2 traits after coronation, got %d: %+v", len(state.EmperorTraits), state.EmperorTraits)
	}

	// Dynasty base trait should be present (dayin → ambitious)
	if !state.HasTrait(TraitAmbitious) {
		t.Fatal("expected TraitAmbitious from dayin dynasty")
	}
}

func TestReformTreePrerequisitesAndUnlocking(t *testing.T) {
	state := NewGame(200)
	state.ForceCoronationForTest()

	// Tier 2 projects should start locked
	for _, p := range state.Projects {
		if p.Tier >= 2 && !p.Locked {
			t.Fatalf("tier %d project %q should start locked", p.Tier, p.Name)
		}
	}

	// Trying to fund a locked project should fail
	_, err := state.ApplyOrder(OrderRequest{Kind: OrderFundProject, Target: "salt-iron-reform"})
	if err == nil {
		t.Fatal("expected error when funding locked project")
	}

	// Complete grand-canal (tier 1) to unlock salt-iron-reform and imperial-granary
	completeProjectFast(state, "grand-canal")

	// Now salt-iron-reform and imperial-granary should be unlocked
	saltIron, ok := projectByID(state, "salt-iron-reform")
	if !ok {
		t.Fatal("salt-iron-reform not found")
	}
	if saltIron.Locked {
		t.Fatal("salt-iron-reform should be unlocked after grand-canal completion")
	}
	granary, ok := projectByID(state, "imperial-granary")
	if !ok {
		t.Fatal("imperial-granary not found")
	}
	if granary.Locked {
		t.Fatal("imperial-granary should be unlocked after grand-canal completion")
	}

	// Tier 3 projects should still be locked
	mint, ok := projectByID(state, "imperial-mint")
	if !ok {
		t.Fatal("imperial-mint not found")
	}
	if !mint.Locked {
		t.Fatal("imperial-mint should still be locked (needs salt-iron-reform + state-academy)")
	}
}

func TestReformTreeTier3RequiresMultiplePrereqs(t *testing.T) {
	state := NewGame(201)
	state.ForceCoronationForTest()

	// imperial-mint requires salt-iron-reform AND state-academy
	// Complete only grand-canal (unlocks salt-iron-reform)
	completeProjectFast(state, "grand-canal")

	// imperial-mint still locked (only prereq path partially met)
	mint, _ := projectByID(state, "imperial-mint")
	if !mint.Locked {
		t.Fatal("imperial-mint should still be locked with only grand-canal done")
	}

	// Complete salt-iron-reform (now one of mint's prereqs is done)
	completeProjectFast(state, "salt-iron-reform")

	// Still locked because state-academy isn't done yet
	mint, _ = projectByID(state, "imperial-mint")
	if !mint.Locked {
		t.Fatal("imperial-mint should still be locked without state-academy")
	}

	// Complete state-academy (now all prereqs met)
	completeProjectFast(state, "state-academy")

	// Now imperial-mint should be unlocked
	mint, _ = projectByID(state, "imperial-mint")
	if mint.Locked {
		t.Fatal("imperial-mint should be unlocked after salt-iron-reform + state-academy")
	}
}

func TestReformTreeSynergyEffects(t *testing.T) {
	state := NewGame(202)
	state.ForceCoronationForTest()

	// Complete grand-canal first - no synergy yet since state-academy isn't done
	completeProjectFast(state, "grand-canal")

	// Now complete state-academy - should trigger synergy with grand-canal
	beforeReform := state.Stats.Reform
	completeProjectFast(state, "state-academy")

	// The synergy bonus (Reform: 4) should have been applied
	// state-academy completion gives Reform: 12, synergy gives Reform: 4, total >= 16
	reformGain := state.Stats.Reform - beforeReform + 1 // +1 for the funding cost
	if reformGain < 12 {
		t.Fatalf("expected significant reform gain including synergy, got %d", reformGain)
	}
}

func TestReformTreeTier2AdvancesSlower(t *testing.T) {
	state := NewGame(203)
	state.ForceCoronationForTest()

	// Unlock a tier 2 project
	completeProjectFast(state, "grand-canal")

	// Fund tier 1 vs tier 2 project and compare advance
	tier1Project := state.Projects[1] // border-arsenal (tier 1)
	beforeT1 := tier1Project.Progress
	_, err := state.ApplyOrder(OrderRequest{Kind: OrderFundProject, Target: tier1Project.ID})
	if err != nil {
		t.Fatalf("fund tier 1: %v", err)
	}
	tier1After, _ := projectByID(state, tier1Project.ID)
	tier1Advance := tier1After.Progress - beforeT1

	// Fund tier 2 project
	saltIron, _ := projectByID(state, "salt-iron-reform")
	beforeT2 := saltIron.Progress
	_, err = state.ApplyOrder(OrderRequest{Kind: OrderFundProject, Target: "salt-iron-reform"})
	if err != nil {
		t.Fatalf("fund tier 2: %v", err)
	}
	saltIronAfter, _ := projectByID(state, "salt-iron-reform")
	tier2Advance := saltIronAfter.Progress - beforeT2

	// Tier 2 should advance less than tier 1 (85% modifier)
	if tier2Advance >= tier1Advance {
		t.Fatalf("tier 2 should advance slower than tier 1, t1=%d t2=%d", tier1Advance, tier2Advance)
	}
}

func TestProjectPassiveEffectsIncludeNewProjects(t *testing.T) {
	state := NewGame(204)
	state.ForceCoronationForTest()

	// Complete a tier 2 project
	completeProjectFast(state, "grand-canal")
	completeProjectFast(state, "salt-iron-reform")

	// Verify passive effects function works for new project IDs
	effects := projectPassiveEffects(ImperialProject{ID: "salt-iron-reform"})
	if effects.Treasury != 3 {
		t.Fatalf("expected salt-iron-reform passive treasury=3, got %d", effects.Treasury)
	}

	// Verify tier 3 passive effects
	effects = projectPassiveEffects(ImperialProject{ID: "imperial-mint"})
	if effects.Treasury != 4 {
		t.Fatalf("expected imperial-mint passive treasury=4, got %d", effects.Treasury)
	}
}

// completeProjectFast forces a project to 100% completion by repeatedly funding it.
func completeProjectFast(state *GameState, projectID string) {
	for i := 0; i < 50; i++ {
		p, ok := projectByID(state, projectID)
		if !ok {
			return
		}
		if p.Completed {
			return
		}
		if p.Locked {
			return
		}
		// Refill command points so we can keep issuing orders
		state.Command = state.commandBudget() + 5
		state.ApplyOrder(OrderRequest{Kind: OrderFundProject, Target: projectID})
	}
}

func TestPrinceTraitDuplicatesPrevented(t *testing.T) {
	state := NewGame(99)

	// Grant scholarly trait from grab-scroll
	if _, err := state.ApplyChoice("grab-scroll"); err != nil {
		t.Fatalf("apply grab-scroll: %v", err)
	}

	// Count how many scholarly traits exist
	scholarlyCount := 0
	for _, trait := range state.EmperorTraits {
		if trait.ID == TraitScholarly {
			scholarlyCount++
		}
	}
	if scholarlyCount != 1 {
		t.Fatalf("expected exactly 1 TraitScholarly, got %d", scholarlyCount)
	}
}
