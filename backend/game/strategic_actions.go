package game

import (
	"errors"
	"fmt"
	"strings"
)

func (s *GameState) applyStrategicAction(req ActionRequest) (*Resolution, error) {
	if s == nil {
		return nil, errors.New("game state is nil")
	}
	s.ensureRNG()
	if s.Ending != nil {
		return nil, errors.New("game has already ended")
	}
	if s.Phase != PhaseEmperor {
		return nil, errors.New("only an emperor can command the strategic map")
	}
	if s.Command <= 0 {
		return nil, errors.New("no command points remain this season")
	}
	req.Target = strings.TrimSpace(req.Target)
	req.Mode = strings.TrimSpace(req.Mode)
	if req.Target == "" {
		return nil, errors.New("missing strategic action target")
	}
	s.ensureStrategicSystems()

	effects, summary, err := s.applyStrategicActionToWorld(req)
	if err != nil {
		return nil, err
	}

	s.Command--
	s.applyEffects(effects)
	s.updateObjectives()
	s.Ending = s.checkEnding()
	if s.Ending != nil {
		s.Scene = nil
	}
	s.History = append(s.History, HistoryEntry{
		Turn:    s.Turn,
		Age:     s.Age,
		Phase:   s.Phase,
		Choice:  strategicActionLabel(req),
		Summary: summary,
		Effects: effects,
	})
	s.addStrategyLog(strategicActionLabel(req), summary, 48)

	return &Resolution{
		Summary: summary,
		Effects: effects,
		Scene:   s.Scene,
		Ending:  s.Ending,
	}, nil
}

func (s *GameState) applyStrategicActionToWorld(req ActionRequest) (Effects, string, error) {
	switch req.Kind {
	case ActionCityDevelop:
		return s.developStrategicCity(req.Target, req.Mode)
	case ActionArmyCommand:
		return s.commandStrategicArmy(req.Target, req.Mode)
	case ActionSiegeCommand:
		return s.commandStrategicSiege(req.Target, req.Mode)
	case ActionGovernorAssign:
		return s.assignStrategicGovernor(req.Target)
	default:
		return Effects{}, "", fmt.Errorf("unknown strategic action kind %q", req.Kind)
	}
}

func (s *GameState) developStrategicCity(cityID, mode string) (Effects, string, error) {
	i, ok := s.Strategy.cityIndex(cityID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic city %q", cityID)
	}
	if mode == "" {
		mode = "farm"
	}
	city := s.Strategy.Cities[i]
	if city.OwnerID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not controlled by the court", city.Name)
	}
	var effects Effects
	var summary string
	switch mode {
	case "farm":
		city.Agriculture = clamp(city.Agriculture+8, 0, 120)
		city.Grain = clamp(city.Grain+14, 0, 180)
		city.Gold = clamp(city.Gold-4, 0, 180)
		effects = Effects{Treasury: -3, Grain: 3, Populace: 1}
		summary = fmt.Sprintf("你命%s垦田修渠，农业升至%d，城中粮草增至%d。", city.Name, city.Agriculture, city.Grain)
	case "market":
		city.Commerce = clamp(city.Commerce+8, 0, 120)
		city.Gold = clamp(city.Gold+12, 0, 180)
		city.Order = clamp(city.Order-2, 0, 100)
		effects = Effects{Treasury: 5, Diplomacy: 1}
		summary = fmt.Sprintf("%s市肆重开，商税入库，商业升至%d，府库增至%d。", city.Name, city.Commerce, city.Gold)
	case "fortify":
		city.Defense = clamp(city.Defense+10, 0, 120)
		city.Grain = clamp(city.Grain-5, 0, 180)
		city.Gold = clamp(city.Gold-8, 0, 180)
		effects = Effects{Treasury: -6, Grain: -2, BorderThreat: -2}
		summary = fmt.Sprintf("%s修墙筑垒，城防升至%d，前线多了一道能拖住敌军的硬骨头。", city.Name, city.Defense)
	case "relief":
		city.Disaster = clamp(city.Disaster-16, 0, 100)
		city.Order = clamp(city.Order+7, 0, 100)
		city.Grain = clamp(city.Grain-12, 0, 180)
		effects = Effects{Grain: -5, Populace: 4, Stability: 2}
		summary = fmt.Sprintf("%s开仓赈灾，灾害降至%d，治安回升到%d。", city.Name, city.Disaster, city.Order)
	case "patrol":
		city.Order = clamp(city.Order+10, 0, 100)
		city.Disaster = clamp(city.Disaster-3, 0, 100)
		effects = Effects{Influence: 2, Stability: 2}
		summary = fmt.Sprintf("巡按与捕盗营进驻%s，治安升至%d，流寇线人被拔掉一批。", city.Name, city.Order)
	case "levy":
		city.Troops = max(0, city.Troops+2600)
		city.Order = clamp(city.Order-6, 0, 100)
		city.Population = clamp(city.Population-2, 0, 120)
		effects = Effects{Army: 4, Populace: -2, Stability: -1}
		summary = fmt.Sprintf("%s募兵守城，城防兵增至%d，但坊里怨声也多了一层。", city.Name, city.Troops)
	default:
		return Effects{}, "", fmt.Errorf("unknown city develop mode %q", mode)
	}
	s.Strategy.Cities[i] = city
	s.syncProvinceFromStrategicCity(city)
	return effects, summary, nil
}

