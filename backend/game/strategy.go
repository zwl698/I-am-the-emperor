package game

import (
	"fmt"
	"strings"
)

type ImperialProject struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Domain      Domain   `json:"domain"`
	Tier        int      `json:"tier"` // 1=基础, 2=进阶, 3=鼎新
	Stage       string   `json:"stage"`
	Progress    int      `json:"progress"`
	Investment  int      `json:"investment"`
	Risk        int      `json:"risk"`
	Completed   bool     `json:"completed"`
	Locked      bool     `json:"locked"`  // 是否锁定（前置条件未满足）
	Prereqs     []string `json:"prereqs"` // 前置项目ID列表（全部完成才解锁）
	Unlocks     []string `json:"unlocks"` // 完成后解锁的项目ID列表
	Reward      string   `json:"reward"`
	Description string   `json:"description"`
	Synergy     string   `json:"synergy"` // 协同效应描述
}

type StandingPolicy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      Domain `json:"domain"`
	Active      bool   `json:"active"`
	Upkeep      int    `json:"upkeep"`
	Strain      int    `json:"strain"`
	Description string `json:"description"`
}

type Relation struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Bond        string `json:"bond"`
	Trust       int    `json:"trust"`
	Tension     int    `json:"tension"`
	Description string `json:"description"`
}

