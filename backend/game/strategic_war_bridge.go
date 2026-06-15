package game

import "fmt"

type strategicWarProfile struct {
	FrontCityID  string
	TargetCityID string
	FactionID    string
	ArmyID       string
}

func (s *GameState) applyStrategicWarTactic(kind OrderKind, war WarCampaign) string {
	if s == nil {
		return ""
	}
	s.ensureStrategicSystems()
	if len(s.Strategy.Cities) == 0 || len(s.Strategy.Armies) == 0 {
		return ""
	}
	profile := strategicProfileForWar(war)
	if profile.TargetCityID == "" && profile.FrontCityID == "" {
		return ""
	}
	switch kind {
	case OrderMobilize:
		return s.mobilizeStrategicWarFront(profile)
	case OrderCampaign:
		return s.campaignStrategicWarFront(profile)
	case OrderFortify:
		return s.fortifyStrategicWarFront(profile)
	case OrderTruce:
		return s.truceStrategicWarFront(profile)
	default:
		return ""
	}
}

func strategicProfileForWar(war WarCampaign) strategicWarProfile {
	switch war.ID {
	case "snow-ridge":
		return strategicWarProfile{FrontCityID: "north", TargetCityID: "snow-ridge", FactionID: "beidi", ArmyID: "northern-banner"}
	case "western-oath":
		return strategicWarProfile{FrontCityID: "west", TargetCityID: "jade-pass", FactionID: "remnant", ArmyID: "imperial-guard"}
	case "river-bandits":
		return strategicWarProfile{FrontCityID: "river-east", TargetCityID: "river-east", FactionID: "rebels", ArmyID: "imperial-guard"}
	case "jade-pass":
		return strategicWarProfile{FrontCityID: "west", TargetCityID: "jade-pass", FactionID: "remnant", ArmyID: "imperial-guard"}
	default:
		return strategicWarProfile{}
	}
}

func strategicProvinceIDForWar(war WarCampaign) string {
	profile := strategicProfileForWar(war)
	if profile.FrontCityID != "" {
		return strategicProvinceID(profile.FrontCityID)
	}
	if profile.TargetCityID != "" {
		return strategicProvinceID(profile.TargetCityID)
	}
	return "north"
}

func (s *GameState) mobilizeStrategicWarFront(profile strategicWarProfile) string {
	armyIndex, ok := s.courtArmyForStrategicWar(profile)
	if !ok {
		return ""
	}
	army := s.Strategy.Armies[armyIndex]
	army.Grain = clamp(army.Grain+28, 0, 160)
	army.Morale = clamp(army.Morale+7, 0, 100)
	army.Training = clamp(army.Training+5, 0, 100)
	army.Status = "兵棋调粮"
	s.Strategy.Armies[armyIndex] = army
	if cityIndex, ok := s.Strategy.cityIndex(army.Location); ok {
		city := s.Strategy.Cities[cityIndex]
		if city.OwnerID == "court" {
			city.Grain = clamp(city.Grain-8, 0, 180)
			s.Strategy.Cities[cityIndex] = city
		}
	}
	s.addStrategyLog("兵棋调粮", fmt.Sprintf("%s得军饷粮车，军粮升至%d，士气升至%d。", army.Name, army.Grain, army.Morale), 52)
	return fmt.Sprintf("战略地图上，%s已补粮整训。", army.Name)
}

func (s *GameState) campaignStrategicWarFront(profile strategicWarProfile) string {
	armyIndex, ok := s.courtArmyForStrategicWar(profile)
	if !ok {
		return ""
	}
	army := s.Strategy.Armies[armyIndex]
	targetIndex, targetOK := s.Strategy.cityIndex(profile.TargetCityID)
	if targetOK {
		target := s.Strategy.Cities[targetIndex]
		if target.OwnerID != "court" && s.Strategy.AreAdjacent(army.Location, target.ID) {
			_, summary, err := s.resolveCourtAssault(armyIndex, targetIndex)
			if err == nil {
				return summary
			}
		}
	}
	next := s.nextFriendlyStepToward(army.Location, profile.TargetCityID)
	if next != "" {
		old := army.Location
		road := s.Strategy.roadBetween(old, next)
		army.Location = next
		army.Target = profile.TargetCityID
		army.Grain = clamp(army.Grain-road.Distance*4, 0, 160)
		army.Morale = clamp(army.Morale-road.Risk/24, 0, 100)
		army.Status = "兵棋进军"
		s.Strategy.Armies[armyIndex] = army
		s.addStrategyLog("兵棋进军", fmt.Sprintf("%s由%s进至%s，准备压向%s。", army.Name, cityDisplayName(s.Strategy, old), cityDisplayName(s.Strategy, next), cityDisplayName(s.Strategy, profile.TargetCityID)), 58)
		return fmt.Sprintf("战略地图上，%s已向%s推进。", army.Name, cityDisplayName(s.Strategy, profile.TargetCityID))
	}
	army.Target = profile.TargetCityID
	army.Grain = clamp(army.Grain-8, 0, 160)
	army.Morale = clamp(army.Morale+2, 0, 100)
	army.Status = "兵棋压迫"
	s.Strategy.Armies[armyIndex] = army
	if targetOK {
		target := s.Strategy.Cities[targetIndex]
		target.Troops = max(0, target.Troops-900)
		target.Order = clamp(target.Order-3, 0, 100)
		target.Front = true
		s.Strategy.Cities[targetIndex] = target
		s.addBattleReport(BattleReport{
			Title:        fmt.Sprintf("%s战役压迫", target.Name),
			CityID:       target.ID,
			Attacker:     "court",
			Defender:     target.OwnerID,
			Outcome:      "pressure",
			AttackerLoss: max(200, army.Troops/80),
			DefenderLoss: 900,
			Participants: []string{army.ID},
			Summary:      fmt.Sprintf("%s按沙盘推演压迫%s，守军折损，城中秩序动摇。", army.Name, target.Name),
			Severity:     62,
		})
		return fmt.Sprintf("战略地图上，%s已压迫%s。", army.Name, target.Name)
	}
	return fmt.Sprintf("战略地图上，%s进入战役机动。", army.Name)
}