func (s *GameState) commandStrategicArmy(target, mode string) (Effects, string, error) {
	if mode == "" {
		mode = "train"
	}
	switch mode {
	case "train":
		return s.trainStrategicArmy(target)
	case "supply":
		return s.supplyStrategicArmy(target)
	case "march":
		armyID, cityID, err := splitStrategicTarget(target)
		if err != nil {
			return Effects{}, "", err
		}
		return s.marchStrategicArmy(armyID, cityID)
	case "assault":
		armyID, cityID, err := splitStrategicTarget(target)
		if err != nil {
			return Effects{}, "", err
		}
		return s.assaultStrategicCity(armyID, cityID)
	case "recruit":
		return s.recruitStrategicCity(target)
	default:
		return Effects{}, "", fmt.Errorf("unknown army command mode %q", mode)
	}
}

func (s *GameState) trainStrategicArmy(armyID string) (Effects, string, error) {
	i, ok := s.Strategy.armyIndex(armyID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic army %q", armyID)
	}
	army := s.Strategy.Armies[i]
	if army.FactionID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not a court army", army.Name)
	}
	army.Morale = clamp(army.Morale+9, 0, 100)
	army.Training = clamp(army.Training+8, 0, 100)
	army.Grain = clamp(army.Grain-5, 0, 160)
	army.Status = "整训"
	s.Strategy.Armies[i] = army
	return Effects{Grain: -3, Army: 2, Martial: 1}, fmt.Sprintf("%s闭营整训，士气升至%d，训练升至%d。", army.Name, army.Morale, army.Training), nil
}

func (s *GameState) supplyStrategicArmy(armyID string) (Effects, string, error) {
	i, ok := s.Strategy.armyIndex(armyID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic army %q", armyID)
	}
	army := s.Strategy.Armies[i]
	cityIndex, ok := s.Strategy.cityIndex(army.Location)
	if !ok {
		return Effects{}, "", fmt.Errorf("army %s is not in a known city", army.Name)
	}
	city := s.Strategy.Cities[cityIndex]
	if army.FactionID != "court" || city.OwnerID != "court" {
		return Effects{}, "", fmt.Errorf("%s cannot receive court supply here", army.Name)
	}
	transfer := min(28, city.Grain)
	city.Grain = clamp(city.Grain-transfer, 0, 180)
	army.Grain = clamp(army.Grain+transfer, 0, 160)
	army.Status = "补给"
	s.Strategy.Cities[cityIndex] = city
	s.Strategy.Armies[i] = army
	return Effects{Grain: -3}, fmt.Sprintf("%s向%s转运%d石军粮，军粮升至%d。", city.Name, army.Name, transfer, army.Grain), nil
}

