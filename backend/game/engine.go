package game

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
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
	DomainReform    Domain = "reform"
	DomainIntrigue  Domain = "intrigue"
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
	Summary  string `json:"summary"`
}

type Objective struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Progress    int    `json:"progress"`
	Target      int    `json:"target"`
	Completed   bool   `json:"completed"`
	Reward      string `json:"reward"`
}

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
	Year    string   `json:"year"`
	Mood    string   `json:"mood"`
	Art     string   `json:"art"`
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
}

type GameState struct {
	ID         string         `json:"id"`
	Seed       int64          `json:"seed"`
	Turn       int            `json:"turn"`
	Age        int            `json:"age"`
	Phase      Phase          `json:"phase"`
	Dynasty    Dynasty        `json:"dynasty"`
	Assets     Assets         `json:"assets"`
	ReignYear  int            `json:"reignYear"`
	Season     string         `json:"season"`
	Stats      Stats          `json:"stats"`
	Factions   []Faction      `json:"factions"`
	Court      []Minister     `json:"court"`
	Provinces  []Province     `json:"provinces"`
	Crisis     Crisis         `json:"crisis"`
	Objectives []Objective    `json:"objectives"`
	Scene      *Scene         `json:"scene,omitempty"`
	Ending     *Ending        `json:"ending,omitempty"`
	History    []HistoryEntry `json:"history"`

	rng *rand.Rand
}

func AvailableDynasties() []Dynasty {
	return slices.Clone(dynasties())
}

func NewGame(seed int64) *GameState {
	state, _ := NewGameWithDynasty("dayin", seed)
	return state
}

func NewGameWithDynasty(dynastyID string, seed int64) (*GameState, error) {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	dynasty, ok := findDynasty(dynastyID)
	if !ok {
		return nil, fmt.Errorf("unknown dynasty %q", dynastyID)
	}

	state := &GameState{
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
		},
		Season:     "春",
		Stats:      dynasty.Initial,
		Factions:   startingFactions(dynasty.ID),
		Court:      startingCourt(),
		Provinces:  startingProvinces(dynasty.ID),
		Crisis:     startingCrisis(dynasty.ID),
		Objectives: startingObjectives(dynasty.ID),
		rng:        rand.New(rand.NewSource(seed)),
	}
	state.updateObjectives()
	state.Scene = princeScene(0, state)
	return state, nil
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
	}

	s.applyEffects(choice.Effects)
	s.applyChoiceToWorld(choice)
	s.Turn++
	s.advanceAfter()
	s.applyWorldPressure(choice.Domain)
	s.updateObjectives()
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
	s.Phase = PhaseEmperor
	s.Age = 18
	s.Turn = 5
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
	s.updateObjectives()
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
	s.Stats.Martial = clamp(s.Stats.Martial+e.Martial, 0, 100)
	s.Stats.Charisma = clamp(s.Stats.Charisma+e.Charisma, 0, 100)
	s.Stats.Influence = clamp(s.Stats.Influence+e.Influence, 0, 100)
	s.Stats.Treasury = clamp(s.Stats.Treasury+e.Treasury, 0, 160)
	s.Stats.Grain = clamp(s.Stats.Grain+e.Grain, 0, 160)
	s.Stats.Populace = clamp(s.Stats.Populace+e.Populace, 0, 100)
	s.Stats.Army = clamp(s.Stats.Army+e.Army, 0, 140)
	s.Stats.Diplomacy = clamp(s.Stats.Diplomacy+e.Diplomacy, 0, 100)
	s.Stats.Stability = clamp(s.Stats.Stability+e.Stability, 0, 100)
	s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+e.BorderThreat, 0, 100)
	s.Stats.Reform = clamp(s.Stats.Reform+e.Reform, 0, 100)
}

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

func (s *GameState) advanceAfter() {
	if s.Phase == PhasePrince {
		ages := []int{6, 10, 14, 16, 18}
		s.Age = ages[min(s.Turn, len(ages)-1)]
		if s.Turn >= 5 {
			s.coronate()
		}
		return
	}
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
		s.Age++
	}
	s.Season = seasons[idx]
	if s.ReignYear < 1 {
		s.ReignYear = 1
	}
}

