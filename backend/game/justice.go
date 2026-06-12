package game

import "fmt"

type PublicOpinion struct {
	Popular   int    `json:"popular"`
	Elite     int    `json:"elite"`
	Rumor     int    `json:"rumor"`
	Fear      int    `json:"fear"`
	Justice   int    `json:"justice"`
	LastEdict string `json:"lastEdict"`
}

type LegalCase struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Domain          Domain `json:"domain"`
	Accuser         string `json:"accuser"`
	Defendant       string `json:"defendant"`
	Charge          string `json:"charge"`
	Stakes          string `json:"stakes"`
	Source          string `json:"source"`
	Portrait        string `json:"portrait"`
	Heat            int    `json:"heat"`
	Evidence        int    `json:"evidence"`
	FactionPressure int    `json:"factionPressure"`
	PublicPressure  int    `json:"publicPressure"`
	Resolved        bool   `json:"resolved"`
	Verdict         string `json:"verdict"`
}

func startingPublicOpinion(dynastyID string) PublicOpinion {
	opinion := PublicOpinion{Popular: 52, Elite: 48, Rumor: 38, Fear: 24, Justice: 46, LastEdict: "新朝法司待命，京城茶楼仍在猜测第一道大案。"}
	switch dynastyID {
	case "dayin":
		opinion.Fear += 8
		opinion.Elite -= 4
	case "jingyao":
		opinion.Popular += 8
		opinion.Rumor += 6
	case "chengping":
		opinion.Rumor += 12
		opinion.Justice -= 6
	case "xuanshuo":
		opinion.Fear += 6
		opinion.Popular -= 4
	}
	return opinion
}

func startingLegalCases(dynastyID string) []LegalCase {
	cases := []LegalCase{
		{ID: "granary-ledger", Title: "常平仓亏空案", Domain: DomainDomestic, Accuser: "河道巡按", Defendant: "江北粮道", Charge: "赈粮漂没", Stakes: "灾民、地方官与户部互相攀咬", Source: "seed:granary", Portrait: "farmer", Heat: 45, Evidence: 58, FactionPressure: 42, PublicPressure: 64},
		{ID: "salt-guild", Title: "盐引私售案", Domain: DomainEconomy, Accuser: "户部给事中", Defendant: "漕运商帮", Charge: "私卖盐引", Stakes: "财政回血会触动商帮账本", Source: "seed:salt", Portrait: "merchant", Heat: 52, Evidence: 50, FactionPressure: 60, PublicPressure: 48},
		{ID: "palace-seal", Title: "宫印误用案", Domain: DomainCourt, Accuser: "内廷总管", Defendant: "失宠外戚", Charge: "伪传懿旨", Stakes: "后宫名分、储位流言与外戚体面交织", Source: "seed:palace", Portrait: "empress", Heat: 48, Evidence: 46, FactionPressure: 66, PublicPressure: 52},
		{ID: "frontier-pay", Title: "边饷冒领案", Domain: DomainMilitary, Accuser: "五军都督府", Defendant: "边镇牙将", Charge: "虚冒军籍", Stakes: "若审得太狠，边镇会说朝廷薄待军功", Source: "seed:frontier", Portrait: "general", Heat: 50, Evidence: 54, FactionPressure: 58, PublicPressure: 45},
	}
	switch dynastyID {
	case "chengping":
		cases[0].Heat += 10
		cases[1].PublicPressure += 8
	case "jingyao":
		cases[1].Heat += 8
		cases[2].FactionPressure += 6
	case "xuanshuo":
		cases[3].Heat += 12
		cases[3].FactionPressure += 8
	case "dayin":
		cases[3].Evidence += 8
	}
	return cases
}

func (s *GameState) applyJusticeOrder(req OrderRequest) (Effects, string, bool, error) {
	switch req.Kind {
	case OrderOpenTrial:
		effects, summary, err := s.openTrial(req.Target)
		return effects, summary, true, err
	case OrderClemency:
		effects, summary, err := s.grantClemency(req.Target)
		return effects, summary, true, err
	case OrderCensorRumor:
		effects, summary, err := s.censorRumor(req.Target)
		return effects, summary, true, err
	case OrderProclaimVerdict:
		effects, summary, err := s.proclaimVerdict(req.Target)
		return effects, summary, true, err
	default:
		return Effects{}, "", false, nil
	}
}