func (s *GameState) marchStrategicArmy(armyID, cityID string) (Effects, string, error) {
	armyIndex, ok := s.Strategy.armyIndex(armyID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic army %q", armyID)
	}
	city, ok := s.Strategy.City(cityID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic city %q", cityID)
	}
	army := s.Strategy.Armies[armyIndex]
	if army.FactionID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not a court army", army.Name)
	}
	if !s.Strategy.AreAdjacent(army.Location, cityID) {
		return Effects{}, "", fmt.Errorf("%s is not adjacent to %s", army.Location, cityID)
	}
	if city.OwnerID != "court" {
		return Effects{}, "", fmt.Errorf("%s is hostile; use assault instead of march", city.Name)
	}
	road := s.Strategy.roadBetween(army.Location, cityID)
	army.Location = cityID
	army.Target = ""
	army.Grain = clamp(army.Grain-road.Distance*4, 0, 160)
	army.Morale = clamp(army.Morale-road.Risk/20, 0, 100)
	army.Status = "行军抵达"
	s.Strategy.Armies[armyIndex] = army
	return Effects{Grain: -1, Martial: 1}, fmt.Sprintf("%s沿%s进抵%s，军粮余%d，士气%d。", army.Name, road.Terrain, city.Name, army.Grain, army.Morale), nil
}

func (s *GameState) assaultStrategicCity(armyID, cityID string) (Effects, string, error) {
	armyIndex, ok := s.Strategy.armyIndex(armyID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic army %q", armyID)
	}
	cityIndex, ok := s.Strategy.cityIndex(cityID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic city %q", cityID)
	}
	army := s.Strategy.Armies[armyIndex]
	city := s.Strategy.Cities[cityIndex]
	if army.FactionID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not a court army", army.Name)
	}
	if city.OwnerID == "court" {
		return Effects{}, "", fmt.Errorf("%s is already controlled by the court", city.Name)
	}
	if !s.Strategy.AreAdjacent(army.Location, cityID) {
		return Effects{}, "", fmt.Errorf("%s is not adjacent to %s", army.Location, cityID)
	}
	return s.resolveCourtAssault(armyIndex, cityIndex)
}

func (s *GameState) recruitStrategicCity(cityID string) (Effects, string, error) {
	i, ok := s.Strategy.cityIndex(cityID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic city %q", cityID)
	}
	city := s.Strategy.Cities[i]
	if city.OwnerID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not controlled by the court", city.Name)
	}
	city.Troops += 2800
	city.Order = clamp(city.Order-5, 0, 100)
	city.Population = clamp(city.Population-2, 0, 120)
	s.Strategy.Cities[i] = city
	return Effects{Army: 3, Populace: -2}, fmt.Sprintf("%s募兵入伍，守军增至%d。", city.Name, city.Troops), nil
}

func (s *GameState) commandStrategicSiege(target, mode string) (Effects, string, error) {
	armyID, cityID, err := splitStrategicTarget(target)
	if err != nil {
		return Effects{}, "", err
	}
	if mode == "" {
		mode = "besiege"
	}
	if mode != "besiege" {
		return Effects{}, "", fmt.Errorf("unknown siege command mode %q", mode)
	}
	armyIndex, ok := s.Strategy.armyIndex(armyID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic army %q", armyID)
	}
	cityIndex, ok := s.Strategy.cityIndex(cityID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic city %q", cityID)
	}
	army := s.Strategy.Armies[armyIndex]
	city := s.Strategy.Cities[cityIndex]
	if army.FactionID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not a court army", army.Name)
	}
	if city.OwnerID == "court" {
		return Effects{}, "", fmt.Errorf("%s is already controlled by the court", city.Name)
	}
	if !s.Strategy.AreAdjacent(army.Location, cityID) {
		return Effects{}, "", fmt.Errorf("%s is not adjacent to %s", army.Location, cityID)
	}
	army.Target = cityID
	army.Siege = clamp(army.Siege+18+army.Training/10, 0, 100)
	army.Grain = clamp(army.Grain-8, 0, 160)
	army.Status = "围城"
	city.Grain = clamp(city.Grain-12, 0, 180)
	city.Order = clamp(city.Order-5, 0, 100)
	s.Strategy.Armies[armyIndex] = army
	s.Strategy.Cities[cityIndex] = city
	if army.Siege >= 60 && (city.Grain <= 0 || city.Order <= 12) || army.Siege >= 85 {
		effects, summary := s.forceStrategicSurrender(armyIndex, cityIndex)
		return effects, summary, nil
	}
	return Effects{Grain: -4, BorderThreat: -3}, fmt.Sprintf("%s围困%s，城中粮草降至%d，围城进度%d。", army.Name, city.Name, city.Grain, army.Siege), nil
}

