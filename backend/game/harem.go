package game

import "fmt"

type Consort struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Rank        string   `json:"rank"`
	Clan        string   `json:"clan"`
	Trait       string   `json:"trait"`
	Favor       int      `json:"favor"`
	FamilyPower int      `json:"familyPower"`
	Ambition    int      `json:"ambition"`
	Influence   int      `json:"influence"`
	Health      int      `json:"health"`
	Children    []string `json:"children"`
	Portrait    string   `json:"portrait"`
}

type Heir struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MotherID string `json:"motherId"`
	Age      int    `json:"age"`
	Talent   int    `json:"talent"`
	Ambition int    `json:"ambition"`
	Support  int    `json:"support"`
	Health   int    `json:"health"`
	Named    bool   `json:"named"`
	Portrait string `json:"portrait"`
}

type Succession struct {
	NamedHeirID        string `json:"namedHeirId"`
	Stability          int    `json:"stability"`
	Dispute            int    `json:"dispute"`
	MaternalClanPower  int    `json:"maternalClanPower"`
	LastSuccessionMove string `json:"lastSuccessionMove"`
}

func startingHarem(dynastyID string) []Consort {
	harem := []Consort{
		{ID: "empress-lu", Name: "陆皇后", Rank: "皇后", Clan: "江左陆氏", Trait: "端肃", Favor: 52, FamilyPower: 56, Ambition: 38, Influence: 58, Health: 72, Children: []string{"heir-crown"}, Portrait: "empress"},
		{ID: "consort-xue", Name: "薛贵妃", Rank: "贵妃", Clan: "关陇薛氏", Trait: "明艳", Favor: 64, FamilyPower: 48, Ambition: 70, Influence: 52, Health: 68, Children: []string{"heir-second"}, Portrait: "consort"},
		{ID: "consort-chen", Name: "陈昭仪", Rank: "昭仪", Clan: "海陵陈氏", Trait: "善谋", Favor: 44, FamilyPower: 42, Ambition: 58, Influence: 45, Health: 75, Children: []string{"princess-qing"}, Portrait: "diplomat"},
		{ID: "consort-yan", Name: "颜才人", Rank: "才人", Clan: "寒门颜氏", Trait: "清雅", Favor: 36, FamilyPower: 25, Ambition: 28, Influence: 30, Health: 80, Portrait: "maid"},
	}
	switch dynastyID {
	case "dayin":
		harem[1].FamilyPower += 10
		harem[1].Ambition += 6
	case "jingyao":
		harem[2].Favor += 10
		harem[2].Influence += 8
	case "chengping":
		harem[0].FamilyPower += 8
		harem[1].Ambition += 10
	case "xuanshuo":
		harem[1].FamilyPower += 8
		harem[1].Influence += 7
	}
	return harem
}

func startingHeirs(dynastyID string) []Heir {
	heirs := []Heir{
		{ID: "heir-crown", Name: "萧承璟", MotherID: "empress-lu", Age: 6, Talent: 66, Ambition: 44, Support: 58, Health: 70, Named: true, Portrait: "teen-prince"},
		{ID: "heir-second", Name: "萧承曜", MotherID: "consort-xue", Age: 4, Talent: 76, Ambition: 68, Support: 46, Health: 74, Portrait: "infant-prince"},
		{ID: "princess-qing", Name: "萧清河", MotherID: "consort-chen", Age: 3, Talent: 72, Ambition: 36, Support: 38, Health: 78, Portrait: "princess"},
	}
	switch dynastyID {
	case "jingyao":
		heirs[0].Support += 8
		heirs[2].Talent += 8
	case "chengping":
		heirs[0].Support -= 8
		heirs[1].Support += 6
	case "xuanshuo":
		heirs[1].Support += 8
		heirs[1].Ambition += 6
	}
	return heirs
}

func startingSuccession(dynastyID string) Succession {
	succession := Succession{NamedHeirID: "heir-crown", Stability: 56, Dispute: 28, MaternalClanPower: 52, LastSuccessionMove: "太子年幼，中宫与东宫暂时同心。"}
	switch dynastyID {
	case "jingyao":
		succession.Stability += 10
		succession.Dispute -= 6
	case "chengping":
		succession.Stability -= 12
		succession.Dispute += 10
	case "xuanshuo":
		succession.MaternalClanPower += 8
		succession.Dispute += 6
	}
	return succession
}

