package game

import "fmt"

type ForeignState struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ruler    string `json:"ruler"`
	Attitude string `json:"attitude"`
	Relation int    `json:"relation"`
	Threat   int    `json:"threat"`
	Tribute  int    `json:"tribute"`
	Leverage int    `json:"leverage"`
	Treaty   string `json:"treaty"`
	Envoy    string `json:"envoy"`
	Portrait string `json:"portrait"`
}

func startingForeignStates(dynastyID string) []ForeignState {
	states := []ForeignState{
		{ID: "beidi", Name: "北狄诸部", Ruler: "阿史那乌勒", Attitude: "骑墙观望", Relation: 36, Threat: 68, Tribute: 18, Leverage: 22, Envoy: "黑毡使", Portrait: "khan"},
		{ID: "xiyu", Name: "西域诸国", Ruler: "龟兹王女", Attitude: "逐利通商", Relation: 54, Threat: 36, Tribute: 35, Leverage: 42, Envoy: "胡商译官", Portrait: "diplomat"},
		{ID: "haiguo", Name: "东海诸岛", Ruler: "海国摄政", Attitude: "试探海贸", Relation: 48, Threat: 32, Tribute: 28, Leverage: 38, Envoy: "海舶正使", Portrait: "envoy"},
		{ID: "nanman", Name: "南岭盟寨", Ruler: "火藤大首领", Attitude: "边贸未定", Relation: 42, Threat: 44, Tribute: 22, Leverage: 30, Envoy: "银环使", Portrait: "rebel"},
	}
	switch dynastyID {
	case "xuanshuo":
		states[0].Threat += 12
		states[0].Relation -= 8
	case "jingyao":
		states[1].Relation += 10
		states[2].Tribute += 10
	case "dayin":
		states[3].Threat += 8
	case "chengping":
		states[2].Relation -= 6
	}
	return states
}

func (s *GameState) applyForeignOrder(req OrderRequest) (Effects, string, bool, error) {
	switch req.Kind {
	case OrderEmbassy:
		effects, summary, err := s.sendEmbassy(req.Target)
		return effects, summary, true, err
	case OrderTreaty:
		effects, summary, err := s.signTreaty(req.Target)
		return effects, summary, true, err
	default:
		return Effects{}, "", false, nil
	}
}

func (s *GameState) sendEmbassy(id string) (Effects, string, error) {
	i, ok := s.findForeignIndex(id)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown foreign state target %q", id)
	}
	foreign := s.ForeignStates[i]
	foreign.Relation = clamp(foreign.Relation+14+s.Stats.Charisma/20, 0, 100)
	foreign.Threat = clamp(foreign.Threat-8-s.Stats.Diplomacy/25, 0, 100)
	foreign.Leverage = clamp(foreign.Leverage+5, 0, 100)
	foreign.Attitude = foreignAttitude(foreign)
	s.ForeignStates[i] = foreign
	effects := Effects{Treasury: -5, Diplomacy: 5, BorderThreat: -2}
	return effects, fmt.Sprintf("你遣%s出使%s，带去金册、丝帛与两句留白。邦交升至%d，敌意降至%d。", foreign.Envoy, foreign.Name, foreign.Relation, foreign.Threat), nil
}

func (s *GameState) signTreaty(id string) (Effects, string, error) {
	i, ok := s.findForeignIndex(id)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown foreign state target %q", id)
	}
	foreign := s.ForeignStates[i]
	if foreign.Relation < 55 {
		return Effects{}, "", fmt.Errorf("%s relation is too low for treaty", foreign.Name)
	}
	foreign.Treaty = "互市盟约"
	foreign.Relation = clamp(foreign.Relation+8, 0, 100)
	foreign.Threat = clamp(foreign.Threat-10, 0, 100)
	foreign.Tribute = clamp(foreign.Tribute+10, 0, 100)
	foreign.Leverage = clamp(foreign.Leverage+12, 0, 100)
	foreign.Attitude = foreignAttitude(foreign)
	s.ForeignStates[i] = foreign
	s.shiftRelationsForDomain(DomainDiplomacy, 4, -2)
	effects := Effects{Treasury: -6, Diplomacy: 8, BorderThreat: -5, Stability: 1}
	return effects, fmt.Sprintf("%s在国书上落印，“%s”生效。贡贸升至%d，边境威胁降至%d。", foreign.Ruler, foreign.Treaty, foreign.Tribute, foreign.Threat), nil
}

func (s *GameState) applyForeignPressure(domain Domain) {
	if len(s.ForeignStates) == 0 {
		return
	}
	for i, foreign := range s.ForeignStates {
		if foreign.Treaty != "" {
			s.applyEffects(Effects{Treasury: foreign.Tribute / 25, Diplomacy: 1, BorderThreat: -1})
			foreign.Relation = clamp(foreign.Relation+1, 0, 100)
			foreign.Threat = clamp(foreign.Threat-1, 0, 100)
		} else {
			delta := 2
			if domain == DomainDiplomacy {
				delta = -2
			}
			foreign.Threat = clamp(foreign.Threat+delta, 0, 100)
			if foreign.Relation < 42 {
				foreign.Threat = clamp(foreign.Threat+2, 0, 100)
			}
		}
		foreign.Attitude = foreignAttitude(foreign)
		s.ForeignStates[i] = foreign
		if foreign.Threat >= 78 {
			s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+2, 0, 100)
			s.Crisis.Severity = clamp(s.Crisis.Severity+1, 0, 100)
		}
	}
}

func (s *GameState) ensureForeignSystems() {
	if len(s.ForeignStates) == 0 {
		s.ForeignStates = startingForeignStates(s.Dynasty.ID)
	}
}

func (s *GameState) findForeignIndex(id string) (int, bool) {
	for i, foreign := range s.ForeignStates {
		if foreign.ID == id {
			return i, true
		}
	}
	return 0, false
}

func foreignAttitude(foreign ForeignState) string {
	switch {
	case foreign.Treaty != "":
		return "盟约在案"
	case foreign.Threat >= 72:
		return "磨刀观望"
	case foreign.Relation >= 65:
		return "礼厚可交"
	case foreign.Relation <= 35:
		return "疑惧生隙"
	default:
		return "骑墙观望"
	}
}