func (s *GameState) assignStrategicGovernor(target string) (Effects, string, error) {
	cityID, ministerID, err := splitStrategicTarget(target)
	if err != nil {
		return Effects{}, "", err
	}
	cityIndex, ok := s.Strategy.cityIndex(cityID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown strategic city %q", cityID)
	}
	minister, ok := s.ministerByID(ministerID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown minister %q", ministerID)
	}
	city := s.Strategy.Cities[cityIndex]
	if city.OwnerID != "court" {
		return Effects{}, "", fmt.Errorf("%s is not controlled by the court", city.Name)
	}
	city.GovernorID = ministerID
	city.Order = clamp(city.Order+minister.Integrity/18+minister.Ability/24, 0, 100)
	city.Commerce = clamp(city.Commerce+minister.Ability/30, 0, 120)
	city.Agriculture = clamp(city.Agriculture+minister.Integrity/32, 0, 120)
	s.Strategy.Cities[cityIndex] = city
	return Effects{Stability: 2, Influence: 1}, fmt.Sprintf("你任%s镇守%s，治安升至%d，府县开始按新章办事。", minister.Name, city.Name, city.Order), nil
}

func splitStrategicTarget(target string) (string, string, error) {
	parts := strings.Split(target, ":")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("strategic target should be left:right, got %q", target)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func (m StrategicState) roadBetween(a, b string) StrategicRoad {
	for _, road := range m.Roads {
		if (road.From == a && road.To == b) || (road.From == b && road.To == a) {
			return road
		}
	}
	return StrategicRoad{From: a, To: b, Terrain: "道路", Risk: 20, Distance: 2}
}

func (s *GameState) commanderAbility(id string) int {
	if minister, ok := s.ministerByID(id); ok {
		return minister.Ability
	}
	return 50
}

func (s *GameState) ministerByID(id string) (Minister, bool) {
	for _, minister := range s.Court {
		if minister.ID == id {
			return minister, true
		}
	}
	return Minister{}, false
}

func (s *GameState) syncProvinceFromStrategicCity(city StrategicCity) {
	provinceID := strategicProvinceID(city.ID)
	for i, province := range s.Provinces {
		if province.ID != provinceID {
			continue
		}
		province.Wealth = clamp((province.Wealth+city.Commerce)/2, 0, 100)
		province.Order = clamp((province.Order+city.Order)/2, 0, 100)
		province.Defense = clamp((province.Defense+city.Defense)/2, 0, 100)
		province.Disaster = clamp((province.Disaster+city.Disaster)/2, 0, 100)
		s.Provinces[i] = province
		return
	}
}

func strategicProvinceID(cityID string) string {
	switch cityID {
	case "south", "canal", "dockyard", "east-sea":
		return "south"
	case "north", "snow-ridge", "river-east":
		return "north"
	case "west", "jade-pass", "bashu", "mountain-pass", "nanling":
		return "west"
	default:
		return "capital"
	}
}

func strategicActionLabel(req ActionRequest) string {
	for _, action := range ActionCatalog() {
		if action.Kind == req.Kind && action.Mode == req.Mode {
			return action.Label
		}
	}
	switch req.Kind {
	case ActionCityDevelop:
		return "城池经营"
	case ActionArmyCommand:
		return "军团指令"
	case ActionSiegeCommand:
		return "围城军令"
	case ActionGovernorAssign:
		return "任命太守"
	default:
		return "战略行动"
	}
}

func (s *GameState) addStrategyLog(title, summary string, severity int) {
	s.Strategy.Logs = append([]StrategyLog{{
		Turn:     s.Turn,
		Season:   s.Season,
		Title:    title,
		Summary:  summary,
		Severity: clamp(severity, 0, 100),
	}}, s.Strategy.Logs...)
	if len(s.Strategy.Logs) > 12 {
		s.Strategy.Logs = s.Strategy.Logs[:12]
	}
}
