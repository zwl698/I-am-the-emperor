package game

import "testing"

func TestResolveBattleWonMatchesLegacyModel(t *testing.T) {
	cases := []struct {
		name                                           string
		attackArms, defendArms, attackFood, defendFood int
		randv                                          int
		want                                           bool
	}{
		{"attacker double, roll 69 wins", 1000, 400, 0, 0, 69, true},
		{"attacker double, roll 70 loses", 1000, 400, 0, 0, 70, false},
		{"attacker more + more food, roll 59 wins", 600, 500, 100, 50, 59, true},
		{"attacker more + more food, roll 60 loses", 600, 500, 100, 50, 60, false},
		{"attacker more, less food, roll 39 wins", 600, 500, 10, 50, 39, true},
		{"attacker more, less food, roll 40 loses", 600, 500, 10, 50, 40, false},
		{"attacker far weaker, roll 1 wins", 100, 500, 0, 0, 1, true},
		{"attacker far weaker, roll 2 loses", 100, 500, 0, 0, 2, false},
		{"attacker weaker + more food, roll 29 wins", 400, 500, 100, 0, 29, true},
		{"attacker weaker + more food, roll 30 loses", 400, 500, 100, 0, 30, false},
		{"attacker weaker, less food, roll 9 wins", 400, 500, 0, 100, 9, true},
		{"attacker weaker, less food, roll 10 loses", 400, 500, 0, 100, 10, false},
		{"attacker zero arms always loses", 0, 500, 0, 0, 0, false},
		{"defender zero arms always wins", 500, 0, 0, 0, 99, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveBattleWon(tc.attackArms, tc.defendArms, tc.attackFood, tc.defendFood, tc.randv)
			if got != tc.want {
				t.Errorf("resolveBattleWon(%d,%d,%d,%d,%d) = %v, want %v",
					tc.attackArms, tc.defendArms, tc.attackFood, tc.defendFood, tc.randv, got, tc.want)
			}
		})
	}
}

// newBattleTestState builds a tiny 2-city scenario for player "p1" vs "p2".
func newBattleTestState() *GameState {
	return &GameState{
		ScenarioID: "test",
		PlayerID:   "p1",
		Date:       Date{Year: 190, Month: 1},
		Rulers: []Ruler{
			{ID: "p1", Name: "甲"},
			{ID: "p2", Name: "乙"},
			{ID: "neutral", Name: "空城"},
		},
		Cities: []City{
			{ID: "c1", Name: "我城", OwnerID: "p1", Food: 500, PeopleDevotion: 50},
			{ID: "c2", Name: "敌城", OwnerID: "p2", Food: 100, PeopleDevotion: 50, Garrison: 200},
			{ID: "c3", Name: "远城", OwnerID: "p2", Food: 100},
		},
		Generals: []General{
			{ID: "g1", Name: "我将", OwnerID: "p1", CityID: "c1", Force: 80, Stamina: 100, Soldiers: 2000},
			{ID: "g2", Name: "敌将", OwnerID: "p2", CityID: "c2", Force: 60, Stamina: 100, Soldiers: 300},
		},
		Routes: []Route{{From: "c1", To: "c2"}},
	}
}

func TestApplyBattleCaptureOnWin(t *testing.T) {
	s := newBattleTestState()
	battleRand = func(int) int { return 0 } // force a win roll
	defer func() { battleRand = defaultBattleRand }()

	outcome, err := s.ApplyBattle("c1", "g1", "c2")
	if err != nil {
		t.Fatalf("ApplyBattle error = %v", err)
	}
	if !outcome.Won || !outcome.Captured {
		t.Fatalf("expected win+capture, got %+v", outcome)
	}
	if outcome.FromCityName != "我城" || outcome.TargetCityName != "敌城" || outcome.GeneralName != "我将" {
		t.Fatalf("battle labels = from %q target %q general %q", outcome.FromCityName, outcome.TargetCityName, outcome.GeneralName)
	}
	if outcome.AttackerRulerName != "甲" || outcome.DefenderRulerName != "乙" {
		t.Fatalf("ruler labels = attacker %q defender %q", outcome.AttackerRulerName, outcome.DefenderRulerName)
	}
	if len(outcome.DefenderGenerals) != 1 || outcome.DefenderGenerals[0] != "敌将" {
		t.Fatalf("defender generals = %v, want [敌将]", outcome.DefenderGenerals)
	}
	if len(outcome.CapturedGenerals) != 1 || outcome.CapturedGenerals[0] != "敌将" {
		t.Fatalf("captured generals = %v, want [敌将]", outcome.CapturedGenerals)
	}
	if outcome.AttackPower <= 0 || outcome.DefensePower <= 0 {
		t.Fatalf("battle power not populated: attack=%d defense=%d", outcome.AttackPower, outcome.DefensePower)
	}
	if outcome.Message == "" {
		t.Fatal("expected detailed battle message")
	}
	if got := s.findCity("c2").OwnerID; got != "p1" {
		t.Errorf("city c2 owner = %q, want p1", got)
	}
	if g := s.findGeneral("g1"); g.CityID != "c2" {
		t.Errorf("general g1 city = %q, want c2 (moved in)", g.CityID)
	}
	if g := s.findGeneral("g2"); g.Soldiers != 0 || g.OwnerID != "p1" || !g.Captive {
		t.Errorf("defender g2 = %+v, want captured by p1 with 0 soldiers", g)
	}
	if g := s.findGeneral("g1"); g.Stamina != 96 {
		t.Errorf("attacker stamina = %d, want 96", g.Stamina)
	}
}

func TestApplyBattleRepelledOnLoss(t *testing.T) {
	s := newBattleTestState()
	// Make attacker far weaker than defender so loss is near-certain, and roll high.
	s.findGeneral("g1").Soldiers = 100
	s.findCity("c2").Garrison = 5000
	battleRand = func(int) int { return 100 }
	defer func() { battleRand = defaultBattleRand }()

	outcome, err := s.ApplyBattle("c1", "g1", "c2")
	if err != nil {
		t.Fatalf("ApplyBattle error = %v", err)
	}
	if outcome.Won || outcome.Captured {
		t.Fatalf("expected loss, got %+v", outcome)
	}
	if got := s.findCity("c2").OwnerID; got != "p2" {
		t.Errorf("city c2 owner = %q, want p2 (unchanged)", got)
	}
}

func TestApplyBattleRejectsNonAdjacent(t *testing.T) {
	s := newBattleTestState()
	if _, err := s.ApplyBattle("c1", "g1", "c3"); err == nil {
		t.Fatal("expected non-adjacent error, got nil")
	}
}

func TestApplyBattleRejectsOwnCity(t *testing.T) {
	s := newBattleTestState()
	s.Routes = append(s.Routes, Route{From: "c1", To: "c1"})
	if _, err := s.ApplyBattle("c1", "g1", "c1"); err == nil {
		t.Fatal("expected same-owner error, got nil")
	}
}

func TestApplyBattleRejectsExhaustedGeneral(t *testing.T) {
	s := newBattleTestState()
	s.findGeneral("g1").Stamina = 3
	if _, err := s.ApplyBattle("c1", "g1", "c2"); err == nil {
		t.Fatal("expected stamina error, got nil")
	}
}

func TestAdjacentCityIDs(t *testing.T) {
	s := newBattleTestState()
	adj := s.AdjacentCityIDs("c1")
	if len(adj) != 1 || adj[0] != "c2" {
		t.Errorf("AdjacentCityIDs(c1) = %v, want [c2]", adj)
	}
}
