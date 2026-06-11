package game

import (
	"fmt"
	"strings"
)

type Office struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Domain      Domain `json:"domain"`
	HolderID    string `json:"holderId"`
	Authority   int    `json:"authority"`
	VacancyRisk int    `json:"vacancyRisk"`
	Seat        string `json:"seat"`
}

func startingOffices(dynastyID string) []Office {
	offices := []Office{
		{ID: "grand-secretariat", Title: "内阁首辅", Domain: DomainReform, HolderID: "gu", Authority: 72, VacancyRisk: 14, Seat: "票拟、封驳、新法"},
		{ID: "revenue-ministry", Title: "户部尚书", Domain: DomainEconomy, HolderID: "shen", Authority: 68, VacancyRisk: 18, Seat: "国库、税粮、漕运"},
		{ID: "grand-general", Title: "五军都督", Domain: DomainMilitary, HolderID: "huo", Authority: 70, VacancyRisk: 20, Seat: "京营、边镇、军功"},
		{ID: "diplomatic-court", Title: "鸿胪寺卿", Domain: DomainDiplomacy, HolderID: "princess", Authority: 58, VacancyRisk: 22, Seat: "盟约、册封、互市"},
		{ID: "censorate", Title: "都察院左都御史", Domain: DomainIntrigue, Authority: 46, VacancyRisk: 44, Seat: "弹劾、巡按、密档"},
		{ID: "palace-secretary", Title: "内廷总管", Domain: DomainCourt, Authority: 50, VacancyRisk: 35, Seat: "宫闱、诏令、礼仪"},
	}
	switch dynastyID {
	case "dayin":
		offices[2].Authority += 10
		offices[4].VacancyRisk += 8
	case "jingyao":
		offices[1].Authority += 8
		offices[5].VacancyRisk += 10
	case "chengping":
		offices[0].VacancyRisk += 12
		offices[1].VacancyRisk += 10
	case "xuanshuo":
		offices[2].Authority += 8
		offices[3].VacancyRisk += 8
	}
	return offices
}

func (s *GameState) applyCourtOrder(req OrderRequest) (Effects, string, bool, error) {
	switch req.Kind {
	case OrderAppoint:
		effects, summary, err := s.appointOffice(req.Target)
		return effects, summary, true, err
	case OrderDismiss:
		effects, summary, err := s.dismissOffice(req.Target)
		return effects, summary, true, err
	case OrderNameHeir:
		effects, summary, err := s.nameHeir(req.Target)
		return effects, summary, true, err
	case OrderFavor:
		effects, summary, err := s.favorConsort(req.Target)
		return effects, summary, true, err
	case OrderMarriage:
		effects, summary, err := s.arrangeMarriageAlliance(req.Target)
		return effects, summary, true, err
	default:
		return Effects{}, "", false, nil
	}
}

func (s *GameState) appointOffice(target string) (Effects, string, error) {
	officeID, ministerID, ok := strings.Cut(target, ":")
	if !ok || officeID == "" || ministerID == "" {
		return Effects{}, "", fmt.Errorf("appointment target must be officeID:ministerID")
	}
	oi, ok := s.findOfficeIndex(officeID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown office target %q", officeID)
	}
	mi, ok := s.findMinisterIndex(ministerID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown minister target %q", ministerID)
	}

	office := s.Offices[oi]
	minister := s.Court[mi]
	if office.HolderID != "" && office.HolderID != ministerID {
		if previous, ok := s.findMinisterIndex(office.HolderID); ok {
			s.adjustMinister(previous, -5, 8)
		}
	}
	authorityGain := minister.Ability/18 + minister.Integrity/30 - minister.Ambition/45
	office.HolderID = ministerID
	office.Authority = clamp(office.Authority+authorityGain, 0, 100)
	office.VacancyRisk = clamp(office.VacancyRisk-24, 0, 100)
	s.Offices[oi] = office
	s.adjustMinister(mi, 5, 9+office.Authority/25)

	effects := officeEffects(office.Domain, minister)
	summary := fmt.Sprintf("你命%s出任%s，掌%s。官署权威升至%d，空转风险降至%d；%s压力也随差事上升。", minister.Name, office.Title, office.Seat, office.Authority, office.VacancyRisk, minister.Name)
	return effects, summary, nil
}

func (s *GameState) dismissOffice(officeID string) (Effects, string, error) {
	oi, ok := s.findOfficeIndex(officeID)
	if !ok {
		return Effects{}, "", fmt.Errorf("unknown office target %q", officeID)
	}
	office := s.Offices[oi]
	holderName := "空缺"
	if office.HolderID != "" {
		if mi, ok := s.findMinisterIndex(office.HolderID); ok {
			holderName = s.Court[mi].Name
			s.adjustMinister(mi, -8, 6)
		}
	}
	office.HolderID = ""
	office.VacancyRisk = clamp(office.VacancyRisk+26, 0, 100)
	office.Authority = clamp(office.Authority-8, 0, 100)
	s.Offices[oi] = office
	effects := Effects{Influence: 4, Stability: -4, Reform: -1}
	return effects, fmt.Sprintf("你罢免%s的%s，留中待补。短期震慑群臣，%s空转风险升至%d。", office.Title, holderName, office.Title, office.VacancyRisk), nil
}

func officeEffects(domain Domain, minister Minister) Effects {
	bonus := max(1, minister.Ability/30)
	switch domain {
	case DomainDomestic:
		return Effects{Populace: bonus + 1, Stability: 1}
	case DomainEconomy:
		return Effects{Treasury: bonus + 2, Reform: 1}
	case DomainMilitary:
		return Effects{Army: bonus + 2, BorderThreat: -2, Martial: 1}
	case DomainDiplomacy:
		return Effects{Diplomacy: bonus + 2, BorderThreat: -1}
	case DomainReform:
		return Effects{Reform: bonus + 2, Stability: -1}
	case DomainIntrigue:
		return Effects{Influence: bonus + 2, Stability: -2}
	case DomainCourt:
		return Effects{Stability: bonus + 1, Influence: 1}
	default:
		return Effects{Influence: bonus}
	}
}

func (s *GameState) applyOfficePressure(domain Domain) {
	if len(s.Offices) == 0 {
		return
	}
	totalRisk := 0
	for i, office := range s.Offices {
		if office.HolderID == "" {
			office.VacancyRisk = clamp(office.VacancyRisk+5, 0, 100)
			office.Authority = clamp(office.Authority-2, 0, 100)
		} else {
			office.VacancyRisk = clamp(office.VacancyRisk-1, 0, 100)
			if mi, ok := s.findMinisterIndex(office.HolderID); ok && domain == office.Domain {
				s.adjustMinister(mi, 1, 2)
			}
		}
		totalRisk += office.VacancyRisk
		s.Offices[i] = office
	}
	if totalRisk/len(s.Offices) >= 58 {
		s.Stats.Stability = clamp(s.Stats.Stability-2, 0, 100)
		s.Stats.Influence = clamp(s.Stats.Influence-1, 0, 100)
	}
}

func (s *GameState) findOfficeIndex(id string) (int, bool) {
	for i, office := range s.Offices {
		if office.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (s *GameState) findMinisterIndex(id string) (int, bool) {
	for i, minister := range s.Court {
		if minister.ID == id {
			return i, true
		}
	}
	return 0, false
}
