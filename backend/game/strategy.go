package game

import "fmt"

type ImperialProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      Domain `json:"domain"`
	Stage       string `json:"stage"`
	Progress    int    `json:"progress"`
	Investment  int    `json:"investment"`
	Risk        int    `json:"risk"`
	Completed   bool   `json:"completed"`
	Reward      string `json:"reward"`
	Description string `json:"description"`
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
		{ID: "grand-canal", Name: "天下漕渠", Domain: DomainEconomy, Stage: "勘河", Progress: 18, Investment: 10, Risk: 34, Reward: "粮草与国库长期收益", Description: "贯通江南粮道，降低灾荒和财政波动。"},
		{ID: "border-arsenal", Name: "北府武库", Domain: DomainMilitary, Stage: "铸炉", Progress: 14, Investment: 11, Risk: 38, Reward: "军力与边防长期收益", Description: "扩建武库与马政，支撑外战后半程。"},
		{ID: "state-academy", Name: "太学新院", Domain: DomainReform, Stage: "选址", Progress: 20, Investment: 8, Risk: 26, Reward: "新政与人才长期收益", Description: "培养新法官僚，让改革不只靠几位能臣。"},
		{ID: "palace-code", Name: "宫禁礼法", Domain: DomainCourt, Stage: "议礼", Progress: 16, Investment: 7, Risk: 32, Reward: "继承与后宫长期稳定", Description: "重定宫闱礼制，压低外戚和储位冲突。"},
		{ID: "secret-archive", Name: "缇骑密档库", Domain: DomainIntrigue, Stage: "清卷", Progress: 12, Investment: 9, Risk: 45, Reward: "暗线检定与派系压制", Description: "整合密档、线人和巡按，提前发现阴谋。"},
	}
	switch dynastyID {
	case "chengping":
		projects[0].Progress -= 8
		projects[0].Risk += 10
		projects[2].Progress += 6
	case "xuanshuo":
		projects[1].Progress += 8
		projects[1].Risk += 8
	case "jingyao":
		projects[0].Progress += 8
		projects[3].Risk += 8
	case "dayin":
		projects[1].Progress += 6
	}
	return projects
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
	authority := bestOfficeAuthorityForDomain(s.Offices, project.Domain)
	advance := 11 + authority/18 + s.Stats.Reform/25
	if project.Domain == DomainMilitary {
		advance += s.Stats.Martial / 30
	}
	if project.Domain == DomainEconomy {
		advance += averageProvinceWealth(s.Provinces) / 35
	}
	project.Progress = clamp(project.Progress+advance, 0, 100)
	project.Risk = clamp(project.Risk-5+project.Investment/12, 0, 100)
	project.Stage = projectStage(project.Progress)
	completedNow := project.Progress >= 100
	project.Completed = completedNow
	s.Projects[i] = project
	s.shiftRelationsForDomain(project.Domain, 4, -2)
	effects := projectFundingEffects(project)
	summary := fmt.Sprintf("你拨给%s专款与人手，工程推进到%d/100，阶段进入“%s”。", project.Name, project.Progress, project.Stage)
	if completedNow {
		bonus := projectCompletionEffects(project)
		effects = mergeEffects(effects, bonus)
		summary = fmt.Sprintf("%s终于告成。%s开始兑现：%s。", project.Name, project.Description, project.Reward)
	}
	return effects, summary, nil
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
	default:
		return Effects{Stability: 4}
	}
}

func projectPassiveEffects(project ImperialProject) Effects {
	switch project.ID {
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
