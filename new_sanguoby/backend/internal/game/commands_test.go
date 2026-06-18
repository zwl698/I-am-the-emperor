package game

import (
	"errors"
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
	if general.Stamina != beforeStamina-8 {
		t.Fatalf("stamina = %d, want %d", general.Stamina, beforeStamina-8)
	}
}

func TestApplyCommandRejectsEnemyCity(t *testing.T) {
	state := NewGame("dongzhuo", "caocao")
	err := state.ApplyCommand("luoyang", "cao-cao", "assart")
	if !errors.Is(err, ErrCityNotPlayable) {
		t.Fatalf("error = %v, want %v", err, ErrCityNotPlayable)
	}
}
