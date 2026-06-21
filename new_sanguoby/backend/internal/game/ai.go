package game

import (
	"fmt"
	"sort"
)

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
	rulerName := s.rulerName(rulerID)

	// Collect this ruler's actionable generals in a stable order.
	generalIDs := make([]string, 0)
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.OwnerID == rulerID && !g.Captive && g.Stamina >= battleStaminaCost && g.Soldiers > 0 {
			generalIDs = append(generalIDs, g.ID)
		}
	}
	sort.Strings(generalIDs)
	if len(generalIDs) == 0 {
		s.prependLog(fmt.Sprintf("诸侯行动：%s 本月按兵不动，暂无可行动武将。", rulerName))
		return 0
	}

	acted := 0
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
			acted++
			continue
		}

		// No worthwhile attack: develop the home city instead.
		s.prependLog(fmt.Sprintf("诸侯行动：%s %s", rulerName, s.aiDevelopCity(general, from)))
		acted++
	}
	if acted == 0 {
		s.prependLog(fmt.Sprintf("诸侯行动：%s 本月军令未动。", rulerName))
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
func (s *GameState) aiDevelopCity(general *General, city *City) string {
	general.Stamina = maxInt(0, general.Stamina-4)
	gain := 10 + general.Intellect/2 + general.Level*2

	switch {
	case city.Farming < city.FarmingLimit:
		before := city.Farming
		city.Farming = minInt(city.FarmingLimit, city.Farming+gain)
		return fmt.Sprintf("令 %s 在 %s 开垦，农业 +%d。", general.Name, city.Name, city.Farming-before)
	case city.Commerce < city.CommerceLimit:
		before := city.Commerce
		city.Commerce = minInt(city.CommerceLimit, city.Commerce+gain)
		return fmt.Sprintf("令 %s 在 %s 招商，商业 +%d。", general.Name, city.Name, city.Commerce-before)
	case city.PeopleDevotion < 100:
		before := city.PeopleDevotion
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+4+general.Intellect/12)
		return fmt.Sprintf("令 %s 巡抚 %s，民忠 +%d。", general.Name, city.Name, city.PeopleDevotion-before)
	default:
		// Fully developed: reinforce the garrison from population.
		if city.Population > 4000 {
			recruits := 100 + general.Force*4
			city.Population -= recruits * 2
			city.Garrison += recruits
			return fmt.Sprintf("令 %s 在 %s 募兵入城，后备 +%d。", general.Name, city.Name, recruits)
		}
	}
	return fmt.Sprintf("令 %s 驻守 %s，城中暂不动员。", general.Name, city.Name)
}
