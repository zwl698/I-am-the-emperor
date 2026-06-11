package game

import "fmt"

func dynasties() []Dynasty {
	return []Dynasty{
		{ID: "dayin", Name: "大胤", Era: "开国元年", Background: "旧都新定，功臣拥兵，百废待兴。", Features: []string{"开国功臣强势", "国库充实但朝制未稳", "军功路线收益更高"}, Challenge: "用刀剑打下天下后，如何让刀剑回鞘。", Asset: "/assets/dynasty-scroll.png", Palette: "ember", Initial: Stats{Legitimacy: 58, Health: 74, Learning: 20, Martial: 28, Charisma: 24, Influence: 20, Treasury: 78, Grain: 62, Populace: 48, Army: 82, Diplomacy: 36, Stability: 42, BorderThreat: 46, Reform: 12}},
		{ID: "jingyao", Name: "景曜", Era: "盛世中叶", Background: "漕运通达，市井繁华，盛世的每一道裂缝都藏在金粉下。", Features: []string{"财政与外交基础优秀", "改革阻力较小", "奢靡会放大危机"}, Challenge: "在歌舞升平里提前看见衰败。", Asset: "/assets/dynasty-scroll.png", Palette: "gold", Initial: Stats{Legitimacy: 64, Health: 72, Learning: 26, Martial: 18, Charisma: 30, Influence: 20, Treasury: 92, Grain: 88, Populace: 76, Army: 58, Diplomacy: 70, Stability: 72, BorderThreat: 26, Reform: 20}},
		{ID: "chengping", Name: "承平", Era: "暮年危局", Background: "库银亏空，兼并成风，灾民与朋党一起挤进奏章。", Features: []string{"财政压力极高", "新法收益更大", "民变风险更快累积"}, Challenge: "在旧制度的裂缝里硬生生开出新路。", Asset: "/assets/dynasty-scroll.png", Palette: "storm", Initial: Stats{Legitimacy: 46, Health: 68, Learning: 28, Martial: 16, Charisma: 22, Influence: 18, Treasury: 36, Grain: 38, Populace: 34, Army: 48, Diplomacy: 42, Stability: 30, BorderThreat: 52, Reform: 8}},
		{ID: "xuanshuo", Name: "玄朔", Era: "北境烽烟", Background: "雪岭烽火连年，边镇半独立，朝廷每一次迟疑都会变成战报。", Features: []string{"边患开局最高", "军务外交回报更高", "粮草消耗更凶"}, Challenge: "一手握兵符，一手还要稳住中原民心。", Asset: "/assets/dynasty-scroll.png", Palette: "frost", Initial: Stats{Legitimacy: 52, Health: 72, Learning: 20, Martial: 30, Charisma: 20, Influence: 22, Treasury: 58, Grain: 46, Populace: 46, Army: 76, Diplomacy: 34, Stability: 44, BorderThreat: 72, Reform: 10}},
	}
}

func sceneGalleryPaths() []string {
	names := []string{
		"birth-chamber",
		"east-palace-study",
		"winter-hunt",
		"flood-levee",
		"succession-hall",
		"throne-court",
		"granary-relief",
		"tax-office",
		"frontier-fortress",
		"envoy-pass",
		"reform-archive",
		"secret-tribunal",
		"banquet-hall",
		"jiangnan-canal",
		"northern-battlefield",
		"desert-market",
		"imperial-garden",
		"rain-corridor",
		"ancestral-temple",
		"ministry-office",
		"dockyard-fleet",
		"drill-ground",
		"rebel-village",
		"silk-market",
		"mountain-monastery",
		"exam-hall",
		"map-room",
		"palace-dawn",
		"diplomatic-tent",
		"festival-night",
	}
	return assetPaths("/assets/scenes/scene", names)
}

func portraitGalleryPaths() []string {
	names := []string{
		"infant-prince",
		"teen-prince",
		"young-emperor",
		"elder-emperor",
		"stern-tutor",
		"frontier-general",
		"finance-minister",
		"grand-princess",
		"noble-consort",
		"young-empress",
		"queen-dowager",
		"palace-maid",
		"eunuch-spymaster",
		"scholar-official",
		"reformist-official",
		"corrupt-magistrate",
		"merchant-leader",
		"foreign-envoy",
		"nomad-khan",
		"monk-strategist",
		"female-diplomat",
		"guard-captain",
		"rebel-leader",
		"river-engineer",
		"imperial-physician",
		"astrologer",
		"poet",
		"court-painter",
		"farmer-representative",
		"masked-assassin",
	}
	return assetPaths("/assets/portraits/portrait", names)
}

func assetPaths(prefix string, names []string) []string {
	paths := make([]string, len(names))
	for i, name := range names {
		paths[i] = fmt.Sprintf("%s-%02d-%s.png", prefix, i+1, name)
	}
	return paths
}

func findDynasty(id string) (Dynasty, bool) {
	for _, dynasty := range dynasties() {
		if dynasty.ID == id {
			return dynasty, true
		}
	}
	return Dynasty{}, false
}