func (s *GameState) openTrial(caseID string) (Effects, string, error) {
	i, ok := s.findLegalCaseIndex(caseID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown legal case target %q", caseID)
	}
	c := s.LegalCases[i]
	if c.Resolved {
		return Effects{}, "", fmt.Errorf("legal case %q is already resolved", c.Title)
	}
	authority := officeAuthority(s.Offices, "censorate")
	strength := c.Evidence + authority/3 + s.Stats.Legitimacy/8 - c.FactionPressure/8
	c.Resolved = true
	if strength >= 68 {
		c.Verdict = "明正典刑"
		s.applyCaseVerdict(c, 5, -10, 8, 4)
		s.shiftRelationsForDomain(c.Domain, -2, 5)
	} else {
		c.Verdict = "证据未尽"
		s.applyCaseVerdict(c, -2, -4, 2, -4)
		s.shiftRelationsForDomain(c.Domain, -1, 3)
	}
	c.Heat = clamp(c.Heat-24, 0, 100)
	c.FactionPressure = clamp(c.FactionPressure+6, 0, 100)
	s.LegalCases[i] = c
	effects := Effects{Influence: 4, Stability: -1, Legitimacy: 2}
	if c.Verdict == "证据未尽" {
		effects = Effects{Influence: 2, Stability: -3, Legitimacy: -1}
	}
	return effects, fmt.Sprintf("三司会审%s，%s被押上堂。判词为“%s”，舆论法度升至%d，谣言降至%d。", c.Title, c.Defendant, c.Verdict, s.PublicOpinion.Justice, s.PublicOpinion.Rumor), nil
}

func (s *GameState) grantClemency(caseID string) (Effects, string, error) {
	i, ok := s.findLegalCaseIndex(caseID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown legal case target %q", caseID)
	}
	c := s.LegalCases[i]
	if c.Resolved {
		return Effects{}, "", fmt.Errorf("legal case %q is already resolved", c.Title)
	}
	c.Resolved = true
	c.Verdict = "宽赦留用"
	c.Heat = clamp(c.Heat-16, 0, 100)
	c.FactionPressure = clamp(c.FactionPressure-14, 0, 100)
	s.LegalCases[i] = c
	s.PublicOpinion.Popular = clamp(s.PublicOpinion.Popular+4, 0, 100)
	s.PublicOpinion.Elite = clamp(s.PublicOpinion.Elite-3, 0, 100)
	s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor+4, 0, 100)
	s.PublicOpinion.Fear = clamp(s.PublicOpinion.Fear-6, 0, 100)
	s.PublicOpinion.LastEdict = fmt.Sprintf("%s从轻发落，朝野暂得喘息。", c.Title)
	s.shiftRelationsForDomain(c.Domain, 4, -4)
	effects := Effects{Stability: 4, Legitimacy: -2, Influence: -1}
	return effects, fmt.Sprintf("你宽赦%s，保住了%s的体面。民间称仁，清议却开始追问法度。", c.Title, c.Defendant), nil
}

func (s *GameState) censorRumor(target string) (Effects, string, error) {
	before := s.PublicOpinion.Rumor
	cut := 16 + s.Stats.Influence/12 + officeAuthority(s.Offices, "palace-secretary")/24
	s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor-cut, 0, 100)
	s.PublicOpinion.Fear = clamp(s.PublicOpinion.Fear+10, 0, 100)
	s.PublicOpinion.Elite = clamp(s.PublicOpinion.Elite-2, 0, 100)
	if i, ok := s.findLegalCaseIndex(target); ok {
		c := s.LegalCases[i]
		c.Heat = clamp(c.Heat-12, 0, 100)
		c.PublicPressure = clamp(c.PublicPressure-8, 0, 100)
		s.LegalCases[i] = c
		s.PublicOpinion.LastEdict = fmt.Sprintf("%s相关传帖被禁，京城忽然安静。", c.Title)
	} else {
		s.PublicOpinion.LastEdict = "茶楼讲史、坊间传帖与宫门小报被一夜清查。"
	}
	effects := Effects{Influence: 3, Legitimacy: -3, Stability: -1}
	return effects, fmt.Sprintf("禁谣令下，流言从%d压到%d；但百姓说话更轻，畏惧升至%d。", before, s.PublicOpinion.Rumor, s.PublicOpinion.Fear), nil
}

