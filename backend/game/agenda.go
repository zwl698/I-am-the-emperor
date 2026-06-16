package game

import (
	"fmt"
	"strings"
)

type agendaOption struct {
	key         string
	text        string
	detail      string
	outcome     string
	effects     Effects
	strategicFn func(s *GameState, opt agendaOption) // 回写战略地图的副作用
}

// domainPressure computes a 0-100 urgency score for each agenda domain,
// based on Stats, Crisis, Strategy, Wars, and other world state.
func domainPressure(s *GameState, domain Domain) int {
	switch domain {
	case DomainDomestic:
		p := worstProvinceDisaster(s.Provinces) + worstStrategicCityDisaster(s.Strategy.Cities)/2
		p += max(0, 60-s.Stats.Populace) + max(0, 50-s.Stats.Grain)/2
		if s.Stats.Grain < 30 {
			p += 20
		}
		return clamp(p, 0, 100)
	case DomainEconomy:
		p := max(0, 70-s.Stats.Treasury) + max(0, 50-totalStrategicGold(s.Strategy.Cities))/3
		if s.Stats.Treasury < 30 {
			p += 25
		}
		if s.Stats.Grain < 35 {
			p += 10
		}
		return clamp(p, 0, 100)
	case DomainMilitary:
		p := s.Stats.BorderThreat + maxWarThreat(s.Wars)/2 + s.strategicMilitaryPressure()/3
		if courtArmyGrainLow(s.Strategy.Armies) {
			p += 15
		}
		if s.Stats.Army < 50 {
			p += 15
		}
		return clamp(p, 0, 100)
	case DomainDiplomacy:
		p := s.Stats.BorderThreat/2 + maxStrategicFactionThreat(s.Strategy.Factions)/2
		p += max(0, 50-s.Stats.Diplomacy)
		if s.Stats.BorderThreat >= 60 && s.Stats.Diplomacy < 50 {
			p += 20
		}
		return clamp(p, 0, 100)
	case DomainReform:
		p := s.Stats.Reform/2 + averageOfficeVacancyRisk(s.Offices)/2
		if s.Stats.Reform >= 50 {
			p += 15 // 高改革值产生推进动力
		}
		if s.Stats.Stability < 40 {
			p -= 10 // 朝不稳则改革阻力大
		}
		return clamp(p, 0, 100)
	case DomainIntrigue:
		p := s.Crisis.Severity/2 + maxFactionPower(s.Factions)/3
		p += len(s.Plots) * 5
		if s.Crisis.Clock >= 5 {
			p += 20
		}
		return clamp(p, 0, 100)
	case DomainCourt:
		p := s.Succession.Dispute/2 + strongestConsortPower(s.Harem)/3
		p += averageOfficeVacancyRisk(s.Offices) / 4
		if s.Succession.Dispute >= 50 {
			p += 15
		}
		return clamp(p, 0, 100)
	default:
		return 30
	}
}

type domainChoice struct {
	domain   Domain
	choice   Choice
	pressure int
}

