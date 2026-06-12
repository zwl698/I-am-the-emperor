package game

import "testing"

func TestNewGameIncludesForeignStatesAndPlots(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 92)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()

	if len(state.ForeignStates) < 4 {
		t.Fatalf("expected foreign states, got %+v", state.ForeignStates)
	}
	if len(state.Plots) < 3 {
		t.Fatalf("expected active intrigue plots, got %+v", state.Plots)
	}
	for _, foreign := range state.ForeignStates {
		if foreign.ID == "" || foreign.Name == "" || foreign.Ruler == "" || foreign.Attitude == "" {
			t.Fatalf("foreign state should expose identity and attitude: %+v", foreign)
		}
		if foreign.Relation <= 0 || foreign.Threat < 0 {
			t.Fatalf("foreign state should expose relation and threat: %+v", foreign)
		}
	}
	for _, plot := range state.Plots {
		if plot.ID == "" || plot.Title == "" || plot.Sponsor == "" || plot.Target == "" {
			t.Fatalf("plot should expose identity, sponsor and target: %+v", plot)
		}
		if plot.Progress <= 0 || plot.Secrecy <= 0 || plot.Danger <= 0 {
			t.Fatalf("plot should expose progress, secrecy and danger: %+v", plot)
		}
	}
}

func TestEmbassyOrderImprovesForeignRelation(t *testing.T) {
	state := NewGame(93)
	state.ForceCoronationForTest()
	foreign := state.ForeignStates[0]
	beforeCommand := state.Command

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderEmbassy, Target: foreign.ID}); err != nil {
		t.Fatalf("embassy order: %v", err)
	}

	afterForeign, ok := foreignByID(state, foreign.ID)
	if !ok {
		t.Fatalf("foreign state missing after embassy: %q", foreign.ID)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease, before %d after %d", beforeCommand, state.Command)
	}
	if afterForeign.Relation <= foreign.Relation {
		t.Fatalf("expected relation to improve, before %+v after %+v", foreign, afterForeign)
	}
	if afterForeign.Threat >= foreign.Threat {
		t.Fatalf("expected threat to fall, before %+v after %+v", foreign, afterForeign)
	}
}

func TestTreatyOrderCreatesLastingForeignPact(t *testing.T) {
	state := NewGame(94)
	state.ForceCoronationForTest()
	state.ForeignStates[0].Relation = 68
	foreign := state.ForeignStates[0]

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderTreaty, Target: foreign.ID}); err != nil {
		t.Fatalf("treaty order: %v", err)
	}

	afterForeign, ok := foreignByID(state, foreign.ID)
	if !ok {
		t.Fatalf("foreign state missing after treaty: %q", foreign.ID)
	}
	if afterForeign.Treaty == "" {
		t.Fatalf("expected treaty after diplomatic pact: %+v", afterForeign)
	}
	if afterForeign.Leverage <= foreign.Leverage {
		t.Fatalf("expected leverage to improve, before %+v after %+v", foreign, afterForeign)
	}
}

func TestIntrigueOrdersRevealAndResolvePlot(t *testing.T) {
	state := NewGame(95)
	state.ForceCoronationForTest()
	plot := state.Plots[0]

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderInvestigatePlot, Target: plot.ID}); err != nil {
		t.Fatalf("investigate plot: %v", err)
	}
	afterInvestigate, ok := plotByID(state, plot.ID)
	if !ok {
		t.Fatalf("plot missing after investigation: %q", plot.ID)
	}
	if !afterInvestigate.Exposed {
		t.Fatalf("expected plot to be exposed, before %+v after %+v", plot, afterInvestigate)
	}
	if afterInvestigate.Secrecy >= plot.Secrecy {
		t.Fatalf("expected secrecy to fall, before %+v after %+v", plot, afterInvestigate)
	}

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderSuppressPlot, Target: plot.ID}); err != nil {
		t.Fatalf("suppress plot: %v", err)
	}
	afterSuppress, ok := plotByID(state, plot.ID)
	if !ok {
		t.Fatalf("plot missing after suppression: %q", plot.ID)
	}
	if !afterSuppress.Resolved {
		t.Fatalf("expected plot to be resolved: %+v", afterSuppress)
	}
}

func TestHeirEducationOrderImprovesHeir(t *testing.T) {
	state := NewGame(96)
	state.ForceCoronationForTest()
	heir := state.Heirs[0]
	beforeCommand := state.Command

	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderEducateHeir, Target: heir.ID + ":study"}); err != nil {
		t.Fatalf("educate heir: %v", err)
	}

	afterHeir, ok := heirByID(state, heir.ID)
	if !ok {
		t.Fatalf("heir missing after education: %q", heir.ID)
	}
	if state.Command != beforeCommand-1 {
		t.Fatalf("expected command to decrease, before %d after %d", beforeCommand, state.Command)
	}
	if afterHeir.Talent <= heir.Talent {
		t.Fatalf("expected heir talent to improve, before %+v after %+v", heir, afterHeir)
	}
	if state.Succession.Stability <= 0 {
		t.Fatalf("expected succession to remain valid: %+v", state.Succession)
	}
}

func foreignByID(state *GameState, id string) (ForeignState, bool) {
	for _, foreign := range state.ForeignStates {
		if foreign.ID == id {
			return foreign, true
		}
	}
	return ForeignState{}, false
}

func plotByID(state *GameState, id string) (Plot, bool) {
	for _, plot := range state.Plots {
		if plot.ID == id {
			return plot, true
		}
	}
	return Plot{}, false
}
