package game

import (
	"fmt"
	"sort"
)

type EventCard struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Category    string   `json:"category"`
	Domain      Domain   `json:"domain"`
	Arc         string   `json:"arc"`
	Stage       string   `json:"stage"`
	Summary     string   `json:"summary"`
	Hook        string   `json:"hook"`
	Consequence string   `json:"consequence"`
	Severity    int      `json:"severity"`
	Urgency     int      `json:"urgency"`
	Duration    int      `json:"duration"`
	Tags        []string `json:"tags"`
}

type eventCategorySeed struct {
	id       string
	name     string
	domain   Domain
	arc      string
	stage    string
	tags     []string
	severity int
	titles   []string
}

func EventDeckCatalog() []EventCard {
	seeds := eventCategorySeeds()
	cards := make([]EventCard, 0, 120)
	for _, seed := range seeds {
		for i, title := range seed.titles {
			cards = append(cards, EventCard{
				ID:          fmt.Sprintf("%s-%02d", seed.id, i+1),
				Title:       title,
				Category:    seed.name,
				Domain:      seed.domain,
				Arc:         seed.arc,
				Stage:       seed.stage,
				Summary:     fmt.Sprintf("%s正在发酵：%s。", seed.arc, title),
				Hook:        eventHook(seed, title, i),
				Consequence: eventConsequence(seed.domain, title),
				Severity:    clamp(seed.severity+i%4*3, 1, 100),
				Urgency:     clamp(30+i*5+seed.severity/5, 1, 100),
				Duration:    1 + i%3,
				Tags:        append([]string(nil), seed.tags...),
			})
		}
	}
	return cards
}

