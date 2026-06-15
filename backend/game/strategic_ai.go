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
	s.syncForeignStatesFromStrategicFactions()
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

func (s *GameState) syncForeignStatesFromStrategicFactions() {
	if len(s.ForeignStates) == 0 || len(s.Strategy.Factions) == 0 {
		return
	}
	for i, foreign := range s.ForeignStates {
		factionID := strategicFactionIDForForeign(foreign.ID)
		faction, ok := s.Strategy.Faction(factionID)
		if !ok || faction.IsPlayer {
			continue
		}
		foreign.Threat = max(foreign.Threat, faction.Threat*3/4)
		if faction.Relation < foreign.Relation {
			foreign.Relation = clamp((foreign.Relation*2+faction.Relation)/3, 0, 100)
		} else if faction.Relation > foreign.Relation {
			foreign.Relation = clamp((foreign.Relation+faction.Relation)/2, 0, 100)
		}
		foreign.Attitude = foreignAttitude(foreign)
		s.ForeignStates[i] = foreign
	}
}

func strategicFactionIDForForeign(foreignID string) string {
	switch foreignID {
	case "nanman":
		return "nanling"
	case "xiyu":
		return "remnant"
	default:
		return foreignID
	}
}

func (s *GameState) strategicMilitaryPressure() int {
	if s == nil || len(s.Strategy.Factions) == 0 {
		return 0
	}
	pressure := 0
	for _, faction := range s.Strategy.Factions {
		if !faction.IsPlayer {
			pressure = max(pressure, faction.Threat)
		}
	}
	for _, battle := range s.Strategy.Battles {
		pressure = max(pressure, battle.Severity)
	}
	for _, army := range s.Strategy.Armies {
		if army.FactionID == "court" || army.Troops <= 0 {
			continue
		}
		for _, neighborID := range s.Strategy.Neighbors(army.Location) {
			city, ok := s.Strategy.City(neighborID)
			if !ok || city.OwnerID != "court" {
				continue
			}
			frontPressure := 45 + army.Troops/1600 + max(0, 55-city.Order)/2 + max(0, 60-city.Defense)/2
			pressure = max(pressure, clamp(frontPressure, 0, 100))
		}
	}
	for _, army := range s.Strategy.Armies {
		if army.FactionID == "court" && army.Grain <= 10 {
			pressure = max(pressure, 62)
		}
	}
	return clamp(pressure, 0, 100)
}