func (s *GameState) coronate() {
	s.Phase = PhaseEmperor
	s.Age = max(s.Age, 18)
	s.ReignYear = 1
	s.Season = "春"
	s.Stats.Treasury = clamp(s.Dynasty.Initial.Treasury+s.Stats.Legitimacy/4+s.Stats.Learning/6, 25, 140)
	s.Stats.Grain = clamp(s.Dynasty.Initial.Grain+s.Stats.Charisma/6, 20, 140)
	s.Stats.Populace = clamp(s.Dynasty.Initial.Populace+s.Stats.Charisma/7+s.Stats.Legitimacy/9, 20, 100)
	s.Stats.Army = clamp(s.Dynasty.Initial.Army+s.Stats.Martial/3, 25, 140)
	s.Stats.Diplomacy = clamp(s.Dynasty.Initial.Diplomacy+s.Stats.Charisma/5+s.Stats.Learning/8, 20, 100)
	s.Stats.Stability = clamp(s.Dynasty.Initial.Stability+s.Stats.Influence/5+s.Stats.Legitimacy/8, 15, 100)
	s.Stats.BorderThreat = clamp(s.Dynasty.Initial.BorderThreat-s.Stats.Martial/8, 5, 100)
}

func (s *GameState) applyWorldPressure(domain Domain) {
	if s.Phase != PhaseEmperor {
		return
	}
	s.Stats.Treasury = clamp(s.Stats.Treasury-2+averageProvinceWealth(s.Provinces)/35, 0, 160)
	s.Stats.Grain = clamp(s.Stats.Grain-1-disasterPressure(s.Provinces)/40, 0, 160)
	if domain != DomainMilitary && domain != DomainDiplomacy {
		s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+2+s.rng.Intn(4), 0, 100)
	}
	if averageFactionLoyalty(s.Factions) < 38 {
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
	if averageProvinceOrder(s.Provinces) < 42 {
		s.Stats.Populace = clamp(s.Stats.Populace-3, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-3, 0, 100)
	}
	if s.Stats.Grain < 25 {
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-2, 0, 100)
	}
	if s.Stats.Treasury < 20 {
		s.Stats.Stability = clamp(s.Stats.Stability-3, 0, 100)
	}
	if s.Stats.BorderThreat > s.Stats.Army+20 {
		s.Stats.Populace = clamp(s.Stats.Populace-5, 0, 100)
		s.Stats.Stability = clamp(s.Stats.Stability-4, 0, 100)
	}
}

func (s *GameState) checkEnding() *Ending {
	if s.Stats.Health <= 0 {
		return &Ending{Kind: EndingDeath, Title: "龙驭宾天", Summary: "操劳与暗疾耗尽了你的生命，史官在夜色里合上实录。"}
	}
	if s.Phase == PhaseEmperor && (s.Stats.Stability <= 0 || s.Stats.Populace <= 0 || s.Crisis.Clock >= 8 || (s.Stats.BorderThreat >= 95 && s.Stats.Army < 90)) {
		return &Ending{Kind: EndingCollapse, Title: "山河失守", Summary: "边患、民怨、党争与危机链同时爆发，王朝在你的御座前倾塌。"}
	}
	if s.Phase == PhaseEmperor && s.Turn >= 45 && s.Stats.Stability >= 80 && s.Stats.Populace >= 80 && s.Stats.BorderThreat <= 25 && s.Stats.Reform >= 55 {
		return &Ending{Kind: EndingGolden, Title: "万邦来朝", Summary: "新法成制，仓廪充盈，边境安定，诸国遣使入贡，你开创了被后世反复吟诵的盛世。"}
	}
	return nil
}

func (s *GameState) nextScene() *Scene {
	if s.Phase == PhasePrince {
		return princeScene(s.Turn, s)
	}
	return emperorScene(s)
}

