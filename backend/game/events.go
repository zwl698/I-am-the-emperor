package game

import "fmt"

type EventCategory string

const (
	EventStory     EventCategory = "story_arc"
	EventSystem    EventCategory = "system_pressure"
	EventMicroGame EventCategory = "micro_game"
)

type SeasonEvent struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Summary  string        `json:"summary"`
	Detail   string        `json:"detail"`
	Category EventCategory `json:"category"`
	Domain   Domain        `json:"domain"`
	Severity int           `json:"severity"`
	Check    string        `json:"check,omitempty"`
	Roll     int           `json:"roll,omitempty"`
	Target   int           `json:"target,omitempty"`
	Success  bool          `json:"success,omitempty"`
	Effects  Effects       `json:"effects"`
	Tags     []string      `json:"tags"`
	Art      string        `json:"art"`
	Portrait string        `json:"portrait"`
}

type eventTemplate struct {
	id             string
	title          string
	summary        string
	detail         string
	category       EventCategory
	domain         Domain
	severity       int
	check          string
	target         int
	successSummary string
	failSummary    string
	effects        Effects
	successEffects Effects
	failEffects    Effects
	tags           []string
	art            string
	portrait       string
	pressure       func(*GameState, Domain) int
	checkScore     func(*GameState) int
	after          func(*GameState, bool)
}

func (s *GameState) triggerSeasonalEvents(domain Domain) []SeasonEvent {
	if s.Phase != PhaseEmperor {
		s.RecentEvents = nil
		return nil
	}
	events := []SeasonEvent{
		s.resolveRandomEvent(pickEventTemplate(s, EventStory, domain, 3)),
		s.resolveRandomEvent(pickEventTemplate(s, EventSystem, domain, 7)),
		s.resolveRandomEvent(pickEventTemplate(s, EventMicroGame, domain, 11)),
	}
	s.RecentEvents = events
	s.EventLog = append(s.EventLog, events...)
	if len(s.EventLog) > 36 {
		s.EventLog = s.EventLog[len(s.EventLog)-36:]
	}
	return events
}

func (s *GameState) resolveRandomEventForTest(id string) SeasonEvent {
	for _, template := range randomEventTemplates() {
		if template.id == id {
			event := s.resolveRandomEvent(template)
			s.RecentEvents = []SeasonEvent{event}
			s.EventLog = append(s.EventLog, event)
			return event
		}
	}
	return SeasonEvent{}
}

func (s *GameState) resolveRandomEvent(template eventTemplate) SeasonEvent {
	success := true
	roll := 0
	target := template.target
	effects := template.effects
	summary := template.summary
	if template.category == EventMicroGame {
		roll = template.checkScore(s) + s.eventJitter(template.id)
		success = roll >= target
		if success {
			summary = template.successSummary
			effects = template.successEffects
		} else {
			summary = template.failSummary
			effects = template.failEffects
		}
	}
	event := SeasonEvent{
		ID:       template.id,
		Title:    template.title,
		Summary:  summary,
		Detail:   template.detail,
		Category: template.category,
		Domain:   template.domain,
		Severity: clamp(template.severity+template.pressure(s, template.domain)/12, 1, 100),
		Check:    template.check,
		Roll:     roll,
		Target:   target,
		Success:  success,
		Effects:  effects,
		Tags:     append([]string(nil), template.tags...),
		Art:      sceneArt(s, eventSceneIndex(template.domain)),
		Portrait: portraitKey(template.portrait),
	}
	s.applyEffects(effects)
	if template.after != nil {
		template.after(s, success)
	}
	s.History = append(s.History, HistoryEntry{
		Turn:    s.Turn,
		Age:     s.Age,
		Phase:   s.Phase,
		Choice:  "随机事件：" + template.title,
		Summary: event.Summary,
		Effects: effects,
	})
	return event
}

func pickEventTemplate(s *GameState, category EventCategory, domain Domain, salt int) eventTemplate {
	templates := randomEventTemplates()
	bestScore := -1
	var best eventTemplate
	for _, template := range templates {
		if template.category != category {
			continue
		}
		pressure := template.pressure(s, domain)
		score := pressure*10 + s.eventJitter(template.id) + salt
		if template.domain == domain {
			score += 35
		}
		if recentEventID(s.RecentEvents, template.id) {
			score -= 1000
		}
		if score > bestScore {
			bestScore = score
			best = template
		}
	}
	return best
}

func recentEventID(events []SeasonEvent, id string) bool {
	for _, event := range events {
		if event.ID == id {
			return true
		}
	}
	return false
}

