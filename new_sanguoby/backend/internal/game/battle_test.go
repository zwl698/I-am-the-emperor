package game

import "testing"

func TestResolveBattleWonMatchesLegacyModel(t *testing.T) {
	cases := []struct {
		name                                           string
		attackArms, defendArms, attackFood, defendFood int
		randv                                          int
		want                                           bool
	}{
		{"attacker double, roll 29 loses", 1000, 400, 0, 0, 29, false},
		{"attacker double, roll 30 wins", 1000, 400, 0, 0, 30, true},
		{"attacker more + more food, roll 39 loses", 600, 500, 100, 50, 39, false},
		{"attacker more + more food, roll 40 wins", 600, 500, 100, 50, 40, true},
		{"attacker more, less food, roll 59 loses", 600, 500, 10, 50, 59, false},
		{"attacker more, less food, roll 60 wins", 600, 500, 10, 50, 60, true},
		{"attacker far weaker, roll 2 wins", 100, 500, 0, 0, 2, true},
		{"attacker far weaker, roll 3 loses", 100, 500, 0, 0, 3, false},
		{"attacker weaker + more food, roll 30 wins", 400, 500, 100, 0, 30, true},
		{"attacker weaker + more food, roll 31 loses", 400, 500, 100, 0, 31, false},
		{"attacker weaker, less food, roll 10 wins", 400, 500, 0, 100, 10, true},
		{"attacker weaker, less food, roll 11 loses", 400, 500, 0, 100, 11, false},
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
	battleRand = func(int) int { return 100 } // force a strong-side win roll
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

func TestApplyBattlePlanConsumesSuppliesAndMovesSelectedGenerals(t *testing.T) {
	s := newBattleTestState()
	s.findCity("c1").Money = 300
	s.findCity("c2").Farming = 100
	s.findCity("c2").Commerce = 100
	s.findCity("c2").Money = 100
	s.Generals = append(s.Generals, General{ID: "g3", Name: "副将", OwnerID: "p1", CityID: "c1", Force: 70, Stamina: 100, Soldiers: 1000})
	battleRand = func(int) int { return 100 }
	defer func() { battleRand = defaultBattleRand }()

	outcome, err := s.ApplyBattlePlan("c1", []string{"g1", "g3"}, "c2", 80, 120, 90, 20)
	if err != nil {
		t.Fatalf("ApplyBattlePlan error = %v", err)
	}
	if !outcome.Won || !outcome.Captured {
		t.Fatalf("expected planned win+capture, got %+v", outcome)
	}
	if outcome.Money != 80 || outcome.Food != 120 || outcome.RemainingFood != 90 || outcome.FieldAdvantage != 20 {
		t.Fatalf("planned resources not reflected: %+v", outcome)
	}
	if len(outcome.GeneralNames) != 2 || outcome.GeneralNames[0] != "我将" || outcome.GeneralNames[1] != "副将" {
		t.Fatalf("general names = %v, want [我将 副将]", outcome.GeneralNames)
	}
	if city := s.findCity("c1"); city.Money != 220 || city.Food != 380 {
		t.Fatalf("origin supplies = money %d food %d, want 220/380", city.Money, city.Food)
	}
	if city := s.findCity("c2"); city.Food != 190 || city.Farming != 95 || city.Commerce != 95 || city.Money != 95 || city.PeopleDevotion != 45 {
		t.Fatalf("battlefield aftermath = food %d farming %d commerce %d money %d devotion %d, want 190/95/95/95/45",
			city.Food, city.Farming, city.Commerce, city.Money, city.PeopleDevotion)
	}
	for _, id := range []string{"g1", "g3"} {
		g := s.findGeneral(id)
		if g.CityID != "c2" {
			t.Errorf("general %s city = %q, want c2", id, g.CityID)
		}
		if g.Stamina != 96 {
			t.Errorf("general %s stamina = %d, want 96", id, g.Stamina)
		}
	}
}

func TestApplyBattlePlanRejectsMissingFood(t *testing.T) {
	s := newBattleTestState()
	s.findCity("c1").Money = 300
	if _, err := s.ApplyBattlePlan("c1", []string{"g1"}, "c2", 10, 0, 0, 0); err == nil {
		t.Fatal("expected no-food error, got nil")
	}
}

func TestApplyBattlePlanOccupiesEmptyCityWithoutCombatLoss(t *testing.T) {
	s := newBattleTestState()
	s.findCity("c1").Money = 300
	s.findCity("c3").OwnerID = "neutral"
	s.findCity("c3").PeopleDevotion = 30
	s.Routes = append(s.Routes, Route{From: "c1", To: "c3"})
	beforeSoldiers := s.findGeneral("g1").Soldiers

	outcome, err := s.ApplyBattlePlan("c1", []string{"g1"}, "c3", 0, 80, 80, 10)
	if err != nil {
		t.Fatalf("ApplyBattlePlan empty city error = %v", err)
	}
	if !outcome.Won || !outcome.Captured || outcome.AttackerLosses != 0 || outcome.DefenderLosses != 0 {
		t.Fatalf("expected peaceful empty-city occupation, got %+v", outcome)
	}
	if city := s.findCity("c3"); city.OwnerID != "p1" || city.Food != 100 || city.PeopleDevotion != 30 {
		t.Fatalf("empty city = owner %q food %d devotion %d, want p1/100/30", city.OwnerID, city.Food, city.PeopleDevotion)
	}
	if g := s.findGeneral("g1"); g.CityID != "c3" || g.Soldiers != beforeSoldiers || g.Stamina != 100 {
		t.Fatalf("attacker after empty occupation = %+v, want moved without loss/stamina cost", g)
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
