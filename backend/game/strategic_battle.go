package game

import "fmt"

func (s *GameState) resolveCourtAssault(armyIndex, cityIndex int) (Effects, string, error) {
	army := s.Strategy.Armies[armyIndex]
	city := s.Strategy.Cities[cityIndex]
	supportIndexes := s.Strategy.supportingArmyIndexes(armyIndex)
	attackerIndexes := append([]int{armyIndex}, supportIndexes...)
	defenderIndexes := s.Strategy.defendingArmyIndexes(city.ID, city.OwnerID)

	attackScore := s.armyBattlePower(army)
	for _, index := range supportIndexes {
		attackScore += s.armyBattlePower(s.Strategy.Armies[index]) * 7 / 10
	}
	defenseScore := s.cityBattleDefense(city)
	for _, index := range defenderIndexes {
		defenseScore += s.armyBattlePower(s.Strategy.Armies[index]) * 35 / 100
	}

	if attackScore > defenseScore {
		totalTroops := s.totalArmyTroops(attackerIndexes)
		attackerLoss := clamp(defenseScore/7, 900, max(1200, totalTroops/2))
		defenderLoss := clamp(attackScore/8, 900, max(1200, city.Troops*2/3))
		s.distributeArmyLosses(attackerIndexes, attackerLoss, "攻城伤亡")
		s.damageDefendingArmies(defenderIndexes, defenderLoss/2, city.OwnerID, city.ID)

		army = s.Strategy.Armies[armyIndex]
		army.Location = city.ID
		army.Target = ""
		army.Grain = clamp(army.Grain-14, 0, 160)
		army.Morale = clamp(army.Morale+8, 0, 100)
		army.Status = "攻城凯旋"
		s.Strategy.Armies[armyIndex] = army
		for _, index := range supportIndexes {
			support := s.Strategy.Armies[index]
			support.Grain = clamp(support.Grain-8, 0, 160)
			support.Morale = clamp(support.Morale+4, 0, 100)
			support.Status = "侧翼支援"
			s.Strategy.Armies[index] = support
		}

		oldOwner := city.OwnerID
		city.OwnerID = "court"
		city.Troops = max(1800, city.Troops-defenderLoss)
		city.Order = clamp(city.Order-8, 0, 100)
		city.Front = true
		s.Strategy.Cities[cityIndex] = city
		s.syncProvinceFromStrategicCity(city)
		s.adjustActiveWar(18, -4, 6, -16, 1)
		summary := fmt.Sprintf("%s攻破%s，%d支友军参战，%s归入朝廷版图。", army.Name, city.Name, len(supportIndexes), city.Name)
		s.addBattleReport(BattleReport{
			Title:        fmt.Sprintf("%s攻城战", city.Name),
			CityID:       city.ID,
			Attacker:     army.FactionID,
			Defender:     oldOwner,
			Outcome:      "capture",
			AttackerLoss: attackerLoss,
			DefenderLoss: defenderLoss,
			Participants: s.armyIDs(attackerIndexes),
			Summary:      summary,
			Severity:     76,
		})
		return Effects{Army: -3, Grain: -6, BorderThreat: -12, Legitimacy: 3, Martial: 2}, summary, nil
	}

	armyLoss := clamp(city.Defense*90+len(defenderIndexes)*400, 900, max(1000, army.Troops/3))
	cityLoss := clamp(attackScore/10, 600, max(700, city.Troops/3))
	s.distributeArmyLosses(attackerIndexes, armyLoss, "攻城受挫")
	city.Troops = max(0, city.Troops-cityLoss)
	city.Order = clamp(city.Order-4, 0, 100)
	s.Strategy.Cities[cityIndex] = city
	for _, index := range attackerIndexes {
		attacker := s.Strategy.Armies[index]
		attacker.Grain = clamp(attacker.Grain-10, 0, 160)
		attacker.Morale = clamp(attacker.Morale-10, 0, 100)
		attacker.Status = "攻城受挫"
		s.Strategy.Armies[index] = attacker
	}
	s.adjustActiveWar(5, -8, -5, 4, 1)
	summary := fmt.Sprintf("%s强攻%s受挫，城上火油滚木齐下，军心下滑。", army.Name, city.Name)
	s.addBattleReport(BattleReport{
		Title:        fmt.Sprintf("%s攻城受挫", city.Name),
		CityID:       city.ID,
		Attacker:     army.FactionID,
		Defender:     city.OwnerID,
		Outcome:      "repelled",
		AttackerLoss: armyLoss,
		DefenderLoss: cityLoss,
		Participants: s.armyIDs(attackerIndexes),
		Summary:      summary,
		Severity:     68,
	})
	return Effects{Army: -6, Grain: -6, BorderThreat: 4, Stability: -2}, summary, nil
}

