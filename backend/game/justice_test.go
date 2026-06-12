package game

import "testing"

func TestNewGameIncludesJusticeAndPublicOpinion(t *testing.T) {
	state := NewGame(1201)
	state.ForceCoronationForTest()

	if len(state.LegalCases) < 3 {
		t.Fatalf("expected legal cases, got %+v", state.LegalCases)
	}
	if state.PublicOpinion.Popular <= 0 || state.PublicOpinion.Elite <= 0 || state.PublicOpinion.Justice <= 0 {
		t.Fatalf("expected populated public opinion, got %+v", state.PublicOpinion)
	}
	for _, c := range state.LegalCases {
		if c.ID == "" || c.Title == "" || c.Defendant == "" || c.Charge == "" {
			t.Fatalf("legal case should expose identity and charge: %+v", c)
		}
		if c.Heat <= 0 || c.Evidence <= 0 || c.FactionPressure <= 0 {
			t.Fatalf("legal case should expose pressure values: %+v", c)
		}
	}
}

func TestOpenTrialResolvesCaseAndChangesOpinion(t *testing.T) {
	state := NewGame(1202)
	state.ForceCoronationForTest()
	c := state.LegalCases[0]
	beforeOpinion := state.PublicOpinion
	beforeCommand := state.Command

	resolution, err := state.ApplyOrder(OrderRequest{Kind: OrderOpenTrial, Target: c.ID})
	if err != nil {
		t.Fatalf("open trial: %v", err)
	}

	afterCase, ok := legalCaseByID(state, c.ID)
	if !ok {
		t.Fatalf("case missing after trial: %q", c.ID)
	}
	if !afterCase.Resolved {
		t.Fatalf("expected case to be resolved: %+v", afterCase)
	}
	if afterCase.Verdict == "" {
		t.Fatalf("expected verdict text: %+v", afterCase)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease, before %d after %d", beforeCommand, state.Command)
	}
	if state.PublicOpinion.Justice <= beforeOpinion.Justice {
		t.Fatalf("expected justice sentiment to improve, before %+v after %+v", beforeOpinion, state.PublicOpinion)
	}
	if state.PublicOpinion.Rumor >= beforeOpinion.Rumor {
		t.Fatalf("expected rumor to cool, before %+v after %+v", beforeOpinion, state.PublicOpinion)
	}
	if resolution.Summary == "" {
		t.Fatalf("expected summary")
	}
}

func TestCensorRumorLowersRumorButRaisesFear(t *testing.T) {
	state := NewGame(1203)
	state.ForceCoronationForTest()
	state.PublicOpinion.Rumor = 82
	state.PublicOpinion.Fear = 18
	beforeOpinion := state.PublicOpinion
	beforeLegitimacy := state.Stats.Legitimacy

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderCensorRumor, Target: "public"}); err != nil {
		t.Fatalf("censor rumor: %v", err)
	}

	if state.PublicOpinion.Rumor >= beforeOpinion.Rumor {
		t.Fatalf("expected rumor to fall, before %+v after %+v", beforeOpinion, state.PublicOpinion)
	}
	if state.PublicOpinion.Fear <= beforeOpinion.Fear {
		t.Fatalf("expected fear to rise, before %+v after %+v", beforeOpinion, state.PublicOpinion)
	}
	if state.Stats.Legitimacy >= beforeLegitimacy {
		t.Fatalf("expected legitimacy cost, before %d after %d", beforeLegitimacy, state.Stats.Legitimacy)
	}
}

func TestExposedPlotCreatesLinkedLegalCase(t *testing.T) {
	state := NewGame(1204)
	state.ForceCoronationForTest()
	plot := state.Plots[0]
	beforeCases := len(state.LegalCases)

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderInvestigatePlot, Target: plot.ID}); err != nil {
		t.Fatalf("investigate plot: %v", err)
	}

	if len(state.LegalCases) <= beforeCases {
		t.Fatalf("expected linked case after exposing plot, before %d after %+v", beforeCases, state.LegalCases)
	}
	linked, ok := legalCaseBySource(state, "plot:"+plot.ID)
	if !ok {
		t.Fatalf("expected legal case linked to plot %q, cases %+v", plot.ID, state.LegalCases)
	}
	if linked.Defendant != plot.Sponsor {
		t.Fatalf("expected sponsor as defendant, plot %+v case %+v", plot, linked)
	}
}

func TestJusticePressureEscalatesOpenCases(t *testing.T) {
	state := NewGame(1205)
	state.ForceCoronationForTest()
	state.LegalCases[0].Heat = 92
	state.LegalCases[0].Resolved = false
	beforeOpinion := state.PublicOpinion
	beforeSeverity := state.Crisis.Severity

	state.applyJusticePressure(DomainIntrigue)

	afterCase := state.LegalCases[0]
	if afterCase.Heat <= 92 {
		t.Fatalf("expected heat to grow under pressure, got %+v", afterCase)
	}
	if state.PublicOpinion.Rumor <= beforeOpinion.Rumor {
		t.Fatalf("expected rumor pressure to grow, before %+v after %+v", beforeOpinion, state.PublicOpinion)
	}
	if state.Crisis.Severity <= beforeSeverity {
		t.Fatalf("expected crisis severity to rise, before %d after %d", beforeSeverity, state.Crisis.Severity)
	}
}

func TestJusticeRandomEventsUseOpinionAndCasePressure(t *testing.T) {
	state := NewGame(1206)
	state.ForceCoronationForTest()
	state.PublicOpinion.Rumor = 84
	state.LegalCases[0].Heat = 88
	beforeRumor := state.PublicOpinion.Rumor
	beforeHeat := state.LegalCases[0].Heat

	rumorEvent := state.resolveRandomEventForTest("capital-tabloid")
	if rumorEvent.ID != "capital-tabloid" || !hasTag(rumorEvent.Tags, "舆论") {
		t.Fatalf("expected opinion event, got %+v", rumorEvent)
	}
	if state.PublicOpinion.Rumor <= beforeRumor {
		t.Fatalf("expected event to feed rumor pressure, before %d after %+v", beforeRumor, state.PublicOpinion)
	}

	caseEvent := state.resolveRandomEventForTest("ministry-case-deadline")
	if caseEvent.ID != "ministry-case-deadline" || !hasTag(caseEvent.Tags, "刑狱") {
		t.Fatalf("expected case deadline event, got %+v", caseEvent)
	}
	if state.LegalCases[0].Heat <= beforeHeat {
		t.Fatalf("expected event to heat open cases, before %d after %+v", beforeHeat, state.LegalCases[0])
	}
}

func legalCaseByID(state *GameState, id string) (LegalCase, bool) {
	for _, c := range state.LegalCases {
		if c.ID == id {
			return c, true
		}
	}
	return LegalCase{}, false
}

func legalCaseBySource(state *GameState, source string) (LegalCase, bool) {
	for _, c := range state.LegalCases {
		if c.Source == source {
			return c, true
		}
	}
	return LegalCase{}, false
}

func hasTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}
