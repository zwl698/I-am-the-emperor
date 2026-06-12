package game

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"time"
)

func AvailableDynasties() []Dynasty {
	return slices.Clone(dynasties())
}

func NewGame(seed int64) *GameState {
	state, _ := NewGameWithDynasty("dayin", seed)
	return state
}

func NewGameWithDynasty(dynastyID string, seed int64) (*GameState, error) {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	dynasty, ok := findDynasty(dynastyID)
	if !ok {
		return nil, fmt.Errorf("unknown dynasty %q", dynastyID)
	}

	state := &GameState{
		ID:      fmt.Sprintf("dragon-%x", seed^time.Now().UnixNano()),
		Seed:    seed,
		Turn:    0,
		Age:     1,
		Phase:   PhasePrince,
		Dynasty: dynasty,
		Assets: Assets{
			Hero:            "/assets/palace-hero.png",
			Dynasties:       "/assets/dynasty-scroll.png",
			Characters:      "/assets/characters.png",
			SceneGallery:    sceneGalleryPaths(),
			PortraitGallery: portraitGalleryPaths(),
		},
		Season:        "春",
		Stats:         dynasty.Initial,
		Factions:      startingFactions(dynasty.ID),
		Court:         startingCourt(),
		TalentPool:    startingTalentPool(dynasty.ID),
		Harem:         startingHarem(dynasty.ID),
		Heirs:         startingHeirs(dynasty.ID),
		Succession:    startingSuccession(dynasty.ID),
		Offices:       startingOffices(dynasty.ID),
		Projects:      startingProjects(dynasty.ID),
		Policies:      startingPolicies(dynasty.ID),
		Relations:     startingRelations(dynasty.ID),
		ForeignStates: startingForeignStates(dynasty.ID),
		Plots:         startingPlots(dynasty.ID),
		LegalCases:    startingLegalCases(dynasty.ID),
		PublicOpinion: startingPublicOpinion(dynasty.ID),
		Provinces:     startingProvinces(dynasty.ID),
		Wars:          startingWars(dynasty.ID),
		Strategy:      startingStrategicState(dynasty.ID),
		Crisis:        startingCrisis(dynasty.ID),
		Objectives:    startingObjectives(dynasty.ID),
		rng:           rand.New(rand.NewSource(seed)),
	}
	state.updateObjectives()
	state.Scene = princeScene(0, state)
	return state, nil
}

func (s *GameState) ApplyChoice(choiceID string) (*Resolution, error) {
	if s == nil {
		return nil, errors.New("game state is nil")
	}
	s.ensureRNG()
	if s.Ending != nil {
		return nil, errors.New("game has already ended")
	}
	if s.Scene == nil {
		return nil, errors.New("game has no active scene")
	}

	choice, ok := s.findChoice(choiceID)
	if !ok {
		return nil, fmt.Errorf("unknown choice %q", choiceID)
	}

	s.applyEffects(choice.Effects)
	s.applyChoiceToWorld(choice)
	s.Turn++
	s.advanceAfter()
	s.applyWorldPressure(choice.Domain)
	events := s.triggerSeasonalEvents(choice.Domain)
	s.dealEventHand()
	s.updateObjectives()
	s.Ending = s.checkEnding()
	if s.Ending == nil {
		s.Scene = s.nextScene()
	} else {
		s.Scene = nil
	}

	s.History = append(s.History, HistoryEntry{
		Turn:    s.Turn,
		Age:     s.Age,
		Phase:   s.Phase,
		Choice:  choice.Text,
		Summary: choice.Outcome,
		Effects: choice.Effects,
	})

	return &Resolution{
		Summary: strings.TrimSpace(choice.Outcome + " " + describeEvents(events)),
		Effects: choice.Effects,
		Scene:   s.Scene,
		Ending:  s.Ending,
	}, nil
}

func (s *GameState) ApplyOrder(req OrderRequest) (*Resolution, error) {
	if s == nil {
		return nil, errors.New("game state is nil")
	}
	s.ensureRNG()
	if s.Ending != nil {
		return nil, errors.New("game has already ended")
	}
	if s.Phase != PhaseEmperor {
		return nil, errors.New("only an emperor can issue orders")
	}
	if s.Command <= 0 {
		return nil, errors.New("no command points remain this season")
	}
	req.Target = strings.TrimSpace(req.Target)
	if req.Kind == "" {
		return nil, errors.New("missing order kind")
	}
	if req.Target == "" {
		return nil, errors.New("missing order target")
	}

	effects, summary, err := s.applyOrderToWorld(req)
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
		Choice:  orderLabel(req.Kind),
		Summary: summary,
		Effects: effects,
	})

	return &Resolution{
		Summary: summary,
		Effects: effects,
		Scene:   s.Scene,
		Ending:  s.Ending,
	}, nil
}

