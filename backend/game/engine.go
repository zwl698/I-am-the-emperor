package game

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Phase string

const (
	PhasePrince  Phase = "prince"
	PhaseEmperor Phase = "emperor"
)

type Domain string

const (
	DomainStory     Domain = "story"
	DomainDomestic  Domain = "domestic"
	DomainEconomy   Domain = "economy"

	DomainMilitary  Domain = "military"
	DomainDiplomacy Domain = "diplomacy"
DomainCourt     Domain = "court"
	D
	DomainCourt     Domain = "court"
)

type EndingKind string

const (
	EndingCollapse EndingKind = "collapse"
	EndingDeath    EndingKind = "death"
	EndingGolden   EndingKind = "golden_age"
)

type Stats struct {
	Legitimacy   int `json:"legitimacy"`
	Health       int `json:"health"`
	Learning     int `json:"learning"`
	Martial      int `json:"martial"`
	Charisma     int `json:"charisma"`
	Influence    int `json:"influence"`
	Treasury     int `json:"treasury"`
	Grain        int `json:"grain"`
	Populace     int `json:"populace"`
	Army         int `json:"army"`
	Diplomacy    int `json:"diplomacy"`
	Stability    int `json:"stability"`
	BorderThreat int `json:"borderThreat"`
	Reform       int `json:"reform"`
}

type Effects struct {
	Legitimacy   int `json:"legitimacy,omitempty"`
	Health       int `json:"health,omitempty"`
	Learning     int `json:"learning,omitempty"`
	Martial      int `json:"martial,omitempty"`
	Charisma     int `json:"charisma,omitempty"`
	Influence    int `json:"influence,omitempty"`
	Treasury     int `json:"treasury,omitempty"`
	Grain        int `json:"grain,omitempty"`
	Populace     int `json:"populace,omitempty"`
	Army         int `json:"army,omitempty"`
	Diplomacy    int `json:"diplomacy,omitempty"`
	Stability    int `json:"stability,omitempty"`
	BorderThreat int `json:"borderThreat,omitempty"`
	Reform       int `json:"reform,omitempty"`
}
   int `json:"reform,omitempty"`
}

type Dynasty struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Era        string   `json:"era"`
	Background string   `json:"background"`
	Features   []string `json:"features"`
	Challenge  string   `json:"challenge"`
	Asset      string   `json:"asset"`
	Palette    string   `json:"palette"`
	Initial    Stats    `json:"initial"`
}

type Assets struct {
	Hero       string `json:"hero"`
	Dynasties  string `json:"dynasties"`
	Characters string `json:"characters"`
}

type Faction struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Leader   string `json:"leader"`
	Power    int    `json:"power"`
	Loyalty  int    `json:"loyalty"`
	Agenda   string `json:"agenda"`
	Portrait string `json:"portrait"`
}

type Minister struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Trait    string `json:"trait"`
	Loyalty  int    `json:"loyalty"`
	Portrait string `json:"portrait"`
}

type Province struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Focus    string `json:"focus"`
	Wealth   int    `json:"wealth"`
	Order    int    `json:"order"`
	Defense  int    `json:"defense"`
	Disaster int    `json:"disaster"`
}

type Crisis struct {
	Title    string `json:"title"`
	Severity int    `json:"severity"`
	Clock    int    `json:"clock"`
	Summary

type Choice struct {
	ID      string  `json:"id"`
	Text    string  `json:"text"`
	Detail  string  `json:"detail"`
	Domain  Domain  `json:"domain"`
	Effects Effects `json:"effects"`
	Outcome string  `json:"outcome"`
}

type Scene struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Body    string   `json:"body"`
   string   `json:"body"`
	Year
	Year    string   `json:"year"`
	Mood    string   `json:"mood"`
	Choices []Choice `json:"choices"`
}

type Ending struct {
	Kind    EndingKind `json:"kind"`
	Title   string     `json:"title"`
	Summary string     `json:"summary"`
}