func startingFactions(dynastyID string) []Faction {
	factions := []Faction{
		{ID: "scholar", Name: "清流士林", Leader: "顾太傅", Power: 45, Loyalty: 58, Agenda: "重礼法、轻苛政", Portrait: "tutor"},
		{ID: "border", Name: "边镇武勋", Leader: "霍骁", Power: 48, Loyalty: 52, Agenda: "要粮饷、要军功", Portrait: "general"},
		{ID: "merchant", Name: "漕运商帮", Leader: "沈万策", Power: 42, Loyalty: 46, Agenda: "求盐铁、通关市", Portrait: "minister"},
		{ID: "clan", Name: "宗室外戚", Leader: "长公主", Power: 40, Loyalty: 50, Agenda: "保爵位、稳宫闱", Portrait: "consort"},
	}
	switch dynastyID {
	case "dayin":
		factions[1].Power += 12
	case "jingyao":
		factions[2].Power += 10
		factions[2].Loyalty += 8
	case "chengping":
		factions[0].Loyalty -= 8
		factions[2].Power += 8
	case "xuanshuo":
		factions[1].Power += 15
		factions[1].Loyalty += 7
	}
	return factions
}

func startingCourt() []Minister {
	return []Minister{
		{ID: "gu", Name: "顾怀章", Role: "太傅", Trait: "刚正", Loyalty: 62, Ability: 82, Ambition: 32, Integrity: 88, Stress: 18, Portrait: "tutor"},
		{ID: "huo", Name: "霍骁", Role: "大将军", Trait: "敢战", Loyalty: 55, Ability: 78, Ambition: 64, Integrity: 54, Stress: 24, Portrait: "general"},
		{ID: "shen", Name: "沈万策", Role: "户部尚书", Trait: "精算", Loyalty: 48, Ability: 86, Ambition: 58, Integrity: 46, Stress: 22, Portrait: "minister"},
		{ID: "princess", Name: "昭宁", Role: "长公主", Trait: "纵横", Loyalty: 56, Ability: 74, Ambition: 70, Integrity: 62, Stress: 20, Portrait: "consort"},
	}
}

func startingWars(dynastyID string) []WarCampaign {
	switch dynastyID {
	case "xuanshuo":
		return []WarCampaign{
			{ID: "snow-ridge", Name: "雪岭北伐", Enemy: "北狄诸部", Front: "北境雪岭", Stage: "压境", Threat: 78, Supply: 48, Morale: 56, Progress: 18, Duration: 2},
		}
	case "dayin":
		return []WarCampaign{
			{ID: "western-oath", Name: "西陲归附战", Enemy: "旧朝残部", Front: "西陲商路", Stage: "拉锯", Threat: 48, Supply: 62, Morale: 60, Progress: 34, Duration: 1},
		}
	case "chengping":
		return []WarCampaign{
			{ID: "river-bandits", Name: "南河剿乱", Enemy: "流寇水寨", Front: "南河漕道", Stage: "粮道危急", Threat: 58, Supply: 35, Morale: 44, Progress: 22, Duration: 3},
		}
	case "jingyao":
		return []WarCampaign{
			{ID: "jade-pass", Name: "玉门互市危机", Enemy: "西域联军", Front: "玉门关市", Stage: "拉锯", Threat: 36, Supply: 70, Morale: 62, Progress: 42, Duration: 1},
		}
	default:
		return nil
	}
}

func startingProvinces(dynastyID string) []Province {
	provinces := []Province{
		{ID: "capital", Name: "京畿", Focus: "朝堂与税源", Wealth: 60, Order: 58, Defense: 52, Disaster: 18},
		{ID: "south", Name: "江南", Focus: "漕运与粮仓", Wealth: 72, Order: 55, Defense: 38, Disaster: 22},
		{ID: "north", Name: "北境", Focus: "边防与马政", Wealth: 38, Order: 48, Defense: 70, Disaster: 28},
		{ID: "west", Name: "西陲", Focus: "商路与藩部", Wealth: 44, Order: 50, Defense: 48, Disaster: 20},
	}
	switch dynastyID {
	case "chengping":
		for i := range provinces {
			provinces[i].Order -= 14
			provinces[i].Disaster += 12
		}
	case "xuanshuo":
		provinces[2].Defense -= 12
		provinces[2].Disaster += 18
	case "jingyao":
		provinces[1].Wealth += 12
		provinces[0].Order += 8
	}
	return provinces
}

func startingCrisis(dynastyID string) Crisis {
	switch dynastyID {
	case "dayin":
		return Crisis{Title: "功臣难驯", Severity: 44, Clock: 2, Summary: "开国诸将仍握重兵，封赏稍慢便会生怨。"}
	case "jingyao":
		return Crisis{Title: "盛世暗蚀", Severity: 24, Clock: 1, Summary: "繁华掩盖了土地兼并与奢靡风气。"}
	case "chengping":
		return Crisis{Title: "民变将起", Severity: 64, Clock: 4, Summary: "灾民、亏空和党争正在互相点燃。"}
	case "xuanshuo":
		return Crisis{Title: "北境压城", Severity: 68, Clock: 4, Summary: "雪岭诸部集结，边镇只等朝廷粮饷。"}
	default:
		return Crisis{Title: "朝局未稳", Severity: 40, Clock: 2, Summary: "新君面前没有小事。"}
	}
}