func startingProjects(dynastyID string) []ImperialProject {
	projects := []ImperialProject{
		// ─── 第一层：基础工程（无前置，开局可造） ───
		{ID: "grand-canal", Name: "天下漕渠", Domain: DomainEconomy, Tier: 1, Stage: "勘河", Progress: 18, Investment: 10, Risk: 34, Locked: false, Prereqs: nil, Unlocks: []string{"salt-iron-reform", "imperial-granary"}, Reward: "粮草与国库长期收益", Description: "贯通江南粮道，降低灾荒和财政波动。", Synergy: "与太学新院协同：新法官僚可管理漕运商税，新政+2/季"},
		{ID: "border-arsenal", Name: "北府武库", Domain: DomainMilitary, Tier: 1, Stage: "铸炉", Progress: 14, Investment: 11, Risk: 38, Locked: false, Prereqs: nil, Unlocks: []string{"fortress-chain", "elite-cavalry"}, Reward: "军力与边防长期收益", Description: "扩建武库与马政，支撑外战后半程。", Synergy: "与缇骑密档库协同：密档提供敌军动向，边患额外-1/季"},
		{ID: "state-academy", Name: "太学新院", Domain: DomainReform, Tier: 1, Stage: "选址", Progress: 20, Investment: 8, Risk: 26, Locked: false, Prereqs: nil, Unlocks: []string{"civil-service", "codified-law"}, Reward: "新政与人才长期收益", Description: "培养新法官僚，让改革不只靠几位能臣。", Synergy: "与宫禁礼法协同：礼法与科举并行，稳定+1/季"},
		{ID: "palace-code", Name: "宫禁礼法", Domain: DomainCourt, Tier: 1, Stage: "议礼", Progress: 16, Investment: 7, Risk: 32, Locked: false, Prereqs: nil, Unlocks: []string{"censorate-reform", "harem-regulation"}, Reward: "继承与后宫长期稳定", Description: "重定宫闱礼制，压低外戚和储位冲突。", Synergy: "与太学新院协同：礼法与科举并行，稳定+1/季"},
		{ID: "secret-archive", Name: "缇骑密档库", Domain: DomainIntrigue, Tier: 1, Stage: "清卷", Progress: 12, Investment: 9, Risk: 45, Locked: false, Prereqs: nil, Unlocks: []string{"shadow-network", "loyalty-audit"}, Reward: "暗线检定与派系压制", Description: "整合密档、线人和巡按，提前发现阴谋。", Synergy: "与北府武库协同：密档提供敌军动向，边患额外-1/季"},

		// ─── 第二层：进阶工程（需完成前置） ───
		{ID: "salt-iron-reform", Name: "盐铁新制", Domain: DomainEconomy, Tier: 2, Stage: "筹备", Progress: 0, Investment: 14, Risk: 48, Locked: true, Prereqs: []string{"grand-canal"}, Unlocks: []string{"imperial-mint"}, Reward: "国库大幅增收，商帮势力受控", Description: "收归盐铁专卖，统一定价与转运，断豪商之利。", Synergy: "与市舶章程协同：海陆商路并控，国库额外+2/季"},
		{ID: "imperial-granary", Name: "常平大仓", Domain: DomainEconomy, Tier: 2, Stage: "筹备", Progress: 0, Investment: 12, Risk: 36, Locked: true, Prereqs: []string{"grand-canal"}, Unlocks: []string{}, Reward: "灾荒损失减半，民心长期稳定", Description: "丰年收储、荒年平粜，让粮价不再随天意翻覆。", Synergy: "与轻徭薄赋协同：仓廪充则民心安，民心额外+1/季"},
		{ID: "fortress-chain", Name: "烽燧连城", Domain: DomainMilitary, Tier: 2, Stage: "筹备", Progress: 0, Investment: 13, Risk: 42, Locked: true, Prereqs: []string{"border-arsenal"}, Unlocks: []string{"great-wall"}, Reward: "边患大幅下降，北境长治久安", Description: "沿边修筑连环堡垒与烽燧，使游骑无处渗透。", Synergy: "与边备常制协同：堡垒需常备军驻守，军力+1/季"},
		{ID: "elite-cavalry", Name: "铁骑营", Domain: DomainMilitary, Tier: 2, Stage: "筹备", Progress: 0, Investment: 15, Risk: 50, Locked: true, Prereqs: []string{"border-arsenal"}, Unlocks: []string{}, Reward: "外战推进加速，敌军威胁骤降", Description: "精选良马锐士，练成一支可深入草原的精锐骑兵。", Synergy: "与烽燧连城协同：骑出击、堡据守，边患额外-2/季"},
		{ID: "civil-service", Name: "科举取士", Domain: DomainReform, Tier: 2, Stage: "筹备", Progress: 0, Investment: 10, Risk: 35, Locked: true, Prereqs: []string{"state-academy"}, Unlocks: []string{"meritocracy"}, Reward: "人才池持续扩充，派系权力自动再平衡", Description: "以考试取代门荫，让寒门才俊入仕为新法输血。", Synergy: "与肃贪巡按协同：新进士子可充任巡按，势力额外+1/季"},
		{ID: "codified-law", Name: "律令成典", Domain: DomainReform, Tier: 2, Stage: "筹备", Progress: 0, Investment: 11, Risk: 40, Locked: true, Prereqs: []string{"state-academy"}, Unlocks: []string{}, Reward: "法度长期上升，阴谋和派系失控减少", Description: "将数百年散乱敕令编纂为一部统一法典。", Synergy: "与宫禁礼法协同：法典约束内外，稳定额外+1/季"},
		{ID: "censorate-reform", Name: "都察院改制", Domain: DomainCourt, Tier: 2, Stage: "筹备", Progress: 0, Investment: 10, Risk: 38, Locked: true, Prereqs: []string{"palace-code"}, Unlocks: []string{}, Reward: "派系权力自动受控，朝稳长期上升", Description: "将都察院从派系工具变为独立监察，打断权臣结网。", Synergy: "与科举取士协同：新御史由科举出身，不依附派系，稳定+2/季"},
		{ID: "harem-regulation", Name: "后宫改制", Domain: DomainCourt, Tier: 2, Stage: "筹备", Progress: 0, Investment: 8, Risk: 30, Locked: true, Prereqs: []string{"palace-code"}, Unlocks: []string{}, Reward: "储位争端大幅减少，外戚势力受控", Description: "限定后族干政范围，裁撤冗余宫职，统一储位礼制。", Synergy: "与宫闱均恩协同：制度与恩宠并行，储位压力减半"},
		{ID: "shadow-network", Name: "暗桩天下", Domain: DomainIntrigue, Tier: 2, Stage: "筹备", Progress: 0, Investment: 13, Risk: 55, Locked: true, Prereqs: []string{"secret-archive"}, Unlocks: []string{}, Reward: "所有阴谋提前暴露，暗线检定大幅加强", Description: "在每一支派系、每一处边镇都埋下听风之人。", Synergy: "与都察院改制协同：暗桩提供线索，御史弹劾，势力额外+2/季"},
		{ID: "loyalty-audit", Name: "忠诚清丈", Domain: DomainIntrigue, Tier: 2, Stage: "筹备", Progress: 0, Investment: 11, Risk: 44, Locked: true, Prereqs: []string{"secret-archive"}, Unlocks: []string{}, Reward: "派系忠诚自动回升，权臣野心受压制", Description: "定期清查官员田产与交往，让暗通款曲者无处藏身。", Synergy: "与肃贪巡按协同：清丈发现贪腐，巡按跟进，双管齐下"},

		// ─── 第三层：鼎新工程（需完成多条前置链） ───
		{ID: "imperial-mint", Name: "皇家铸币监", Domain: DomainEconomy, Tier: 3, Stage: "筹备", Progress: 0, Investment: 18, Risk: 55, Locked: true, Prereqs: []string{"salt-iron-reform", "state-academy"}, Unlocks: []string{}, Reward: "国库与新政长期大幅提升，铸币权归中央", Description: "统一天下铸币，夺回地方私铸之利，彻底掌控通货。", Synergy: "与盐铁新制、科举取士三联：财政-官僚-货币三位一体，国库额外+3/季"},
		{ID: "great-wall", Name: "万里长城", Domain: DomainMilitary, Tier: 3, Stage: "筹备", Progress: 0, Investment: 20, Risk: 60, Locked: true, Prereqs: []string{"fortress-chain", "grand-canal"}, Unlocks: []string{}, Reward: "边患几乎归零，北方长治久安", Description: "连缀旧墙、新筑关隘，使北境防线成为不可逾越的天堑。", Synergy: "与烽燧连城、常平大仓三联：兵精粮足墙高，边患额外-3/季"},
		{ID: "meritocracy", Name: "唯才是举", Domain: DomainReform, Tier: 3, Stage: "筹备", Progress: 0, Investment: 16, Risk: 52, Locked: true, Prereqs: []string{"civil-service", "codified-law"}, Unlocks: []string{}, Reward: "所有官署能力提升，派系自动弱化", Description: "以考课与政绩取代出身与恩荫，让天下官位归有能者。", Synergy: "与科举取士、都察院改制三联：选、考、监闭环，新政额外+3/季"},
	}
	switch dynastyID {
	case "chengping":
		adjustProjectProgress(projects, "grand-canal", -8)
		adjustProjectRisk(projects, "grand-canal", 10)
		adjustProjectProgress(projects, "state-academy", 6)
	case "xuanshuo":
		adjustProjectProgress(projects, "border-arsenal", 8)
		adjustProjectRisk(projects, "border-arsenal", 8)
	case "jingyao":
		adjustProjectProgress(projects, "grand-canal", 8)
		adjustProjectRisk(projects, "palace-code", 8)
	case "dayin":
		adjustProjectProgress(projects, "border-arsenal", 6)
	}
	return projects
}