func (s *GameState) proclaimVerdict(caseID string) (Effects, string, error) {
	i, ok := s.findLegalCaseIndex(caseID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown legal case target %q", caseID)
	}
	c := s.LegalCases[i]
	if !c.Resolved {
		return Effects{}, "", fmt.Errorf("legal case %q has no verdict to proclaim", c.Title)
	}
	boost := 6 + s.Stats.Charisma/20
	s.PublicOpinion.Popular = clamp(s.PublicOpinion.Popular+boost, 0, 100)
	s.PublicOpinion.Elite = clamp(s.PublicOpinion.Elite+3, 0, 100)
	s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor-8, 0, 100)
	s.PublicOpinion.Justice = clamp(s.PublicOpinion.Justice+4, 0, 100)
	s.PublicOpinion.LastEdict = fmt.Sprintf("判词榜示天下：%s，%s。", c.Title, c.Verdict)
	effects := Effects{Legitimacy: 4, Populace: 2, Stability: 1}
	return effects, fmt.Sprintf("你命礼部将%s榜示天下。民望升至%d，士论升至%d。", c.Title, s.PublicOpinion.Popular, s.PublicOpinion.Elite), nil
}

func (s *GameState) applyJusticePressure(domain Domain) {
	if len(s.LegalCases) == 0 {
		return
	}
	highestHeat := 0
	openCases := 0
	for i, c := range s.LegalCases {
		if c.Resolved {
			c.Heat = clamp(c.Heat-2, 0, 100)
			s.LegalCases[i] = c
			continue
		}
		openCases++
		advance := 3 + c.PublicPressure/24 + c.FactionPressure/28
		if domain == DomainIntrigue || domain == c.Domain {
			advance--
		}
		c.Heat = clamp(c.Heat+advance, 0, 100)
		highestHeat = max(highestHeat, c.Heat)
		if c.Heat >= 96 {
			s.Crisis.Severity = clamp(s.Crisis.Severity+3, 0, 100)
			s.Crisis.Clock = clamp(s.Crisis.Clock+1, 0, 8)
			s.Stats.Stability = clamp(s.Stats.Stability-2, 0, 100)
		}
		s.LegalCases[i] = c
	}
	s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor+openCases+highestHeat/35, 0, 100)
	if s.PublicOpinion.Rumor >= 70 {
		s.Stats.Stability = clamp(s.Stats.Stability-2, 0, 100)
	}
	if s.PublicOpinion.Fear >= 70 {
		s.Stats.Populace = clamp(s.Stats.Populace-2, 0, 100)
	}
	if s.PublicOpinion.Justice >= 70 && s.PublicOpinion.Rumor <= 45 {
		s.Stats.Legitimacy = clamp(s.Stats.Legitimacy+1, 0, 100)
	}
	s.seedSystemicCases()
}

func (s *GameState) seedCaseFromPlot(plot Plot) {
	if plot.ID == "" || caseSourceExists(s.LegalCases, "plot:"+plot.ID) {
		return
	}
	c := LegalCase{
		ID:              "case-" + plot.ID,
		Title:           plot.Title + "案",
		Domain:          plot.Domain,
		Accuser:         "缇骑密档",
		Defendant:       plot.Sponsor,
		Charge:          plot.Target + "谋逆牵连",
		Stakes:          plot.Summary,
		Source:          "plot:" + plot.ID,
		Portrait:        "spy",
		Heat:            clamp(46+plot.Danger/3, 0, 100),
		Evidence:        clamp(72-plot.Secrecy/3, 25, 100),
		FactionPressure: clamp(38+plot.Danger/4, 0, 100),
		PublicPressure:  clamp(42+plot.Progress/3, 0, 100),
	}
	s.LegalCases = append(s.LegalCases, c)
	s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor+6, 0, 100)
	s.PublicOpinion.LastEdict = fmt.Sprintf("%s浮出水面，法司新立案卷。", c.Title)
}