func emperorChoices(s *GameState) []Choice {
	// Step 1: 为每个领域生成候选选项
	allCandidates := []domainChoice{
		{domain: DomainDomestic, choice: domesticAgenda(s), pressure: domainPressure(s, DomainDomestic)},
		{domain: DomainEconomy, choice: economyAgenda(s), pressure: domainPressure(s, DomainEconomy)},
		{domain: DomainMilitary, choice: militaryAgenda(s), pressure: domainPressure(s, DomainMilitary)},
		{domain: DomainDiplomacy, choice: diplomacyAgenda(s), pressure: domainPressure(s, DomainDiplomacy)},
		{domain: DomainReform, choice: reformAgenda(s), pressure: domainPressure(s, DomainReform)},
		{domain: DomainIntrigue, choice: intrigueAgenda(s), pressure: domainPressure(s, DomainIntrigue)},
		{domain: DomainCourt, choice: courtAgenda(s), pressure: domainPressure(s, DomainCourt)},
	}

	// Step 2: 按压力降序排列
	sortByPressure(allCandidates)

	// Step 3: 筛选核心议题——压力>=35 的领域必选，最多4个
	var core []Choice
	var sidelined []Choice
	for _, c := range allCandidates {
		if c.pressure >= 35 && len(core) < 4 {
			core = append(core, c.choice)
		} else {
			sidelined = append(sidelined, c.choice)
		}
	}
	// 至少保留2个
	if len(core) < 2 {
		for _, c := range allCandidates {
			if len(core) >= 2 {
				break
			}
			found := false
			for _, cc := range core {
				if cc.ID == c.choice.ID {
					found = true
					break
				}
			}
			if !found {
				core = append(core, c.choice)
			}
		}
	}

	// Step 4: 如果危机钟 >= 4，追加一个紧急选项（来自被跳过的领域中压力最高的）
	if s.Crisis.Clock >= 4 && len(sidelined) > 0 {
		core = append(core, sidelined[0])
	}

	return core
}

func sortByPressure(candidates []domainChoice) {
	for i := 1; i < len(candidates); i++ {
		for j := i; j > 0 && candidates[j].pressure > candidates[j-1].pressure; j-- {
			candidates[j], candidates[j-1] = candidates[j-1], candidates[j]
		}
	}
}

// ──────────────────────────────────────────────
// 联动1: 朝堂场景读取战略态势生成选项
// ──────────────────────────────────────────────

func domesticAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "granary", text: "巡抚联名，请开常平仓", detail: "民政：赈济灾省，降低灾害与民变，但耗粮耗银。", effects: Effects{Treasury: -9, Grain: -10, Populace: 13, Stability: 5, Legitimacy: 2}, outcome: "常平仓火漆拆开，灾民开始回村。地方官也明白，朱批会追到县衙。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, city := range s.Strategy.Cities {
					if city.OwnerID == "court" && city.Disaster >= 30 {
						s.Strategy.Cities[i].Disaster = clamp(city.Disaster-14, 0, 100)
						s.Strategy.Cities[i].Order = clamp(city.Order+8, 0, 100)
					}
				}
			}},
		{key: "levee", text: "南河总督请修新堤", detail: "民政：压水患、稳漕运，短期花费大。", effects: Effects{Treasury: -12, Grain: 3, Populace: 8, Stability: 4, Reform: 2}, outcome: "夯土声连到夜里，新堤像一道迟来的承诺横在河岸。",
			strategicFn: func(s *GameState, _ agendaOption) {
				s.Strategy.adjustCity("canal", 0, 0, 4, 4, -8)
				s.Strategy.adjustCity("south", 0, 0, 2, 2, -6)
			}},
		{key: "resettle", text: "议迁流民，给牛种田", detail: "民政：重建人口与秩序，见效慢但能拖长国运。", effects: Effects{Treasury: -7, Grain: -6, Populace: 10, Stability: 6}, outcome: "流民牌册重新登记，荒地上有了第一缕炊烟。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, city := range s.Strategy.Cities {
					if city.OwnerID == "court" && city.Population < 50 {
						s.Strategy.Cities[i].Population = clamp(city.Population+8, 0, 120)
						s.Strategy.Cities[i].Agriculture = clamp(city.Agriculture+4, 0, 120)
					}
				}
			}},
	}

	// 战略联动: 如果战略地图上有城池受灾严重，优先生成赈灾选项
	worstStrategicDisaster := worstStrategicCityDisaster(s.Strategy.Cities)
	if worstProvinceDisaster(s.Provinces) >= 45 || worstStrategicDisaster >= 40 {
		return agendaChoice(s, DomainDomestic, options[0])
	}
	return agendaChoice(s, DomainDomestic, options[agendaPick(s, 1, len(options))])
}

func economyAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "arrears", text: "户部追补亏空，封存旧账", detail: "财政：快速回血，压迫地方豪强。", effects: Effects{Treasury: 16, Populace: -4, Stability: -4, Reform: 2}, outcome: "银车连夜入库，亏空薄了一层，地方士绅的脸色也冷了一层。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, city := range s.Strategy.Cities {
					if city.OwnerID == "court" {
						s.Strategy.Cities[i].Gold = clamp(city.Gold+6, 0, 180)
						s.Strategy.Cities[i].Order = clamp(city.Order-4, 0, 100)
					}
				}
			}},
		{key: "salt", text: "重定盐铁引票，收归榷场", detail: "财政：提高长期税源，激怒商帮。", effects: Effects{Treasury: 12, Reform: 4, Stability: -3, Diplomacy: 1}, outcome: "新盐引盖上朱印，商帮暂时低头，却没有忘记旧价。",
			strategicFn: func(s *GameState, _ agendaOption) {
				s.Strategy.adjustCity("capital", 6, 0, 0, -2, 0)
				s.Strategy.adjustCity("south", 4, 0, 0, 0, 0)
			}},
		{key: "harbor", text: "准江南开港，抽舶脚钱", detail: "财政：商贸与外交齐升，边境窥探增加。", effects: Effects{Treasury: 10, Diplomacy: 5, BorderThreat: 3, Populace: 2}, outcome: "海舶与漕船并泊，库银闻声上涨，远方也有人闻利而来。",
			strategicFn: func(s *GameState, _ agendaOption) {
				s.Strategy.adjustCity("dockyard", 8, 0, 0, 0, 0)
				s.Strategy.adjustCity("east-sea", 4, 0, 0, 0, 0)
			}},
	}

	if s.Stats.Treasury < 45 || totalStrategicGold(s.Strategy.Cities) < 60 {
		return agendaChoice(s, DomainEconomy, options[0])
	}
	return agendaChoice(s, DomainEconomy, options[agendaPick(s, 3, len(options))])
}

func militaryAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "supply", text: "边镇急报，请救前线粮道", detail: "军务：补给战争，降低边患，国库与粮草承压。", effects: Effects{Treasury: -10, Grain: -9, Army: 8, BorderThreat: -10, Martial: 1}, outcome: "粮车压过冻土，前线军心终于不用靠空话支撑。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, army := range s.Strategy.Armies {
					if army.FactionID == "court" {
						s.Strategy.Armies[i].Grain = clamp(army.Grain+18, 0, 160)
						s.Strategy.Armies[i].Morale = clamp(army.Morale+5, 0, 100)
					}
				}
			}},
		{key: "drill", text: "京营三大营合练新阵", detail: "军务：提升军力与武略，耗费显著。", effects: Effects{Treasury: -11, Grain: -3, Army: 14, BorderThreat: -6, Martial: 1}, outcome: "鼓声震动校场，老兵看见了新阵，新兵看见了军纪。",
			strategicFn: func(s *GameState, _ agendaOption) {
				s.Strategy.adjustArmy("imperial-guard", 0, 4, 6)
				s.Strategy.adjustArmy("northern-banner", 0, 4, 6)
			}},
		{key: "frontier", text: "轮戍边镇，撤换骄将", detail: "军务：压边镇尾大不掉，同时可能伤武臣。", effects: Effects{Treasury: -8, Army: 6, BorderThreat: -9, Stability: -2, Influence: 2}, outcome: "边镇军旗换防，骄将递上谢表，字里行间都是不甘。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, city := range s.Strategy.Cities {
					if city.OwnerID == "court" && city.Front {
						s.Strategy.Cities[i].Defense = clamp(city.Defense+8, 0, 120)
						s.Strategy.Cities[i].Troops = max(0, city.Troops+1500)
					}
				}
			}},
	}

	// 战略联动: 如果我方军队缺粮或敌军压境，优先补给
	if activeWarSupply(s.Wars) <= 40 || courtArmyGrainLow(s.Strategy.Armies) || s.strategicMilitaryPressure() >= 60 {
		return agendaChoice(s, DomainMilitary, options[0])
	}
	return agendaChoice(s, DomainMilitary, options[agendaPick(s, 5, len(options))])
}

func diplomacyAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "divide", text: "遣使携金册，离间诸邦", detail: "外交：削边患、增邦交，需要银钱铺路。", effects: Effects{Treasury: -6, Diplomacy: 12, BorderThreat: -10, Stability: 1}, outcome: "使臣在帐中把酒分给两边，敌人的盟约先从宴席上松动。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, faction := range s.Strategy.Factions {
					if !faction.IsPlayer && faction.Threat >= 40 {
						s.Strategy.Factions[i].Threat = clamp(faction.Threat-8, 0, 100)
						s.Strategy.Factions[i].Relation = clamp(faction.Relation+6, 0, 100)
					}
				}
			}},
		{key: "hostage", text: "请藩王世子入京伴读", detail: "外交：稳宗藩，暗藏人质意味。", effects: Effects{Diplomacy: 8, Stability: 3, Influence: 2, Treasury: -3}, outcome: "世子车驾入京，礼乐周全，刀锋藏在笑意背后。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, faction := range s.Strategy.Factions {
					if !faction.IsPlayer {
						s.Strategy.Factions[i].Relation = clamp(faction.Relation+4, 0, 100)
						s.Strategy.Factions[i].Threat = clamp(faction.Threat-4, 0, 100)
					}
				}
			}},
		{key: "market", text: "开互市三月，换马换情报", detail: "外交：贸易换缓冲，商帮得势。", effects: Effects{Treasury: 5, Diplomacy: 7, BorderThreat: -5, Populace: 2}, outcome: "关市人声鼎沸，马匹与密信一同过关。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, faction := range s.Strategy.Factions {
					if !faction.IsPlayer && faction.Relation >= 35 {
						s.Strategy.Factions[i].Relation = clamp(faction.Relation+6, 0, 100)
						s.Strategy.Factions[i].Threat = clamp(faction.Threat-5, 0, 100)
					}
				}
				s.Strategy.adjustCity("west", 4, 0, 0, 0, 0)
			}},
	}

	// 战略联动: 如果战略势力威胁高或边境城池被压，优先外交
	if s.Stats.BorderThreat >= 62 || maxStrategicFactionThreat(s.Strategy.Factions) >= 65 {
		return agendaChoice(s, DomainDiplomacy, options[0])
	}
	return agendaChoice(s, DomainDiplomacy, options[agendaPick(s, 7, len(options))])
}

func reformAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "audit", text: "设考成法，季末核官", detail: "新法：提高行政效率，引发官场反弹。", effects: Effects{Reform: 12, Treasury: 4, Stability: -6, Populace: 3, Legitimacy: 2}, outcome: "官员第一次发现，漂亮奏章不能替他们完成差事。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, city := range s.Strategy.Cities {
					if city.OwnerID == "court" {
						s.Strategy.Cities[i].Order = clamp(city.Order+6, 0, 100)
						s.Strategy.Cities[i].Commerce = clamp(city.Commerce+3, 0, 120)
					}
				}
			}},
		{key: "exam", text: "科举加策论，取能吏入局", detail: "新法：培养改革官僚，短期触动旧学。", effects: Effects{Reform: 9, Learning: 2, Stability: -3, Influence: 2}, outcome: "贡院灯火不灭，新题目像石子投进旧池。"},
		{key: "registry", text: "重造黄册，清丈隐田", detail: "新法：民籍与田亩重整，财政长期更稳。", effects: Effects{Reform: 10, Treasury: 7, Populace: -2, Stability: -5}, outcome: "黄册重修，隐田显形，地方豪右终于听见算盘之外的声音。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, city := range s.Strategy.Cities {
					if city.OwnerID == "court" {
						s.Strategy.Cities[i].Gold = clamp(city.Gold+4, 0, 180)
						s.Strategy.Cities[i].Agriculture = clamp(city.Agriculture+3, 0, 120)
					}
				}
			}},
	}
	return agendaChoice(s, DomainReform, options[agendaPick(s, 11, len(options))])
}

func intrigueAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "faction", text: "缇骑请开密档，查朋党", detail: "暗线：削派系权势，伤名望与稳定。", effects: Effects{Influence: 8, Stability: -6, Legitimacy: -4, Health: -2}, outcome: "密档摊开，朝臣跪得更低，也把话藏得更深。"},
		{key: "palace-ledger", text: "审宫账，追查内帑流向", detail: "暗线：打击宫中贪墨，牵连后宫与内廷。", effects: Effects{Treasury: 6, Influence: 5, Stability: -4, Legitimacy: -2}, outcome: "宫账翻到旧年，烛光照出许多不该出现的名字。"},
		{key: "assassin", text: "暗捕刺客线人，反钓幕后", detail: "暗线：保护皇权，但会制造恐惧。", effects: Effects{Influence: 7, Health: 1, Stability: -5, BorderThreat: -2}, outcome: "刺客的供词比刀锋更冷，幕后的人开始急着切断旧线。",
			strategicFn: func(s *GameState, _ agendaOption) {
				for i, faction := range s.Strategy.Factions {
					if !faction.IsPlayer && faction.Strategy == "复辟骚扰" {
						s.Strategy.Factions[i].Threat = clamp(faction.Threat-6, 0, 100)
					}
				}
			}},
	}
	return agendaChoice(s, DomainIntrigue, options[agendaPick(s, 13, len(options))])
}

func courtAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "office-rotation", text: "官职任免：重排中枢差遣", detail: "宫廷：调整官署权力，降低空转风险，也会抬高臣子压力。", effects: Effects{Influence: 4, Stability: 2, Reform: 1, Treasury: -2}, outcome: "吏部名单送入御前，几张椅子换了主人，朝堂风向也跟着转弯。"},
		{key: "harem-rites", text: "后宫册礼：安抚嫔妃与外戚", detail: "宫廷：稳宫闱、压争宠，但外戚会拿到更多筹码。", effects: Effects{Stability: 4, Health: 1, Treasury: -5, Influence: -1}, outcome: "册礼钟声回荡在内廷，笑容底下的账本又添了一页。"},
		{key: "succession-rite", text: "储位议礼：重申东宫名分", detail: "宫廷：提高继承稳定，可能刺激失宠母族。", effects: Effects{Legitimacy: 3, Stability: 3, Influence: 2, Treasury: -3}, outcome: "太庙香烟升起，储位名分被重新写进礼制，也写进每个母族的心里。"},
	}
	if len(s.Offices) > 0 && averageOfficeVacancyRisk(s.Offices) >= 42 {
		return agendaChoice(s, DomainCourt, options[0])
	}
	if s.Succession.Dispute >= 50 {
		return agendaChoice(s, DomainCourt, options[2])
	}
	return agendaChoice(s, DomainCourt, options[agendaPick(s, 17, len(options))])
}

func agendaChoice(s *GameState, domain Domain, option agendaOption) Choice {
	return Choice{
		ID:      fmt.Sprintf("%s-%s-%d-%s", domain, option.key, s.Turn, s.Season),
		Text:    option.text,
		Detail:  option.detail,
		Domain:  domain,
		Effects: option.effects,
		Outcome: option.outcome,
	}
}

func agendaPick(s *GameState, salt, n int) int {
	if n <= 1 {
		return 0
	}
	return (s.Turn + s.ReignYear*3 + seasonIndex(s.Season) + salt + len(s.History)) % n
}

func seasonIndex(season string) int {
	switch season {
	case "春":
		return 0
	case "夏":
		return 1
	case "秋":
		return 2
	case "冬":
		return 3
	default:
		return 0
	}
}

// ──────────────────────────────────────────────
// 联动2: 朝堂选择效果回写战略地图
// ──────────────────────────────────────────────