func (s *GameState) ForceCoronationForTest() {
	s.Phase = PhaseEmperor
	s.Age = 18
	s.Turn = 5
	s.ReignYear = 1
	s.Season = "春"
	s.Command = s.commandBudget()
	s.Stats.Treasury = max(s.Stats.Treasury, 72)
	s.Stats.Grain = max(s.Stats.Grain, 66)
	s.Stats.Populace = max(s.Stats.Populace, 64)
	s.Stats.Army = max(s.Stats.Army, 58)
	s.Stats.Diplomacy = max(s.Stats.Diplomacy, 52)
	s.Stats.Stability = max(s.Stats.Stability, 60)
	if s.Stats.BorderThreat == 0 {
		s.Stats.BorderThreat = 38
	}
	s.ensureCourtSystems()
	s.dealEventHand()
	s.updateObjectives()
	s.Scene = emperorScene(s)
}

func (s *GameState) findChoice(choiceID string) (Choice, bool) {
	for _, choice := range s.Scene.Choices {
		if choice.ID == choiceID {
			return choice, true
		}
	}
	return Choice{}, false
}

func (s *GameState) ensureRNG() {
	if s.rng == nil {
		s.rng = rand.New(rand.NewSource(s.Seed + int64(s.Turn*7919)))
	}
}

func (s *GameState) applyEffects(e Effects) {
	s.Stats.Legitimacy = clamp(s.Stats.Legitimacy+e.Legitimacy, 0, 100)
	s.Stats.Health = clamp(s.Stats.Health+e.Health, 0, 100)
	s.Stats.Learning = clamp(s.Stats.Learning+e.Learning, 0, 100)
	s.Stats.Martial = clamp(s.Stats.Martial+e.Martial, 0, 100)
	s.Stats.Charisma = clamp(s.Stats.Charisma+e.Charisma, 0, 100)
	s.Stats.Influence = clamp(s.Stats.Influence+e.Influence, 0, 100)
	s.Stats.Treasury = clamp(s.Stats.Treasury+e.Treasury, 0, 160)
	s.Stats.Grain = clamp(s.Stats.Grain+e.Grain, 0, 160)
	s.Stats.Populace = clamp(s.Stats.Populace+e.Populace, 0, 100)
	s.Stats.Army = clamp(s.Stats.Army+e.Army, 0, 140)
	s.Stats.Diplomacy = clamp(s.Stats.Diplomacy+e.Diplomacy, 0, 100)
	s.Stats.Stability = clamp(s.Stats.Stability+e.Stability, 0, 100)
	s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+e.BorderThreat, 0, 100)
	s.Stats.Reform = clamp(s.Stats.Reform+e.Reform, 0, 100)
}

func (s *GameState) applyChoiceToWorld(choice Choice) {
	if s.Phase != PhaseEmperor {
		return
	}
	targetProvince := s.Turn % max(1, len(s.Provinces))
	targetFaction := s.Turn % max(1, len(s.Factions))

	switch choice.Domain {
	case DomainDomestic:
		s.adjustProvince(targetProvince, -8, 5, 0, -8)
		s.adjustFaction(targetFaction, -2, 2)
		s.adjustMinisterByRole("户部尚书", 1, 3)
	case DomainEconomy:
		s.adjustProvince(targetProvince, 8, -2, 0, 2)
		s.adjustFaction(targetFaction, 5, -4)
		s.adjustMinisterByRole("户部尚书", -3, 6)
	case DomainMilitary:
		s.adjustProvince(targetProvince, -2, 0, 9, 0)
		s.adjustFactionByID("border", 6, 5)
		s.adjustMinisterByRole("大将军", 4, 7)
		s.adjustActiveWar(8, -7, -6, 5, 1)
	case DomainDiplomacy:
		s.adjustFactionByID("clan", 4, 4)
		s.adjustProvince(targetProvince, 2, 2, 0, 0)
		s.adjustMinisterByRole("长公主", 3, 3)
		s.adjustActiveWar(4, -5, 2, 2, 0)
	case DomainReform:
		s.adjustProvince(0, 6, -6, 0, -3)
		s.adjustFaction(0, -7, -8)
		s.adjustMinisterByRole("太傅", -2, 8)
	case DomainIntrigue:
		s.adjustFaction(targetFaction, -8, -5)
		s.adjustProvince(targetProvince, -1, -2, 0, 0)
		s.adjustMinisterByRole("太傅", -4, 5)
	case DomainCourt:
		s.adjustFaction(targetFaction, 3, 5)
		s.adjustMinister(targetFaction, 5, -2)
		s.applyCourtAgendaOutcome(choice)
	}

	s.Crisis.Severity = clamp(s.Crisis.Severity+crisisDelta(choice.Domain), 0, 100)
	if s.Crisis.Severity < 35 {
		s.Crisis.Clock = max(0, s.Crisis.Clock-1)
	} else {
		s.Crisis.Clock = clamp(s.Crisis.Clock+1, 0, 8)
	}
}

