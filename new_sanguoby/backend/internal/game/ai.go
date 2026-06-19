package game

import "sort"

// ai.go drives the non-player rulers each turn so the campaign feels alive,
// mirroring the legacy "群雄逐鹿" loop where rival warlords develop their cities
// and march on weaker neighbours between the player's strategy phases.

// aiBattleStaminaCost mirrors the player's battle stamina requirement.
const aiAttackArmsAdvantage = 120 // attack only when attacker arms >= defender * 120%

// RunEnemyTurns lets every non-player, non-neutral ruler act once. Each of their
// generals either attacks a clearly weaker adjacent enemy city or develops its
// home city's economy. Returns the number of cities that changed hands.
func (s *GameState) RunEnemyTurns() int {
	captures := 0

	// Stable ruler order keeps turns deterministic for tests/replays.
	rulerIDs := make([]string, 0, len(s.Rulers))
	for _, r := range s.Rulers {
		if r.ID == s.PlayerID || r.ID == "neutral" || r.ID == "" {
			continue
		}
		rulerIDs = append(rulerIDs, r.ID)
	}
	sort.Strings(rulerIDs)

	for _, rulerID := range rulerIDs {
		captures += s.runRulerTurn(rulerID)
	}
	return captures
}

// runRulerTurn acts for a single AI ruler and returns the number of captures.
func (s *GameState) runRulerTurn(rulerID string) int {
	captures := 0

	// Collect this ruler's actionable generals in a stable order.
	generalIDs := make([]string, 0)
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.OwnerID == rulerID && !g.Captive && g.Stamina >= battleStaminaCost && g.Soldiers > 0 {
			generalIDs = append(generalIDs, g.ID)
		}
	}
	sort.Strings(generalIDs)

	for _, gid := range generalIDs {
		general := s.findGeneral(gid)
		if general == nil || general.Captive || general.Stamina < battleStaminaCost || general.Soldiers <= 0 {
			continue
		}
		from := s.findCity(general.CityID)
		if from == nil || from.OwnerID != rulerID {
			continue
		}

		target := s.pickAttackTarget(rulerID, from, general)
		if target != nil {
			outcome := s.resolveAttack(from, general, target)
			if outcome.Captured {
				captures++
			}
			continue
		}

		// No worthwhile attack: develop the home city instead.
		s.aiDevelopCity(general, from)
	}
	return captures
}

// pickAttackTarget returns the best adjacent enemy city to attack, or nil if no
// target is clearly weaker than the attacking force. Neutral/player/other-ruler
// cities are all valid prey; the strongest relative advantage is chosen.
func (s *GameState) pickAttackTarget(rulerID string, from *City, general *General) *City {
	attackArms := general.Soldiers + general.Soldiers*general.Force/200

	neighbourIDs := s.AdjacentCityIDs(from.ID)
	sort.Strings(neighbourIDs)

	var best *City
	bestMargin := 0
	for _, nbID := range neighbourIDs {
		nb := s.findCity(nbID)
		if nb == nil || nb.OwnerID == rulerID {
			continue
		}
		defendArms := s.cityTroops(nbID)
		defendArms += defendArms * nb.PeopleDevotion / 200
		// Require a meaningful arms advantage before committing to a siege.
		if attackArms*100 < defendArms*aiAttackArmsAdvantage {
			continue
		}
		margin := attackArms - defendArms
		if best == nil || margin > bestMargin {
			best = nb
			bestMargin = margin
		}
	}
	return best
}

// aiDevelopCity invests an AI general's action into the home city's economy,
// reusing the same growth shape as the player's 内政 commands.
func (s *GameState) aiDevelopCity(general *General, city *City) {
	general.Stamina = maxInt(0, general.Stamina-4)
	gain := 10 + general.Intellect/2 + general.Level*2

	switch {
	case city.Farming < city.FarmingLimit:
		city.Farming = minInt(city.FarmingLimit, city.Farming+gain)
	case city.Commerce < city.CommerceLimit:
		city.Commerce = minInt(city.CommerceLimit, city.Commerce+gain)
	case city.PeopleDevotion < 100:
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+4+general.Intellect/12)
	default:
		// Fully developed: reinforce the garrison from population.
		if city.Population > 4000 {
			recruits := 100 + general.Force*4
			city.Population -= recruits * 2
			city.Garrison += recruits
		}
	}
}
