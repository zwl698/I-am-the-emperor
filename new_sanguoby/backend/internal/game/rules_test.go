package game

import "testing"

func TestAdvanceMonthAppliesCampaignEconomy(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	state.Date.Month = 2

	xuchang := state.CityByID("xuchang")
	xuchang.Money = 1000
	xuchang.Commerce = 400
	xuchang.Food = 1000
	xuchang.Farming = 800
	xuchang.Population = 90_000
	xuchang.PopulationLimit = 90_020

	caoCao := state.GeneralByID("cao-cao")
	caoCao.Stamina = 98

	state.AdvanceMonth()

	if state.Date.Month != 3 {
		t.Fatalf("month = %d, want 3", state.Date.Month)
	}
	if xuchang.Money != 1200 {
		t.Fatalf("money = %d, want quarterly commerce revenue 1200", xuchang.Money)
	}
	if xuchang.Food != 980 {
		t.Fatalf("food = %d, want upkeep without harvest 980", xuchang.Food)
	}
	if xuchang.Population != 90_020 {
		t.Fatalf("population = %d, want capped growth 90020", xuchang.Population)
	}
	if caoCao.Stamina != 100 {
		t.Fatalf("stamina = %d, want capped recovery 100", caoCao.Stamina)
	}
}

func TestAdvanceMonthHarvestsInJuneAndOctober(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	state.Date.Month = 5

	xuchang := state.CityByID("xuchang")
	xuchang.Food = 1000
	xuchang.Farming = 800
	xuchang.Garrison = 0
	for i := range state.Generals {
		if state.Generals[i].CityID == xuchang.ID {
			state.Generals[i].Soldiers = 0
		}
	}

	state.AdvanceMonth()

	if state.Date.Month != 6 {
		t.Fatalf("month = %d, want 6", state.Date.Month)
	}
	if xuchang.Food != 1200 {
		t.Fatalf("food = %d, want harvest food 1200", xuchang.Food)
	}
}

func TestAdvanceMonthWrapsYearAfterDecember(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	state.Date.Year = 190
	state.Date.Month = 12

	state.AdvanceMonth()

	if state.Date.Year != 191 || state.Date.Month != 1 {
		t.Fatalf("date = %d/%d, want 191/1", state.Date.Year, state.Date.Month)
	}
}

func TestAdvanceMonthTurnsStarvingCityToFamine(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	xuchang := state.CityByID("xuchang")
	xuchang.Food = 1
	xuchang.Garrison = 0

	caoCao := state.GeneralByID("cao-cao")
	caoCao.Soldiers = 1000
	before := caoCao.Soldiers

	state.AdvanceMonth()

	if xuchang.Food != 0 {
		t.Fatalf("food = %d, want 0 after famine", xuchang.Food)
	}
	if xuchang.State != CityStateFamine {
		t.Fatalf("city state = %s, want famine", xuchang.State)
	}
	if caoCao.Soldiers != before/2 {
		t.Fatalf("soldiers = %d, want %d", caoCao.Soldiers, before/2)
	}
}