func (s *GameState) eventJitter(id string) int {
	total := int(s.Seed%97) + s.Turn*17 + s.ReignYear*11 + seasonIndex(s.Season)*13
	for _, r := range id {
		total += int(r)
	}
	return total % 31
}

func randomEventTemplates() []eventTemplate {
	return []eventTemplate{
		{
			id: "frontier-elegy", title: "边城哀歌", category: EventStory, domain: DomainMilitary, severity: 48,
			summary: "北境童谣开始传进京城：雪线以北，母亲把战报缝进孩子棉衣。民间已经把边患想象成一场迟早落下的雪。",
			detail:  "剧情弧：外战压力会改变民间叙事，军务不再只是军力数字。",
			effects: Effects{Legitimacy: -1, BorderThreat: 2}, tags: []string{"战争", "民心"}, portrait: "general",
			pressure: func(s *GameState, d Domain) int {
				return s.Stats.BorderThreat + max(maxWarThreat(s.Wars), s.strategicMilitaryPressure())
			},
		},
		{
			id: "market-rumor", title: "市井银荒", category: EventStory, domain: DomainEconomy, severity: 42,
			summary: "京城钱铺开始收紧放账，茶楼里有人说国库的锁声比更鼓更响。财政危机会先变成谣言，再变成价格。",
			detail:  "剧情弧：财政不是单独资源，它会从商路、物价和民心上溢。",
			effects: Effects{Treasury: -2, Stability: -1}, tags: []string{"财政", "谣言"}, portrait: "merchant",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 90-s.Stats.Treasury) + factionPower(s.Factions, "merchant")/2
			},
		},
		{
			id: "heir-fable", title: "东宫寓言", category: EventStory, domain: DomainCourt, severity: 44,
			summary: "太学童子把几位皇嗣编进寓言：有人是玉，有人是火。童言传得太快，背后通常有大人的手。",
			detail:  "剧情弧：继承压力会通过传闻发酵，不等皇帝开口。",
			effects: Effects{Stability: -1, Influence: 1}, tags: []string{"继承", "后宫"}, portrait: "prince",
			pressure: func(s *GameState, d Domain) int { return s.Succession.Dispute + s.Succession.MaternalClanPower/2 },
		},
		{
			id: "flood-ballad", title: "河工夜谣", category: EventStory, domain: DomainDomestic, severity: 40,
			summary: "河工在夜里唱起新谣，唱的是水位，也是地方官吞下去的银子。灾情开始有了名字。",
			detail:  "剧情弧：灾害会把民政、吏治、财政三条线拧在一起。",
			effects: Effects{Populace: -1, Reform: 1}, tags: []string{"灾害", "地方"}, portrait: "engineer",
			pressure: func(s *GameState, d Domain) int {
				return worstProvinceDisaster(s.Provinces) + max(0, 65-averageProvinceOrder(s.Provinces))
			},
		},
		{
			id: "new-law-poem", title: "新法入诗", category: EventStory, domain: DomainReform, severity: 36,
			summary: "有年轻士子把新法写进诗里，也有老臣在诗后批了四个字：扰乱祖制。改革第一次拥有了拥趸和敌人。",
			detail:  "剧情弧：新法会形成社会声量，推动或反噬改革。",
			effects: Effects{Reform: 1, Stability: -1}, tags: []string{"新法", "士林"}, portrait: "poet",
			pressure: func(s *GameState, d Domain) int { return s.Stats.Reform + factionPower(s.Factions, "scholar")/2 },
		},
		{
			id: "office-backlog", title: "官署积牍", category: EventSystem, domain: DomainReform, severity: 52,
			summary: "中枢文书堆到窗下，几份急奏被压过了三日。官职不是名号，空转会吞掉帝国的反应速度。",
			detail:  "系统风暴：官署空转会伤稳定与新政，任免系统能缓解。",
			effects: Effects{Stability: -3, Reform: -2}, tags: []string{"官职", "行政"}, portrait: "minister",
			pressure: func(s *GameState, d Domain) int {
				return averageOfficeVacancyRisk(s.Offices) + max(0, 70-averageMinisterLoyalty(s.Court))
			},
			after: func(s *GameState, success bool) {
				for i := range s.Offices {
					s.Offices[i].VacancyRisk = clamp(s.Offices[i].VacancyRisk+3, 0, 100)
				}
			},
		},
		{
			id: "clan-petition", title: "外戚联名", category: EventSystem, domain: DomainCourt, severity: 54,
			summary: "几家母族同时递帖，语气恭顺，落款却整齐得像兵阵。后宫与继承开始合谋成一张网。",
			detail:  "系统风暴：外戚、宠爱、储位争议会互相推高。",
			effects: Effects{Influence: -2, Stability: -2}, tags: []string{"后宫", "继承"}, portrait: "empress",
			pressure: func(s *GameState, d Domain) int { return strongestConsortPower(s.Harem) + s.Succession.Dispute },
			after: func(s *GameState, success bool) {
				s.Succession.Dispute = clamp(s.Succession.Dispute+4, 0, 100)
			},
		},
		{
			id: "capital-tabloid", title: "京城小报", category: EventSystem, domain: DomainIntrigue, severity: 50,
			summary: "茶楼小报把宫中旧案、边镇军饷和东宫闲话写成连篇话本。传闻不再只是传闻，它正在替朝廷解释朝廷。",
			detail:  "系统风暴：舆论热度会放大案卷和密谋压力，禁谣能压火，宣判能转化口碑。",
			effects: Effects{Legitimacy: -1, Stability: -2}, tags: []string{"舆论", "谣言"}, portrait: "merchant",
			pressure: func(s *GameState, d Domain) int {
				return s.PublicOpinion.Rumor + maxLegalCaseHeat(s.LegalCases)/2 + len(s.Plots)*4
			},
			after: func(s *GameState, success bool) {
				s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor+5, 0, 100)
				s.PublicOpinion.Elite = clamp(s.PublicOpinion.Elite-2, 0, 100)
			},
		},
		{
			id: "ministry-case-deadline", title: "刑部限期", category: EventSystem, domain: DomainIntrigue, severity: 55,
			summary: "刑部、大理寺、都察院互相移文催案。案卷越拖，证词越多，派系也越有时间把自己摘出去。",
			detail:  "系统风暴：未结案件会持续升温，明审、宽赦或宣判能把热度转成明确后果。",
			effects: Effects{Influence: -1, Stability: -2}, tags: []string{"刑狱", "官职", "舆论"}, portrait: "minister",
			pressure: func(s *GameState, d Domain) int {
				return openLegalCasePressure(s.LegalCases) + averageOfficeVacancyRisk(s.Offices)/2
			},
			after: func(s *GameState, success bool) {
				for i := range s.LegalCases {
					if !s.LegalCases[i].Resolved {
						s.LegalCases[i].Heat = clamp(s.LegalCases[i].Heat+5, 0, 100)
					}
				}
				s.PublicOpinion.Justice = clamp(s.PublicOpinion.Justice-2, 0, 100)
			},
		},
		{
			id: "deserter-ledger", title: "逃卒名册", category: EventSystem, domain: DomainMilitary, severity: 58,
			summary: "兵部送来逃卒名册，墨迹未干。粮道一弱，士气就会先从纸面上掉下去。",
			detail:  "系统风暴：战争补给不足会直接拖累军力与危机钟。",
			effects: Effects{Army: -4, BorderThreat: 3, Stability: -1}, tags: []string{"战争", "粮道"}, portrait: "guard",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 90-activeWarSupply(s.Wars)) + max(maxWarThreat(s.Wars), s.strategicMilitaryPressure())
			},
			after: func(s *GameState, success bool) { s.adjustActiveWar(-2, -3, -5, 3, 0) },
		},
		{
			id: "grain-mold", title: "仓粮霉变", category: EventSystem, domain: DomainDomestic, severity: 48,
			summary: "南仓开封时有霉气扑面。粮草数字还在账上，百姓的碗却不会被账本填满。",
			detail:  "系统风暴：粮草与灾害共同决定民心安全线。",
			effects: Effects{Grain: -8, Populace: -2, Stability: -1}, tags: []string{"粮草", "灾害"}, portrait: "farmer",
			pressure: func(s *GameState, d Domain) int { return max(0, 85-s.Stats.Grain) + worstProvinceDisaster(s.Provinces) },
		},
		{
			id: "merchant-strike", title: "商帮停船", category: EventSystem, domain: DomainEconomy, severity: 46,
			summary: "漕运商帮忽然以修船为名停泊三日。没有人明说抗旨，但每只空船都在讨价还价。",
			detail:  "系统风暴：财政压榨和商帮权势会影响粮银周转。",
			effects: Effects{Treasury: -5, Grain: -3, Stability: -2}, tags: []string{"商帮", "财政"}, portrait: "merchant",
			pressure: func(s *GameState, d Domain) int {
				return factionPower(s.Factions, "merchant") + max(0, 60-factionLoyalty(s.Factions, "merchant"))
			},
		},
		{
			id: "audit-sprint", title: "御前考成", category: EventMicroGame, domain: DomainReform, severity: 50,
			detail: "微玩法：以新政、首辅能力、官署权威进行检定。通过则改革加速，失败则官场拖延。",
			check:  "新政 + 首辅能力 + 官署权威", target: 82,
			successSummary: "你在御前连问七案，首辅当场改票拟。官员第一次发现拖延也会被计入考绩。",
			failSummary:    "御前考成被旧例缠住，几位老臣用章程把锋芒磨钝。新法声势一时受挫。",
			successEffects: Effects{Reform: 5, Treasury: 2, Influence: 2},
			failEffects:    Effects{Reform: -2, Stability: -3},
			tags:           []string{"微玩法", "新法"}, portrait: "reformer",
			pressure: func(s *GameState, d Domain) int { return s.Stats.Reform + averageOfficeAuthority(s.Offices) },
			checkScore: func(s *GameState) int {
				return s.Stats.Reform/2 + ministerAbilityByRole(s.Court, "太傅")/2 + officeAuthority(s.Offices, "grand-secretariat")/4
			},
		},
		{
			id: "war-tabletop", title: "沙盘推演", category: EventMicroGame, domain: DomainMilitary, severity: 55,
			detail: "微玩法：以武略、军力、士气进行检定。通过可抢占战机，失败会暴露粮道。",
			check:  "武略 + 军力 + 前线士气", target: 86,
			successSummary: "沙盘上的红旗推进三寸，前线照此奇袭，敌军斥候被迫后撤。",
			failSummary:    "推演漏算了雪路，前线粮车慢了一日，战机从指缝里滑走。",
			successEffects: Effects{Army: 3, BorderThreat: -5, Martial: 1},
			failEffects:    Effects{Army: -3, BorderThreat: 4, Grain: -3},
			tags:           []string{"微玩法", "战争"}, portrait: "general",
			pressure: func(s *GameState, d Domain) int {
				return max(maxWarThreat(s.Wars), s.strategicMilitaryPressure()) + s.Stats.BorderThreat
			},
			checkScore: func(s *GameState) int { return s.Stats.Martial + s.Stats.Army/3 + activeWarMorale(s.Wars)/2 },
			after: func(s *GameState, success bool) {
				if success {
					s.adjustActiveWar(5, -1, 2, -4, 0)
				} else {
					s.adjustActiveWar(-1, -5, -4, 4, 0)
				}
			},
		},
		{
			id: "palace-whisper", title: "宫闱辨声", category: EventMicroGame, domain: DomainCourt, severity: 52,
			detail: "微玩法：以魅力、势力、储位稳定进行检定。通过可压下传闻，失败会扩大继承争议。",
			check:  "魅力 + 势力 + 储位稳定", target: 78,
			successSummary: "你在家宴上只问了三句话，几家母族便明白今夜不宜再传话。",
			failSummary:    "家宴笑声很满，散席后的耳语更满。东宫传闻从内廷漏到外朝。",
			successEffects: Effects{Stability: 3, Influence: 2},
			failEffects:    Effects{Stability: -3, Influence: -1},
			tags:           []string{"微玩法", "后宫"}, portrait: "consort",
			pressure:   func(s *GameState, d Domain) int { return s.Succession.Dispute + strongestConsortPower(s.Harem) },
			checkScore: func(s *GameState) int { return s.Stats.Charisma/2 + s.Stats.Influence/2 + s.Succession.Stability/3 },
			after: func(s *GameState, success bool) {
				delta := -5
				if !success {
					delta = 7
				}
				s.Succession.Dispute = clamp(s.Succession.Dispute+delta, 0, 100)
			},
		},
		{
			id: "granary-race", title: "驰驿开仓", category: EventMicroGame, domain: DomainDomestic, severity: 47,
			detail: "微玩法：以民心、粮草、地方秩序进行检定。通过则灾民回流，失败则赈济被截留。",
			check:  "民心 + 粮草 + 省份秩序", target: 76,
			successSummary: "驿马比雨云更快，仓门打开时，灾民还没来得及结成乱队。",
			failSummary:    "赈粮在路上被层层截留，灾民看见的是官旗，不是热粥。",
			successEffects: Effects{Populace: 5, Stability: 2, Grain: -3},
			failEffects:    Effects{Populace: -3, Stability: -4, Grain: -4},
			tags:           []string{"微玩法", "灾害"}, portrait: "farmer",
			pressure: func(s *GameState, d Domain) int { return worstProvinceDisaster(s.Provinces) + max(0, 90-s.Stats.Grain) },
			checkScore: func(s *GameState) int {
				return s.Stats.Populace/2 + s.Stats.Grain/3 + averageProvinceOrder(s.Provinces)/2
			},
		},
		{
			id: "spy-cipher", title: "密档译码", category: EventMicroGame, domain: DomainIntrigue, severity: 49,
			detail: "微玩法：以势力、学识、暗线官署进行检定。通过可提前发现阴谋，失败会误伤清议。",
			check:  "势力 + 学识 + 都察院权威", target: 80,
			successSummary: "密语被译出时，刺客还在等下一封信。你先一步拿到了名字。",
			failSummary:    "译码错了一字，牵连错了一人。恐惧仍有效，却不再精准。",
			successEffects: Effects{Influence: 5, Stability: 1, Health: 1},
			failEffects:    Effects{Influence: 2, Stability: -4, Legitimacy: -2},
			tags:           []string{"微玩法", "暗线"}, portrait: "spy",
			pressure: func(s *GameState, d Domain) int { return s.Crisis.Severity + maxFactionPower(s.Factions) },
			checkScore: func(s *GameState) int {
				return s.Stats.Influence/2 + s.Stats.Learning/2 + officeAuthority(s.Offices, "censorate")/3
			},
		},
		{
			id: "envoy-gambit", title: "使节对弈", category: EventMicroGame, domain: DomainDiplomacy, severity: 45,
			detail: "微玩法：以邦交、魅力、外战威胁进行检定。通过可换来缓冲，失败会被外邦试探。",
			check:  "邦交 + 魅力 - 边患压力", target: 72,
			successSummary: "你让使臣带去两封语气不同的国书，敌盟在回信前已经互相猜疑。",
			failSummary:    "外邦收下礼物，却把退兵二字写得含糊。边境斥候反而更多了。",
			successEffects: Effects{Diplomacy: 5, BorderThreat: -4},
			failEffects:    Effects{Treasury: -3, BorderThreat: 3, Diplomacy: -1},
			tags:           []string{"微玩法", "外交"}, portrait: "diplomat",
			pressure: func(s *GameState, d Domain) int { return s.Stats.BorderThreat + max(0, 80-s.Stats.Diplomacy) },
			checkScore: func(s *GameState) int {
				return s.Stats.Diplomacy/2 + s.Stats.Charisma/2 + max(0, 80-s.Stats.BorderThreat)/2
			},
		},
	}
}

