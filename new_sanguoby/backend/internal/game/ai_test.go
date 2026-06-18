package game

import "testing"

// newAITestState builds a 3-city line: player p1 (c1) — enemy p2 (c2) — neutral (c3).
func newAITestState() *GameState {
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
			{ID: "c1", Name: "甲城", OwnerID: "p1", Food: 200, PeopleDevotion: 50, Garrison: 100,
				Farming: 100, FarmingLimit: 1000, Commerce: 100, CommerceLimit: 1000, Population: 50000, PopulationLimit: 80000},
			{ID: "c2", Name: "乙城", OwnerID: "p2", Food: 500, PeopleDevotion: 50,
				Farming: 100, FarmingLimit: 1000, Commerce: 100, CommerceLimit: 1000, Population: 50000, PopulationLimit: 80000},
			{ID: "c3", Name: "空城", OwnerID: "neutral", Food: 100, PeopleDevotion: 30},
		},
		Generals: []General{
			{ID: "g1", Name: "甲将", OwnerID: "p1", CityID: "c1", Force: 70, Stamina: 100, Soldiers: 300},
			{ID: "g2", Name: "乙将", OwnerID: "p2", CityID: "c2", Force: 80, Stamina: 100, Soldiers: 5000},
		},
		Routes: []Route{{From: "c1", To: "c2"}, {From: "c2", To: "c3"}},
	}
}

func TestRunEnemyTurnsAttacksWeakerNeighbour(t *testing.T) {
	s := newAITestState()
	// Remove the neutral city so the only target is the weak player city c1.
	s.Cities = s.Cities[:2]
	s.Routes = []Route{{From: "c1", To: "c2"}}
	battleRand = func(int) int { return 0 } // force AI win
	defer func() { battleRand = defaultBattleRand }()

	captures := s.RunEnemyTurns()
	if captures == 0 {
		t.Fatal("expected AI to capture at least one city")
	}
	// p2's strong general should have taken the weak player city c1.
	if got := s.findCity("c1").OwnerID; got != "p2" {
		t.Errorf("c1 owner = %q, want p2 (captured by AI)", got)
	}
}

func TestRunEnemyTurnsDevelopsWhenNoTarget(t *testing.T) {
	s := newAITestState()
	// Only the player city neighbours the AI, and it is too strong to attack.
	s.Cities = s.Cities[:2]
	s.Routes = []Route{{From: "c1", To: "c2"}}
	s.findGeneral("g2").Soldiers = 10
	s.findCity("c1").Garrison = 100000
	beforeFarming := s.findCity("c2").Farming

	captures := s.RunEnemyTurns()
	if captures != 0 {
		t.Fatalf("expected no captures, got %d", captures)
	}
	if s.findCity("c2").Farming <= beforeFarming {
		t.Errorf("AI should have developed c2 farming: before=%d after=%d", beforeFarming, s.findCity("c2").Farming)
	}
}

func TestRunEnemyTurnsSkipsPlayerAndNeutral(t *testing.T) {
	s := newAITestState()
	battleRand = func(int) int { return 0 }
	defer func() { battleRand = defaultBattleRand }()

	playerGeneralStaminaBefore := s.findGeneral("g1").Stamina
	s.RunEnemyTurns()
	// Player's general must not have acted (stamina unchanged by AI).
	if s.findGeneral("g1").Stamina != playerGeneralStaminaBefore {
		t.Errorf("player general acted during enemy turn: stamina %d -> %d",
			playerGeneralStaminaBefore, s.findGeneral("g1").Stamina)
	}
}

func TestEvaluateVictoryPlayerWins(t *testing.T) {
	s := newAITestState()
	// Give every owned city to the player.
	for i := range s.Cities {
		if s.Cities[i].OwnerID == "p2" {
			s.Cities[i].OwnerID = "p1"
		}
	}
	s.Log = nil
	s.evaluateVictory()
	if len(s.Log) == 0 || s.Log[0] != "天下归一，主公成就霸业！" {
		t.Errorf("expected victory log, got %v", s.Log)
	}
}

func TestEvaluateVictoryPlayerLoses(t *testing.T) {
	s := newAITestState()
	for i := range s.Cities {
		if s.Cities[i].OwnerID == "p1" {
			s.Cities[i].OwnerID = "p2"
		}
	}
	s.Log = nil
	s.evaluateVictory()
	if len(s.Log) == 0 || s.Log[0] != "大势已去，主公基业尽失！" {
		t.Errorf("expected defeat log, got %v", s.Log)
	}
}

func TestEndStrategyPhaseRunsEnemyTurns(t *testing.T) {
	s := newAITestState()
	s.Cities = s.Cities[:2]
	s.Routes = []Route{{From: "c1", To: "c2"}}
	battleRand = func(int) int { return 0 }
	defer func() { battleRand = defaultBattleRand }()

	s.EndStrategyPhase()
	// After a full turn the strong AI should have expanded from c2.
	if got := s.findCity("c1").OwnerID; got != "p2" {
		t.Errorf("c1 owner = %q, want p2 after AI-driven turn", got)
	}
	if s.Date.Month != 2 {
		t.Errorf("month = %d, want 2", s.Date.Month)
	}
}