func (s *GameState) nameHeir(heirID string) (Effects, string, error) {
	i, ok := s.findHeirIndex(heirID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown heir target %q", heirID)
	}
	heir := s.Heirs[i]
	previous := s.Succession.NamedHeirID
	for j := range s.Heirs {
		if s.Heirs[j].ID == heirID {
			s.Heirs[j].Named = true
			s.Heirs[j].Support = clamp(s.Heirs[j].Support+14, 0, 100)
		} else {
			s.Heirs[j].Named = false
			s.Heirs[j].Support = clamp(s.Heirs[j].Support-4, 0, 100)
		}
	}
	if ci, ok := s.findConsortIndex(heir.MotherID); ok {
		s.Harem[ci].FamilyPower = clamp(s.Harem[ci].FamilyPower+8, 0, 100)
		s.Harem[ci].Influence = clamp(s.Harem[ci].Influence+5, 0, 100)
		s.Harem[ci].Ambition = clamp(s.Harem[ci].Ambition+3, 0, 100)
		s.Succession.MaternalClanPower = clamp(s.Harem[ci].FamilyPower, 0, 100)
	}

	disputeDelta := 8
	stabilityDelta := 5 + heir.Talent/25 + heir.Support/30 - heir.Ambition/28
	if previous != "" && previous != heirID {
		disputeDelta += 6
		stabilityDelta -= 6
	}
	if heir.Age < 5 {
		disputeDelta += 4
		stabilityDelta -= 2
	}
	s.Succession.NamedHeirID = heirID
	s.Succession.Dispute = clamp(s.Succession.Dispute+disputeDelta, 0, 100)
	s.Succession.Stability = clamp(s.Succession.Stability+stabilityDelta, 0, 100)
	s.Succession.LastSuccessionMove = fmt.Sprintf("册立%s，东宫名分重写。", heir.Name)

	effects := Effects{Legitimacy: 2, Stability: stabilityDelta, Influence: 2}
	summary := fmt.Sprintf("礼部在太庙宣读册文，%s成为储君。拥护升至%d，储位稳定为%d，争议也升到%d。", heir.Name, s.Heirs[i].Support, s.Succession.Stability, s.Succession.Dispute)
	return effects, summary, nil
}

func (s *GameState) favorConsort(consortID string) (Effects, string, error) {
	i, ok := s.findConsortIndex(consortID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown consort target %q", consortID)
	}
	for j := range s.Harem {
		if j == i {
			s.Harem[j].Favor = clamp(s.Harem[j].Favor+16, 0, 100)
			s.Harem[j].Influence = clamp(s.Harem[j].Influence+6, 0, 100)
			s.Harem[j].FamilyPower = clamp(s.Harem[j].FamilyPower+4, 0, 100)
			s.Harem[j].Ambition = clamp(s.Harem[j].Ambition+3, 0, 100)
		} else {
			s.Harem[j].Favor = clamp(s.Harem[j].Favor-3, 0, 100)
		}
	}
	consort := s.Harem[i]
	dispute := 1
	if consort.Ambition >= 70 || consort.FamilyPower >= 70 {
		dispute = 4
	}
	s.Succession.Dispute = clamp(s.Succession.Dispute+dispute, 0, 100)
	s.Succession.Stability = clamp(s.Succession.Stability-dispute/2, 0, 100)
	effects := Effects{Health: 1, Stability: -dispute / 2, Influence: 1}
	return effects, fmt.Sprintf("你留宿%s宫，%s宠爱升至%d；%s也把这份恩宠记进外戚账本。", consort.Rank, consort.Name, consort.Favor, consort.Clan), nil
}

func (s *GameState) arrangeMarriageAlliance(target string) (Effects, string, error) {
	i, ok := s.findConsortIndex(target)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown marriage alliance target %q", target)
	}
	consort := s.Harem[i]
	s.Harem[i].FamilyPower = clamp(consort.FamilyPower+10, 0, 100)
	s.Harem[i].Influence = clamp(consort.Influence+8, 0, 100)
	s.Succession.MaternalClanPower = clamp(s.Succession.MaternalClanPower+6, 0, 100)
	s.Succession.Dispute = clamp(s.Succession.Dispute+3, 0, 100)
	effects := Effects{Diplomacy: 5, Stability: 2, Treasury: -4}
	return effects, fmt.Sprintf("你以%s为纽带与%s重修姻亲。朝堂暂稳，外戚势力也随之抬头。", consort.Name, consort.Clan), nil
}

func (s *GameState) applyPalacePressure(domain Domain) {
	if len(s.Harem) == 0 || len(s.Heirs) == 0 {
		return
	}
	if s.Season == "春" && s.Phase == PhaseEmperor {
		for i := range s.Heirs {
			s.Heirs[i].Age++
			if s.Heirs[i].Age >= 8 {
				s.Heirs[i].Support = clamp(s.Heirs[i].Support+s.Heirs[i].Talent/40, 0, 100)
			}
		}
	}
	if domain == DomainCourt {
		s.Succession.Stability = clamp(s.Succession.Stability+1, 0, 100)
	} else if s.Succession.Dispute >= 65 {
		s.Succession.Stability = clamp(s.Succession.Stability-2, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-1, 0, 100)
	}
	for i, consort := range s.Harem {
		if consort.Favor >= 70 {
			s.Harem[i].Influence = clamp(consort.Influence+1, 0, 100)
		}
		if consort.Ambition+consort.FamilyPower >= 145 {
			s.Succession.Dispute = clamp(s.Succession.Dispute+1, 0, 100)
		}
	}
}

func (s *GameState) findConsortIndex(id string) (int, bool) {
	for i, consort := range s.Harem {
		if consort.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) findHeirIndex(id string) (int, bool) {
	for i, heir := range s.Heirs {
		if heir.ID == id {
			return i, true
		}
	}
	return 0, false
}