func (s *GameState) dealEventHand() {
	if s == nil || s.Phase != PhaseEmperor {
		s.EventHand = nil
		return
	}
	type scoredCard struct {
		card  EventCard
		score int
	}
	catalog := EventDeckCatalog()
	scored := make([]scoredCard, 0, len(catalog))
	for _, card := range catalog {
		pressure := s.eventCardPressure(card)
		jitter := s.eventCardJitter(card.ID)
		card.Urgency = clamp(card.Urgency+pressure/8+jitter%15, 1, 100)
		card.Severity = clamp(card.Severity+pressure/12, 1, 100)
		scored = append(scored, scoredCard{card: card, score: pressure*3 + card.Urgency + jitter})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	hand := make([]EventCard, 0, 5)
	seenCategory := map[string]bool{}
	for _, item := range scored {
		if len(hand) >= 5 {
			break
		}
		if seenCategory[item.card.Category] && len(hand) < 3 {
			continue
		}
		hand = append(hand, item.card)
		seenCategory[item.card.Category] = true
	}
	s.EventHand = hand
}

func (s *GameState) eventCardPressure(card EventCard) int {
	if s == nil {
		return 0
	}
	switch card.Domain {
	case DomainDomestic:
		// 联动4: 加入战略地图城池灾情压力
		strategicDisaster := worstStrategicCityDisaster(s.Strategy.Cities)
		strategicOrder := averageStrategicCityOrder(s.Strategy.Cities)
		return worstProvinceDisaster(s.Provinces) + strategicDisaster/2 + max(0, 65-averageProvinceOrder(s.Provinces)) + max(0, 50-strategicOrder)/2
	case DomainEconomy:
		// 联动4: 加入战略地图城池商贸/黄金压力
		strategicGold := totalStrategicGold(s.Strategy.Cities)
		return max(0, 90-s.Stats.Treasury) + factionPower(s.Factions, "merchant")/2 + max(0, 60-strategicGold/4)
	case DomainMilitary:
		// 联动4: 军事压力已含strategicMilitaryPressure，增加军队缺粮维度
		armyPressure := 0
		if courtArmyGrainLow(s.Strategy.Armies) {
			armyPressure = 25
		}
		return s.Stats.BorderThreat + max(maxWarThreat(s.Wars), s.strategicMilitaryPressure()) + armyPressure
	case DomainDiplomacy:
		// 联动4: 加入战略势力关系/威胁维度
		strategicFactionPressure := maxStrategicFactionThreat(s.Strategy.Factions)
		return s.Stats.BorderThreat/2 + max(maxForeignThreat(s.ForeignStates), strategicFactionPressure) + max(0, 70-s.Stats.Diplomacy)
	case DomainCourt:
		return s.Succession.Dispute + strongestConsortPower(s.Harem)/2
	case DomainReform:
		return s.Stats.Reform + averageOfficeVacancyRisk(s.Offices)/2
	case DomainIntrigue:
		return maxPlotProgress(s.Plots) + maxLegalCaseHeat(s.LegalCases) + s.PublicOpinion.Rumor/2
	default:
		return s.Crisis.Severity
	}
}

func (s *GameState) eventCardJitter(id string) int {
	total := int(s.Seed%101) + s.Turn*29 + s.ReignYear*17 + seasonIndex(s.Season)*13
	for _, r := range id {
		total += int(r)
	}
	return total % 97
}

func eventHook(seed eventCategorySeed, title string, index int) string {
	hooks := []string{
		"若放任一季，相关势力会把它变成筹码。",
		"可用御令、任官、审案或战术提前拆解。",
		"处理方式会改变下一季事件牌权重。",
		"它会牵连至少两个系统，不是孤立奏章。",
	}
	return fmt.Sprintf("%s：%s", title, hooks[index%len(hooks)])
}

func eventConsequence(domain Domain, title string) string {
	switch domain {
	case DomainMilitary:
		return title + "会推高敌势、消耗粮道，或改变战役阶段。"
	case DomainEconomy:
		return title + "会影响国库、商帮和民间物价。"
	case DomainDomestic:
		return title + "会牵动民心、灾情和地方秩序。"
	case DomainCourt:
		return title + "会牵动后宫、储位、臣子忠诚和朝稳。"
	case DomainDiplomacy:
		return title + "会改变外邦关系、贡贸和边境威慑。"
	case DomainReform:
		return title + "会推动或反噬新法、官署和士林关系。"
	case DomainIntrigue:
		return title + "会生成案卷、密谋或舆论压力。"
	default:
		return title + "会改变王朝危机线。"
	}
}

func eventCategorySeeds() []eventCategorySeed {
	return []eventCategorySeed{
		{id: "prince", name: "皇子成长", domain: DomainCourt, arc: "少年帝王线", stage: "成长", tags: []string{"皇子", "夺嫡"}, severity: 28, titles: []string{"出生异象", "乳母密语", "东宫初读", "兄弟争席", "御花园落水", "冬狩惊马", "太傅密考", "母族请托", "宗室观望", "夺嫡夜诏"}},
		{id: "disaster", name: "内政灾害", domain: DomainDomestic, arc: "民生灾异线", stage: "地方", tags: []string{"灾害", "民心"}, severity: 44, titles: []string{"河堤决口", "蝗灾入境", "粮仓霉变", "灾民围县", "疫病流坊", "井盐断供", "漕船翻沉", "荒田复垦", "乡绅抗粮", "常平仓亏空"}},
		{id: "treasury", name: "财政商路", domain: DomainEconomy, arc: "国库商帮线", stage: "财政", tags: []string{"财政", "商帮"}, severity: 42, titles: []string{"盐引私卖", "银荒挤兑", "海贸走私", "商帮联保", "矿税暴动", "漕运加价", "钱庄倒闭", "皇庄兼并", "市舶争税", "户部亏空"}},
		{id: "office", name: "官职朝堂", domain: DomainReform, arc: "中枢官署线", stage: "朝堂", tags: []string{"官职", "朋党"}, severity: 40, titles: []string{"首辅封驳", "御史弹劾", "空署积牍", "官员告病", "朋党会饮", "科道联名", "边臣索饷", "吏部卖缺", "清流上疏", "顾命争权"}},
		{id: "harem", name: "后宫外戚", domain: DomainCourt, arc: "宫闱母族线", stage: "宫廷", tags: []string{"后宫", "外戚"}, severity: 39, titles: []string{"中宫赐宴", "贵妃争宠", "宫账暗亏", "内侍传信", "外戚请封", "皇后病笺", "宫女失踪", "母族联姻", "冷宫旧案", "宫印误用"}},
		{id: "succession", name: "继承东宫", domain: DomainCourt, arc: "储位继承线", stage: "东宫", tags: []string{"继承", "东宫"}, severity: 45, titles: []string{"太子伴读", "皇嗣骑射", "师傅党争", "童谣立储", "母族押注", "东宫失仪", "兄弟结社", "册文争字", "储位谤书", "太庙议礼"}},
		{id: "intrigue", name: "密谋刑狱", domain: DomainIntrigue, arc: "暗线案卷线", stage: "密档", tags: []string{"密谋", "刑狱"}, severity: 47, titles: []string{"刺客供词", "丝账暗线", "边书私递", "宫酒疑云", "清议黑榜", "密档失窃", "三司会审", "证人翻供", "禁谣风波", "判词榜示"}},
		{id: "war", name: "对外战争", domain: DomainMilitary, arc: "边境战役线", stage: "战局", tags: []string{"战争", "边患"}, severity: 52, titles: []string{"敌骑试探", "粮道断续", "雪夜奇袭", "将领争功", "城寨失火", "俘虏交换", "援军迟至", "前线瘟疫", "决战请命", "凯旋分赏"}},
		{id: "foreign", name: "外交诸邦", domain: DomainDiplomacy, arc: "万邦纵横线", stage: "外邦", tags: []string{"外交", "外邦"}, severity: 41, titles: []string{"北狄求亲", "西域贡使", "东海海盗", "南岭盟寨", "藩王世子", "互市争价", "国书错译", "边境会盟", "间使离间", "贡船失踪"}},
		{id: "reform", name: "新法改革", domain: DomainReform, arc: "变法制度线", stage: "新政", tags: []string{"新法", "士林"}, severity: 43, titles: []string{"黄册重造", "考成法", "太学新院", "反新法诗案", "巡按肃贪", "丈田冲突", "役法试点", "工坊招募", "旧党抵制", "新法成例"}},
		{id: "opinion", name: "舆论民心", domain: DomainIntrigue, arc: "京城风声线", stage: "舆情", tags: []string{"舆论", "民心"}, severity: 38, titles: []string{"京城小报", "茶楼传谣", "灾民歌谣", "清议诗会", "榜文争读", "神童预言", "士子罢考", "民间祈雨", "边城童谣", "盛世烟火"}},
		{id: "endgame", name: "终局危机", domain: DomainStory, arc: "王朝命运线", stage: "终局", tags: []string{"危机", "终局"}, severity: 58, titles: []string{"党争裂朝", "边患压城", "国库枯竭", "储位崩盘", "民变连营", "藩镇坐大", "外邦合围", "瘟疫入京", "皇帝病危", "万邦来朝"}},
	}
}

func maxForeignThreat(states []ForeignState) int {
	highest := 0
	for _, state := range states {
		highest = max(highest, state.Threat)
	}
	return highest
}

func maxPlotProgress(plots []Plot) int {
	highest := 0
	for _, plot := range plots {
		if !plot.Resolved {
			highest = max(highest, plot.Progress)
		}
	}
	return highest
}