func (s *GameState) applyOrderToWorld(req OrderRequest) (Effects, string, error) {
	if effects, summary, ok, err := s.applyCourtOrder(req); ok || err != nil {
		return effects, summary, err
	}
	if effects, summary, ok, err := s.applyGrandStrategyOrder(req); ok || err != nil {
		return effects, summary, err
	}
	if effects, summary, ok, err := s.applyForeignOrder(req); ok || err != nil {
		return effects, summary, err
	}
	if effects, summary, ok, err := s.applyPlotOrder(req); ok || err != nil {
		return effects, summary, err
	}
	if effects, summary, ok, err := s.applyJusticeOrder(req); ok || err != nil {
		return effects, summary, err
	}
	switch req.Kind {
	case OrderRelief:
		i, ok := s.findProvinceIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown province target %q", req.Target)
		}
		province := s.Provinces[i]
		beforeDisaster := province.Disaster
		s.adjustProvince(i, -1, 12, 0, -16)
		s.adjustCrisis(-4, -1)
		effects := Effects{Treasury: -6, Grain: -8, Populace: 6, Stability: 3, Legitimacy: 1}
		return effects, fmt.Sprintf("你命%s开仓设粥棚、修堤坝。灾情从%d压到%d，百姓开始把年号写进门联。", province.Name, beforeDisaster, s.Provinces[i].Disaster), nil
	case OrderGarrison:
		i, ok := s.findProvinceIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown province target %q", req.Target)
		}
		province := s.Provinces[i]
		s.adjustProvince(i, -2, 4, 15, -3)
		s.adjustFactionByID("border", 4, 4)
		s.adjustCrisis(-2, 0)
		effects := Effects{Treasury: -7, Army: -2, BorderThreat: -6, Martial: 1}
		return effects, fmt.Sprintf("禁军换防%s，烽燧重新点亮。守备升至%d，边报上的红印少了一角。", province.Name, s.Provinces[i].Defense), nil
	case OrderTax:
		i, ok := s.findProvinceIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown province target %q", req.Target)
		}
		province := s.Provinces[i]
		s.adjustProvince(i, -8, -7, 0, 4)
		s.adjustFactionByID("merchant", 4, -6)
		s.adjustCrisis(2, 0)
		effects := Effects{Treasury: 14, Populace: -5, Stability: -3, Reform: 1}
		return effects, fmt.Sprintf("户部清丈%s田亩，银库一夜充盈；地方豪强却开始把账本藏进暗格。", province.Name), nil
	case OrderInspect:
		if i, ok := s.findFactionIndex(req.Target); ok {
			faction := s.Factions[i]
			s.adjustFaction(i, -10, -5)
			s.adjustCrisis(1, 0)
			effects := Effects{Influence: 5, Legitimacy: -1, Stability: -2}
			return effects, fmt.Sprintf("密档送入御前，%s被迫交出几条旧线。权势降至%d，但怨气也在殿外结霜。", faction.Name, s.Factions[i].Power), nil
		}
		i, ok := s.findProvinceIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown order target %q", req.Target)
		}
		province := s.Provinces[i]
		s.adjustProvince(i, 0, 7, 0, -5)
		effects := Effects{Influence: 3, Stability: 1}
		return effects, fmt.Sprintf("巡按暗访%s，胥吏名单换了一批。地方秩序升至%d。", province.Name, s.Provinces[i].Order), nil
	case OrderAppease:
		i, ok := s.findFactionIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown faction target %q", req.Target)
		}
		faction := s.Factions[i]
		s.adjustFaction(i, 4, 15)
		s.adjustCrisis(-1, 0)
		effects := Effects{Treasury: -5, Stability: 4, Influence: -1}
		return effects, fmt.Sprintf("你召见%s，赐宴、授差、留台阶。%s忠诚升至%d，朝会气氛松了一寸。", faction.Leader, faction.Name, s.Factions[i].Loyalty), nil
	case OrderPurge:
		i, ok := s.findFactionIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown faction target %q", req.Target)
		}
		faction := s.Factions[i]
		s.adjustFaction(i, -17, -14)
		s.adjustCrisis(4, 1)
		effects := Effects{Influence: 7, Stability: -6, Legitimacy: -3, Health: -1}
		return effects, fmt.Sprintf("廷杖声落在%s门前，%s权势骤降到%d。短期无人敢言，长期无人敢忘。", faction.Leader, faction.Name, s.Factions[i].Power), nil
	case OrderCanal:
		i, ok := s.findProvinceIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown province target %q", req.Target)
		}
		province := s.Provinces[i]
		s.adjustProvince(i, 12, 2, 0, -6)
		effects := Effects{Treasury: -10, Grain: 5, Reform: 4, Populace: 2}
		return effects, fmt.Sprintf("%s新渠开挖，粮船有了第二条路。地方财富升至%d，新政图纸也多了一道墨线。", province.Name, s.Provinces[i].Wealth), nil
	case OrderTrade:
		i, ok := s.findProvinceIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown province target %q", req.Target)
		}
		province := s.Provinces[i]
		s.adjustProvince(i, 14, -2, 0, 1)
		s.adjustFactionByID("merchant", 5, 7)
		effects := Effects{Treasury: 9, Diplomacy: 4, BorderThreat: 2}
		return effects, fmt.Sprintf("你准%s开互市，驼队与货船挤满关津。商帮称颂圣明，边境也多了几双试探的眼睛。", province.Name), nil
	case OrderMobilize:
		i, ok := s.findWarIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown war target %q", req.Target)
		}
		war := s.Wars[i]
		s.adjustWar(i, 3, 12, 10, -3, 0)
		s.adjustFactionByID("border", 5, 4)
		s.adjustMinisterByRole("大将军", 3, 8)
		effects := Effects{Treasury: -9, Grain: -7, Army: 4, BorderThreat: -3}
		return effects, fmt.Sprintf("你向%s增发军饷与粮车，%s粮道升至%d、士气升至%d。", war.Front, war.Name, s.Wars[i].Supply, s.Wars[i].Morale), nil
	case OrderCampaign:
		i, ok := s.findWarIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown war target %q", req.Target)
		}
		war := s.Wars[i]
		advance := 10 + s.Stats.Martial/12 + s.Wars[i].Morale/18
		threatDrop := 7 + s.Wars[i].Supply/25
		s.adjustWar(i, advance, -6, -5, -threatDrop, 1)
		s.adjustProvinceByID("north", -2, 1, 4, 0)
		s.adjustMinisterByRole("大将军", 2, 10)
		effects := Effects{Treasury: -12, Grain: -8, Army: -4, BorderThreat: -threatDrop, Martial: 1}
		return effects, fmt.Sprintf("你准%s出塞决战，战役推进到%d/%d，敌势降至%d。", war.Name, s.Wars[i].Progress, 100, s.Wars[i].Threat), nil
	case OrderFortify:
		i, ok := s.findWarIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown war target %q", req.Target)
		}
		war := s.Wars[i]
		s.adjustWar(i, 2, 5, 4, -10, 0)
		s.adjustProvinceByID("north", -3, 2, 12, -2)
		effects := Effects{Treasury: -8, Grain: -4, BorderThreat: -8, Stability: 1}
		return effects, fmt.Sprintf("%s沿线筑堡、修烽燧、屯军粮。敌军威胁降至%d，北境防务更稳。", war.Front, s.Wars[i].Threat), nil
	case OrderTruce:
		i, ok := s.findWarIndex(req.Target)
		if !ok {
			return Effects{}, "", fmt.Errorf("unknown war target %q", req.Target)
		}
		war := s.Wars[i]
		s.adjustWar(i, 5, -2, -4, -12, 1)
		s.adjustFactionByID("border", -4, -3)
		s.adjustMinisterByRole("长公主", 4, 4)
		effects := Effects{Treasury: -6, Diplomacy: 8, BorderThreat: -10, Legitimacy: -1}
		return effects, fmt.Sprintf("使团在%s外设帐议和，%s退去一营骑兵；武臣不满，但边患暂缓。", war.Front, war.Enemy), nil
	default:
		return Effects{}, "", fmt.Errorf("unknown order kind %q", req.Kind)
	}
}

