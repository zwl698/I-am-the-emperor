package game

import (
	"fmt"
	"strings"
)

type Plot struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Sponsor  string `json:"sponsor"`
	Target   string `json:"target"`
	Domain   Domain `json:"domain"`
	Stage    string `json:"stage"`
	Summary  string `json:"summary"`
	Secrecy  int    `json:"secrecy"`
	Progress int    `json:"progress"`
	Danger   int    `json:"danger"`
	Exposed  bool   `json:"exposed"`
	Resolved bool   `json:"resolved"`
}

func startingPlots(dynastyID string) []Plot {
	plots := []Plot{
		{ID: "silk-ledger", Title: "丝账暗线", Sponsor: "漕运商帮", Target: "户部", Domain: DomainEconomy, Stage: "暗递名帖", Summary: "商帮用旧账牵住几名户部郎官。", Secrecy: 56, Progress: 24, Danger: 45},
		{ID: "palace-poison", Title: "宫酒疑云", Sponsor: "失宠外戚", Target: "东宫", Domain: DomainCourt, Stage: "买通内侍", Summary: "内廷酒食采买中多出陌生印记。", Secrecy: 52, Progress: 22, Danger: 58},
		{ID: "frontier-letter", Title: "边书私递", Sponsor: "骄兵悍将", Target: "兵部", Domain: DomainMilitary, Stage: "截留军报", Summary: "边镇有书信绕过中枢直入勋贵府。", Secrecy: 60, Progress: 28, Danger: 62},
		{ID: "censor-list", Title: "清议黑榜", Sponsor: "旧党御史", Target: "新法官员", Domain: DomainReform, Stage: "联名酝酿", Summary: "御史台有人准备集中弹劾新法骨干。", Secrecy: 50, Progress: 20, Danger: 48},
	}
	switch dynastyID {
	case "chengping":
		plots[0].Progress += 12
		plots[3].Danger += 8
	case "xuanshuo":
		plots[2].Progress += 12
	case "jingyao":
		plots[1].Secrecy += 8
	case "dayin":
		plots[2].Danger += 8
	}
	return plots
}

func (s *GameState) applyPlotOrder(req OrderRequest) (Effects, string, bool, error) {
	switch req.Kind {
	case OrderInvestigatePlot:
		effects, summary, err := s.investigatePlot(req.Target)
		return effects, summary, true, err
	case OrderSuppressPlot:
		effects, summary, err := s.suppressPlot(req.Target)
		return effects, summary, true, err
	case OrderEducateHeir:
		effects, summary, err := s.educateHeir(req.Target)
		return effects, summary, true, err
	default:
		return Effects{}, "", false, nil
	}
}

func (s *GameState) investigatePlot(plotID string) (Effects, string, error) {
	i, ok := s.findPlotIndex(plotID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown plot target %q", plotID)
	}
	plot := s.Plots[i]
	if plot.Resolved {
		return Effects{}, "", fmt.Errorf("plot %q is already resolved", plot.Title)
	}
	reveal := 24 + s.Stats.Influence/12 + officeAuthority(s.Offices, "censorate")/18
	plot.Secrecy = clamp(plot.Secrecy-reveal, 0, 100)
	plot.Progress = clamp(plot.Progress-8, 0, 100)
	plot.Exposed = plot.Secrecy <= 42
	plot.Stage = plotStage(plot)
	s.Plots[i] = plot
	if plot.Exposed {
		s.seedCaseFromPlot(plot)
	}
	effects := Effects{Treasury: -3, Influence: 3, Stability: -1}
	return effects, fmt.Sprintf("缇骑顺着%s追查%s，隐秘降至%d，进度压到%d。", plot.Sponsor, plot.Title, plot.Secrecy, plot.Progress), nil
}

func (s *GameState) suppressPlot(plotID string) (Effects, string, error) {
	i, ok := s.findPlotIndex(plotID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown plot target %q", plotID)
	}
	plot := s.Plots[i]
	if plot.Resolved {
		return Effects{}, "", fmt.Errorf("plot %q is already resolved", plot.Title)
	}
	if !plot.Exposed {
		return Effects{}, "", fmt.Errorf("plot %q is not exposed", plot.Title)
	}
	plot.Resolved = true
	plot.Progress = 0
	plot.Secrecy = 0
	plot.Stage = "结案"
	s.Plots[i] = plot
	s.shiftRelationsForDomain(plot.Domain, -2, 5)
	effects := Effects{Influence: 5, Stability: 2, Legitimacy: -1}
	return effects, fmt.Sprintf("你准奏处置%s，%s的线被连根拔起。朝堂安静了一些，也更怕你。", plot.Title, plot.Sponsor), nil
}

