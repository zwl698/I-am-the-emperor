package game

import (
	"errors"
	"strings"
	"testing"
)

func TestApplyCommandDevelopsPlayerCity(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	city := state.CityByID("xuchang")
	general := state.GeneralByID("cao-cao")
	beforeFarming := city.Farming
	beforeMoney := city.Money
	beforeStamina := general.Stamina

	if err := state.ApplyCommand(city.ID, general.ID, "assart"); err != nil {
		t.Fatalf("ApplyCommand error = %v", err)
	}
	if city.Farming <= beforeFarming {
		t.Fatalf("farming = %d, want > %d", city.Farming, beforeFarming)
	}
	if city.Money != beforeMoney-50 {
		t.Fatalf("money = %d, want %d", city.Money, beforeMoney-50)
	}
	if general.Stamina != beforeStamina-4 {
		t.Fatalf("stamina = %d, want %d", general.Stamina, beforeStamina-4)
	}
}

func TestApplyCommandRejectsEnemyCity(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	err := state.ApplyCommand("luoyang", "cao-cao", "assart")
	if !errors.Is(err, ErrCityNotPlayable) {
		t.Fatalf("error = %v, want %v", err, ErrCityNotPlayable)
	}
}

func TestApplyCommandMovesGeneralToFriendlyAdjacentCity(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	general := state.GeneralByID("cao-cao")

	if err := state.ApplyCommandWithTarget("xuchang", general.ID, "move", "chenliu"); err != nil {
		t.Fatalf("ApplyCommandWithTarget move error = %v", err)
	}
	if general.CityID != "chenliu" {
		t.Fatalf("general city = %q, want chenliu", general.CityID)
	}
}

func TestApplyCommandTransportsSuppliesToFriendlyAdjacentCity(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	from := state.CityByID("xuchang")
	target := state.CityByID("chenliu")
	general := state.GeneralByID("cao-cao")
	beforeFood := target.Food
	beforeMoney := target.Money

	if err := state.ApplyCommandWithTarget(from.ID, general.ID, "transportation", target.ID); err != nil {
		t.Fatalf("ApplyCommandWithTarget transportation error = %v", err)
	}
	if target.Food <= beforeFood || target.Money <= beforeMoney {
		t.Fatalf("target resources food=%d money=%d, want increases from %d/%d", target.Food, target.Money, beforeFood, beforeMoney)
	}
}

func TestApplyCommandKillsCaptiveGeneral(t *testing.T) {
	state := newBattleTestState()
	state.PlayerID = "p1"
	city := state.CityByID("c1")
	executor := state.GeneralByID("g1")
	captive := state.GeneralByID("g2")
	captive.OwnerID = "p1"
	captive.CityID = city.ID
	captive.Captive = true
	captive.Soldiers = 0

	if err := state.ApplyCommandDetailed(city.ID, executor.ID, "kill", "", captive.ID); err != nil {
		t.Fatalf("ApplyCommandDetailed kill error = %v", err)
	}
	if captive.CityID != "" || captive.OwnerID != "" || captive.Captive {
		t.Fatalf("captive after kill = %+v, want removed from campaign", captive)
	}
}

func TestApplyCommandRewardsSpecifiedGeneral(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	city := state.CityByID("xuchang")
	executor := state.GeneralByID("cao-cao")
	target := state.GeneralByID("xiahou-dun")
	target.CityID = city.ID
	target.Loyalty = 60
	beforeStamina := executor.Stamina

	if err := state.ApplyCommandDetailed(city.ID, executor.ID, "largess", "", target.ID); err != nil {
		t.Fatalf("ApplyCommandDetailed largess error = %v", err)
	}
	if target.Loyalty != 68 {
		t.Fatalf("target loyalty = %d, want 68", target.Loyalty)
	}
	if executor.Stamina != beforeStamina-4 {
		t.Fatalf("executor stamina = %d, want %d", executor.Stamina, beforeStamina-4)
	}
}

func TestReconnoitreReportsAdjacentEnemyNames(t *testing.T) {
	state := newBattleTestState()
	state.PlayerID = "p1"
	state.CityByID("c1").Money = 100

	if err := state.ApplyCommand("c1", "g1", "reconnoitre"); err != nil {
		t.Fatalf("ApplyCommand reconnoitre error = %v", err)
	}
	if len(state.Log) == 0 || !strings.Contains(state.Log[0], "敌城(乙)") {
		t.Fatalf("reconnoitre log = %v, want adjacent enemy city and ruler", state.Log)
	}
}