func eventSceneIndex(domain Domain) int {
	switch domain {
	case DomainDomestic:
		return 6
	case DomainEconomy:
		return 7
	case DomainMilitary:
		return 14
	case DomainDiplomacy:
		return 28
	case DomainReform:
		return 10
	case DomainIntrigue:
		return 11
	case DomainCourt:
		return 16
	default:
		return 5
	}
}

func portraitKey(key string) string {
	if key == "" {
		return "emperor"
	}
	return key
}

func maxWarThreat(wars []WarCampaign) int {
	threat := 0
	for _, war := range wars {
		threat = max(threat, war.Threat)
	}
	return threat
}

func activeWarMorale(wars []WarCampaign) int {
	if len(wars) == 0 {
		return 60
	}
	for _, war := range wars {
		if war.Stage != "凯旋" {
			return war.Morale
		}
	}
	return wars[0].Morale
}

func averageOfficeAuthority(offices []Office) int {
	if len(offices) == 0 {
		return 0
	}
	total := 0
	for _, office := range offices {
		total += office.Authority
	}
	return total / len(offices)
}

func officeAuthority(offices []Office, id string) int {
	for _, office := range offices {
		if office.ID == id {
			return office.Authority
		}
	}
	return 0
}

func ministerAbilityByRole(court []Minister, role string) int {
	for _, minister := range court {
		if minister.Role == role {
			return minister.Ability
		}
	}
	return 0
}

func averageMinisterLoyalty(court []Minister) int {
	if len(court) == 0 {
		return 100
	}
	total := 0
	for _, minister := range court {
		total += minister.Loyalty
	}
	return total / len(court)
}

func strongestConsortPower(harem []Consort) int {
	power := 0
	for _, consort := range harem {
		power = max(power, consort.FamilyPower+consort.Ambition+consort.Influence)
	}
	return power / 3
}

func factionPower(factions []Faction, id string) int {
	for _, faction := range factions {
		if faction.ID == id {
			return faction.Power
		}
	}
	return 0
}

func factionLoyalty(factions []Faction, id string) int {
	for _, faction := range factions {
		if faction.ID == id {
			return faction.Loyalty
		}
	}
	return 50
}

func maxFactionPower(factions []Faction) int {
	power := 0
	for _, faction := range factions {
		power = max(power, faction.Power)
	}
	return power
}

func describeEvents(events []SeasonEvent) string {
	if len(events) == 0 {
		return ""
	}
	return fmt.Sprintf("另有%d件突发奏报入档。", len(events))
}