func (s *GameState) adjustProvince(i, wealth, order, defense, disaster int) {
	if len(s.Provinces) == 0 {
		return
	}
	i = clamp(i, 0, len(s.Provinces)-1)
	p := s.Provinces[i]
	p.Wealth = clamp(p.Wealth+wealth, 0, 100)
	p.Order = clamp(p.Order+order, 0, 100)
	p.Defense = clamp(p.Defense+defense, 0, 100)
	p.Disaster = clamp(p.Disaster+disaster, 0, 100)
	s.Provinces[i] = p
}

func (s *GameState) adjustProvinceByID(id string, wealth, order, defense, disaster int) {
	if i, ok := s.findProvinceIndex(id); ok {
		s.adjustProvince(i, wealth, order, defense, disaster)
	}
}

func (s *GameState) adjustCrisis(severity, clock int) {
	s.Crisis.Severity = clamp(s.Crisis.Severity+severity, 0, 100)
	s.Crisis.Clock = clamp(s.Crisis.Clock+clock, 0, 8)
}

func (s *GameState) findProvinceIndex(id string) (int, bool) {
	for i, province := range s.Provinces {
		if province.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) findFactionIndex(id string) (int, bool) {
	for i, faction := range s.Factions {
		if faction.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) findWarIndex(id string) (int, bool) {
	for i, war := range s.Wars {
		if war.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) adjustFaction(i, power, loyalty int) {
	if len(s.Factions) == 0 {
		return
	}
	i = clamp(i, 0, len(s.Factions)-1)
	f := s.Factions[i]
	f.Power = clamp(f.Power+power, 0, 100)
	f.Loyalty = clamp(f.Loyalty+loyalty, 0, 100)
	s.Factions[i] = f
}

func (s *GameState) adjustFactionByID(id string, power, loyalty int) {
	for i := range s.Factions {
		if s.Factions[i].ID == id {
			s.adjustFaction(i, power, loyalty)
			return
		}
	}
}

func (s *GameState) adjustMinister(i, loyalty, stress int) {
	if len(s.Court) == 0 {
		return
	}
	i = clamp(i, 0, len(s.Court)-1)
	minister := s.Court[i]
	minister.Loyalty = clamp(minister.Loyalty+loyalty, 0, 100)
	minister.Stress = clamp(minister.Stress+stress, 0, 100)
	s.Court[i] = minister
}

func (s *GameState) adjustMinisterByRole(role string, loyalty, stress int) {
	for i := range s.Court {
		if s.Court[i].Role == role {
			s.adjustMinister(i, loyalty, stress)
			return
		}
	}
}

func (s *GameState) adjustWar(i, progress, supply, morale, threat, duration int) {
	if len(s.Wars) == 0 {
		return
	}
	i = clamp(i, 0, len(s.Wars)-1)
	war := s.Wars[i]
	war.Progress = clamp(war.Progress+progress, 0, 100)
	war.Supply = clamp(war.Supply+supply, 0, 100)
	war.Morale = clamp(war.Morale+morale, 0, 100)
	war.Threat = clamp(war.Threat+threat, 0, 100)
	war.Duration = max(0, war.Duration+duration)
	switch {
	case war.Progress >= 100 && war.Threat <= 20:
		war.Stage = "凯旋"
	case war.Progress >= 72:
		war.Stage = "反攻"
	case war.Threat >= 80:
		war.Stage = "压境"
	case war.Supply <= 25:
		war.Stage = "粮道危急"
	default:
		war.Stage = "拉锯"
	}
	s.Wars[i] = war
}

func (s *GameState) adjustActiveWar(progress, supply, morale, threat, duration int) {
	if len(s.Wars) == 0 {
		return
	}
	active := 0
	for i, war := range s.Wars {
		if war.Stage != "凯旋" {
			active = i
			break
		}
	}
	s.adjustWar(active, progress, supply, morale, threat, duration)
}

func crisisDelta(domain Domain) int {
	switch domain {
	case DomainDomestic:
		return -4
	case DomainEconomy:
		return 1
	case DomainMilitary:
		return -6
	case DomainDiplomacy:
		return -5
	case DomainReform:
		return 5
	case DomainIntrigue:
		return 3
	default:
		return 1
	}
}

func (s *GameState) advanceAfter() {
	if s.Phase == PhasePrince {
		ages := []int{6, 10, 14, 16, 18}
		s.Age = ages[min(s.Turn, len(ages)-1)]
		if s.Turn >= 5 {
			s.coronate()
		}
		return
	}
	s.advanceCalendar()
}

func (s *GameState) advanceCalendar() {
	seasons := []string{"春", "夏", "秋", "冬"}
	idx := 0
	for i, season := range seasons {
		if s.Season == season {
			idx = i
			break
		}
	}
	idx++
	if idx >= len(seasons) {
		idx = 0
		s.ReignYear++
		s.Age++
	}
	s.Season = seasons[idx]
	if s.ReignYear < 1 {
		s.ReignYear = 1
	}
	s.Command = s.commandBudget()
}

func (s *GameState) coronate() {
	s.Phase = PhaseEmperor
	s.Age = max(s.Age, 18)
	s.ReignYear = 1
	s.Season = "春"
	s.Stats.Treasury = clamp(s.Dynasty.Initial.Treasury+s.Stats.Legitimacy/4+s.Stats.Learning/6, 25, 140)
	s.Stats.Grain = clamp(s.Dynasty.Initial.Grain+s.Stats.Charisma/6, 20, 140)
	s.Stats.Populace = clamp(s.Dynasty.Initial.Populace+s.Stats.Charisma/7+s.Stats.Legitimacy/9, 20, 100)
	s.Stats.Army = clamp(s.Dynasty.Initial.Army+s.Stats.Martial/3, 25, 140)
	s.Stats.Diplomacy = clamp(s.Dynasty.Initial.Diplomacy+s.Stats.Charisma/5+s.Stats.Learning/8, 20, 100)
	s.Stats.Stability = clamp(s.Dynasty.Initial.Stability+s.Stats.Influence/5+s.Stats.Legitimacy/8, 15, 100)
	s.Stats.BorderThreat = clamp(s.Dynasty.Initial.BorderThreat-s.Stats.Martial/8, 5, 100)
	s.Command = s.commandBudget()
	s.ensureCourtSystems()
	s.dealEventHand()
}

func (s *GameState) ensureCourtSystems() {
	if len(s.Harem) == 0 {
		s.Harem = startingHarem(s.Dynasty.ID)
	}
	if len(s.Heirs) == 0 {
		s.Heirs = startingHeirs(s.Dynasty.ID)
	}
	if s.Succession.Stability <= 0 {
		s.Succession = startingSuccession(s.Dynasty.ID)
	}
	if len(s.Offices) == 0 {
		s.Offices = startingOffices(s.Dynasty.ID)
	}
	s.ensureTalentPool()
	s.ensureGrandStrategySystems()
	s.ensureForeignSystems()
	s.ensurePlotSystems()
	s.ensureJusticeSystems()
	s.ensureStrategicSystems()
}

func (s *GameState) commandBudget() int {
	budget := 3
	if s.Stats.Influence >= 55 {
		budget++
	}
	if s.Stats.Reform >= 60 {
		budget++
	}
	if s.Stats.Health < 25 {
		budget--
	}
	return clamp(budget, 2, 5)
}

func (s *GameState) applyWorldPressure(domain Domain) {
	if s.Phase != PhaseEmperor {
		return
	}
	s.Stats.Treasury = clamp(s.Stats.Treasury-2+averageProvinceWealth(s.Provinces)/35, 0, 160)
	s.Stats.Grain = clamp(s.Stats.Grain-1-disasterPressure(s.Provinces)/40, 0, 160)
	if domain != DomainMilitary && domain != DomainDiplomacy {
		s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+2+s.rng.Intn(4), 0, 100)
	}
	if averageFactionLoyalty(s.Factions) < 38 {
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
	if averageProvinceOrder(s.Provinces) < 42 {
		s.Stats.Populace = clamp(s.Stats.Populace-3, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-3, 0, 100)
	}
	if s.Stats.Grain < 25 {
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-2, 0, 100)
	}
	if s.Stats.Treasury < 20 {
		s.Stats.Stability = clamp(s.Stats.Stability-3, 0, 100)
	}
	if s.Stats.BorderThreat > s.Stats.Army+20 {
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
	s.applyWarPressure(domain)
	s.applyCourtPressure(domain)
	s.applyGrandStrategyPressure(domain)
	s.applyForeignPressure(domain)
	s.applyPlotPressure(domain)
	s.applyJusticePressure(domain)
	s.applyStrategicPressure(domain)
}

func (s *GameState) applyWarPressure(domain Domain) {
	if len(s.Wars) == 0 {
		return
	}
	for i, war := range s.Wars {
		if war.Stage == "凯旋" {
			s.Stats.BorderThreat = clamp(s.Stats.BorderThreat-2, 0, 100)
			continue
		}
		threat := 2
		supply := -2
		morale := -1
		if domain == DomainMilitary {
			threat = -2
			supply = -1
			morale = 1
		}
		if domain == DomainDiplomacy {
			threat = -1
			morale = 0
		}
		s.adjustWar(i, 0, supply, morale, threat, 1)
		if war.Threat >= 75 {
			s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+3, 0, 100)
			s.Crisis.Severity = clamp(s.Crisis.Severity+2, 0, 100)
		}
		if war.Supply <= 20 {
			s.Stats.Army = clamp(s.Stats.Army-3, 0, 140)
			s.Crisis.Clock = clamp(s.Crisis.Clock+1, 0, 8)
		}
	}
}

func (s *GameState) applyCourtPressure(domain Domain) {
	for i, minister := range s.Court {
		stress := 1
		if domain == DomainCourt {
			stress = -2
		}
		if minister.Stress >= 80 {
			s.Court[i].Loyalty = clamp(minister.Loyalty-2, 0, 100)
		}
		s.Court[i].Stress = clamp(s.Court[i].Stress+stress, 0, 100)
	}
	s.applyOfficePressure(domain)
	s.applyPalacePressure(domain)
}

func (s *GameState) checkEnding() *Ending {
	if s.Stats.Health <= 0 {
		return &Ending{Kind: EndingDeath, Title: "龙驭宾天", Summary: "操劳与暗疾耗尽了你的生命，史官在夜色里合上实录。"}
	}
	if s.Phase == PhaseEmperor && (s.Stats.Stability <= 0 || s.Stats.Populace <= 0 || s.Crisis.Clock >= 8 || (s.Stats.BorderThreat >= 95 && s.Stats.Army < 90) || warCanBreakDynasty(s.Wars, s.Stats.Army)) {
		return &Ending{Kind: EndingCollapse, Title: "山河失守", Summary: "边患、民怨、党争与危机链同时爆发，王朝在你的御座前倾塌。"}
	}
	if s.Phase == PhaseEmperor && s.Turn >= 72 && s.Stats.Stability >= 80 && s.Stats.Populace >= 80 && s.Stats.BorderThreat <= 25 && s.Stats.Reform >= 55 && averageWarProgress(s.Wars) >= 85 {
		return &Ending{Kind: EndingGolden, Title: "万邦来朝", Summary: "新法成制，仓廪充盈，边境安定，诸国遣使入贡，你开创了被后世反复吟诵的盛世。"}
	}
	return nil
}

func (s *GameState) nextScene() *Scene {
	if s.Phase == PhasePrince {
		return princeScene(s.Turn, s)
	}
	return emperorScene(s)
}

func startingObjectives(dynastyID string) []Objective {
	objectives := []Objective{
		{ID: "secure_throne", Title: "稳坐龙椅", Description: "完成皇子成长与夺嫡，正式登基。", Target: 100, Reward: "开启六部朝政与天下棋盘。"},
		{ID: "stabilize_realm", Title: "安民定国", Description: "让朝稳、民心、粮草都回到可持续水平。", Target: 80, Reward: "降低危机钟推进速度。"},
		{ID: "reform_state", Title: "鼎新旧制", Description: "推行新法，建立足以撑起盛世的新制度。", Target: 80, Reward: "提升长期财政与民生收益。"},
		{ID: "pacify_borders", Title: "靖平边患", Description: "用军务与外交压低边患，稳住北境与西陲。", Target: 80, Reward: "解锁盛世结局条件之一。"},
		{ID: "win_external_war", Title: "外战定鼎", Description: "经营粮道、士气、攻势与议和，把边境战事打成可控胜利。", Target: 100, Reward: "边患长期下降，大将军与边镇归心。"},
		{ID: "long_reign", Title: "十八年长治", Description: "熬过夺嫡后的漫长统治，把一时胜利变成制度惯性。", Target: 72, Reward: "进入盛世终局评定。"},
	}
	if dynastyID == "chengping" {
		objectives = append(objectives, Objective{ID: "restore_treasury", Title: "补回亏空", Description: "让国库脱离危险线，阻止财政崩盘。", Target: 70, Reward: "财政线消耗降低。"})
	}
	if dynastyID == "xuanshuo" {
		objectives = append(objectives, Objective{ID: "hold_north", Title: "守住雪岭", Description: "提升北境防御，避免烽火压城。", Target: 75, Reward: "边镇武勋忠诚提高。"})
	}
	return objectives
}

func (s *GameState) updateObjectives() {
	for i := range s.Objectives {
		objective := s.Objectives[i]
		switch objective.ID {
		case "secure_throne":
			if s.Phase == PhaseEmperor {
				objective.Progress = objective.Target
			} else {
				objective.Progress = clamp(s.Turn*20, 0, objective.Target)
			}
		case "stabilize_realm":
			objective.Progress = clamp((s.Stats.Stability+s.Stats.Populace+s.Stats.Grain)/3, 0, objective.Target)
		case "reform_state":
			objective.Progress = clamp(s.Stats.Reform, 0, objective.Target)
		case "pacify_borders":
			objective.Progress = clamp(80-s.Stats.BorderThreat+s.Stats.Army/4+s.Stats.Diplomacy/6, 0, objective.Target)
		case "win_external_war":
			objective.Progress = clamp(averageWarProgress(s.Wars), 0, objective.Target)
		case "long_reign":
			objective.Progress = clamp(s.Turn, 0, objective.Target)
		case "restore_treasury":
			objective.Progress = clamp(s.Stats.Treasury, 0, objective.Target)
		case "hold_north":
			objective.Progress = clamp(provinceDefense(s.Provinces, "north")-s.Stats.BorderThreat/3+s.Stats.Army/5, 0, objective.Target)
		}
		objective.Completed = objective.Progress >= objective.Target
		s.Objectives[i] = objective
	}
}

func provinceDefense(provinces []Province, id string) int {
	for _, province := range provinces {
		if province.ID == id {
			return province.Defense
		}
	}
	return 0
}

func crisisLine(s *GameState) string {
	return fmt.Sprintf("%s：%s 当前烈度 %d，危机钟 %d/8。", s.Crisis.Title, s.Crisis.Summary, s.Crisis.Severity, s.Crisis.Clock)
}

func emperorMood(stats Stats) string {
	score := stats.Stability + stats.Populace + stats.Army + stats.Diplomacy - stats.BorderThreat + stats.Reform/2
	switch {
	case score >= 285:
		return "盛世"
	case score >= 210:
		return "可治"
	case score >= 145:
		return "暗涌"
	default:
		return "危局"
	}
}

func cloneScene(scene Scene) *Scene {
	choices := make([]Choice, len(scene.Choices))
	copy(choices, scene.Choices)
	scene.Choices = choices
	return &scene
}

func (e Effects) Describe() string {
	parts := make([]string, 0, 8)
	add := func(name string, value int) {
		if value == 0 {
			return
		}
		sign := "+"
		if value < 0 {
			sign = ""
		}
		parts = append(parts, fmt.Sprintf("%s%s%d", name, sign, value))
	}
	add("名望", e.Legitimacy)
	add("健康", e.Health)
	add("学识", e.Learning)
	add("武略", e.Martial)
	add("魅力", e.Charisma)
	add("势力", e.Influence)
	add("国库", e.Treasury)
	add("粮草", e.Grain)
	add("民心", e.Populace)
	add("军力", e.Army)
	add("邦交", e.Diplomacy)
	add("朝稳", e.Stability)
	add("边患", e.BorderThreat)
	add("新政", e.Reform)
	return strings.Join(parts, "、")
}

func averageProvinceWealth(provinces []Province) int {
	if len(provinces) == 0 {
		return 0
	}
	total := 0
	for _, p := range provinces {
		total += p.Wealth
	}
	return total / len(provinces)
}

func averageProvinceOrder(provinces []Province) int {
	if len(provinces) == 0 {
		return 0
	}
	total := 0
	for _, p := range provinces {
		total += p.Order
	}
	return total / len(provinces)
}

func disasterPressure(provinces []Province) int {
	if len(provinces) == 0 {
		return 0
	}
	total := 0
	for _, p := range provinces {
		total += p.Disaster
	}
	return total / len(provinces)
}

func averageFactionLoyalty(factions []Faction) int {
	if len(factions) == 0 {
		return 100
	}
	total := 0
	for _, f := range factions {
		total += f.Loyalty
	}
	return total / len(factions)
}

func averageWarProgress(wars []WarCampaign) int {
	if len(wars) == 0 {
		return 100
	}
	total := 0
	for _, war := range wars {
		total += war.Progress + max(0, 60-war.Threat)/3 + war.Supply/8 + war.Morale/8
	}
	return clamp(total/len(wars), 0, 100)
}

func warCanBreakDynasty(wars []WarCampaign, army int) bool {
	for _, war := range wars {
		if war.Stage != "凯旋" && war.Threat >= 92 && war.Supply <= 15 && army < 75 {
			return true
		}
	}
	return false
}

func clamp(v, low, high int) int {
	return int(math.Max(float64(low), math.Min(float64(high), float64(v))))
}