func (s *GameState) seedSystemicCases() {
	if s.Succession.Dispute >= 72 && !caseSourceExists(s.LegalCases, "system:succession") {
		s.LegalCases = append(s.LegalCases, LegalCase{ID: "case-succession-rumor", Title: "东宫谤书案", Domain: DomainCourt, Accuser: "詹事府", Defendant: "失名清客", Charge: "散布储位谤书", Stakes: "审轻了压不住流言，审重了会牵动母族", Source: "system:succession", Portrait: "prince", Heat: 58, Evidence: 44, FactionPressure: 70, PublicPressure: 68})
	}
	if averageOfficeVacancyRisk(s.Offices) >= 64 && !caseSourceExists(s.LegalCases, "system:office") {
		s.LegalCases = append(s.LegalCases, LegalCase{ID: "case-office-backlog", Title: "官署积牍问责案", Domain: DomainReform, Accuser: "内阁票拟房", Defendant: "空缺官署", Charge: "压奏误政", Stakes: "问责能立威，也会暴露中枢空转", Source: "system:office", Portrait: "minister", Heat: 50, Evidence: 62, FactionPressure: 48, PublicPressure: 55})
	}
	if worstProvinceDisaster(s.Provinces) >= 72 && !caseSourceExists(s.LegalCases, "system:disaster") {
		s.LegalCases = append(s.LegalCases, LegalCase{ID: "case-disaster-embezzle", Title: "灾银侵吞案", Domain: DomainDomestic, Accuser: "灾民联名", Defendant: "赈务胥吏", Charge: "侵吞灾银", Stakes: "灾民盯着判词，地方官盯着风向", Source: "system:disaster", Portrait: "farmer", Heat: 72, Evidence: 55, FactionPressure: 38, PublicPressure: 82})
	}
}

func (s *GameState) ensureJusticeSystems() {
	if len(s.LegalCases) == 0 {
		s.LegalCases = startingLegalCases(s.Dynasty.ID)
	}
	if s.PublicOpinion.Popular <= 0 && s.PublicOpinion.Elite <= 0 {
		s.PublicOpinion = startingPublicOpinion(s.Dynasty.ID)
	}
}

func (s *GameState) findLegalCaseIndex(id string) (int, bool) {
	for i, c := range s.LegalCases {
		if c.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) applyCaseVerdict(c LegalCase, popular, rumor, justice, elite int) {
	s.PublicOpinion.Popular = clamp(s.PublicOpinion.Popular+popular+c.PublicPressure/40, 0, 100)
	s.PublicOpinion.Rumor = clamp(s.PublicOpinion.Rumor+rumor, 0, 100)
	s.PublicOpinion.Justice = clamp(s.PublicOpinion.Justice+justice, 0, 100)
	s.PublicOpinion.Elite = clamp(s.PublicOpinion.Elite+elite-c.FactionPressure/35, 0, 100)
	s.PublicOpinion.Fear = clamp(s.PublicOpinion.Fear+3+c.FactionPressure/35, 0, 100)
	s.PublicOpinion.LastEdict = fmt.Sprintf("%s判为%s。", c.Title, c.Verdict)
}

func caseSourceExists(cases []LegalCase, source string) bool {
	for _, c := range cases {
		if c.Source == source {
			return true
		}
	}
	return false
}

func maxLegalCaseHeat(cases []LegalCase) int {
	heat := 0
	for _, c := range cases {
		if !c.Resolved {
			heat = max(heat, c.Heat)
		}
	}
	return heat
}

func openLegalCasePressure(cases []LegalCase) int {
	total := 0
	for _, c := range cases {
		if !c.Resolved {
			total += c.Heat/2 + c.PublicPressure/3 + c.FactionPressure/4
		}
	}
	return total
}
