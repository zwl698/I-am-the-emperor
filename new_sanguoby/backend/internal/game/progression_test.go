package game

import "testing"

func TestMaxArmsFormula(t *testing.T) {
	// PlcArmsMaxP: Level*100 + Force*10 + IQ*10
	g := General{Level: 5, Force: 80, Intellect: 60}
	want := 5*100 + 80*10 + 60*10 // 500 + 800 + 600 = 1900
	if got := g.MaxArms(); got != want {
		t.Fatalf("MaxArms() = %d, want %d", got, want)
	}
}

func TestLevelUpCappedAtMax(t *testing.T) {
	g := General{Level: maxLevel}
	g.LevelUp()
	if g.Level != maxLevel {
		t.Fatalf("LevelUp() exceeded cap: got %d, want %d", g.Level, maxLevel)
	}
}

func TestGainExperienceLevelsUp(t *testing.T) {
	g := General{Level: 1, Experience: 0}
	// 250 exp at 100/level -> +2 levels, 50 residual
	levels := g.gainExperience(250)
	if levels != 2 {
		t.Fatalf("gainExperience levels = %d, want 2", levels)
	}
	if g.Level != 3 {
		t.Fatalf("level = %d, want 3", g.Level)
	}
	if g.Experience != 50 {
		t.Fatalf("residual experience = %d, want 50", g.Experience)
	}
}

func TestGainExperienceStopsAtMaxLevel(t *testing.T) {
	g := General{Level: maxLevel, Experience: 0}
	g.gainExperience(500)
	if g.Level != maxLevel {
		t.Fatalf("level exceeded cap: %d", g.Level)
	}
	if g.Experience > fgtExpMax {
		t.Fatalf("residual experience overflowed: %d", g.Experience)
	}
}

func TestBattleExpFormula(t *testing.T) {
	// hurt=400 -> sqrt=20, /4 = 5; same level -> exp = 5 + 2 = 7
	if got := battleExp(400, 5, 5); got != 7 {
		t.Fatalf("battleExp(400, equal level) = %d, want 7", got)
	}
	// attacker much higher level: levelDiff penalty
	// hurt=400 -> base 5; diff=10 -> 5<=10 so exp=0; +2 = 2
	if got := battleExp(400, 15, 5); got != 2 {
		t.Fatalf("battleExp with high level diff = %d, want 2", got)
	}
}

func TestKillBonusExp(t *testing.T) {
	if got := killBonusExp(5, 5); got != killExpSameLvl {
		t.Fatalf("same level kill bonus = %d, want %d", got, killExpSameLvl)
	}
	if got := killBonusExp(3, 8); got != killExpLowLvl {
		t.Fatalf("lower level kill bonus = %d, want %d", got, killExpLowLvl)
	}
	if got := killBonusExp(8, 3); got != killExpHighLvl {
		t.Fatalf("higher level kill bonus = %d, want %d", got, killExpHighLvl)
	}
}

func TestConscriptUsesDevotionAndMoney(t *testing.T) {
	s := &GameState{}
	g := &General{Name: "将", Force: 70}
	// PeopleDevotion=50 -> 50*20 = 1000 desired; Money=200 -> limit 2000; so 1000
	c := &City{Name: "城", PeopleDevotion: 50, Money: 200}
	got := s.conscript(g, c)
	if got != 1000 {
		t.Fatalf("conscript recruited = %d, want 1000", got)
	}
	if c.MothballArms != 1000 {
		t.Fatalf("MothballArms = %d, want 1000", c.MothballArms)
	}
	if c.Money != 200-100 { // 1000/10 = 100 spent
		t.Fatalf("Money = %d, want 100", c.Money)
	}
}

func TestConscriptLimitedByMoney(t *testing.T) {
	s := &GameState{}
	g := &General{Name: "将"}
	// devotion wants 50*20=1000, but Money=30 -> limit 30*10=300
	c := &City{Name: "城", PeopleDevotion: 50, Money: 30}
	got := s.conscript(g, c)
	if got != 300 {
		t.Fatalf("conscript limited = %d, want 300", got)
	}
	if c.Money != 0 {
		t.Fatalf("Money = %d, want 0", c.Money)
	}
}

func TestDistributeCappedByMaxArms(t *testing.T) {
	s := &GameState{}
	// MaxArms = 1*100 + 50*10 + 50*10 = 1100
	g := &General{Name: "将", Level: 1, Force: 50, Intellect: 50, Soldiers: 100}
	c := &City{Name: "城", MothballArms: 5000}
	moved := s.distribute(g, c)
	wantMoved := 1100 - 100 // capacity = max - current
	if moved != wantMoved {
		t.Fatalf("distribute moved = %d, want %d", moved, wantMoved)
	}
	if g.Soldiers != 1100 {
		t.Fatalf("soldiers = %d, want 1100", g.Soldiers)
	}
	if c.MothballArms != 5000-wantMoved {
		t.Fatalf("remaining reserve = %d, want %d", c.MothballArms, 5000-wantMoved)
	}
}

func TestSearchForTalentRecruitsNeutral(t *testing.T) {
	s := &GameState{
		Generals: []General{
			{ID: "g1", Name: "搜寻者", OwnerID: "p1", CityID: "c1", Intellect: 149},
			{ID: "g2", Name: "在野贤", OwnerID: "neutral", CityID: "c1"},
		},
	}
	city := &City{ID: "c1", Name: "城", OwnerID: "p1"}
	g := &s.Generals[0]
	// 智力149 -> rand(150) almost always < 149; run a few times to ensure recruitment
	recruited := false
	for i := 0; i < 50; i++ {
		s.Generals[1].OwnerID = "neutral"
		s.Generals[1].CityID = "c1"
		s.searchForTalent(g, city)
		if s.Generals[1].OwnerID == "p1" {
			recruited = true
			break
		}
	}
	if !recruited {
		t.Fatal("expected high-intellect general to recruit a neutral talent")
	}
}
