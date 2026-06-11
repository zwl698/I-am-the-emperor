package game

import (
	"fmt"
	"strings"
)

type agendaOption struct {
	key     string
	text    string
	detail  string
	outcome string
	effects Effects
}

func emperorChoices(s *GameState) []Choice {
	return []Choice{
		domesticAgenda(s),
		economyAgenda(s),
		militaryAgenda(s),
		diplomacyAgenda(s),
		reformAgenda(s),
		intrigueAgenda(s),
		courtAgenda(s),
	}
}

func domesticAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "granary", text: "巡抚联名，请开常平仓", detail: "民政：赈济灾省，降低灾害与民变，但耗粮耗银。", effects: Effects{Treasury: -9, Grain: -10, Populace: 13, Stability: 5, Legitimacy: 2}, outcome: "常平仓火漆拆开，灾民开始回村。地方官也明白，朱批会追到县衙。"},
		{key: "levee", text: "南河总督请修新堤", detail: "民政：压水患、稳漕运，短期花费大。", effects: Effects{Treasury: -12, Grain: 3, Populace: 8, Stability: 4, Reform: 2}, outcome: "夯土声连到夜里，新堤像一道迟来的承诺横在河岸。"},
		{key: "resettle", text: "议迁流民，给牛种田", detail: "民政：重建人口与秩序，见效慢但能拖长国运。", effects: Effects{Treasury: -7, Grain: -6, Populace: 10, Stability: 6}, outcome: "流民牌册重新登记，荒地上有了第一缕炊烟。"},
	}
	if worstProvinceDisaster(s.Provinces) >= 45 {
		return agendaChoice(s, DomainDomestic, options[0])
	}
	return agendaChoice(s, DomainDomestic, options[agendaPick(s, 1, len(options))])
}

func economyAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "arrears", text: "户部追补亏空，封存旧账", detail: "财政：快速回血，压迫地方豪强。", effects: Effects{Treasury: 16, Populace: -4, Stability: -4, Reform: 2}, outcome: "银车连夜入库，亏空薄了一层，地方士绅的脸色也冷了一层。"},
		{key: "salt", text: "重定盐铁引票，收归榷场", detail: "财政：提高长期税源，激怒商帮。", effects: Effects{Treasury: 12, Reform: 4, Stability: -3, Diplomacy: 1}, outcome: "新盐引盖上朱印，商帮暂时低头，却没有忘记旧价。"},
		{key: "harbor", text: "准江南开港，抽舶脚钱", detail: "财政：商贸与外交齐升，边境窥探增加。", effects: Effects{Treasury: 10, Diplomacy: 5, BorderThreat: 3, Populace: 2}, outcome: "海舶与漕船并泊，库银闻声上涨，远方也有人闻利而来。"},
	}
	if s.Stats.Treasury < 45 {
		return agendaChoice(s, DomainEconomy, options[0])
	}
	return agendaChoice(s, DomainEconomy, options[agendaPick(s, 3, len(options))])
}

func militaryAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "supply", text: "边镇急报，请救前线粮道", detail: "军务：补给战争，降低边患，国库与粮草承压。", effects: Effects{Treasury: -10, Grain: -9, Army: 8, BorderThreat: -10, Martial: 1}, outcome: "粮车压过冻土，前线军心终于不用靠空话支撑。"},
		{key: "drill", text: "京营三大营合练新阵", detail: "军务：提升军力与武略，耗费显著。", effects: Effects{Treasury: -11, Grain: -3, Army: 14, BorderThreat: -6, Martial: 1}, outcome: "鼓声震动校场，老兵看见了新阵，新兵看见了军纪。"},
		{key: "frontier", text: "轮戍边镇，撤换骄将", detail: "军务：压边镇尾大不掉，同时可能伤武臣。", effects: Effects{Treasury: -8, Army: 6, BorderThreat: -9, Stability: -2, Influence: 2}, outcome: "边镇军旗换防，骄将递上谢表，字里行间都是不甘。"},
	}
	if activeWarSupply(s.Wars) <= 40 {
		return agendaChoice(s, DomainMilitary, options[0])
	}
	return agendaChoice(s, DomainMilitary, options[agendaPick(s, 5, len(options))])
}

func diplomacyAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "divide", text: "遣使携金册，离间诸邦", detail: "外交：削边患、增邦交，需要银钱铺路。", effects: Effects{Treasury: -6, Diplomacy: 12, BorderThreat: -10, Stability: 1}, outcome: "使臣在帐中把酒分给两边，敌人的盟约先从宴席上松动。"},
		{key: "hostage", text: "请藩王世子入京伴读", detail: "外交：稳宗藩，暗藏人质意味。", effects: Effects{Diplomacy: 8, Stability: 3, Influence: 2, Treasury: -3}, outcome: "世子车驾入京，礼乐周全，刀锋藏在笑意背后。"},
		{key: "market", text: "开互市三月，换马换情报", detail: "外交：贸易换缓冲，商帮得势。", effects: Effects{Treasury: 5, Diplomacy: 7, BorderThreat: -5, Populace: 2}, outcome: "关市人声鼎沸，马匹与密信一同过关。"},
	}
	if s.Stats.BorderThreat >= 62 {
		return agendaChoice(s, DomainDiplomacy, options[0])
	}
	return agendaChoice(s, DomainDiplomacy, options[agendaPick(s, 7, len(options))])
}

func reformAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "audit", text: "设考成法，季末核官", detail: "新法：提高行政效率，引发官场反弹。", effects: Effects{Reform: 12, Treasury: 4, Stability: -6, Populace: 3, Legitimacy: 2}, outcome: "官员第一次发现，漂亮奏章不能替他们完成差事。"},
		{key: "exam", text: "科举加策论，取能吏入局", detail: "新法：培养改革官僚，短期触动旧学。", effects: Effects{Reform: 9, Learning: 2, Stability: -3, Influence: 2}, outcome: "贡院灯火不灭，新题目像石子投进旧池。"},
		{key: "registry", text: "重造黄册，清丈隐田", detail: "新法：民籍与田亩重整，财政长期更稳。", effects: Effects{Reform: 10, Treasury: 7, Populace: -2, Stability: -5}, outcome: "黄册重修，隐田显形，地方豪右终于听见算盘之外的声音。"},
	}
	return agendaChoice(s, DomainReform, options[agendaPick(s, 11, len(options))])
}

func intrigueAgenda(s *GameState) Choice {
	options := []agendaOption{
		{key: "faction", text: "缇骑请开密档，查朋党", detail: "暗线：削派系权势，伤名望与稳定。", effects: Effects{Influence: 8, Stability: -6, Legitimacy: -4, Health: -2}, outcome: "密档摊开，朝臣跪得更低，也把话藏得更深。"},
		{key: "palace-ledger", text: "审宫账，追查内帑流向", detail: "暗线：打击宫中贪墨，牵连后宫与内廷。", effects: Effects{Treasury: 6, Influence: 5, Stability: -4, Legitimacy: -2}, outcome: "宫账翻到旧年，烛光照出许多不该出现的名字。"},
		{key: "assassin", text: "暗捕刺客线人，反钓幕后", detail: "暗线：保护皇权，但会制造恐惧。", effects: Effects{Influence: 7, Health: 1, Stability: -5, BorderThreat: -2}, outcome: "刺客的供词比刀锋更冷，幕后的人开始急着切断旧线。"},
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

func worstProvinceDisaster(provinces []Province) int {
	worst := 0
	for _, province := range provinces {
		worst = max(worst, province.Disaster)
	}
	return worst
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