func adjustProjectProgress(projects []ImperialProject, id string, delta int) {
	for i := range projects {
		if projects[i].ID == id {
			projects[i].Progress = clamp(projects[i].Progress+delta, 0, 100)
			return
		}
	}
}

func adjustProjectRisk(projects []ImperialProject, id string, delta int) {
	for i := range projects {
		if projects[i].ID == id {
			projects[i].Risk = clamp(projects[i].Risk+delta, 0, 100)
			return
		}
	}
}

func startingPolicies(dynastyID string) []StandingPolicy {
	policies := []StandingPolicy{
		{ID: "light-tax", Name: "轻徭薄赋", Domain: DomainDomestic, Upkeep: 5, Strain: 1, Description: "每季耗国库，换民心与稳定。"},
		{ID: "frontier-ready", Name: "边备常制", Domain: DomainMilitary, Upkeep: 6, Strain: 2, Description: "每季耗粮银，压边患并稳士气。"},
		{ID: "anti-corruption", Name: "肃贪巡按", Domain: DomainIntrigue, Upkeep: 4, Strain: 4, Description: "提高势力和新政，增加官场压力。"},
		{ID: "market-charter", Name: "市舶章程", Domain: DomainEconomy, Upkeep: 2, Strain: 3, Description: "增加财政外交，商帮权势上升。"},
		{ID: "palace-balance", Name: "宫闱均恩", Domain: DomainCourt, Upkeep: 3, Strain: 2, Description: "压低储位争议，略损个人势力。"},
	}
	if dynastyID == "xuanshuo" {
		policies[1].Active = true
	}
	if dynastyID == "jingyao" {
		policies[3].Active = true
	}
	return policies
}

