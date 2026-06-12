package game

import "fmt"

func (s *GameState) applyStrategicPressure(domain Domain) {
	if s.Phase != PhaseEmperor {
		return
	}
	s.ensureStrategicSystems()
	s.applyStrategicCityProduction()
	s.applyStrategicArmySupply()
	s.applyStrategicEnemyPressure(domain)
}

func (s *GameState) applyStrategicCityProduction() {
	treasuryGain := 0
	grainGain := 0
	for i, city := range s.Strategy.Cities {
		if city.OwnerID != "court" {
			continue
		}
		gold := max(0, city.Commerce/22+city.Order/35-city.Disaster/30)
		grain := max(0, city.Agriculture/20+city.Order/45-city.Disaster/28)
		city.Gold = clamp(city.Gold+gold, 0, 180)
		city.Grain = clamp(city.Grain+grain, 0, 180)
		treasuryGain += gold
		grainGain += grain
		s.Strategy.Cities[i] = city
	}
	if treasuryGain > 0 || grainGain > 0 {
		s.Stats.Treasury = clamp(s.Stats.Treasury+treasuryGain/18, 0, 160)
		s.Stats.Grain = clamp(s.Stats.Grain+grainGain/18, 0, 160)
	}
}

func (s *GameState) applyStrategicArmySupply() {
	for i, army := range s.Strategy.Armies {
		need := max(1, army.Troops/9000)
		if army.Grain <= 0 {
			loss := max(250, army.Troops/22)
			army.Troops = max(0, army.Troops-loss)
			army.Morale = clamp(army.Morale-9, 0, 100)
			army.Status = "缺粮减员"
			s.Strategy.Armies[i] = army
			if army.FactionID == "court" {
				s.Stats.Army = clamp(s.Stats.Army-2, 0, 140)
				s.Crisis.Clock = clamp(s.Crisis.Clock+1, 0, 8)
			}
			s.addStrategyLog("军粮断续", fmt.Sprintf("%s军粮耗尽，士气降至%d，兵力余%d。", army.Name, army.Morale, army.Troops), 72)
			continue
		}
		army.Grain = clamp(army.Grain-need, 0, 160)
		if army.Grain <= 10 {
			army.Morale = clamp(army.Morale-2, 0, 100)
			army.Status = "粮道吃紧"
		}
		s.Strategy.Armies[i] = army
	}
}

func (s *GameState) applyStrategicEnemyPressure(domain Domain) {
	threatDelta := 2
	if domain == DomainMilitary {
		threatDelta = -1
	}
	if domain == DomainDiplomacy {
		threatDelta = -2
	}
	for i, faction := range s.Strategy.Factions {
		if faction.IsPlayer {
			continue
		}
		faction.Threat = clamp(faction.Threat+threatDelta, 0, 100)
		s.Strategy.Factions[i] = faction
	}
	for armyIndex := range s.Strategy.Armies {
		army := s.Strategy.Armies[armyIndex]
		if army.FactionID == "court" || army.Troops <= 0 {
			continue
		}
		faction, _ := s.Strategy.Faction(army.FactionID)
		for _, neighborID := range s.Strategy.Neighbors(army.Location) {
			cityIndex, ok := s.Strategy.cityIndex(neighborID)
			if !ok {
				continue
			}
			city := s.Strategy.Cities[cityIndex]
			if city.OwnerID != "court" {
				continue
			}
			if s.resolveEnemyAssault(armyIndex, cityIndex, faction) {
				break
			}
			pressure := clamp(faction.Threat/25+army.Troops/14000, 1, 8)
			city.Order = clamp(city.Order-pressure, 0, 100)
			city.Defense = clamp(city.Defense-max(1, pressure/3), 0, 120)
			city.Front = true
			s.Strategy.Cities[cityIndex] = city
			s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+max(1, pressure/2), 0, 100)
			if pressure >= 4 {
				s.Crisis.Severity = clamp(s.Crisis.Severity+1, 0, 100)
			}
			s.addStrategyLog("敌军压境", fmt.Sprintf("%s在%s外施压，%s治安降至%d，城防降至%d。", army.Name, city.Name, city.Name, city.Order, city.Defense), 55+pressure*5)
		}
	}
}