type HistoryEntry struct {
	Turn    int     `json:"turn"`
	Age     int     `json:"age"`
	Phase   Phase   `json:"phase"`
	Choice  string  `json:"choice"`
	Summary string  `json:"summary"`
	Effects Effects `json:"effects"`
}

type Resolution struct {
	Summary string  `json:"summary"`
	Effects Effects `json:"effects"`
	Scene   *Scene  `json:"scene,omitempty"`
	Ending  *Ending `json:"ending,omitempty"`
	ID        string         `json:"id"`
	Seed      int64          `json:"seed"`
	Turn      int            `json:"turn"`
	Age       int            `json:"age"`
	Phase     Phase          `json:"phase"`
	Dynasty   Dynasty        `json:"dynasty"`
	Assets    Assets         `json:"assets"`
	ReignYear int            `json:"reignYear"`
	Season    string         `json:"season"`
	Stats     Stats          `json:"stats"`
	Factions  []Faction      `json:"factions"`
	Court     []Minister     `json:"court"`
	Provinces []Province     `json:"provinces"`
	Crisis    Crisis         `json:"crisis"`
	Scene     *Scene         `json:"scene,omitempty"`
	Ending    *Ending        `json:"ending,omitempty"`
	History   []HistoryEntry `json:"history"`
	Scene   *Scene         `json:"scene,omitempty"`
	Ending  *Ending        `json:"ending,omitempty"`
	History []HistoryEntry `json:"history"`

e([]Dynasty, len(all))
	copy(out, all)
	return out
}

func NewGame(seed int64) *GameState {
	state, _ := N
	rng *rand.Rand
("dayin", seed)
	return state
}

func NewGameWithDynasty(dynastyID string, seed int64)
}