// applyStrategicConsequences 在 applyChoiceToWorld 中调用，
// 将朝堂选择的战略副作用写回 Strategy 状态。
func (s *GameState) applyStrategicConsequences(choice Choice) {
	if s.Phase != PhaseEmperor {
		return
	}
	s.ensureStrategicSystems()

	// 查找原始 agendaOption 中注册的 strategicFn
	opt := s.findAgendaOption(choice)
	if opt != nil && opt.strategicFn != nil {
		opt.strategicFn(s, *opt)
	}

	// 按领域施加通用战略影响
	switch choice.Domain {
	case DomainDomestic:
		for i, city := range s.Strategy.Cities {
			if city.OwnerID == "court" && city.Disaster >= 20 {
				s.Strategy.Cities[i].Order = clamp(city.Order+3, 0, 100)
			}
		}
	case DomainEconomy:
		for i, city := range s.Strategy.Cities {
			if city.OwnerID == "court" {
				s.Strategy.Cities[i].Commerce = clamp(city.Commerce+2, 0, 120)
			}
		}
	case DomainMilitary:
		for i, army := range s.Strategy.Armies {
			if army.FactionID == "court" {
				s.Strategy.Armies[i].Training = clamp(army.Training+2, 0, 100)
			}
		}
		for i, faction := range s.Strategy.Factions {
			if !faction.IsPlayer {
				s.Strategy.Factions[i].Threat = clamp(faction.Threat-2, 0, 100)
			}
		}
	case DomainDiplomacy:
		for i, faction := range s.Strategy.Factions {
			if !faction.IsPlayer {
				s.Strategy.Factions[i].Relation = clamp(faction.Relation+3, 0, 100)
			}
		}
	case DomainReform:
		for i, city := range s.Strategy.Cities {
			if city.OwnerID == "court" {
				s.Strategy.Cities[i].Order = clamp(city.Order+2, 0, 100)
			}
		}
	case DomainIntrigue:
		for i, faction := range s.Strategy.Factions {
			if !faction.IsPlayer && faction.Threat >= 50 {
				s.Strategy.Factions[i].Threat = clamp(faction.Threat-3, 0, 100)
			}
		}
	}
}

// findAgendaOption 在所有 agenda 选项池中按 choice ID 查找对应的 agendaOption
func (s *GameState) findAgendaOption(choice Choice) *agendaOption {
	allOptions := s.allAgendaOptions()
	for i := range allOptions {
		if strings.Contains(choice.ID, allOptions[i].key) {
			return &allOptions[i]
		}
	}
	return nil
}

