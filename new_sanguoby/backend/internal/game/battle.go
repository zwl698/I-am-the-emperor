package game

import (
	"fmt"
	"math/rand"
)

// battle.go implements the 出征/攻城 (campaign attack) resolution.
//
// The win/lose probability model is a faithful port of the legacy engine's
// FgtCountWon (FgtCount.c): the outcome is decided by comparing total troop
// strength (兵力) and provisions (兵粮) between attacker and defender.
//
//	攻方兵力 > 守方兵力:
//	    攻方 > 守方*2          -> 70% 胜
//	    否则 攻方粮 > 守方粮    -> 60% 胜
//	    否则                  -> 40% 胜
//	攻方兵力 <= 守方兵力:
//	    攻方 < 守方/2          -> 2%  胜
//	    否则 攻方粮 > 守方粮    -> 30% 胜
//	    否则                  -> 10% 胜

var (
	ErrBattleSameOwner    = fmt.Errorf("%w: cannot attack your own city", ErrInvalidCommand)
	ErrBattleNotAdjacent  = fmt.Errorf("%w: target city is not adjacent", ErrInvalidCommand)
	ErrBattleNoSoldiers   = fmt.Errorf("%w: general has no soldiers", ErrGeneralNotReady)
	ErrBattleTargetNoCity = fmt.Errorf("%w: target city not found", ErrInvalidCommand)
)

// BattleOutcome describes the result of one 出征 resolution for the API/UI.
type BattleOutcome struct {
	Won              bool     `json:"won"`
	FromCityID       string   `json:"fromCityId"`
	TargetCityID     string   `json:"targetCityId"`
	GeneralID        string   `json:"generalId"`
	AttackerLosses   int      `json:"attackerLosses"`
	DefenderLosses   int      `json:"defenderLosses"`
	Captured         bool     `json:"captured"`
	CapturedGenerals []string `json:"capturedGenerals"`
	Message          string   `json:"message"`
}

const battleStaminaCost = 4

// battleRand is the source of randomness; overridable in tests.
var defaultBattleRand = rand.Intn
var battleRand = defaultBattleRand

// AdjacentCityIDs returns the set of city IDs reachable from cityID via a single
// route hop (the campaign road network is undirected).
func (s *GameState) AdjacentCityIDs(cityID string) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range s.Routes {
		var other string
		switch cityID {
		case r.From:
			other = r.To
		case r.To:
			other = r.From
		default:
			continue
		}
		if other != "" && !seen[other] {
			seen[other] = true
			out = append(out, other)
		}
	}
	return out
}

func (s *GameState) isAdjacent(fromID, targetID string) bool {
	for _, id := range s.AdjacentCityIDs(fromID) {
		if id == targetID {
			return true
		}
	}
	return false
}

// cityTroops sums every soldier defending a city (garrison + non-captive generals).
func (s *GameState) cityTroops(cityID string) int {
	city := s.findCity(cityID)
	if city == nil {
		return 0
	}
	total := city.Garrison
	for i := range s.Generals {
		if s.Generals[i].CityID == cityID && !s.Generals[i].Captive {
			total += s.Generals[i].Soldiers
		}
	}
	return total
}

// resolveBattleWon ports FgtCountWon: given strengths and provisions, returns
// whether the attacker wins. randv is a 0..100 roll.
func resolveBattleWon(attackArms, defendArms, attackFood, defendFood, randv int) bool {
	if attackArms == 0 {
		return false
	}
	if defendArms == 0 {
		return true
	}
	if attackArms > defendArms {
		switch {
		case attackArms>>1 > defendArms:
			return randv < 70
		case attackFood > defendFood:
			return randv < 60
		default:
			return randv < 40
		}
	}
	switch {
	case attackArms < defendArms>>1:
		return randv < 2
	case attackFood > defendFood:
		return randv < 30
	default:
		return randv < 10
	}
}

// ApplyBattle launches an attack from a player-owned city against an adjacent
// city, using the chosen general's army plus the origin garrison as the strike
// force. The outcome follows the legacy probability model.
func (s *GameState) ApplyBattle(fromCityID, generalID, targetCityID string) (*BattleOutcome, error) {
	from := s.findCity(fromCityID)
	if from == nil {
		return nil, fmt.Errorf("%w: city %s", ErrInvalidCommand, fromCityID)
	}
	if from.OwnerID != s.PlayerID {
		return nil, ErrCityNotPlayable
	}

	general := s.findGeneral(generalID)
	if general == nil || general.CityID != fromCityID || general.OwnerID != s.PlayerID || general.Captive {
		return nil, ErrGeneralNotReady
	}
	if general.Stamina < battleStaminaCost {
		return nil, ErrGeneralNotReady
	}
	if general.Soldiers <= 0 {
		return nil, ErrBattleNoSoldiers
	}

	target := s.findCity(targetCityID)
	if target == nil {
		return nil, ErrBattleTargetNoCity
	}
	if target.OwnerID == s.PlayerID {
		return nil, ErrBattleSameOwner
	}
	if !s.isAdjacent(fromCityID, targetCityID) {
		return nil, ErrBattleNotAdjacent
	}

	return s.resolveAttack(from, general, target), nil
}