func NewGame(seed int64) *GameState {
ed = time.Now().UnixNano()
	}
	dynasty, ok := findDynasty(dynastyID)
	if !ok {
		return nil, fmt.Errorf("unknown dynasty %q", dynastyI
	if seed == 0 {
		ID:      fmt.Sprintf("dragon-%x", seed^time.Now().UnixNano()),
		Seed:    seed,
		Turn:    0,
		Age:     1,
		Phase:   PhasePrince,
		Dynasty: dynasty,
		Assets: Assets{
			Hero:       "/assets/palace-hero.png",
			Dynasties:  "/assets/dynasty-scroll.png",
			Characters: "/assets/characters.png",
			Martial:    18,
		Season:    "春",
		ReignYear: 0,
		Stats:     dynasty.Initial,
		Factions:  startingFactions(dynasty.ID),
		Court:     startingCourt(),
		Provinces: startingProvinces(dynasty.ID),
		Crisis:    startingCrisis(dynasty.ID),
		rng:       rand.New(rand.NewSource(seed)),
			Influence:  18,
	state.Scene = princeScene(0, state)
	return state, nil
	}
	state.Scene = princeScene(0, state.Stats)
	return state
}

func (s *GameState) ApplyChoice(choiceID string) (*Resolution, error) {
	if s == nil {
		return nil, errors.New("game state is nil")
	}
	s.ensureRNG()
	if s.Ending != nil {
		return nil, errors.New("game has already ended")
	}
	if s.Scene == nil {
		return nil, errors.New("game has no active scene")
	}

	choice, ok := s.findChoice(choiceID)
	if !ok {
		return nil, fmt.Errorf("unknown choice %q", choiceID)
.a
	}

	s.applyEffects(choice.Effects)
	s.Turn++
	s.advanceAfter(choice)
	s.applyWorldPressure(choice.Domain)
	s.Ending = s.checkEnding()
	if s.Ending == nil {
		s.Scene = s.nextScene()
	} else {
		s.Scene = nil
	}

	s.History = append(s.History, HistoryEntry{
		Turn:    s.Turn,
		Age:     s.Age,
		Phase:   s.Phase,
		Choice:  choice.Text,
		Summary: choice.Outcome,
		Effects: choice.Effects,
	})

	return &Resolution{
		Summary: choice.Outcome,
		Effects: choice.Effects,
		Scene:   s.Scene,
		Ending:  s.Ending,
	}, nil
}

func (s *GameState) ForceCoronationForTest() {
	s.ReignYear = 1
	s.Season = "春"
	s.Stats.Treasury = max(s.Stats.Treasury, 72)
	s.Stats.Grain = max(s.Stats.Grain, 66)
	s.Stats.Populace = max(s.Stats.Populace, 64)
	s.Stats.Army = max(s.Stats.Army, 58)
	s.Stats.Diplomacy = max(s.Stats.Diplomacy, 52)
	s.Stats.Stability = max(s.Stats.Stability, 60)
	if s.Stats.BorderThreat == 0 {
		s.Stats.BorderThreat = 38
	}
	s.Stats.Diplomacy = 52
	s.Stats.Stability = 60
	s.Stats.BorderThreat = 38
	s.Scene = emperorScene(s)
}

func (s *GameState) findChoice(choiceID string) (Choice, bool) {
	for _, choice := range s.Scene.Choices {
		if choice.ID == choiceID {
			return choice, true
		}
	}
	return Choice{}, false
}

func (s *GameState) ensureRNG() {
	if s.rng == nil {
		s.rng = rand.New(rand.NewSource(s.Seed + int64(s.Turn*7919)))
	}
}

func (s *GameState) applyEffects(e Effects) {
	s.Stats.Legitimacy = clamp(s.Stats.Legitimacy+e.Legitimacy, 0, 100)
	s.Stats.Health = clamp(s.Stats.Health+e.Health, 0, 100)
	s.Stats.Learning = clamp(s.Stats.Learning+e.Learning, 0, 100)
	s.Stats.Treasury = clamp(s.Stats.Treasury+e.Treasury, 0, 160)
	s.Stats.Grain = clamp(s.Stats.Grain+e.Grain, 0, 160)
	s.Stats.Influence = clamp(s.Stats.Influence+e.Influence, 0, 100)
	s.Stats.Army = clamp(s.Stats.Army+e.Army, 0, 140)
	s.Stats.Grain = clamp(s.Stats.Grain+e.Grain, 0, 140)
	s.Stats.Populace = clamp(s.Stats.Populace+e.Populace, 0, 100)
	s.Stats.Army = clamp(s.Stats.Army+e.Army, 0, 120)
	s.Stats.Diplomacy = clamp(s.Stats.Diplomacy+e.Diplomacy, 0, 100)
	s.Stats.Stability = clamp(s.Stats.Stability+e.Stability, 0, 100)
	s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+e.BorderThreat, 0, 100)
func (s *GameState) applyChoiceToWorld(choice Choice) {
	if s.Phase != PhaseEmperor {
		return
	}
	targetProvince := s.Turn % max(1, len(s.Provinces))
	targetFaction := s.Turn % max(1, len(s.Factions))

	switch choice.Domain {
	case DomainDomestic:
		s.adjustProvince(targetProvince, -8, 5, 0, -8)
		s.adjustFaction(targetFaction, -2, 2)
	case DomainEconomy:
		s.adjustProvince(targetProvince, 8, -2, 0, 2)
		s.adjustFaction(targetFaction, 5, -4)
	case DomainMilitary:
		s.adjustProvince(targetProvince, -2, 0, 9, 0)
		s.adjustFactionByID("border", 6, 5)
	case DomainDiplomacy:
		s.adjustFactionByID("clan", 4, 4)
		s.adjustProvince(targetProvince, 2, 2, 0, 0)
	case DomainReform:
		s.adjustProvince(0, 6, -6, 0, -3)
		s.adjustFaction(0, -7, -8)
	case DomainIntrigue:
		s.adjustFaction(targetFaction, -8, -5)
		s.adjustProvince(targetProvince, -1, -2, 0, 0)
	case DomainCourt:
		s.adjustFaction(targetFaction, 3, 5)
	}

	s.Crisis.Severity = clamp(s.Crisis.Severity+crisisDelta(choice.Domain), 0, 100)
	if s.Crisis.Severity < 35 {
		s.Crisis.Clock = max(0, s.Crisis.Clock-1)
	} else {
		s.Crisis.Clock = clamp(s.Crisis.Clock+1, 0, 8)
	}
}

func (s *GameState) adjustProvince(i, wealth, order, defense, disaster int) {
	if len(s.Provinces) == 0 {
		return
	}
	i = clamp(i, 0, len(s.Provinces)-1)
	p := s.Provinces[i]
	p.Wealth = clamp(p.Wealth+wealth, 0, 100)
	p.Order = clamp(p.Order+order, 0, 100)
	p.Defense = clamp(p.Defense+defense, 0, 100)
	p.Disaster = clamp(p.Disaster+disaster, 0, 100)
	s.Provinces[i] = p
}

func (s *GameState) adjustFaction(i, power, loyalty int) {
	if len(s.Factions) == 0 {
		return
	}
	i = clamp(i, 0, len(s.Factions)-1)
	f := s.Factions[i]
	f.Power = clamp(f.Power+power, 0, 100)
	f.Loyalty = clamp(f.Loyalty+loyalty, 0, 100)
	s.Factions[i] = f
}

func (s *GameState) adjustFactionByID(id string, power, loyalty int) {
	for i := range s.Factions {
		if s.Factions[i].ID == id {
			s.adjustFaction(i, power, loyalty)
			return
		}
	}
}

func crisisDelta(domain Domain) int {
	switch domain {
	case DomainDomestic:
		return -4
	case DomainEconomy:
		return 1
	case DomainMilitary:
		return -6
	case DomainDiplomacy:
		return -5
	case DomainReform:
		return 5
	case DomainIntrigue:
		return 3
	default:
		return 1
	}
}

func (s *GameState) advanceAfter(choice Choice) {
	if s.Phase == PhasePrince {
		ages := []int{6, 10, 14, 16, 18}
		s.Age = ages[min(s.Turn, len(ages)-1)]
		if s.Turn >= 4 {
		if s.Turn >= 4 {
			s.coronate()
		}
		return
	s.advanceCalendar()
}

func (s *GameState) advanceCalendar() {
	seasons := []string{"春", "夏", "秋", "冬"}
	idx := 0
	for i, season := range seasons {
		if s.Season == season {
			idx = i
			break
		}
	}
	idx++
	if idx >= len(seasons) {
		idx = 0
		s.ReignYear++
	s.advanceCalendar()
}
if s.ReignYear < 1 {
		s.ReignYear = 1
	}
}

func (s *GameState) coronate() {
	s.Phase = PhaseEmpero

func (s *GameState) advanceCalendar() {
	s.ReignYear = 1
	s.Season = "春"
	s.Stats.Treasury = clamp(s.Dynasty.Initial.Treasury+s.Stats.Legitimacy/4+s.Stats.Learning/6, 25, 140)
	s.Stats.Grain = clamp(s.Dynasty.Initial.Grain+s.Stats.Charisma/6, 20, 140)
	s.Stats.Populace = clamp(s.Dynasty.Initial.Populace+s.Stats.Charisma/7+s.Stats.Legitimacy/9, 20, 100)
	s.Stats.Army = clamp(s.Dynasty.Initial.Army+s.Stats.Martial/3, 25, 140)
	s.Stats.Diplomacy = clamp(s.Dynasty.Initial.Diplomacy+s.Stats.Charisma/5+s.Stats.Learning/8, 20, 100)
	s.Stats.Stability = clamp(s.Dynasty.Initial.Stability+s.Stats.Influence/5+s.Stats.Legitimacy/8, 15, 100)
	s.Stats.BorderThreat = clamp(s.Dynasty.Initial.BorderThreat-s.Stats.Martial/8, 5, 100)
		if s.Turn >= 5 {
			s.coronate()
		}
		return
	}
	if s.Turn%2 == 0 {
		s.Age++
	s.Stats.Treasury = clamp(s.Stats.Treasury-2+averageProvinceWealth(s.Provinces)/35, 0, 160)
	s.Stats.Grain = clamp(s.Stats.Grain-1-disasterPressure(s.Provinces)/40, 0, 160)
	if domain != DomainMilitary && domain != DomainDiplomacy {
	s.Phase = PhaseEmperor
	s.Age = max(s.Age, 18)
erThreat+2+s.rng.Intn(4), 0, 100)
	}
	if averageFactionLoyalty(s.Factions) < 38 {
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
	if averageProvinceOrder(s.Provinces) < 42 {
		s.Stats.Populace = clamp(s.Stats.Populace-3, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-3, 0, 100)
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
	if s.Stats.Grain < 25 {
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-2, 0, 100)
	}
	if s.Stats.Treasury < 20 {
		s.Stats.Stab
	s.Stats.Treasury = clamp(50+s.Stats.Legitimacy/3+s.Stats.Learning/5, 35, 120)
	s.Stats.Grain = clamp(48+s.Stats.Charisma/4, 35, 120)
	s.Stats.Populace = clamp(48+s.Stats.Charisma/5+s.Stats.Legitimacy/6, 30, 100)
	s.Stats.Army = clamp(42+s.Stats.Martial/2, 30, 120)
	s.Stats.Diplomacy = clamp(38+s.Stats.Charisma/3+s.Stats.Learning/6, 25, 100)
	s.Stats.Stability = clamp(42+s.Stats.Influence/3+s.Stats.Legitimacy/6, 25, 100)
	s.Stats.BorderThreat = clamp(48-s.Stats.Martial/5, 15, 75)
		return &Ending{Kind: EndingDeath, Title: "龙驭宾天", Summary: "操劳与暗疾耗尽了你的生命，史官在夜色里合上实录。"}
	}
	if s.Phase == PhaseEmperor && (s.Stats.Stability <= 0 || s.Stats.Populace <= 0 || s.Crisis.Clock >= 8 || (s.Stats.BorderThreat >= 95 && s.Stats.Army < 90)) {
		return &Ending{Kind: EndingCollapse, Title: "山河失守", Summary: "边患、民怨、党争与危机链同时爆发，王朝在你的御座前倾塌。"}
	}
	if s.Phase == PhaseEmperor && s.Turn >= 45 && s.Stats.Stability >= 80 && s.Stats.Populace >= 80 && s.Stats.BorderThreat <= 25 && s.Stats.Reform >= 55 {
		return &Ending{Kind: EndingGolden, Title: "万邦来朝", Summary: "新法成制，仓廪充盈，边境安定，诸国遣使入贡，你开创了被后世反复吟诵的盛世。"}
	}
	if s.Stats.BorderThreat > s.Stats.Army+20 {
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
}
		return princeScene(s.Turn, s)
func (s *GameState) checkEnding() *Ending {
	if s.Stats.Health <= 0 {
		return &Ending{
			Kind:    EndingDeath,
func princeScene(turn int, state *GameState) *Scene {
			Summary: "操劳与暗疾耗尽了你的生命，史官在夜色里合上实录。",
		}
	}
	if s.Phase == PhaseEmperor && (s.Stats.Stability <= 0 || s.Stats.Populace <= 0 || (s.Stats.BorderThreat >= 95 && s.Stats.Army < 80)) {
			Year:  state.Dynasty.Era,
			Kind:    EndingCollapse,
			Art:   state.Assets.Characters,
			Body:  fmt.Sprintf("你出生在%s的风雪清晨。%s 宫中每一句吉言都可能变成刀锋。", state.Dynasty.Name, state.Dynasty.Background),
			Summary: "边患、民怨与朝局崩裂同时袭来，王朝在你的御座前倾塌。",
		}
	}
	if s.Phase == PhaseEmperor && s.Turn >= 45 && s.Stats.Stability >= 80 && s.Stats.Populace >= 80 && s.Stats.BorderThreat <= 25 {
		return &Ending{
			Kind:    EndingGolden,
			Title:   "万邦来朝",
			ID: "study-yard", Title: "东宫书院", Year: "六岁", Mood: "养成", Art: state.Assets.Characters,
			Body: "皇子们第一次同席读书。太傅问你：治国先治什么？兄弟们都在等你出错。",

func (s *GameState) nextScene() *Scene {
	if s.Phase == PhasePrince {
		return princeScene(s.Turn, s.Stats)
	}
	return emperorScene(s)
}
			ID: "winter-hunt", Title: "皇家冬狩", Year: "十岁", Mood: "锋芒", Art: state.Assets.Dynasties,
			Body: "猎场上，三皇子故意惊马。你摔在雪里，侍卫们一瞬间不敢动。",
			Title: "紫宸宫中的啼哭",
			Year:  "皇子元年",
			Mood:  "启蒙",
			Body:  "你出生在风雪初停的清晨。太史说紫气绕阙，贵妃却知道，宫里每一句吉言都可能变成刀锋。",
			Choices: []Choice{
				{ID: "grab-scroll", Text: "抓起案上的竹简", Detail: "让太傅记住你早慧的一面。", Domain: DomainStory, Effects: Effects{Learning: 8, Legitimacy: 2, Health: -1}, Outcome: "你咿呀抓住竹简，满殿笑声里，多了一个“好学皇子”的传闻。"},
				{ID: "smile-consort", Text: "向皇后展露笑容", Detail: "讨得中宫欢心，但母妃心中不安。", Domain: DomainStory, Effects: Effects{Charisma: 7, Influence: 4}, Outcome: "皇后轻抚你的额头，宫人们很快学会了对你多行半礼。"},
			ID: "flood-memorial", Title: "南河急报", Year: "十四岁", Mood: "试政", Art: state.Assets.Dynasties,
			Body: "南河决堤，朝堂争论赈灾银从何处来。父皇将奏章推到你面前，要你试拟朱批。",
			Title: "东宫书院",
			Year:  "六岁",
			Mood:  "养成",
			Body:  "皇子们第一次同席读书。太傅问你：治国先治什么？兄弟们都在等你出错。",
			Choices: []Choice{
				{ID: "answer-people", Text: "答：先安百姓", Detail: "赢得清流赞许。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 4, Legitimacy: 3}, Outcome: "太傅捻须点头，清流大臣开始把你的名字写进密札。"},
				{ID: "answer-army", Text: "答：先强兵甲", Detail: "让武臣另眼相看。", Domain: DomainStory, Effects: Effects{Martial: 8, Influence: 2}, Outcome: "武学师傅当场请你试弓，你拉不开弓，却拉来了将门的好感。"},
			ID: "succession-night", Title: "烛影摇红", Year: "十六岁", Mood: "夺嫡", Art: state.Assets.Hero,
			Body: fmt.Sprintf("父皇病重，诸王入宫。你手中有学识 %d、武略 %d、声望 %d。最后一夜，谁先动，谁就可能坐上明日的朝堂。", state.Stats.Learning, state.Stats.Martial, state.Stats.Legitimacy),
			Title: "皇家冬狩",
			Year:  "十岁",
			Mood:  "锋芒",
			Body:  "猎场上，三皇子故意惊马。你摔在雪里，侍卫们一瞬间不敢动。",
			Choices: []Choice{
				{ID: "mount-again", Text: "忍痛重新上马", Detail: "以勇气换取威望。", Domain: DomainStory, Effects: Effects{Martial: 9, Legitimacy: 4, Health: -4}, Outcome: "你带伤上马，雪原上响起军士的喝彩。三皇子的笑容僵住了。"},
				{ID: "protect-servant", Text: "先扶起被撞倒的小内侍", Detail: "仁名会悄悄传开。", Domain: DomainStory, Effects: Effects{Charisma: 8, Populace: 2, Legitimacy: 1}, Outcome: "一个小小内侍救不了天下，却能让天下相信你会低头看人。"},
				{ID: "accuse-brother", Text: "当众指认三皇子", Detail: "直接开战，风险很高。", Domain: DomainStory, Effects: Effects{Influence: 8, Martial: 3, Stability: -3}, Outcome: "猎场瞬间安静。你赢得了一批拥护者，也让夺嫡提早见血。"},
			},
		},
		{
	year := fmt.Sprintf("登基%d年 · %s", max(1, s.ReignYear), s.Season)
	pressure := crisisLine(s)
		},
		{
		Title: "太和朝议",
			Title: "烛影摇红",
			Year:  "十六岁",
		Art:   s.Assets.Hero,
		Body:  pressure + " 六部、边军、宗室、清流与商帮都在等你落子。选择不是单纯加减数值，派系与省份会记住你的每一道旨意。",
			Body:  fmt.Sprintf("父皇病重，诸王入宫。你手中有学识 %d、武略 %d、声望 %d。最后一夜，谁先动，谁就可能坐上明日的朝堂。", stats.Learning, stats.Martial, stats.Legitimacy),
			{ID: fmt.Sprintf("relief-%d", s.Turn), Text: "户部开仓，巡抚赈济", Detail: "民政线：压灾情、保民心，但消耗粮银。", Domain: DomainDomestic, Effects: Effects{Treasury: -10, Grain: -8, Populace: 12, Stability: 5, Legitimacy: 2}, Outcome: "粥棚沿官道铺开，流民队伍短了，户部账册却重得像铁。"},
			{ID: fmt.Sprintf("tax-%d", s.Turn), Text: "重估田亩，整顿盐铁", Detail: "财政线：增加国库，激怒豪强商帮。", Domain: DomainEconomy, Effects: Effects{Treasury: 18, Populace: -4, Stability: -4, Reform: 3}, Outcome: "银车入库，盐引重发。商帮笑着谢恩，袖中却攥紧了旧账。"},
			{ID: fmt.Sprintf("train-%d", s.Turn), Text: "拨银练兵，轮戍边镇", Detail: "军务线：提升军力、压边患，耗费国库。", Domain: DomainMilitary, Effects: Effects{Treasury: -12, Army: 15, BorderThreat: -13, Martial: 1}, Outcome: "军营号角重新响亮，边将呈上的地图终于少了几个红圈。"},
			{ID: fmt.Sprintf("envoy-%d", s.Turn), Text: "遣使联姻，分化诸邦", Detail: "外交线：缓冲战争，稳住宗室和外邦。", Domain: DomainDiplomacy, Effects: Effects{Diplomacy: 13, BorderThreat: -8, Treasury: -4, Stability: 1}, Outcome: "使团携金册出关，远方可汗收下礼物，也收起了一半刀锋。"},
			{ID: fmt.Sprintf("reform-%d", s.Turn), Text: "设考成法，裁撤冗官", Detail: "新法线：长期强国，短期引爆党争。", Domain: DomainReform, Effects: Effects{Reform: 12, Treasury: 5, Stability: -7, Populace: 3, Legitimacy: 2}, Outcome: "新法像春雷落进官场。有人称颂清明，也有人开始深夜串门。"},
			{ID: fmt.Sprintf("spy-%d", s.Turn), Text: "开缇骑密档，夜审朋党", Detail: "暗线：削派系权势，但损声望与稳定。", Domain: DomainIntrigue, Effects: Effects{Influence: 7, Stability: -6, Legitimacy: -4, Health: -2}, Outcome: "宫门后的灯亮到三更，第二天朝会上少了几张熟悉的脸。"},
			{ID: fmt.Sprintf("banquet-%d", s.Turn), Text: "大宴群臣，粉饰太平", Detail: "宫廷线：短期安抚派系，长期损害国本。", Domain: DomainCourt, Effects: Effects{Treasury: -9, Grain: -5, Populace: -7, Stability: -5, BorderThreat: 5, Health: -2}, Outcome: "钟鼓响彻宫城，杯盏遮住了奏章。散席后，问题仍在殿外等你。"},
		},
	}
	return cloneScene(scenes[min(turn, len(scenes)-1)])
}

func emperorScene(s *GameState) *Scene {
	year := fmt.Sprintf("登基第%d年", max(1, s.Turn-4))
	pressure := "边疆尚稳，朝堂却没有真正安静的一天。"
	if s.Stats.BorderThreat > 70 {
		pressure = "北境烽烟频传，兵部的奏报一封比一封急。"
	} else if s.Stats.Grain < 30 {
		pressure = "粮价上涨，市井已有怨声，户部请求立刻决断。"
	} else if s.Stats.Stability < 35 {
		pressure = "言官互攻，外戚与清流争得满殿冷汗。"
	}

	return &Scene{
		ID:    fmt.Sprintf("court-%d", s.Turn),
		Title: "乾清晨议",
		Year:  year,
		Mood:  emperorMood(s.Stats),
		Body:  pressure + " 你坐在御座上，看见内政、军务、邦交与宫廷都在同一张棋盘上移动。",
		Choices: []Choice{
			{ID: fmt.Sprintf("relief-%d", s.Turn), Text: "减赋开仓，安抚州县", Detail: "提升民心与粮政，但国库承压。", Domain: DomainDomestic, Effects: Effects{Treasury: -10, Grain: -7, Populace: 12, Stability: 5, Legitimacy: 2}, Outcome: "诏书驰往各省，粥棚前的哭声少了，户部尚书的白发多了。"},
			{ID: fmt.Sprintf("reform-%d", s.Turn), Text: "整顿吏治，推行新法", Detail: "增加改革进度，短期刺激既得利益者。", Domain: DomainDomestic, Effects: Effects{Reform: 10, Treasury: 5, Stability: -5, Populace: 3}, Outcome: "新法像春雷落进官场。有人称颂清明，也有人开始深夜串门。"},
			{ID: fmt.Sprintf("train-%d", s.Turn), Text: "拨银练兵，巡视边防", Detail: "强化军力，压低边患，耗费明显。", Domain: DomainMilitary, Effects: Effects{Treasury: -12, Army: 14, BorderThreat: -12, Martial: 1}, Outcome: "军营号角重新响亮，边将呈上的地图终于少了几个红圈。"},
			{ID: fmt.Sprintf("envoy-%d", s.Turn), Text: "遣使联姻，分化诸邦", Detail: "用外交缓冲战争，也会被鹰派批评。", Domain: DomainDiplomacy, Effects: Effects{Diplomacy: 12, BorderThreat: -7, Treasury: -4, Stability: 1}, Outcome: "使团携金册出关，远方可汗收下礼物，也收起了一半刀锋。"},
			{ID: fmt.Sprintf("banquet-%d", s.Turn), Text: "大宴群臣，粉饰太平", Detail: "短期笼络朝臣，长期损害国本。", Domain: DomainCourt, Effects: Effects{Treasury: -9, Grain: -5, Populace: -7, Stability: -8, BorderThreat: 7, Health: -2}, Outcome: "钟鼓响彻宫城，杯盏遮住了奏章。散席后，问题仍在殿外等你。"},
		},
	}
}

func emperorMood(stats Stats) string {
	score := stats.Stability + stats.Populace + stats.Army + stats.Diplomacy - stats.BorderThreat
	switch {
	case score >= 260:
		return "盛世"
	case score >= 190:
		return "可治"
	case score >= 130:
		return "暗涌"
	default:
		return "危局"
	}
}

func cloneScene(scene Scene) *Scene {
	choices := make([]Choice, len(scene.Choices))
	copy(choices, scene.Choices)
	scene.Choices = choices
	return &scene
}

func (e Effects) Describe() string {
	parts := make([]string, 0, 8)
	add := func(name string, value int) {
		if value == 0 {
			return
		}
		sign := "+"
		if value < 0 {
			sign = ""
		}
		parts = append(parts, fmt.Sprintf("%s%s%d", name, sign, value))
	}
	add("名望", e.Legitimacy)
	add("健康", e.Health)
	add("学识", e.Learning)
	add("武略", e.Martial)
	add("魅力", e.Charisma)
	add("势力", e.Influence)
	add("国库", e.Treasury)
	add("粮草", e.Grain)
	add("民心", e.Populace)
	add("军力", e.Army)
	add("邦交", e.Diplomacy)
	add("朝稳", e.Stability)
	add("边患", e.BorderThreat)
	add("新政", e.Reform)
	return strings.Join(parts, "、")
}

func clamp(v, low, high int) int {
	return int(math.Max(float64(low), math.Min(float64(high), float64(v))))
}