// allAgendaOptions 返回当前朝堂所有可能的选项池，用于回溯查找
func (s *GameState) allAgendaOptions() []agendaOption {
	return []agendaOption{
		// domestic
		{key: "granary", strategicFn: func(s *GameState, _ agendaOption) {
			for i, city := range s.Strategy.Cities {
				if city.OwnerID == "court" && city.Disaster >= 30 {
					s.Strategy.Cities[i].Disaster = clamp(city.Disaster-14, 0, 100)
					s.Strategy.Cities[i].Order = clamp(city.Order+8, 0, 100)
				}
			}
		}},
		{key: "levee", strategicFn: func(s *GameState, _ agendaOption) {
			s.Strategy.adjustCity("canal", 0, 0, 4, 4, -8)
			s.Strategy.adjustCity("south", 0, 0, 2, 2, -6)
		}},
		{key: "resettle", strategicFn: func(s *GameState, _ agendaOption) {
			for i, city := range s.Strategy.Cities {
				if city.OwnerID == "court" && city.Population < 50 {
					s.Strategy.Cities[i].Population = clamp(city.Population+8, 0, 120)
					s.Strategy.Cities[i].Agriculture = clamp(city.Agriculture+4, 0, 120)
				}
			}
		}},
		// economy
		{key: "arrears", strategicFn: func(s *GameState, _ agendaOption) {
			for i, city := range s.Strategy.Cities {
				if city.OwnerID == "court" {
					s.Strategy.Cities[i].Gold = clamp(city.Gold+6, 0, 180)
					s.Strategy.Cities[i].Order = clamp(city.Order-4, 0, 100)
				}
			}
		}},
		{key: "salt", strategicFn: func(s *GameState, _ agendaOption) {
			s.Strategy.adjustCity("capital", 6, 0, 0, -2, 0)
			s.Strategy.adjustCity("south", 4, 0, 0, 0, 0)
		}},
		{key: "harbor", strategicFn: func(s *GameState, _ agendaOption) {
			s.Strategy.adjustCity("dockyard", 8, 0, 0, 0, 0)
			s.Strategy.adjustCity("east-sea", 4, 0, 0, 0, 0)
		}},
		// military
		{key: "supply", strategicFn: func(s *GameState, _ agendaOption) {
			for i, army := range s.Strategy.Armies {
				if army.FactionID == "court" {
					s.Strategy.Armies[i].Grain = clamp(army.Grain+18, 0, 160)
					s.Strategy.Armies[i].Morale = clamp(army.Morale+5, 0, 100)
				}
			}
		}},
		{key: "drill", strategicFn: func(s *GameState, _ agendaOption) {
			s.Strategy.adjustArmy("imperial-guard", 0, 4, 6)
			s.Strategy.adjustArmy("northern-banner", 0, 4, 6)
		}},
		{key: "frontier", strategicFn: func(s *GameState, _ agendaOption) {
			for i, city := range s.Strategy.Cities {
				if city.OwnerID == "court" && city.Front {
					s.Strategy.Cities[i].Defense = clamp(city.Defense+8, 0, 120)
					s.Strategy.Cities[i].Troops = max(0, city.Troops+1500)
				}
			}
		}},
		// diplomacy
		{key: "divide", strategicFn: func(s *GameState, _ agendaOption) {
			for i, faction := range s.Strategy.Factions {
				if !faction.IsPlayer && faction.Threat >= 40 {
					s.Strategy.Factions[i].Threat = clamp(faction.Threat-8, 0, 100)
					s.Strategy.Factions[i].Relation = clamp(faction.Relation+6, 0, 100)
				}
			}
		}},
		{key: "hostage", strategicFn: func(s *GameState, _ agendaOption) {
			for i, faction := range s.Strategy.Factions {
				if !faction.IsPlayer {
					s.Strategy.Factions[i].Relation = clamp(faction.Relation+4, 0, 100)
					s.Strategy.Factions[i].Threat = clamp(faction.Threat-4, 0, 100)
				}
			}
		}},
		{key: "market", strategicFn: func(s *GameState, _ agendaOption) {
			for i, faction := range s.Strategy.Factions {
				if !faction.IsPlayer && faction.Relation >= 35 {
					s.Strategy.Factions[i].Relation = clamp(faction.Relation+6, 0, 100)
					s.Strategy.Factions[i].Threat = clamp(faction.Threat-5, 0, 100)
				}
			}
			s.Strategy.adjustCity("west", 4, 0, 0, 0, 0)
		}},
		// reform
		{key: "audit", strategicFn: func(s *GameState, _ agendaOption) {
			for i, city := range s.Strategy.Cities {
				if city.OwnerID == "court" {
					s.Strategy.Cities[i].Order = clamp(city.Order+6, 0, 100)
					s.Strategy.Cities[i].Commerce = clamp(city.Commerce+3, 0, 120)
				}
			}
		}},
		{key: "registry", strategicFn: func(s *GameState, _ agendaOption) {
			for i, city := range s.Strategy.Cities {
				if city.OwnerID == "court" {
					s.Strategy.Cities[i].Gold = clamp(city.Gold+4, 0, 180)
					s.Strategy.Cities[i].Agriculture = clamp(city.Agriculture+3, 0, 120)
				}
			}
		}},
		// intrigue
		{key: "assassin", strategicFn: func(s *GameState, _ agendaOption) {
			for i, faction := range s.Strategy.Factions {
				if !faction.IsPlayer && faction.Strategy == "复辟骚扰" {
					s.Strategy.Factions[i].Threat = clamp(faction.Threat-6, 0, 100)
				}
			}
		}},
	}
}