// resolveAttack performs the actual siege math and mutation for any attacker
// (player or AI). Callers are responsible for validating ownership/adjacency
// and ensuring the general can act. The conquered city is transferred to the
// attacking general's owner.
func (s *GameState) resolveAttack(from *City, general *General, target *City) *BattleOutcome {
	general.Stamina -= battleStaminaCost

	// Attacker strength: the marching general's army, boosted by 武力/民忠 morale.
	attackArms := general.Soldiers + general.Soldiers*general.Force/200
	defendArms := s.cityTroops(target.ID)
	// Defender morale bonus from city devotion (民心向背).
	defendArms += defendArms * target.PeopleDevotion / 200

	randv := battleRand(101)
	won := resolveBattleWon(attackArms, defendArms, from.Food, target.Food, randv)

	outcome := &BattleOutcome{
		FromCityID:   from.ID,
		TargetCityID: target.ID,
		GeneralID:    general.ID,
		Won:          won,
	}

	if won {
		// Attacker loses a modest share; defender is routed.
		attackerLoss := minInt(general.Soldiers, general.Soldiers*(20+randv/4)/100)
		general.Soldiers -= attackerLoss
		outcome.AttackerLosses = attackerLoss
		outcome.DefenderLosses = s.routDefenders(target.ID)
		outcome.CapturedGenerals = s.captureCity(target, general)
		outcome.Captured = true
		if len(outcome.CapturedGenerals) == 1 {
			outcome.Message = fmt.Sprintf("%s 攻克 %s，俘虏 %s！", general.Name, target.Name, outcome.CapturedGenerals[0])
		} else if len(outcome.CapturedGenerals) > 1 {
			outcome.Message = fmt.Sprintf("%s 攻克 %s，俘虏 %s 等 %d 将！", general.Name, target.Name, outcome.CapturedGenerals[0], len(outcome.CapturedGenerals))
		} else {
			outcome.Message = fmt.Sprintf("%s 攻克 %s！", general.Name, target.Name)
		}
	} else {
		// Attacker is repelled with heavier losses; defender takes light losses.
		attackerLoss := minInt(general.Soldiers, general.Soldiers*(40+randv/3)/100)
		general.Soldiers -= attackerLoss
		outcome.AttackerLosses = attackerLoss
		outcome.DefenderLosses = s.lightDefenderLosses(target.ID, randv)
		outcome.Message = fmt.Sprintf("%s 进攻 %s 失利，退回 %s。", general.Name, target.Name, from.Name)
	}

	s.prependLog(outcome.Message)
	return outcome
}

// routDefenders wipes most of a routed city's defenders and returns the losses.
func (s *GameState) routDefenders(cityID string) int {
	city := s.findCity(cityID)
	losses := 0
	if city != nil {
		losses += city.Garrison
		city.Garrison = 0
	}
	for i := range s.Generals {
		if s.Generals[i].CityID == cityID && !s.Generals[i].Captive {
			losses += s.Generals[i].Soldiers
			s.Generals[i].Soldiers = 0
		}
	}
	return losses
}

// lightDefenderLosses applies a small casualty to defenders after a failed siege.
func (s *GameState) lightDefenderLosses(cityID string, randv int) int {
	losses := 0
	for i := range s.Generals {
		if s.Generals[i].CityID == cityID && !s.Generals[i].Captive {
			loss := s.Generals[i].Soldiers * (5 + randv/10) / 100
			s.Generals[i].Soldiers -= loss
			losses += loss
		}
	}
	return losses
}

// captureCity transfers a city to the attacker, moves the conquering general in,
// captures the defeated officers, and applies the social cost of conquest.
func (s *GameState) captureCity(target *City, general *General) []string {
	capturedGenerals := []string{}
	target.OwnerID = general.OwnerID
	general.CityID = target.ID
	general.Captive = false
	for i := range s.Generals {
		defender := &s.Generals[i]
		if defender.CityID != target.ID || defender.ID == general.ID || defender.OwnerID == general.OwnerID {
			continue
		}
		defender.OwnerID = general.OwnerID
		defender.Captive = true
		defender.Soldiers = 0
		defender.Stamina = 0
		defender.Loyalty = minInt(defender.Loyalty, 45)
		if defender.Loyalty <= 0 {
			defender.Loyalty = 30
		}
		capturedGenerals = append(capturedGenerals, defender.Name)
	}
	// The conquered populace is shaken: devotion drops, calamity risk rises.
	target.PeopleDevotion = maxInt(0, target.PeopleDevotion-20)
	target.AvoidCalamity = maxInt(0, target.AvoidCalamity-10)
	// Defending generals that survived (none after a rout) would defect; here the
	// city's leaderless remnants surrender, leaving the garrison emptied.
	target.Garrison = 0
	return capturedGenerals
}