func (s *GameState) educateHeir(target string) (Effects, string, error) {
	heirID, focus, ok := strings.Cut(target, ":")
	if !ok || heirID == "" {
		return Effects{}, "", fmt.Errorf("education target must be heirID:focus")
	}
	i, ok := s.findHeirIndex(heirID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown heir target %q", heirID)
	}
	heir := s.Heirs[i]
	effects := Effects{Treasury: -3, Legitimacy: 1}
	switch focus {
	case "study":
		heir.Talent = clamp(heir.Talent+7, 0, 100)
		heir.Support = clamp(heir.Support+3, 0, 100)
		effects.Learning = 1
	case "drill":
		heir.Talent = clamp(heir.Talent+4, 0, 100)
		heir.Ambition = clamp(heir.Ambition+5, 0, 100)
		effects.Martial = 1
	case "rites":
		heir.Support = clamp(heir.Support+8, 0, 100)
		s.Succession.Stability = clamp(s.Succession.Stability+3, 0, 100)
		effects.Stability = 2
	default:
		return Effects{}, "", fmt.Errorf("unknown heir education focus %q", focus)
	}
	s.Heirs[i] = heir
	return effects, fmt.Sprintf("你命东宫师傅为%s加开%s课。资质升至%d，拥护升至%d。", heir.Name, educationName(focus), heir.Talent, heir.Support), nil
}

func (s *GameState) applyPlotPressure(domain Domain) {
	for i, plot := range s.Plots {
		if plot.Resolved {
			continue
		}
		advance := 3 + plot.Danger/18 + plot.Secrecy/35
		if plot.Exposed {
			advance -= 3
		}
		if domain == DomainIntrigue {
			advance -= 2
		}
		plot.Progress = clamp(plot.Progress+advance, 0, 100)
		plot.Stage = plotStage(plot)
		if plot.Progress >= 100 {
			s.applyPlotCrisis(plot)
			plot.Resolved = true
			plot.Exposed = true
			plot.Stage = "爆发"
		}
		s.Plots[i] = plot
	}
}

func (s *GameState) applyPlotCrisis(plot Plot) {
	switch plot.Domain {
	case DomainCourt:
		s.Succession.Dispute = clamp(s.Succession.Dispute+12, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-6, 0, 100)
	case DomainMilitary:
		s.Stats.Army = clamp(s.Stats.Army-8, 0, 140)
		s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+8, 0, 100)
	case DomainEconomy:
		s.Stats.Treasury = clamp(s.Stats.Treasury-12, 0, 160)
		s.Stats.Stability = clamp(s.Stats.Stability-3, 0, 100)
	case DomainReform:
		s.Stats.Reform = clamp(s.Stats.Reform-8, 0, 100)
		s.Stats.Legitimacy = clamp(s.Stats.Legitimacy-3, 0, 100)
	default:
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
	s.Crisis.Severity = clamp(s.Crisis.Severity+6, 0, 100)
}

func (s *GameState) ensurePlotSystems() {
	if len(s.Plots) == 0 {
		s.Plots = startingPlots(s.Dynasty.ID)
	}
}

func (s *GameState) findPlotIndex(id string) (int, bool) {
	for i, plot := range s.Plots {
		if plot.ID == id {
			return i, true
		}
	}
	return 0, false
}

func plotStage(plot Plot) string {
	switch {
	case plot.Resolved:
		return "结案"
	case plot.Progress >= 80:
		return "将发"
	case plot.Exposed:
		return "暴露"
	case plot.Progress >= 50:
		return "成形"
	default:
		return "潜伏"
	}
}

func educationName(focus string) string {
	switch focus {
	case "study":
		return "经史"
	case "drill":
		return "骑射"
	case "rites":
		return "礼法"
	default:
		return "课业"
	}
}