func (s *GameState) forceStrategicSurrender(armyIndex, cityIndex int) (Effects, string) {
	army := s.Strategy.Armies[armyIndex]
	city := s.Strategy.Cities[cityIndex]
	oldOwner := city.OwnerID
	defenderLoss := max(900, city.Troops/3)
	attackerLoss := max(120, army.Troops/60)
	army.Troops = max(0, army.Troops-attackerLoss)
	army.Location = city.ID
	army.Target = ""
	army.Siege = 0
	army.Grain = clamp(army.Grain-4, 0, 160)
	army.Morale = clamp(army.Morale+5, 0, 100)
	army.Status = "围城迫降"
	city.OwnerID = "court"
	city.Troops = max(1200, city.Troops-defenderLoss)
	city.Order = clamp(city.Order-6, 0, 100)
	city.Grain = clamp(city.Grain, 0, 180)
	city.Front = true
	s.Strategy.Armies[armyIndex] = army
	s.Strategy.Cities[cityIndex] = city
	s.syncProvinceFromStrategicCity(city)
	s.adjustActiveWar(14, -2, 5, -10, 1)
	summary := fmt.Sprintf("%s久围%s，城中粮断人疲，守将献门投降。", army.Name, city.Name)
	s.addBattleReport(BattleReport{
		Title:        fmt.Sprintf("%s迫降", city.Name),
		CityID:       city.ID,
		Attacker:     army.FactionID,
		Defender:     oldOwner,
		Outcome:      "surrender",
		AttackerLoss: attackerLoss,
		DefenderLoss: defenderLoss,
		Participants: []string{army.ID},
		Summary:      summary,
		Severity:     72,
	})
	return Effects{Army: -1, Grain: -4, BorderThreat: -8, Legitimacy: 2, Martial: 1}, summary
}

func (s *GameState) enemyShouldAssault(army ArmyGroup, city StrategicCity, faction StrategicFaction) bool {
	weakFront := city.Defense <= 28 || city.Order <= 25 || city.Troops <= 5000
	if !weakFront || faction.Threat < 45 {
		return false
	}
	attackScore := s.armyBattlePower(army) + faction.Threat*80
	defenseScore := s.cityBattleDefense(city)
	for _, index := range s.Strategy.defendingArmyIndexes(city.ID, "court") {
		defenseScore += s.armyBattlePower(s.Strategy.Armies[index]) * 45 / 100
	}
	return attackScore > defenseScore*12/10
}

func (s *GameState) resolveEnemyAssault(armyIndex, cityIndex int, faction StrategicFaction) bool {
	army := s.Strategy.Armies[armyIndex]
	city := s.Strategy.Cities[cityIndex]
	if !s.enemyShouldAssault(army, city, faction) {
		return false
	}
	attackerLoss := clamp(s.cityBattleDefense(city)/8, 300, max(500, army.Troops/4))
	defenderLoss := clamp(s.armyBattlePower(army)/7, 800, max(900, city.Troops))
	oldOwner := city.OwnerID
	army.Troops = max(0, army.Troops-attackerLoss)
	army.Location = city.ID
	army.Target = ""
	army.Grain = clamp(army.Grain-9, 0, 160)
	army.Morale = clamp(army.Morale+6, 0, 100)
	army.Status = "破城"
	city.OwnerID = army.FactionID
	city.Troops = max(900, city.Troops-defenderLoss)
	city.Order = clamp(city.Order-10, 0, 100)
	city.Defense = clamp(city.Defense-4, 0, 120)
	city.Front = true
	s.Strategy.Armies[armyIndex] = army
	s.Strategy.Cities[cityIndex] = city
	s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+12, 0, 100)
	s.Crisis.Severity = clamp(s.Crisis.Severity+4, 0, 100)
	s.Crisis.Clock = clamp(s.Crisis.Clock+1, 0, 8)
	s.addBattleReport(BattleReport{
		Title:        fmt.Sprintf("%s失守", city.Name),
		CityID:       city.ID,
		Attacker:     army.FactionID,
		Defender:     oldOwner,
		Outcome:      "enemy_capture",
		AttackerLoss: attackerLoss,
		DefenderLoss: defenderLoss,
		Participants: []string{army.ID},
		Summary:      fmt.Sprintf("%s趁%s空虚突入，%s易帜，朝野震动。", army.Name, city.Name, city.Name),
		Severity:     88,
	})
	s.addStrategyLog("边城失守", fmt.Sprintf("%s被%s攻占，边患陡升。", city.Name, faction.Name), 88)
	return true
}