// ──────────────────────────────────────────────
// 朝堂选择原有效果保留
// ──────────────────────────────────────────────

func (s *GameState) applyCourtAgendaOutcome(choice Choice) {
	if choice.Domain != DomainCourt {
		return
	}
	switch {
	case strings.Contains(choice.ID, "office"):
		for i := range s.Offices {
			s.Offices[i].VacancyRisk = clamp(s.Offices[i].VacancyRisk-5, 0, 100)
			s.Offices[i].Authority = clamp(s.Offices[i].Authority+2, 0, 100)
		}
		for i := range s.Court {
			s.Court[i].Stress = clamp(s.Court[i].Stress+2, 0, 100)
		}
	case strings.Contains(choice.ID, "harem"):
		for i := range s.Harem {
			s.Harem[i].Favor = clamp(s.Harem[i].Favor+3, 0, 100)
			s.Harem[i].Influence = clamp(s.Harem[i].Influence+1, 0, 100)
		}
		s.Succession.Dispute = clamp(s.Succession.Dispute+2, 0, 100)
	case strings.Contains(choice.ID, "succession"):
		s.Succession.Stability = clamp(s.Succession.Stability+6, 0, 100)
		s.Succession.Dispute = clamp(s.Succession.Dispute-6, 0, 100)
		if i, ok := s.findHeirIndex(s.Succession.NamedHeirID); ok {
			s.Heirs[i].Support = clamp(s.Heirs[i].Support+6, 0, 100)
		}
	}
}

// ──────────────────────────────────────────────
// 战略态势查询辅助函数
// ──────────────────────────────────────────────

func worstProvinceDisaster(provinces []Province) int {
	worst := 0
	for _, province := range provinces {
		worst = max(worst, province.Disaster)
	}
	return worst
}

// worstStrategicCityDisaster 返回我方城池中最高灾情
func worstStrategicCityDisaster(cities []StrategicCity) int {
	worst := 0
	for _, city := range cities {
		if city.OwnerID == "court" {
			worst = max(worst, city.Disaster)
		}
	}
	return worst
}

// totalStrategicGold 返回我方城池总黄金
func totalStrategicGold(cities []StrategicCity) int {
	total := 0
	for _, city := range cities {
		if city.OwnerID == "court" {
			total += city.Gold
		}
	}
	return total
}

// courtArmyGrainLow 判断我方军队是否整体缺粮
func courtArmyGrainLow(armies []ArmyGroup) bool {
	for _, army := range armies {
		if army.FactionID == "court" && army.Grain <= 15 {
			return true
		}
	}
	return false
}

// maxStrategicFactionThreat 返回非玩家势力最大威胁
func maxStrategicFactionThreat(factions []StrategicFaction) int {
	maxThreat := 0
	for _, faction := range factions {
		if !faction.IsPlayer {
			maxThreat = max(maxThreat, faction.Threat)
		}
	}
	return maxThreat
}

// averageStrategicCityOrder 返回我方城池平均治安
func averageStrategicCityOrder(cities []StrategicCity) int {
	total := 0
	count := 0
	for _, city := range cities {
		if city.OwnerID == "court" {
			total += city.Order
			count++
		}
	}
	if count == 0 {
		return 50
	}
	return total / count
}

func activeWarSupply(wars []WarCampaign) int {
	if len(wars) == 0 {
		return 100
	}
	lowest := 100
	for _, war := range wars {
		if war.Stage != "凯旋" {
			lowest = min(lowest, war.Supply)
		}
	}
	return lowest
}

func averageOfficeVacancyRisk(offices []Office) int {
	if len(offices) == 0 {
		return 0
	}
	total := 0
	for _, office := range offices {
		total += office.VacancyRisk
	}
	return total / len(offices)
}