func startingRelations(dynastyID string) []Relation {
	relations := []Relation{
		{ID: "scholar-reformer", From: "清流士林", To: "新法官僚", Bond: "理念", Trust: 44, Tension: 42, Description: "士林支持清议，但未必支持新法速度。"},
		{ID: "border-court", From: "边镇武勋", To: "中枢文臣", Bond: "粮饷", Trust: 38, Tension: 55, Description: "边镇要粮要功，中枢怕其尾大不掉。"},
		{ID: "merchant-revenue", From: "漕运商帮", To: "户部", Bond: "税银", Trust: 48, Tension: 38, Description: "商帮能救财政，也能绑架财政。"},
		{ID: "empress-consort", From: "中宫", To: "贵妃母族", Bond: "储位", Trust: 32, Tension: 62, Description: "后宫的礼数越周全，暗处的账越细。"},
		{ID: "heir-ministers", From: "东宫", To: "顾命群臣", Bond: "名分", Trust: 56, Tension: 28, Description: "储君需要群臣背书，群臣也会借储君自保。"},
		{ID: "envoy-frontier", From: "鸿胪寺", To: "边镇将门", Bond: "战和", Trust: 40, Tension: 46, Description: "外交缓战时，武臣总担心军功被谈判偷走。"},
	}
	switch dynastyID {
	case "dayin":
		relations[1].Trust += 8
	case "chengping":
		relations[0].Tension += 8
		relations[2].Tension += 6
	case "xuanshuo":
		relations[1].Tension += 8
	case "jingyao":
		relations[3].Tension += 8
	}
	return relations
}

func (s *GameState) applyGrandStrategyOrder(req OrderRequest) (Effects, string, bool, error) {
	switch req.Kind {
	case OrderFundProject:
		effects, summary, err := s.fundProject(req.Target)
		return effects, summary, true, err
	case OrderEnactPolicy:
		effects, summary, err := s.enactPolicy(req.Target)
		return effects, summary, true, err
	default:
		return Effects{}, "", false, nil
	}
}

func (s *GameState) fundProject(projectID string) (Effects, string, error) {
	i, ok := s.findProjectIndex(projectID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown project target %q", projectID)
	}
	project := s.Projects[i]
	if project.Completed {
		return Effects{}, "", fmt.Errorf("project %q is already completed", project.Name)
	}
	if project.Locked {
		return Effects{}, "", fmt.Errorf("project %q is locked: prerequisites not met", project.Name)
	}

	authority := bestOfficeAuthorityForDomain(s.Offices, project.Domain)
	advance := 11 + authority/18 + s.Stats.Reform/25
	if project.Domain == DomainMilitary {
		advance += s.Stats.Martial / 30
	}
	if project.Domain == DomainEconomy {
		advance += averageProvinceWealth(s.Provinces) / 35
	}
	// Higher tier projects advance slower
	if project.Tier == 2 {
		advance = advance * 85 / 100
	}
	if project.Tier == 3 {
		advance = advance * 70 / 100
	}

	project.Progress = clamp(project.Progress+advance, 0, 100)
	project.Risk = clamp(project.Risk-5+project.Investment/12, 0, 100)
	project.Stage = projectStage(project.Progress)
	completedNow := project.Progress >= 100
	project.Completed = completedNow
	s.Projects[i] = project
	s.shiftRelationsForDomain(project.Domain, 4, -2)
	effects := projectFundingEffects(project)
	summary := fmt.Sprintf("你拨给%s专款与人手，工程推进到%d/100，阶段进入「%s」。", project.Name, project.Progress, project.Stage)

	if completedNow {
		bonus := projectCompletionEffects(project)
		effects = mergeEffects(effects, bonus)

		// Unlock downstream projects
		unlockSummary := s.unlockDownstreamProjects(project)

		// Apply synergy effects
		synergyEffects, synergySummary := s.computeProjectSynergyEffects(project)
		effects = mergeEffects(effects, synergyEffects)

		summary = fmt.Sprintf("%s终于告成。%s开始兑现：%s。%s%s", project.Name, project.Description, project.Reward, unlockSummary, synergySummary)
	}
	return effects, summary, nil
}

// unlockDownstreamProjects unlocks any projects whose prerequisites include the just-completed project.
// Returns a human-readable summary of what was unlocked.
func (s *GameState) unlockDownstreamProjects(completed ImperialProject) string {
	var unlocked []string
	for i, project := range s.Projects {
		if !project.Locked || project.Completed {
			continue
		}
		if s.prereqsMet(project) {
			s.Projects[i].Locked = false
			unlocked = append(unlocked, project.Name)
		}
	}
	if len(unlocked) == 0 {
		return ""
	}
	return fmt.Sprintf("新工程解锁：%s。", strings.Join(unlocked, "、"))
}