func (s *GameState) fortifyStrategicWarFront(profile strategicWarProfile) string {
	cityID := profile.FrontCityID
	if cityID == "" {
		cityID = profile.TargetCityID
	}
	cityIndex, ok := s.Strategy.cityIndex(cityID)
	if !ok {
		return ""
	}
	city := s.Strategy.Cities[cityIndex]
	city.Defense = clamp(city.Defense+12, 0, 120)
	city.Troops = max(0, city.Troops+1600)
	city.Grain = clamp(city.Grain-4, 0, 180)
	city.Front = true
	s.Strategy.Cities[cityIndex] = city
	for i, road := range s.Strategy.Roads {
		if road.From == city.ID || road.To == city.ID {
			road.Risk = clamp(road.Risk-4, 0, 100)
			s.Strategy.Roads[i] = road
		}
	}
	s.syncProvinceFromStrategicCity(city)
	s.addStrategyLog("前线筑垒", fmt.Sprintf("%s城防升至%d，守军增至%d，近路风险下降。", city.Name, city.Defense, city.Troops), 50)
	return fmt.Sprintf("战略地图上，%s已加固为前线堡垒。", city.Name)
}

func (s *GameState) truceStrategicWarFront(profile strategicWarProfile) string {
	factionIndex, ok := s.Strategy.factionIndex(profile.FactionID)
	if !ok {
		return ""
	}
	faction := s.Strategy.Factions[factionIndex]
	faction.Threat = clamp(faction.Threat-16, 0, 100)
	faction.Relation = clamp(faction.Relation+10, 0, 100)
	s.Strategy.Factions[factionIndex] = faction
	for i, army := range s.Strategy.Armies {
		if army.FactionID != faction.ID {
			continue
		}
		army.Target = ""
		army.Morale = clamp(army.Morale-3, 0, 100)
		army.Status = "观望退营"
		s.Strategy.Armies[i] = army
	}
	s.addStrategyLog("边帐议和", fmt.Sprintf("%s暂缓攻势，威胁降至%d，邦交升至%d。", faction.Name, faction.Threat, faction.Relation), 46)
	return fmt.Sprintf("战略地图上，%s转入观望。", faction.Name)
}

func (s *GameState) courtArmyForStrategicWar(profile strategicWarProfile) (int, bool) {
	if profile.ArmyID != "" {
		if index, ok := s.Strategy.armyIndex(profile.ArmyID); ok && s.Strategy.Armies[index].FactionID == "court" {
			return index, true
		}
	}
	bestIndex := -1
	bestScore := -1
	for i, army := range s.Strategy.Armies {
		if army.FactionID != "court" || army.Troops <= 0 {
			continue
		}
		score := army.Troops/1000 + army.Morale/4 + army.Training/4 + army.Grain/5
		if army.Location == profile.FrontCityID {
			score += 80
		}
		if s.Strategy.AreAdjacent(army.Location, profile.TargetCityID) {
			score += 60
		}
		if army.Location == profile.TargetCityID {
			score += 40
		}
		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	if bestIndex < 0 {
		return 0, false
	}
	return bestIndex, true
}

func (s *GameState) nextFriendlyStepToward(from, target string) string {
	for _, neighborID := range s.Strategy.Neighbors(from) {
		if neighborID == target {
			continue
		}
		city, ok := s.Strategy.City(neighborID)
		if !ok || city.OwnerID != "court" {
			continue
		}
		if s.Strategy.AreAdjacent(neighborID, target) {
			return neighborID
		}
	}
	for _, neighborID := range s.Strategy.Neighbors(from) {
		city, ok := s.Strategy.City(neighborID)
		if ok && city.OwnerID == "court" {
			return neighborID
		}
	}
	return ""
}

func cityDisplayName(strategy StrategicState, cityID string) string {
	if city, ok := strategy.City(cityID); ok {
		return city.Name
	}
	return cityID
}

func withStrategicSummary(base, strategic string) string {
	if strategic == "" {
		return base
	}
	return base + " " + strategic
}
