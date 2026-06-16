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
	// CrisisBranch marks this event as a player-choice crisis event
	CrisisBranch     bool     `json:"crisisBranch,omitempty"`
	BranchID         string   `json:"branchId,omitempty"`
	Choices          []Choice `json:"choices,omitempty"`
	Resolved         bool     `json:"resolved,omitempty"`
	ResolvedChoiceID string   `json:"resolvedChoiceId,omitempty"`
	ResolvedOutcome  string   `json:"resolvedOutcome,omitempty"`
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

	// Season-specific event: each season has its own challenges
	if seasonal := s.generateSeasonalEvent(); seasonal.ID != "" {
		events = append(events, seasonal)
	}

	// Crisis branch: when crisis clock >= 5, generate a branching crisis event
	// that requires player choice and will influence next turn's events
	if s.Crisis.Clock >= 5 {
		if crisis := s.generateCrisisBranchEvent(); crisis.ID != "" {
			events = append(events, crisis)
		}
	}

	s.RecentEvents = events
	s.EventLog = append(s.EventLog, events...)
	if len(s.EventLog) > 48 {
		s.EventLog = s.EventLog[len(s.EventLog)-48:]
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
		// ─── 原有剧情事件 ───
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

		// ─── 新增剧情事件：腐败 / 吏治 ───
		{
			id: "corruption-web", title: "贪墨蛛网", category: EventStory, domain: DomainIntrigue, severity: 50,
			summary: "户部账簿被人翻出夹层：每笔修河银两都少了三成，多出来的部分流向了同一个姓氏。腐败不再是个人的手，它织成了一张网。",
			detail:  "剧情弧：高腐败环境下，贪墨会自我复制，从银两蔓延到官职和案卷。",
			effects: Effects{Treasury: -4, Stability: -2, Influence: -1}, tags: []string{"腐败", "吏治"}, portrait: "minister",
			pressure: func(s *GameState, d Domain) int {
				return openLegalCasePressure(s.LegalCases) + max(0, 70-averageMinisterLoyalty(s.Court))
			},
		},
		{
			id: "ghost-soldier", title: "鬼兵名册", category: EventStory, domain: DomainMilitary, severity: 53,
			summary: "兵部名册上有一千人在领饷，营地只看得见六百。吃空饷的人比上阵的人多，将军们在替死人领赏。",
			detail:  "剧情弧：军中腐败直接削弱军力，若不肃清，战时才知道手里没有兵。",
			effects: Effects{Army: -5, Treasury: -3, Stability: -1}, tags: []string{"腐败", "军务"}, portrait: "guard",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 90-s.Stats.Army) + averageOfficeVacancyRisk(s.Offices)/2
			},
		},
		{
			id: "sale-office", title: "卖官鬻爵", category: EventStory, domain: DomainReform, severity: 47,
			summary: "吏部铨选名册被人抄出市价：知县五千两，知府两万两。银子进的不是国库，而是几双看不见的手。",
			detail:  "剧情弧：官职成为商品时，政令就不再听从皇帝，而是听从出价人。",
			effects: Effects{Reform: -3, Stability: -2, Influence: -2}, tags: []string{"腐败", "官职"}, portrait: "minister",
			pressure: func(s *GameState, d Domain) int {
				return averageOfficeVacancyRisk(s.Offices) + max(0, 70-averageMinisterLoyalty(s.Court))
			},
		},

		// ─── 新增剧情事件：低民心 / 民变 ───
		{
			id: "starving-poem", title: "饿殍诗帖", category: EventStory, domain: DomainDomestic, severity: 46,
			summary: "城门上贴出一首无名诗：'朱门酒肉臭，路有冻死骨。'诗帖被撕了三次，又贴了四次。民心不在奏折里，在城门上。",
			detail:  "剧情弧：当民心跌破安全线，民间会用自己的方式让朝廷听见。",
			effects: Effects{Populace: -3, Stability: -2, Legitimacy: -1}, tags: []string{"民心", "舆论"}, portrait: "farmer",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 80-s.Stats.Populace) + worstProvinceDisaster(s.Provinces)/2
			},
		},
		{
			id: "refugee-tide", title: "流民潮涌", category: EventStory, domain: DomainDomestic, severity: 55,
			summary: "三省会合处的官道上，拖家带口的人走成了河。他们不喊冤，不告状，只是走。沉默比哭声更让地方官害怕。",
			detail:  "剧情弧：流民是无声的危机，不处置就会变成有组织的民变。",
			effects: Effects{Populace: -4, Stability: -3, Grain: -3}, tags: []string{"民心", "灾害"}, portrait: "farmer",
			pressure: func(s *GameState, d Domain) int {
				return worstProvinceDisaster(s.Provinces) + max(0, 70-s.Stats.Grain) + max(0, 60-s.Stats.Populace)
			},
		},
		{
			id: "village-militia", title: "乡勇结社", category: EventStory, domain: DomainMilitary, severity: 51,
			summary: "南方几县百姓自组乡勇，名义上防匪，实际上已不受县令节制。民间武装出现的速度比朝廷调兵快。",
			detail:  "剧情弧：朝廷管不到的地方，民间会自己管。但这把刀有两面。",
			effects: Effects{Army: -2, Stability: -3, Populace: 2}, tags: []string{"民变", "军务"}, portrait: "guard",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 65-s.Stats.Stability) + worstProvinceDisaster(s.Provinces)/2 + max(0, 60-averageProvinceOrder(s.Provinces))
			},
		},

		// ─── 新增剧情事件：皇帝特质触发 ───
		{
			id: "paranoia-night", title: "烛影惊心", category: EventStory, domain: DomainCourt, severity: 52,
			summary: "深夜你听见殿外脚步声，叫来禁军搜查却什么也没有。但你知道——也许今晚没有，不代表明晚没有。",
			detail:  "剧情弧：偏执的皇帝会在深夜制造自己的敌人。每多疑一次，真正的忠诚就少一分。",
			effects: Effects{Health: -2, Stability: -2, Influence: 1}, tags: []string{"皇帝特质", "多疑"}, portrait: "emperor",
			pressure: func(s *GameState, d Domain) int {
				if s.HasTrait(TraitParanoid) || s.HasTrait(TraitSuspicious) {
					return s.Crisis.Severity + maxFactionPower(s.Factions)/2 + 30
				}
				return 0
			},
		},
		{
			id: "vanity-monument", title: "功德碑文", category: EventStory, domain: DomainEconomy, severity: 44,
			summary: "工部呈上碑文草稿：'圣德巍巍，四方来朝。'石料从西山运来，工匠从三省征调。百姓说：碑比城墙高。",
			detail:  "剧情弧：好大喜功的皇帝会在盛世中提前消费未来。碑立起来时，国库也在空下去。",
			effects: Effects{Treasury: -6, Legitimacy: 2, Populace: -2}, tags: []string{"皇帝特质", "好大喜功"}, portrait: "emperor",
			pressure: func(s *GameState, d Domain) int {
				if s.HasTrait(TraitVainglorious) {
					return s.Stats.Legitimacy/2 + max(0, 80-s.Stats.Treasury) + 25
				}
				return 0
			},
		},
		{
			id: "scholar-debate", title: "经筵论道", category: EventStory, domain: DomainReform, severity: 38,
			summary: "太学新院的学子们连续论道七日，从经典辩到新政。你旁听了一下午，有人在引用你的话反驳你。",
			detail:  "剧情弧：好学的皇帝会催生一个敢于辩论的士林，他们尊敬你，但不会盲从你。",
			effects: Effects{Learning: 2, Reform: 2, Stability: -1}, tags: []string{"皇帝特质", "好学"}, portrait: "poet",
			pressure: func(s *GameState, d Domain) int {
				if s.HasTrait(TraitScholarly) {
					return s.Stats.Learning + s.Stats.Reform/2 + 20
				}
				return 0
			},
		},
		{
			id: "ruthless-purge-tale", title: "缇骑传说", category: EventStory, domain: DomainIntrigue, severity: 54,
			summary: "民间开始传一个新词：'缇骑夜至'。没有人知道他们是谁，只知道天亮后某个人就不见了。恐惧是一种统治，但它有保质期。",
			detail:  "剧情弧：铁腕的皇帝拥有最快的安全感，也拥有最快的民心流失。",
			effects: Effects{Influence: 3, Populace: -3, Stability: -2}, tags: []string{"皇帝特质", "铁腕"}, portrait: "spy",
			pressure: func(s *GameState, d Domain) int {
				if s.HasTrait(TraitRuthless) {
					return s.Stats.Influence + max(0, 70-s.Stats.Populace) + 20
				}
				return 0
			},
		},
		{
			id: "frugal-plain-meal", title: "素衣朝会", category: EventStory, domain: DomainCourt, severity: 32,
			summary: "你穿素服上朝，取消御膳，令后宫减用。大臣们面面相觑——有人羞愧，有人害怕，有人在计算这是不是试探。",
			detail:  "剧情弧：节俭是美德，但朝廷的排场也是权力的象征。缩减到什么程度，由你决定。",
			effects: Effects{Treasury: 4, Charisma: -1, Legitimacy: 1}, tags: []string{"皇帝特质", "节俭"}, portrait: "emperor",
			pressure: func(s *GameState, d Domain) int {
				if s.HasTrait(TraitFrugal) {
					return max(0, 90-s.Stats.Treasury) + 15
				}
				return 0
			},
		},

		// ─── 新增剧情事件：老年 / 健康 ───
		{
			id: "dragon-illness", title: "龙体欠安", category: EventStory, domain: DomainCourt, severity: 58,
			summary: "早朝时你咳了三声，群臣跪了一地。太医把脉后跪得更低。每一声咳嗽，朝堂的天平就会晃一下。",
			detail:  "剧情弧：皇帝的健康是朝堂最大的变量，比边患和粮荒更难预测。",
			effects: Effects{Health: -3, Stability: -3, Influence: -2}, tags: []string{"健康", "朝堂"}, portrait: "emperor",
			pressure: func(s *GameState, d Domain) int {
				if s.Age >= 50 || s.Stats.Health < 50 {
					return max(0, 100-s.Stats.Health) + (s.Age-40)*2 + s.Succession.Dispute/3
				}
				return 0
			},
		},
		{
			id: "abdication-whisper", title: "禅位风声", category: EventStory, domain: DomainCourt, severity: 60,
			summary: "朝中开始有人低声议论'内禅'。没有人敢当面提，但东宫的灯比御书房亮得更晚了。权力不需要移交，它自己会找下一个人。",
			detail:  "剧情弧：禅位压力既是健康问题，也是继承问题，更是权力问题。",
			effects: Effects{Influence: -3, Stability: -2, Legitimacy: -1}, tags: []string{"健康", "继承"}, portrait: "prince",
			pressure: func(s *GameState, d Domain) int {
				if s.Condition.AbdicationRisk >= 30 {
					return s.Condition.AbdicationRisk + s.Succession.Dispute/2 + max(0, 100-s.Stats.Health)/2
				}
				return 0
			},
		},

		// ─── 新增系统压力事件 ───
		{
			id: "plague-spread", title: "疫疠蔓延", category: EventSystem, domain: DomainDomestic, severity: 56,
			summary: "城外义庄一夜满了三成，太医院请旨封锁疫区。疫病不认官衔，也不等奏报。",
			detail:  "系统风暴：疫病会同时消耗民心、粮草和军力，是唯一能同时打击三条线的系统事件。",
			effects: Effects{Populace: -5, Army: -3, Grain: -4, Stability: -3}, tags: []string{"疫病", "灾害"}, portrait: "farmer",
			pressure: func(s *GameState, d Domain) int {
				return worstProvinceDisaster(s.Provinces) + max(0, 65-averageProvinceOrder(s.Provinces)) + max(0, 70-s.Stats.Populace)
			},
		},
		{
			id: "faction-collision", title: "朋党火并", category: EventSystem, domain: DomainCourt, severity: 57,
			summary: "两派大臣在朝会上当面撕破奏折，互相弹劾。朋党之争终于从暗处走到了明处，而皇帝的裁决只能让一方活下去。",
			detail:  "系统风暴：派系火并会撕裂朝稳，每一次弹劾都会留下仇恨的遗产。",
			effects: Effects{Stability: -5, Influence: -2, Reform: -2}, tags: []string{"朋党", "朝堂"}, portrait: "minister",
			pressure: func(s *GameState, d Domain) int {
				return maxFactionPower(s.Factions) + max(0, 70-averageMinisterLoyalty(s.Court))
			},
			after: func(s *GameState, success bool) {
				for i := range s.Court {
					s.Court[i].Loyalty = clamp(s.Court[i].Loyalty-3, 0, 100)
					s.Court[i].Stress = clamp(s.Court[i].Stress+4, 0, 100)
				}
			},
		},
		{
			id: "trade-route-break", title: "商路断裂", category: EventSystem, domain: DomainEconomy, severity: 50,
			summary: "西线商队全部折返：路匪、关卡、战火各占一成原因，七成原因没人说出口。商路是王朝的血管，断一条就白一截。",
			detail:  "系统风暴：商路断裂会影响财政、外交和民心三个维度。",
			effects: Effects{Treasury: -7, Diplomacy: -2, Populace: -2}, tags: []string{"商路", "财政"}, portrait: "merchant",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 80-s.Stats.Treasury) + s.Stats.BorderThreat/3 + factionPower(s.Factions, "merchant")/2
			},
		},
		{
			id: "eunuch-power", title: "内侍干政", category: EventSystem, domain: DomainCourt, severity: 53,
			summary: "司礼监的批红越来越像圣旨，外朝的奏折先经内廷再转内阁。内侍的影子比大臣更长，而影子不会上奏谢恩。",
			detail:  "系统风暴：宦官势力会侵蚀朝稳和改革进度，且很难用常规手段削弱。",
			effects: Effects{Influence: -3, Stability: -3, Reform: -2}, tags: []string{"宦官", "朝堂"}, portrait: "minister",
			pressure: func(s *GameState, d Domain) int {
				return s.Crisis.Severity + max(0, 65-s.Stats.Influence) + s.Succession.Dispute/2
			},
			after: func(s *GameState, success bool) {
				if !success {
					for i := range s.Offices {
						s.Offices[i].Authority = clamp(s.Offices[i].Authority-3, 0, 100)
					}
				}
			},
		},
		{
			id: "foreign-collusion", title: "外邦合谋", category: EventSystem, domain: DomainDiplomacy, severity: 55,
			summary: "密探截获两邦互通密信，措辞客气但内容惊人：他们准备同时施压边境，看朝廷顾哪头。",
			detail:  "系统风暴：外邦合谋会使边患翻倍，同时牵制外交和军力。",
			effects: Effects{BorderThreat: 6, Diplomacy: -3, Stability: -2}, tags: []string{"外交", "边患"}, portrait: "diplomat",
			pressure: func(s *GameState, d Domain) int {
				return maxForeignThreat(s.ForeignStates) + s.Stats.BorderThreat + max(0, 70-s.Stats.Diplomacy)
			},
		},
		{
			id: "tax-evasion-ring", title: "抗税同盟", category: EventSystem, domain: DomainEconomy, severity: 49,
			summary: "三省士绅联名上书'请免苛税'，附上了厚厚一叠'民意'。抗税的不是百姓，是把百姓当挡箭牌的人。",
			detail:  "系统风暴：有组织的抗税会让国库雪上加霜，强行征税则会推高民怨。",
			effects: Effects{Treasury: -6, Stability: -2, Reform: -1}, tags: []string{"财政", "士绅"}, portrait: "merchant",
			pressure: func(s *GameState, d Domain) int {
				return max(0, 85-s.Stats.Treasury) + factionPower(s.Factions, "merchant")/2 + max(0, 60-s.Stats.Populace)
			},
		},

		// ─── 新增微玩法事件 ───
		{
			id: "imperial-exam", title: "殿试策问", category: EventMicroGame, domain: DomainReform, severity: 44,
			detail: "微玩法：以学识、新政、士林舆论进行检定。通过可选拔良才，失败则科场舞弊。",
			check:  "学识 + 新政 + 士林舆论", target: 78,
			successSummary: "你亲拟策问题目，殿试之上无一人敢舞弊。新科进士中有三人的策论让你读了两遍。",
			failSummary:    "科场传出夹带消息，你不得不取消三人的名次。士林议论的不是你的严明，而是你的失察。",
			successEffects: Effects{Learning: 3, Reform: 4, Stability: 2, Legitimacy: 1},
			failEffects:    Effects{Learning: -1, Stability: -3, Legitimacy: -2},
			tags:           []string{"微玩法", "科举"}, portrait: "poet",
			pressure: func(s *GameState, d Domain) int {
				return s.Stats.Learning + s.Stats.Reform + factionPower(s.Factions, "scholar")/2
			},
			checkScore: func(s *GameState) int {
				return s.Stats.Learning/2 + s.Stats.Reform/3 + factionLoyalty(s.Factions, "scholar")/3
			},
		},
		{
			id: "emergency-relief", title: "急赈调度", category: EventMicroGame, domain: DomainDomestic, severity: 52,
			detail: "微玩法：以民心、国库、官署效率进行检定。通过可精准赈灾，失败则赈银被截留。",
			check:  "民心 + 国库 + 官署效率", target: 80,
			successSummary: "你绕过地方官，直派钦差押银到县。银两在灾民手里，不在中间人的账上。",
			failSummary:    "赈银到县时少了四成，剩下的六成被'暂借'修堤。灾民只看到官员的轿子。",
			successEffects: Effects{Populace: 6, Treasury: -4, Stability: 3},
			failEffects:    Effects{Populace: -3, Treasury: -6, Stability: -4},
			tags:           []string{"微玩法", "赈灾"}, portrait: "farmer",
			pressure: func(s *GameState, d Domain) int {
				return worstProvinceDisaster(s.Provinces) + max(0, 70-s.Stats.Populace) + max(0, 60-s.Stats.Grain)
			},
			checkScore: func(s *GameState) int {
				return s.Stats.Populace/3 + s.Stats.Treasury/4 + averageOfficeAuthority(s.Offices)/3
			},
		},
		{
			id: "court-debate", title: "朝堂廷辩", category: EventMicroGame, domain: DomainCourt, severity: 48,
			detail: "微玩法：以魅力、势力、学识进行检定。通过可压制异议统一路线，失败则朝议分裂。",
			check:  "魅力 + 势力 + 学识", target: 76,
			successSummary: "你在廷辩中引经据典，把反对派说得哑口无言。散朝时，连反对者都在抄你的原话。",
			failSummary:    "廷辩中你被三位老臣联手驳得说不出话。朝议不欢而散，各派回去写自己的奏折。",
			successEffects: Effects{Influence: 4, Stability: 3, Reform: 2},
			failEffects:    Effects{Influence: -2, Stability: -4, Reform: -2},
			tags:           []string{"微玩法", "朝堂"}, portrait: "emperor",
			pressure: func(s *GameState, d Domain) int {
				return s.Stats.Influence + maxFactionPower(s.Factions)/2 + s.Crisis.Severity/2
			},
			checkScore: func(s *GameState) int {
				return s.Stats.Charisma/3 + s.Stats.Influence/3 + s.Stats.Learning/3
			},
		},
		{
			id: "secret-interrogation", title: "密审暗线", category: EventMicroGame, domain: DomainIntrigue, severity: 54,
			detail: "微玩法：以势力、学识、密谋压力进行检定。通过可获取关键情报，失败则打草惊蛇。",
			check:  "势力 + 学识 + 密谋反制", target: 84,
			successSummary: "你在密室中审了三个时辰，犯人终于交代了上线。这张网比你想的更大，但至少现在你有了线头。",
			failSummary:    "审讯走了风声，主犯连夜出城。你拿到了一具空壳，和一条更难追的暗线。",
			successEffects: Effects{Influence: 6, Stability: 2, Health: -1},
			failEffects:    Effects{Influence: -3, Stability: -3, Legitimacy: -2},
			tags:           []string{"微玩法", "暗线"}, portrait: "spy",
			pressure: func(s *GameState, d Domain) int {
				return maxPlotProgress(s.Plots) + s.PublicOpinion.Rumor/2 + s.Crisis.Severity/2
			},
			checkScore: func(s *GameState) int {
				return s.Stats.Influence/2 + s.Stats.Learning/3 + max(0, 100-maxPlotProgress(s.Plots))/3
			},
		},
		{
			id: "military-inspection", title: "阅兵大阅", category: EventMicroGame, domain: DomainMilitary, severity: 46,
			detail: "微玩法：以武略、军力、士气进行检定。通过可振奋军心，失败则暴露军中弊病。",
			check:  "武略 + 军力 + 士气", target: 75,
			successSummary: "你骑马检阅三军，长枪如林。军心在呼号声中凝聚，连老将都说：这一代皇帝，上过阵。",
			failSummary:    "阅兵时一匹战马受惊，冲乱了方阵。你看见的不是军威，是训练不足和粮草短缺的痕迹。",
			successEffects: Effects{Army: 4, BorderThreat: -3, Martial: 1, Legitimacy: 1},
			failEffects:    Effects{Army: -2, Stability: -2, BorderThreat: 2},
			tags:           []string{"微玩法", "军务"}, portrait: "general",
			pressure: func(s *GameState, d Domain) int {
				return s.Stats.BorderThreat + max(0, 80-s.Stats.Army) + max(maxWarThreat(s.Wars), s.strategicMilitaryPressure())/2
			},
			checkScore: func(s *GameState) int {
				return s.Stats.Martial/2 + s.Stats.Army/3 + activeWarMorale(s.Wars)/3
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
	branchCount := 0
	for _, e := range events {
		if e.CrisisBranch {
			branchCount++
		}
	}
	if branchCount > 0 {
		return fmt.Sprintf("另有%d件突发奏报入档，其中%d件需要圣裁。", len(events), branchCount)
	}
	return fmt.Sprintf("另有%d件突发奏报入档。", len(events))
}

// ──────────────────────────────────────────────
// 季节特有事件：每个季节有不同的挑战
// ──────────────────────────────────────────────

type seasonalEventDef struct {
	id       string
	title    string
	summary  string
	domain   Domain
	effects  Effects
	tags     []string
	portrait string
}

func seasonalEvents() map[string][]seasonalEventDef {
	return map[string][]seasonalEventDef{
		"春": {
			{id: "spring-plowing", title: "春耕令", summary: "冰雪消融，各省春耕开始。农时一误，全年粮草都要打折。你准许减免农税，还是催缴旧欠？",
				domain: DomainDomestic, effects: Effects{Grain: 5, Populace: 3, Treasury: -4},
				tags: []string{"季节", "农业"}, portrait: "farmer"},
			{id: "spring-exam", title: "春闱放榜", summary: "贡院放榜，新科进士候旨入仕。新血能注入改革力量，也会冲击旧有官署格局。",
				domain: DomainReform, effects: Effects{Learning: 2, Reform: 3, Stability: -1},
				tags: []string{"季节", "科举"}, portrait: "poet"},
		},
		"夏": {
			{id: "summer-flood", title: "夏汛预警", summary: "梅雨连绵，南河水位上涨。往年此时，堤坝上最少也得有万人巡守。",
				domain: DomainDomestic, effects: Effects{Grain: -4, Treasury: -3, Populace: -2},
				tags: []string{"季节", "灾害"}, portrait: "engineer"},
			{id: "summer-locust", title: "蝗影东来", summary: "山东急报蝗群过境，所到之处寸草不生。粮价已经开始波动。",
				domain: DomainDomestic, effects: Effects{Grain: -8, Populace: -3, Stability: -2},
				tags: []string{"季节", "灾害"}, portrait: "farmer"},
		},
		"秋": {
			{id: "autumn-tax", title: "秋粮解京", summary: "各省秋粮开始解运京师。丰年则仓满，灾年则奏折比银车先到。",
				domain: DomainEconomy, effects: Effects{Treasury: 8, Grain: 6, Populace: -2},
				tags: []string{"季节", "税收"}, portrait: "merchant"},
			{id: "autumn-harvest-festival", title: "秋社祭典", summary: "民间秋社，百姓叩谢天地。一场体面的祭典能收拢民心，但库房不等人。",
				domain: DomainCourt, effects: Effects{Populace: 4, Legitimacy: 2, Treasury: -3},
				tags: []string{"季节", "礼制"}, portrait: "emperor"},
		},
		"冬": {
			{id: "winter-fuel", title: "炭价飞涨", summary: "入冬炭价翻了三倍，京城百姓开始劈旧家具取暖。如果不管，民心会像炭一样烧尽。",
				domain: DomainDomestic, effects: Effects{Populace: -3, Treasury: -4, Stability: -2},
				tags: []string{"季节", "民生"}, portrait: "farmer"},
			{id: "winter-border-raid", title: "雪夜掠边", summary: "北境游骑趁雪夜犯边，边镇烽燧在暴雪中沉默。春暖之前，边境消息会断断续续。",
				domain: DomainMilitary, effects: Effects{BorderThreat: 5, Army: -2, Stability: -2},
				tags: []string{"季节", "边患"}, portrait: "guard"},
		},
	}
}

func (s *GameState) generateSeasonalEvent() SeasonEvent {
	defs, ok := seasonalEvents()[s.Season]
	if !ok || len(defs) == 0 {
		return SeasonEvent{}
	}
	s.ensureRNG()
	pick := s.rng.Intn(len(defs))
	d := defs[pick]
	event := SeasonEvent{
		ID:       fmt.Sprintf("%s-%d-%s", d.id, s.Turn, s.Season),
		Title:    d.title,
		Summary:  d.summary,
		Detail:   fmt.Sprintf("季节事件：%s季特有挑战，忽视会让局势恶化。", s.Season),
		Category: EventSystem,
		Domain:   d.domain,
		Severity: clamp(35+s.Crisis.Severity/4, 20, 80),
		Effects:  d.effects,
		Tags:     append([]string(nil), d.tags...),
		Art:      sceneArt(s, eventSceneIndex(d.domain)),
		Portrait: portraitKey(d.portrait),
	}
	s.applyEffects(d.effects)
	s.History = append(s.History, HistoryEntry{
		Turn: s.Turn, Age: s.Age, Phase: s.Phase,
		Choice: "季节事件：" + d.title, Summary: d.summary, Effects: d.effects,
	})
	return event
}

// ──────────────────────────────────────────────
// 危机分支事件：需要玩家选择，影响后续局势
// ──────────────────────────────────────────────

type crisisBranchDef struct {
	id        string
	title     string
	summary   string
	domain    Domain
	severity  int
	branchA   Choice
	branchB   Choice
	triggerFn func(s *GameState) bool
	portrait  string
}

func crisisBranchDefs() []crisisBranchDef {
	return []crisisBranchDef{
		// ─── 原有危机分支 ───
		{
			id: "crisis-grain-riot", title: "粮荒民变", domain: DomainDomestic, severity: 72,
			summary: "多地粮仓告急，流民开始冲击县衙。是开仓放粮稳住局面，还是调兵弹压以儆效尤？",
			branchA: Choice{ID: "crisis-grain-riot-open", Text: "开仓放粮，安抚灾民",
				Detail: "损失粮银，换回民心与稳定。",
				Domain: DomainDomestic, Effects: Effects{Grain: -12, Treasury: -6, Populace: 8, Stability: 6},
				Outcome: "粮车从常平仓驶出，灾民跪在路旁。这一夜的代价，明天会记在账上。"},
			branchB: Choice{ID: "crisis-grain-riot-suppress", Text: "调兵弹压，以儆效尤",
				Detail: "保住粮银，但民怨加深。",
				Domain: DomainMilitary, Effects: Effects{Army: -3, Populace: -8, Stability: -4, Influence: 3},
				Outcome: "军靴踏过街市，聚众散去，怨恨没有散去。它会变成下一封没有署名的揭帖。"},
			triggerFn: func(s *GameState) bool { return s.Stats.Grain < 35 || worstProvinceDisaster(s.Provinces) >= 50 },
			portrait:  "farmer",
		},
		{
			id: "crisis-border-invasion", title: "边关告急", domain: DomainMilitary, severity: 78,
			summary: "敌军大举压境，边镇求援急报一日三至。是集中兵力决战，还是弃地固守待援？",
			branchA: Choice{ID: "crisis-border-counterattack", Text: "调集主力，御驾亲征之势决战",
				Detail: "高风险高回报，胜则边患大减，败则动摇国本。",
				Domain: DomainMilitary, Effects: Effects{Treasury: -14, Grain: -10, Army: 8, BorderThreat: -12, Martial: 2},
				Outcome: "大军出塞，旌旗遮日。若胜，边患十年可缓；若败，京城再无余兵可用。"},
			branchB: Choice{ID: "crisis-border-defend", Text: "收缩防线，固守待援",
				Detail: "稳妥但丧失主动，边境百姓将承受兵祸。",
				Domain: DomainMilitary, Effects: Effects{Treasury: -6, BorderThreat: -4, Populace: -4, Stability: -2},
				Outcome: "烽燧内撤，百姓随军南迁。失地可以再夺，人命不能复生。"},
			triggerFn: func(s *GameState) bool { return s.Stats.BorderThreat >= 65 || maxWarThreat(s.Wars) >= 70 },
			portrait:  "general",
		},
		{
			id: "crisis-court-coup", title: "宫廷惊变", domain: DomainIntrigue, severity: 75,
			summary: "密探截获废立密诏草稿，朝中重臣牵涉其中。是先发制人清除异己，还是以退为进安抚人心？",
			branchA: Choice{ID: "crisis-court-purge", Text: "先发制人，连夜清洗",
				Detail: "快速消除威胁，但会制造恐惧与怨恨。",
				Domain: DomainIntrigue, Effects: Effects{Influence: 10, Stability: -8, Legitimacy: -4, Health: -2},
				Outcome: "夜色中缇骑四出，密谋者还没来得及烧掉证据。恐惧有效，但忠诚需要更久才能重建。"},
			branchB: Choice{ID: "crisis-court-appease", Text: "以退为进，召见安抚",
				Detail: "降低对立，但密谋可能继续发酵。",
				Domain: DomainDiplomacy, Effects: Effects{Stability: 4, Influence: -4, Legitimacy: 2},
				Outcome: "你在御书房单独召见几位重臣，没有点破，只是给了一条退路。他们会感激，还是会觉得你软弱？"},
			triggerFn: func(s *GameState) bool { return s.Crisis.Severity >= 60 || maxFactionPower(s.Factions) >= 60 },
			portrait:  "spy",
		},
		// ─── 新增危机分支：宦官专权 ───
		{
			id: "crisis-eunuch-usurp", title: "阉党乱政", domain: DomainCourt, severity: 74,
			summary: "司礼监总管已经代批奏章三月有余，六部公文先过内廷再入内阁。是雷霆手段清除阉党，还是借力打力分化内侍？",
			branchA: Choice{ID: "crisis-eunuch-purge", Text: "雷霆手段，清除阉党",
				Detail: "快速夺回批红权，但可能激起内廷反扑。",
				Domain: DomainIntrigue, Effects: Effects{Influence: 8, Stability: -6, Legitimacy: 3, Health: -2},
				Outcome: "你连下三道旨意收回批红权，内廷总管被发配守陵。但宫中人心惶惶，不知下一个会是谁。"},
			branchB: Choice{ID: "crisis-eunuch-divide", Text: "借力打力，分化内侍",
				Detail: "用内侍制衡外朝，但会让宦官势力更深。",
				Domain: DomainCourt, Effects: Effects{Influence: 2, Stability: 2, Reform: -3},
				Outcome: "你提拔了两位新总管互相牵制，批红权暂时回到了你手中。但宫中的暗流比以前更深了。"},
			triggerFn: func(s *GameState) bool { return s.Crisis.Severity >= 55 && max(0, 65-s.Stats.Influence) >= 25 },
			portrait:  "minister",
		},
		// ─── 新增危机分支：藩镇割据 ───
		{
			id: "crisis-warlord", title: "藩镇拥兵", domain: DomainMilitary, severity: 76,
			summary: "边镇大将拥兵自重，截留赋税、自署官吏，奏折中只称'臣'不称'奴才'。是削藩收权不惜一战，还是赐爵安抚换取表面臣服？",
			branchA: Choice{ID: "crisis-warlord-suppress", Text: "削藩收权，不惜一战",
				Detail: "决心铲除割据，但内战会消耗大量国力。",
				Domain: DomainMilitary, Effects: Effects{Treasury: -16, Grain: -12, Army: -6, BorderThreat: 8, Influence: 6},
				Outcome: "削藩令下达当日，三镇起兵两镇观望。这是一场赌上国运的豪赌，赢了天下归心，输了四分五裂。"},
			branchB: Choice{ID: "crisis-warlord-appease", Text: "赐爵安抚，换取臣服",
				Detail: "暂时稳住局面，但藩镇会变得更难撼动。",
				Domain: DomainDiplomacy, Effects: Effects{Treasury: -8, Influence: -4, Stability: 3, Legitimacy: -3},
				Outcome: "你赐下铁券金书，藩镇谢恩的表文写得很漂亮。但所有人都知道，这张纸挡不住刀。"},
			triggerFn: func(s *GameState) bool { return s.Stats.Army < 50 && s.Stats.BorderThreat >= 50 && s.Crisis.Clock >= 4 },
			portrait:  "general",
		},
		// ─── 新增危机分支：国库枯竭 ───
		{
			id: "crisis-treasury-empty", title: "国库枯竭", domain: DomainEconomy, severity: 70,
			summary: "户部尚书跪呈空账：国库仅余三月之银，而各省秋粮尚未解运。是抄没贪官家产充公，还是向商帮借银以解燃眉？",
			branchA: Choice{ID: "crisis-treasury-confiscate", Text: "抄没贪官，杀鸡取卵",
				Detail: "快速充盈国库，但朝中人人自危。",
				Domain: DomainIntrigue, Effects: Effects{Treasury: 18, Stability: -8, Influence: -4, Reform: -2},
				Outcome: "抄家令一下，三品以上官员各怀鬼胎。国库确实满了，但满朝文武的心空了。"},
			branchB: Choice{ID: "crisis-treasury-borrow", Text: "向商帮借银，以盐引作抵",
				Detail: "解燃眉之急，但商帮势力会进一步膨胀。",
				Domain: DomainEconomy, Effects: Effects{Treasury: 12, Stability: 2, Influence: -3, Reform: -1},
				Outcome: "商帮痛快地借了银子，条件是未来三年的盐引专营。你解决了今天的危机，预支了明天的权力。"},
			triggerFn: func(s *GameState) bool { return s.Stats.Treasury <= 25 },
			portrait:  "merchant",
		},
		// ─── 新增危机分支：瘟疫入京 ───
		{
			id: "crisis-plague-capital", title: "瘟疫入京", domain: DomainDomestic, severity: 80,
			summary: "京城出现第一例疫病死者，太医院束手无策。是封锁九门严防扩散，还是开放义诊广施草药？",
			branchA: Choice{ID: "crisis-plague-quarantine", Text: "封锁九门，严防扩散",
				Detail: "有效遏制疫病，但城中百姓将承受困守之苦。",
				Domain: DomainDomestic, Effects: Effects{Populace: -6, Treasury: -8, Stability: -3, Grain: -5, Health: 2},
				Outcome: "九门落锁，京城变成了一座巨大的笼子。疫病死了一百人，困守饿死了三百人。数字不会说话，但史书会。"},
			branchB: Choice{ID: "crisis-plague-treat", Text: "开放义诊，广施草药",
				Detail: "体恤百姓，但疫病可能扩散到更广范围。",
				Domain: DomainDomestic, Effects: Effects{Populace: 4, Treasury: -10, Stability: 2, Health: -3, Army: -4},
				Outcome: "义诊棚搭了七座，草药施了三千剂。疫病没有停住，但百姓看见了皇帝的心。代价是军营也病了。"},
			triggerFn: func(s *GameState) bool { return worstProvinceDisaster(s.Provinces) >= 60 && s.Stats.Populace < 45 },
			portrait:  "farmer",
		},
		// ─── 新增危机分支：储位崩盘 ───
		{
			id: "crisis-succession-collapse", title: "储位崩盘", domain: DomainCourt, severity: 78,
			summary: "太子被弹劾通敌，二皇子手握重兵，三皇子有母族撑腰。朝臣各自站队，朝会变成战场。是废太子另立，还是力挺太子压制诸王？",
			branchA: Choice{ID: "crisis-succession-replace", Text: "废太子，另立贤嗣",
				Detail: "消除争议，但废太子会动摇国本和法统。",
				Domain: DomainCourt, Effects: Effects{Stability: -6, Legitimacy: -4, Influence: 4},
				Outcome: "废太子诏书宣读时，东宫有人自尽了。新太子谢恩的表文很漂亮，但他的手在发抖。"},
			branchB: Choice{ID: "crisis-succession-support", Text: "力挺太子，压制诸王",
				Detail: "维护法统，但太子的污点不会被遗忘。",
				Domain: DomainCourt, Effects: Effects{Stability: 4, Influence: -3, Legitimacy: 2, Health: -1},
				Outcome: "你当众宣布太子不废，诸王俯首。但弹劾太子的奏折还在，它会在你死后被人翻出来。"},
			triggerFn: func(s *GameState) bool { return s.Succession.Dispute >= 65 },
			portrait:  "prince",
		},
	}
}

func (s *GameState) generateCrisisBranchEvent() SeasonEvent {
	defs := crisisBranchDefs()
	candidates := make([]crisisBranchDef, 0, len(defs))
	for _, d := range defs {
		if d.triggerFn(s) {
			candidates = append(candidates, d)
		}
	}
	if len(candidates) == 0 {
		// Fall back to first if none trigger but clock is high
		candidates = defs[:1]
	}
	s.ensureRNG()
	pick := s.rng.Intn(len(candidates))
	d := candidates[pick]

	return SeasonEvent{
		ID:           fmt.Sprintf("%s-%d-%s", d.id, s.Turn, s.Season),
		Title:        d.title,
		Summary:      d.summary,
		Detail:       "危机分支：此事件需要圣裁，不同选择将产生截然不同的后果。",
		Category:     EventStory,
		Domain:       d.domain,
		Severity:     d.severity,
		Effects:      Effects{},
		Tags:         []string{"危机", "圣裁"},
		Art:          sceneArt(s, eventSceneIndex(d.domain)),
		Portrait:     portraitKey(d.portrait),
		CrisisBranch: true,
		BranchID:     d.id,
		Choices:      []Choice{d.branchA, d.branchB},
	}
}