// prereqsMet checks if all prerequisites for a project have been completed.
func (s *GameState) prereqsMet(project ImperialProject) bool {
	for _, prereqID := range project.Prereqs {
		found := false
		for _, p := range s.Projects {
			if p.ID == prereqID && p.Completed {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// computeProjectSynergyEffects checks for synergy bonuses when a project completes.
// Synergy is triggered when both the completed project AND its synergy partner are done.
func (s *GameState) computeProjectSynergyEffects(completed ImperialProject) (Effects, string) {
	synergyMap := projectSynergyPairs()
	partners, ok := synergyMap[completed.ID]
	if !ok {
		return Effects{}, ""
	}

	var totalEffects Effects
	var descriptions []string

	for _, partner := range partners {
		// Check if the partner project is also completed
		partnerCompleted := false
		for _, p := range s.Projects {
			if p.ID == partner.ID && p.Completed {
				partnerCompleted = true
				break
			}
		}
		if !partnerCompleted {
			continue
		}

		// Also check policy-based synergy partners
		if partner.PolicyID != "" {
			active := false
			for _, pol := range s.Policies {
				if pol.ID == partner.PolicyID && pol.Active {
					active = true
					break
				}
			}
			if !active {
				continue
			}
		}

		totalEffects = mergeEffects(totalEffects, partner.Bonus)
		descriptions = append(descriptions, partner.Desc)
	}

	if len(descriptions) == 0 {
		return Effects{}, ""
	}
	return totalEffects, fmt.Sprintf("协同生效：%s。", strings.Join(descriptions, "；"))
}

// synergyPartner describes a synergy relationship.
type synergyPartner struct {
	ID       string  // partner project ID (or empty if policy-based)
	PolicyID string  // partner policy ID (or empty if project-based)
	Bonus    Effects // synergy bonus effects
	Desc     string  // human-readable description
}

// projectSynergyPairs returns the map of project ID -> synergy partners.
func projectSynergyPairs() map[string][]synergyPartner {
	return map[string][]synergyPartner{
		// Tier 1 synergies (project + project)
		"grand-canal": {
			{ID: "state-academy", Bonus: Effects{Reform: 4}, Desc: "漕渠+太学：新法官僚管理漕运商税"},
			{ID: "border-arsenal", Bonus: Effects{BorderThreat: -3}, Desc: "漕渠+武库：粮道支撑边防补给"},
		},
		"border-arsenal": {
			{ID: "secret-archive", Bonus: Effects{BorderThreat: -3, Influence: 2}, Desc: "武库+密档：密档提供敌军动向"},
		},
		"state-academy": {
			{ID: "grand-canal", Bonus: Effects{Reform: 4}, Desc: "太学+漕渠：新法官僚管理漕运商税"},
			{ID: "palace-code", Bonus: Effects{Stability: 3}, Desc: "太学+礼法：礼法与科举并行"},
		},
		"palace-code": {
			{ID: "state-academy", Bonus: Effects{Stability: 3}, Desc: "礼法+太学：礼法与科举并行"},
		},
		"secret-archive": {
			{ID: "border-arsenal", Bonus: Effects{BorderThreat: -3, Influence: 2}, Desc: "密档+武库：密档提供敌军动向"},
		},

		// Tier 2 synergies (project + policy)
		"salt-iron-reform": {
			{PolicyID: "market-charter", Bonus: Effects{Treasury: 6}, Desc: "盐铁+市舶：海陆商路并控"},
		},
		"imperial-granary": {
			{PolicyID: "light-tax", Bonus: Effects{Populace: 3}, Desc: "常平仓+轻徭：仓廪充则民心安"},
		},
		"fortress-chain": {
			{PolicyID: "frontier-ready", Bonus: Effects{Army: 3}, Desc: "烽燧+边备：堡垒需常备军驻守"},
			{ID: "elite-cavalry", Bonus: Effects{BorderThreat: -5, Army: 3}, Desc: "烽燧+铁骑：骑出击、堡据守"},
		},
		"civil-service": {
			{PolicyID: "anti-corruption", Bonus: Effects{Influence: 3}, Desc: "科举+肃贪：新进士子充任巡按"},
		},
		"loyalty-audit": {
			{PolicyID: "anti-corruption", Bonus: Effects{Influence: 2, Stability: 2}, Desc: "清丈+肃贪：清丈发现贪腐，巡按跟进"},
		},
		"harem-regulation": {
			{PolicyID: "palace-balance", Bonus: Effects{Stability: 4}, Desc: "后宫改制+宫闱均恩：制度与恩宠并行"},
		},

		// Tier 2 project + project synergies
		"elite-cavalry": {
			{ID: "fortress-chain", Bonus: Effects{BorderThreat: -5, Army: 3}, Desc: "铁骑+烽燧：骑出击、堡据守"},
		},
		"codified-law": {
			{ID: "palace-code", Bonus: Effects{Stability: 3, Legitimacy: 2}, Desc: "律典+礼法：法典约束内外"},
		},
		"censorate-reform": {
			{ID: "civil-service", Bonus: Effects{Stability: 4, Reform: 3}, Desc: "都察院+科举：新御史不依附派系"},
		},
		"shadow-network": {
			{ID: "censorate-reform", Bonus: Effects{Influence: 5, Stability: 2}, Desc: "暗桩+都察院：暗桩提供线索，御史弹劾"},
		},

		// Tier 3 grand synergies (three-way)
		"imperial-mint": {
			{ID: "salt-iron-reform", Bonus: Effects{Treasury: 8, Reform: 4}, Desc: "铸币+盐铁：财政-货币双统"},
			{ID: "civil-service", Bonus: Effects{Treasury: 5, Reform: 3}, Desc: "铸币+科举：财政-官僚-货币三位一体"},
		},
		"great-wall": {
			{ID: "fortress-chain", Bonus: Effects{BorderThreat: -8, Army: 5}, Desc: "长城+烽燧：兵精粮足墙高"},
			{ID: "imperial-granary", Bonus: Effects{BorderThreat: -4, Grain: 6}, Desc: "长城+常平仓：北方防线粮草无忧"},
		},
		"meritocracy": {
			{ID: "civil-service", Bonus: Effects{Reform: 6, Stability: 4}, Desc: "唯才+科举：选、考、监闭环"},
			{ID: "censorate-reform", Bonus: Effects{Reform: 4, Influence: 4}, Desc: "唯才+都察院：选、考、监闭环"},
		},
	}
}

func (s *GameState) enactPolicy(policyID string) (Effects, string, error) {
	i, ok := s.findPolicyIndex(policyID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown policy target %q", policyID)
	}
	policy := s.Policies[i]
	if policy.Active {
		policy.Active = false
		s.Policies[i] = policy
		s.shiftRelationsForDomain(policy.Domain, -2, -3)
		return Effects{Stability: 1}, fmt.Sprintf("你暂罢“%s”，朝局少了一项长期开销，也少了一层约束。", policy.Name), nil
	}
	active := 0
	for _, item := range s.Policies {
		if item.Active {
			active++
		}
	}
	if active >= 3 {
		return Effects{}, "", fmt.Errorf("standing policy limit reached")
	}
	policy.Active = true
	s.Policies[i] = policy
	s.shiftRelationsForDomain(policy.Domain, 3, policy.Strain)
	effects := policyImmediateEffects(policy)
	return effects, fmt.Sprintf("你颁行“%s”。%s 这会成为每季自动生效的国策。", policy.Name, policy.Description), nil
}

func (s *GameState) applyGrandStrategyPressure(domain Domain) {
	for i, policy := range s.Policies {
		if !policy.Active {
			continue
		}
		effects := policySeasonalEffects(policy)
		s.applyEffects(effects)
		s.Policies[i].Strain = clamp(policy.Strain+1, 0, 100)
		s.shiftRelationsForDomain(policy.Domain, 1, policy.Strain/8)
	}
	for i, project := range s.Projects {
		if project.Completed {
			s.applyEffects(projectPassiveEffects(project))
			continue
		}
		delta := 2
		if project.Domain == domain {
			delta = -2
		}
		s.Projects[i].Risk = clamp(project.Risk+delta, 0, 100)
		if s.Projects[i].Risk >= 75 {
			s.Crisis.Severity = clamp(s.Crisis.Severity+1, 0, 100)
		}
	}
}

func (s *GameState) ensureGrandStrategySystems() {
	if len(s.Projects) == 0 {
		s.Projects = startingProjects(s.Dynasty.ID)
	}
	if len(s.Policies) == 0 {
		s.Policies = startingPolicies(s.Dynasty.ID)
	}
	if len(s.Relations) == 0 {
		s.Relations = startingRelations(s.Dynasty.ID)
	}
}

func (s *GameState) findProjectIndex(id string) (int, bool) {
	for i, project := range s.Projects {
		if project.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) findPolicyIndex(id string) (int, bool) {
	for i, policy := range s.Policies {
		if policy.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) shiftRelationsForDomain(domain Domain, trust, tension int) {
	if len(s.Relations) == 0 {
		return
	}
	for i, relation := range s.Relations {
		if relationTouchesDomain(relation, domain) || i == 0 {
			relation.Trust = clamp(relation.Trust+trust, 0, 100)
			relation.Tension = clamp(relation.Tension+tension, 0, 100)
			s.Relations[i] = relation
		}
	}
}

func relationTouchesDomain(relation Relation, domain Domain) bool {
	switch domain {
	case DomainMilitary:
		return relation.ID == "border-court" || relation.ID == "envoy-frontier"
	case DomainEconomy:
		return relation.ID == "merchant-revenue"
	case DomainCourt:
		return relation.ID == "empress-consort" || relation.ID == "heir-ministers"
	case DomainReform:
		return relation.ID == "scholar-reformer"
	case DomainDiplomacy:
		return relation.ID == "envoy-frontier"
	case DomainIntrigue:
		return relation.ID == "scholar-reformer" || relation.ID == "empress-consort"
	default:
		return false
	}
}

func projectStage(progress int) string {
	switch {
	case progress >= 100:
		return "告成"
	case progress >= 70:
		return "合龙"
	case progress >= 35:
		return "大兴"
	default:
		return "筹备"
	}
}

func bestOfficeAuthorityForDomain(offices []Office, domain Domain) int {
	best := 40
	for _, office := range offices {
		if office.Domain == domain {
			best = max(best, office.Authority)
		}
	}
	return best
}

func projectFundingEffects(project ImperialProject) Effects {
	switch project.Domain {
	case DomainMilitary:
		return Effects{Treasury: -project.Investment, Grain: -4, Army: 2}
	case DomainEconomy:
		return Effects{Treasury: -project.Investment, Grain: 2, Reform: 1}
	case DomainReform:
		return Effects{Treasury: -project.Investment, Reform: 3, Stability: -1}
	case DomainCourt:
		return Effects{Treasury: -project.Investment, Stability: 2, Influence: -1}
	case DomainIntrigue:
		return Effects{Treasury: -project.Investment, Influence: 3, Stability: -1}
	default:
		return Effects{Treasury: -project.Investment}
	}
}

func projectCompletionEffects(project ImperialProject) Effects {
	switch project.ID {
	// Tier 1
	case "grand-canal":
		return Effects{Treasury: 10, Grain: 16, Populace: 4, Stability: 3}
	case "border-arsenal":
		return Effects{Army: 14, BorderThreat: -10, Martial: 2}
	case "state-academy":
		return Effects{Reform: 12, Learning: 4, Stability: 2}
	case "palace-code":
		return Effects{Stability: 8, Legitimacy: 4, Influence: 2}
	case "secret-archive":
		return Effects{Influence: 10, Stability: 2, Legitimacy: -1}
	// Tier 2 - Economy
	case "salt-iron-reform":
		return Effects{Treasury: 18, Grain: 4, Stability: -3, Influence: 3}
	case "imperial-granary":
		return Effects{Grain: 20, Populace: 8, Stability: 6}
	// Tier 2 - Military
	case "fortress-chain":
		return Effects{BorderThreat: -14, Army: 8, Stability: 4}
	case "elite-cavalry":
		return Effects{Army: 16, BorderThreat: -12, Martial: 3, Grain: -6}
	// Tier 2 - Reform
	case "civil-service":
		return Effects{Reform: 10, Learning: 6, Influence: -2, Stability: 2}
	case "codified-law":
		return Effects{Reform: 8, Stability: 6, Influence: 4, Legitimacy: 2}
	// Tier 2 - Court
	case "censorate-reform":
		return Effects{Stability: 8, Influence: 6, Legitimacy: 3}
	case "harem-regulation":
		return Effects{Stability: 6, Legitimacy: 4, Influence: 2}
	// Tier 2 - Intrigue
	case "shadow-network":
		return Effects{Influence: 14, Stability: -2, Legitimacy: -3}
	case "loyalty-audit":
		return Effects{Influence: 8, Stability: 4, Legitimacy: 2}
	// Tier 3 - Grand projects
	case "imperial-mint":
		return Effects{Treasury: 24, Reform: 8, Stability: 4, Legitimacy: 3}
	case "great-wall":
		return Effects{BorderThreat: -20, Army: 12, Stability: 8, Grain: -8}
	case "meritocracy":
		return Effects{Reform: 14, Learning: 6, Stability: 8, Influence: 4}
	default:
		return Effects{Stability: 4}
	}
}

func projectPassiveEffects(project ImperialProject) Effects {
	switch project.ID {
	// Tier 1
	case "grand-canal":
		return Effects{Treasury: 1, Grain: 2}
	case "border-arsenal":
		return Effects{Army: 1, BorderThreat: -1}
	case "state-academy":
		return Effects{Reform: 1}
	case "palace-code":
		return Effects{Stability: 1}
	case "secret-archive":
		return Effects{Influence: 1}
	// Tier 2 - Economy
	case "salt-iron-reform":
		return Effects{Treasury: 3, Influence: 1}
	case "imperial-granary":
		return Effects{Grain: 2, Populace: 1, Stability: 1}
	// Tier 2 - Military
	case "fortress-chain":
		return Effects{BorderThreat: -2, Army: 1}
	case "elite-cavalry":
		return Effects{Army: 2, BorderThreat: -1, Grain: -1}
	// Tier 2 - Reform
	case "civil-service":
		return Effects{Reform: 2, Learning: 1}
	case "codified-law":
		return Effects{Stability: 1, Reform: 1}
	// Tier 2 - Court
	case "censorate-reform":
		return Effects{Stability: 2, Influence: 1}
	case "harem-regulation":
		return Effects{Stability: 1, Legitimacy: 1}
	// Tier 2 - Intrigue
	case "shadow-network":
		return Effects{Influence: 2, Legitimacy: -1}
	case "loyalty-audit":
		return Effects{Influence: 1, Stability: 1}
	// Tier 3 - Grand
	case "imperial-mint":
		return Effects{Treasury: 4, Reform: 2, Stability: 1}
	case "great-wall":
		return Effects{BorderThreat: -3, Army: 2, Stability: 1}
	case "meritocracy":
		return Effects{Reform: 3, Learning: 1, Stability: 1}
	default:
		return Effects{}
	}
}

func policyImmediateEffects(policy StandingPolicy) Effects {
	switch policy.ID {
	case "light-tax":
		return Effects{Populace: 3, Stability: 2, Treasury: -policy.Upkeep}
	case "frontier-ready":
		return Effects{Army: 3, BorderThreat: -4, Treasury: -policy.Upkeep}
	case "anti-corruption":
		return Effects{Influence: 4, Reform: 2, Stability: -1}
	case "market-charter":
		return Effects{Treasury: 4, Diplomacy: 2}
	case "palace-balance":
		return Effects{Stability: 3, Influence: -1}
	default:
		return Effects{Stability: 1}
	}
}

func policySeasonalEffects(policy StandingPolicy) Effects {
	switch policy.ID {
	case "light-tax":
		return Effects{Treasury: -policy.Upkeep, Populace: 2, Stability: 1}
	case "frontier-ready":
		return Effects{Treasury: -policy.Upkeep, Grain: -2, BorderThreat: -2, Army: 1}
	case "anti-corruption":
		return Effects{Treasury: -policy.Upkeep, Influence: 1, Reform: 1, Stability: -1}
	case "market-charter":
		return Effects{Treasury: 3, Diplomacy: 1, BorderThreat: 1}
	case "palace-balance":
		return Effects{Treasury: -policy.Upkeep, Stability: 1}
	default:
		return Effects{}
	}
}

func mergeEffects(a, b Effects) Effects {
	return Effects{
		Legitimacy:   a.Legitimacy + b.Legitimacy,
		Health:       a.Health + b.Health,
		Learning:     a.Learning + b.Learning,
		Martial:      a.Martial + b.Martial,
		Charisma:     a.Charisma + b.Charisma,
		Influence:    a.Influence + b.Influence,
		Treasury:     a.Treasury + b.Treasury,
		Grain:        a.Grain + b.Grain,
		Populace:     a.Populace + b.Populace,
		Army:         a.Army + b.Army,
		Diplomacy:    a.Diplomacy + b.Diplomacy,
		Stability:    a.Stability + b.Stability,
		BorderThreat: a.BorderThreat + b.BorderThreat,
		Reform:       a.Reform + b.Reform,
	}
}