func (s *GameState) armyBattlePower(army ArmyGroup) int {
	commander := 50
	if army.FactionID == "court" {
		commander = s.commanderAbility(army.CommanderID)
	}
	quality := clamp(army.Morale+army.Training+commander/2, 35, 220)
	return max(0, army.Troops*quality/100)
}

func (s *GameState) cityBattleDefense(city StrategicCity) int {
	return max(0, city.Troops+city.Defense*120+city.Order*40+city.Grain*25)
}

func (s *GameState) totalArmyTroops(indexes []int) int {
	total := 0
	for _, index := range indexes {
		if index >= 0 && index < len(s.Strategy.Armies) {
			total += max(0, s.Strategy.Armies[index].Troops)
		}
	}
	return total
}

func (s *GameState) distributeArmyLosses(indexes []int, totalLoss int, status string) {
	totalTroops := s.totalArmyTroops(indexes)
	if totalTroops <= 0 || totalLoss <= 0 {
		return
	}
	remaining := totalLoss
	for position, index := range indexes {
		army := s.Strategy.Armies[index]
		loss := totalLoss * max(0, army.Troops) / totalTroops
		if position == len(indexes)-1 {
			loss = remaining
		}
		loss = min(loss, army.Troops)
		army.Troops = max(0, army.Troops-loss)
		army.Status = status
		remaining = max(0, remaining-loss)
		s.Strategy.Armies[index] = army
	}
}

func (s *GameState) damageDefendingArmies(indexes []int, totalLoss int, ownerID, cityID string) {
	if len(indexes) == 0 || totalLoss <= 0 {
		return
	}
	s.distributeArmyLosses(indexes, totalLoss, "守城溃退")
	retreat := s.Strategy.retreatCityFor(ownerID, cityID)
	for _, index := range indexes {
		army := s.Strategy.Armies[index]
		army.Morale = clamp(army.Morale-12, 0, 100)
		army.Target = ""
		if retreat != "" {
			army.Location = retreat
		}
		army.Status = "守城溃退"
		s.Strategy.Armies[index] = army
	}
}

func (s *GameState) armyIDs(indexes []int) []string {
	ids := []string{}
	for _, index := range indexes {
		if index >= 0 && index < len(s.Strategy.Armies) {
			ids = append(ids, s.Strategy.Armies[index].ID)
		}
	}
	return ids
}

func (s *GameState) addBattleReport(report BattleReport) {
	report.Turn = s.Turn
	report.Season = s.Season
	report.Severity = clamp(report.Severity, 0, 100)
	s.Strategy.Battles = append([]BattleReport{report}, s.Strategy.Battles...)
	if len(s.Strategy.Battles) > 12 {
		s.Strategy.Battles = s.Strategy.Battles[:12]
	}
}

func (m StrategicState) supportingArmyIndexes(primaryIndex int) []int {
	if primaryIndex < 0 || primaryIndex >= len(m.Armies) {
		return nil
	}
	primary := m.Armies[primaryIndex]
	indexes := []int{}
	for i, army := range m.Armies {
		if i == primaryIndex || army.Troops <= 0 {
			continue
		}
		if army.FactionID == primary.FactionID && army.Location == primary.Location {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m StrategicState) defendingArmyIndexes(cityID, ownerID string) []int {
	indexes := []int{}
	for i, army := range m.Armies {
		if army.Troops <= 0 {
			continue
		}
		if army.FactionID == ownerID && army.Location == cityID {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m StrategicState) retreatCityFor(ownerID, lostCityID string) string {
	for _, neighborID := range m.Neighbors(lostCityID) {
		if city, ok := m.City(neighborID); ok && city.OwnerID == ownerID {
			return neighborID
		}
	}
	return ""
}