func princeScene(turn int, state *GameState) *Scene {
	scenes := []Scene{
		{
			ID:    "birth-omen",
			Title: "紫宸宫中的啼哭",
			Year:  state.Dynasty.Era,
			Mood:  "启蒙",
			Art:   state.Assets.Characters,
			Body:  fmt.Sprintf("你出生在%s的风雪清晨。%s 宫中每一句吉言都可能变成刀锋。", state.Dynasty.Name, state.Dynasty.Background),
			Choices: []Choice{
				{ID: "grab-scroll", Text: "抓起案上的竹简", Detail: "让太傅记住你早慧的一面。", Domain: DomainStory, Effects: Effects{Learning: 8, Legitimacy: 2, Health: -1}, Outcome: "你咿呀抓住竹简，满殿笑声里，多了一个“好学皇子”的传闻。"},
				{ID: "smile-consort", Text: "向皇后展露笑容", Detail: "讨得中宫欢心，但母妃心中不安。", Domain: DomainStory, Effects: Effects{Charisma: 7, Influence: 4}, Outcome: "皇后轻抚你的额头，宫人们很快学会了对你多行半礼。"},
				{ID: "cry-loudly", Text: "放声大哭震住众人", Detail: "生命力旺盛，也显得倔强难驯。", Domain: DomainStory, Effects: Effects{Health: 8, Martial: 3, Charisma: -1}, Outcome: "哭声穿过朱门，老内侍低声说：这孩子有帝王的肺腑。"},
			},
		},
		{
			ID: "study-yard", Title: "东宫书院", Year: "六岁", Mood: "养成", Art: state.Assets.Characters,
			Body: "皇子们第一次同席读书。太傅问你：治国先治什么？兄弟们都在等你出错。",
			Choices: []Choice{
				{ID: "answer-people", Text: "答：先安百姓", Detail: "赢得清流赞许。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 4, Legitimacy: 3}, Outcome: "太傅捻须点头，清流大臣开始把你的名字写进密札。"},
				{ID: "answer-army", Text: "答：先强兵甲", Detail: "让武臣另眼相看。", Domain: DomainStory, Effects: Effects{Martial: 8, Influence: 2}, Outcome: "武学师傅当场请你试弓，你拉不开弓，却拉来了将门的好感。"},
				{ID: "answer-father", Text: "答：先顺父皇", Detail: "谨慎稳妥，少惹麻烦。", Domain: DomainStory, Effects: Effects{Influence: 6, Legitimacy: 2, Learning: 2}, Outcome: "父皇听闻后没有评价，但赏下了一方端砚。宫中人都懂沉默的分量。"},
			},
		},
		{
			ID: "winter-hunt", Title: "皇家冬狩", Year: "十岁", Mood: "锋芒", Art: state.Assets.Dynasties,
			Body: "猎场上，三皇子故意惊马。你摔在雪里，侍卫们一瞬间不敢动。",
			Choices: []Choice{
				{ID: "mount-again", Text: "忍痛重新上马", Detail: "以勇气换取威望。", Domain: DomainStory, Effects: Effects{Martial: 9, Legitimacy: 4, Health: -4}, Outcome: "你带伤上马，雪原上响起军士的喝彩。三皇子的笑容僵住了。"},
				{ID: "protect-servant", Text: "先扶起被撞倒的小内侍", Detail: "仁名会悄悄传开。", Domain: DomainStory, Effects: Effects{Charisma: 8, Populace: 2, Legitimacy: 1}, Outcome: "一个小小内侍救不了天下，却能让天下相信你会低头看人。"},
				{ID: "accuse-brother", Text: "当众指认三皇子", Detail: "直接开战，风险很高。", Domain: DomainStory, Effects: Effects{Influence: 8, Martial: 3, Stability: -3}, Outcome: "猎场瞬间安静。你赢得了一批拥护者，也让夺嫡提早见血。"},
			},
		},
		{
			ID: "flood-memorial", Title: "南河急报", Year: "十四岁", Mood: "试政", Art: state.Assets.Dynasties,
			Body: "南河决堤，朝堂争论赈灾银从何处来。父皇将奏章推到你面前，要你试拟朱批。",
			Choices: []Choice{
				{ID: "open-granary", Text: "开仓赈济，严查贪墨", Detail: "仁政与吏治并行。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 5, Reform: 4}, Outcome: "你的朱批被贴到灾区驿站，流民第一次知道京中还有人惦记他们。"},
				{ID: "borrow-merchants", Text: "向皇商借银平灾", Detail: "来钱快，但商人会记账。", Domain: DomainStory, Effects: Effects{Treasury: 6, Influence: 5, Legitimacy: -2}, Outcome: "堤坝保住了，皇商也在你的未来里占了一席。"},
				{ID: "send-army", Text: "调禁军协助筑堤", Detail: "效率高，武臣得势。", Domain: DomainStory, Effects: Effects{Martial: 5, Army: 4, Influence: 3}, Outcome: "铁甲入泥，军令比公文更快。百姓记住了旗号，也记住了你。"},
			},
		},
		{
			ID: "succession-night", Title: "烛影摇红", Year: "十六岁", Mood: "夺嫡", Art: state.Assets.Hero,
			Body: fmt.Sprintf("父皇病重，诸王入宫。你手中有学识 %d、武略 %d、声望 %d。最后一夜，谁先动，谁就可能坐上明日的朝堂。", state.Stats.Learning, state.Stats.Martial, state.Stats.Legitimacy),
			Choices: []Choice{
				{ID: "secure-edict", Text: "请太傅与中书共同护诏", Detail: "以制度和名分夺位。", Domain: DomainCourt, Effects: Effects{Legitimacy: 10, Stability: 5, Influence: 4}, Outcome: "玉玺落印，群臣跪伏。你没有拔剑，却让刀兵失去了名义。"},
				{ID: "control-guards", Text: "联络禁军封锁宫门", Detail: "以速度换取皇位。", Domain: DomainMilitary, Effects: Effects{Martial: 8, Influence: 8, Stability: -4}, Outcome: "宫门在夜色中合拢。天亮时，反对你的人已经错过了进宫的时辰。"},
				{ID: "appeal-clans", Text: "向宗室许诺共治", Detail: "稳住人心，但埋下掣肘。", Domain: DomainDiplomacy, Effects: Effects{Charisma: 7, Legitimacy: 6, Reform: -2}, Outcome: "宗室长者替你说了第一句话。你得到了皇位，也得到了许多双盯着你的眼睛。"},
			},
		},
	}
	return cloneScene(scenes[min(turn, len(scenes)-1)])
}

func emperorScene(s *GameState) *Scene {
	year := fmt.Sprintf("登基%d年 · %s", max(1, s.ReignYear), s.Season)
	return &Scene{
		ID:    fmt.Sprintf("court-%d", s.Turn),
		Title: "太和朝议",
		Year:  year,
		Mood:  emperorMood(s.Stats),
		Art:   s.Assets.Hero,
		Body:  crisisLine(s) + " 六部、边军、宗室、清流与商帮都在等你落子。选择不是单纯加减数值，派系与省份会记住你的每一道旨意。",
		Choices: []Choice{
			{ID: fmt.Sprintf("relief-%d", s.Turn), Text: "户部开仓，巡抚赈济", Detail: "民政线：压灾情、保民心，但消耗粮银。", Domain: DomainDomestic, Effects: Effects{Treasury: -10, Grain: -8, Populace: 12, Stability: 5, Legitimacy: 2}, Outcome: "粥棚沿官道铺开，流民队伍短了，户部账册却重得像铁。"},
			{ID: fmt.Sprintf("tax-%d", s.Turn), Text: "重估田亩，整顿盐铁", Detail: "财政线：增加国库，激怒豪强商帮。", Domain: DomainEconomy, Effects: Effects{Treasury: 18, Populace: -4, Stability: -4, Reform: 3}, Outcome: "银车入库，盐引重发。商帮笑着谢恩，袖中却攥紧了旧账。"},
			{ID: fmt.Sprintf("train-%d", s.Turn), Text: "拨银练兵，轮戍边镇", Detail: "军务线：提升军力、压边患，耗费国库。", Domain: DomainMilitary, Effects: Effects{Treasury: -12, Army: 15, BorderThreat: -13, Martial: 1}, Outcome: "军营号角重新响亮，边将呈上的地图终于少了几个红圈。"},
			{ID: fmt.Sprintf("envoy-%d", s.Turn), Text: "遣使联姻，分化诸邦", Detail: "外交线：缓冲战争，稳住宗室和外邦。", Domain: DomainDiplomacy, Effects: Effects{Diplomacy: 13, BorderThreat: -8, Treasury: -4, Stability: 1}, Outcome: "使团携金册出关，远方可汗收下礼物，也收起了一半刀锋。"},
			{ID: fmt.Sprintf("reform-%d", s.Turn), Text: "设考成法，裁撤冗官", Detail: "新法线：长期强国，短期引爆党争。", Domain: DomainReform, Effects: Effects{Reform: 12, Treasury: 5, Stability: -7, Populace: 3, Legitimacy: 2}, Outcome: "新法像春雷落进官场。有人称颂清明，也有人开始深夜串门。"},
			{ID: fmt.Sprintf("spy-%d", s.Turn), Text: "开缇骑密档，夜审朋党", Detail: "暗线：削派系权势，但损声望与稳定。", Domain: DomainIntrigue, Effects: Effects{Influence: 7, Stability: -6, Legitimacy: -4, Health: -2}, Outcome: "宫门后的灯亮到三更，第二天朝会上少了几张熟悉的脸。"},
			{ID: fmt.Sprintf("banquet-%d", s.Turn), Text: "大宴群臣，粉饰太平", Detail: "宫廷线：短期安抚派系，长期损害国本。", Domain: DomainCourt, Effects: Effects{Treasury: -9, Grain: -5, Populace: -7, Stability: -5, BorderThreat: 5, Health: -2}, Outcome: "钟鼓响彻宫城，杯盏遮住了奏章。散席后，问题仍在殿外等你。"},
		},
	}
}

func dynasties() []Dynasty {
	return []Dynasty{
		{ID: "dayin", Name: "大胤", Era: "开国元年", Background: "旧都新定，功臣拥兵，百废待兴。", Features: []string{"开国功臣强势", "国库充实但朝制未稳", "军功路线收益更高"}, Challenge: "用刀剑打下天下后，如何让刀剑回鞘。", Asset: "/assets/dynasty-scroll.png", Palette: "ember", Initial: Stats{Legitimacy: 58, Health: 74, Learning: 20, Martial: 28, Charisma: 24, Influence: 20, Treasury: 78, Grain: 62, Populace: 48, Army: 82, Diplomacy: 36, Stability: 42, BorderThreat: 46, Reform: 12}},
		{ID: "jingyao", Name: "景曜", Era: "盛世中叶", Background: "漕运通达，市井繁华，盛世的每一道裂缝都藏在金粉下。", Features: []string{"财政与外交基础优秀", "改革阻力较小", "奢靡会放大危机"}, Challenge: "在歌舞升平里提前看见衰败。", Asset: "/assets/dynasty-scroll.png", Palette: "gold", Initial: Stats{Legitimacy: 64, Health: 72, Learning: 26, Martial: 18, Charisma: 30, Influence: 20, Treasury: 92, Grain: 88, Populace: 76, Army: 58, Diplomacy: 70, Stability: 72, BorderThreat: 26, Reform: 20}},
		{ID: "chengping", Name: "承平", Era: "暮年危局", Background: "库银亏空，兼并成风，灾民与朋党一起挤进奏章。", Features: []string{"财政压力极高", "新法收益更大", "民变风险更快累积"}, Challenge: "在旧制度的裂缝里硬生生开出新路。", Asset: "/assets/dynasty-scroll.png", Palette: "storm", Initial: Stats{Legitimacy: 46, Health: 68, Learning: 28, Martial: 16, Charisma: 22, Influence: 18, Treasury: 36, Grain: 38, Populace: 34, Army: 48, Diplomacy: 42, Stability: 30, BorderThreat: 52, Reform: 8}},
		{ID: "xuanshuo", Name: "玄朔", Era: "北境烽烟", Background: "雪岭烽火连年，边镇半独立，朝廷每一次迟疑都会变成战报。", Features: []string{"边患开局最高", "军务外交回报更高", "粮草消耗更凶"}, Challenge: "一手握兵符，一手还要稳住中原民心。", Asset: "/assets/dynasty-scroll.png", Palette: "frost", Initial: Stats{Legitimacy: 52, Health: 72, Learning: 20, Martial: 30, Charisma: 20, Influence: 22, Treasury: 58, Grain: 46, Populace: 46, Army: 76, Diplomacy: 34, Stability: 44, BorderThreat: 72, Reform: 10}},
	}
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
		{ID: "gu", Name: "顾怀章", Role: "太傅", Trait: "刚正", Loyalty: 62, Portrait: "tutor"},
		{ID: "huo", Name: "霍骁", Role: "大将军", Trait: "敢战", Loyalty: 55, Portrait: "general"},
		{ID: "shen", Name: "沈万策", Role: "户部尚书", Trait: "精算", Loyalty: 48, Portrait: "minister"},
		{ID: "princess", Name: "昭宁", Role: "长公主", Trait: "纵横", Loyalty: 56, Portrait: "consort"},
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

func startingObjectives(dynastyID string) []Objective {
	objectives := []Objective{
		{ID: "secure_throne", Title: "稳坐龙椅", Description: "完成皇子成长与夺嫡，正式登基。", Target: 100, Reward: "开启六部朝政与天下棋盘。"},
		{ID: "stabilize_realm", Title: "安民定国", Description: "让朝稳、民心、粮草都回到可持续水平。", Target: 80, Reward: "降低危机钟推进速度。"},
		{ID: "reform_state", Title: "鼎新旧制", Description: "推行新法，建立足以撑起盛世的新制度。", Target: 80, Reward: "提升长期财政与民生收益。"},
		{ID: "pacify_borders", Title: "靖平边患", Description: "用军务与外交压低边患，稳住北境与西陲。", Target: 80, Reward: "解锁盛世结局条件之一。"},
	}
	if dynastyID == "chengping" {
		objectives = append(objectives, Objective{ID: "restore_treasury", Title: "补回亏空", Description: "让国库脱离危险线，阻止财政崩盘。", Target: 70, Reward: "财政线消耗降低。"})
	}
	if dynastyID == "xuanshuo" {
		objectives = append(objectives, Objective{ID: "hold_north", Title: "守住雪岭", Description: "提升北境防御，避免烽火压城。", Target: 75, Reward: "边镇武勋忠诚提高。"})
	}
	return objectives
}

func (s *GameState) updateObjectives() {
	for i := range s.Objectives {
		objective := s.Objectives[i]
		switch objective.ID {
		case "secure_throne":
			if s.Phase == PhaseEmperor {
				objective.Progress = objective.Target
			} else {
				objective.Progress = clamp(s.Turn*20, 0, objective.Target)
			}
		case "stabilize_realm":
			objective.Progress = clamp((s.Stats.Stability+s.Stats.Populace+s.Stats.Grain)/3, 0, objective.Target)
		case "reform_state":
			objective.Progress = clamp(s.Stats.Reform, 0, objective.Target)
		case "pacify_borders":
			objective.Progress = clamp(80-s.Stats.BorderThreat+s.Stats.Army/4+s.Stats.Diplomacy/6, 0, objective.Target)
		case "restore_treasury":
			objective.Progress = clamp(s.Stats.Treasury, 0, objective.Target)
		case "hold_north":
			objective.Progress = clamp(provinceDefense(s.Provinces, "north")-s.Stats.BorderThreat/3+s.Stats.Army/5, 0, objective.Target)
		}
		objective.Completed = objective.Progress >= objective.Target
		s.Objectives[i] = objective
	}
}

func provinceDefense(provinces []Province, id string) int {
	for _, province := range provinces {
		if province.ID == id {
			return province.Defense
		}
	}
	return 0
}

func crisisLine(s *GameState) string {
	return fmt.Sprintf("%s：%s 当前烈度 %d，危机钟 %d/8。", s.Crisis.Title, s.Crisis.Summary, s.Crisis.Severity, s.Crisis.Clock)
}

func emperorMood(stats Stats) string {
	score := stats.Stability + stats.Populace + stats.Army + stats.Diplomacy - stats.BorderThreat + stats.Reform/2
	switch {
	case score >= 285:
		return "盛世"
	case score >= 210:
		return "可治"
	case score >= 145:
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

func averageProvinceWealth(provinces []Province) int {
	if len(provinces) == 0 {
		return 0
	}
	total := 0
	for _, p := range provinces {
		total += p.Wealth
	}
	return total / len(provinces)
}

func averageProvinceOrder(provinces []Province) int {
	if len(provinces) == 0 {
		return 0
	}
	total := 0
	for _, p := range provinces {
		total += p.Order
	}
	return total / len(provinces)
}

func disasterPressure(provinces []Province) int {
	if len(provinces) == 0 {
		return 0
	}
	total := 0
	for _, p := range provinces {
		total += p.Disaster
	}
	return total / len(provinces)
}

func averageFactionLoyalty(factions []Faction) int {
	if len(factions) == 0 {
		return 100
	}
	total := 0
	for _, f := range factions {
		total += f.Loyalty
	}
	return total / len(factions)
}

func clamp(v, low, high int) int {
	return int(math.Max(float64(low), math.Min(float64(high), float64(v))))
}
