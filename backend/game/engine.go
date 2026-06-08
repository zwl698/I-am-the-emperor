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
	DomainMilitary  Domain = "military"
	DomainDiplomacy Domain = "diplomacy"
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
	ID      string         `json:"id"`
	Seed    int64          `json:"seed"`
	Turn    int            `json:"turn"`
	Age     int            `json:"age"`
	Phase   Phase          `json:"phase"`
	Stats   Stats          `json:"stats"`
	Scene   *Scene         `json:"scene,omitempty"`
	Ending  *Ending        `json:"ending,omitempty"`
	History []HistoryEntry `json:"history"`

	rng *rand.Rand
}

func NewGame(seed int64) *GameState {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	state := &GameState{
		ID:    fmt.Sprintf("dragon-%x", seed^time.Now().UnixNano()),
		Seed:  seed,
		Turn:  0,
		Age:   1,
		Phase: PhasePrince,
		Stats: Stats{
			Legitimacy: 56,
			Health:     72,
			Learning:   20,
			Martial:    18,
			Charisma:   24,
			Influence:  18,
		},
		rng: rand.New(rand.NewSource(seed)),
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
	s.Phase = PhaseEmperor
	s.Age = 18
	s.Turn = 5
	s.Stats.Treasury = 72
	s.Stats.Grain = 66
	s.Stats.Populace = 64
	s.Stats.Army = 58
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
	s.Stats.Martial = clamp(s.Stats.Martial+e.Martial, 0, 100)
	s.Stats.Charisma = clamp(s.Stats.Charisma+e.Charisma, 0, 100)
	s.Stats.Influence = clamp(s.Stats.Influence+e.Influence, 0, 100)
	s.Stats.Treasury = clamp(s.Stats.Treasury+e.Treasury, 0, 140)
	s.Stats.Grain = clamp(s.Stats.Grain+e.Grain, 0, 140)
	s.Stats.Populace = clamp(s.Stats.Populace+e.Populace, 0, 100)
	s.Stats.Army = clamp(s.Stats.Army+e.Army, 0, 120)
	s.Stats.Diplomacy = clamp(s.Stats.Diplomacy+e.Diplomacy, 0, 100)
	s.Stats.Stability = clamp(s.Stats.Stability+e.Stability, 0, 100)
	s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+e.BorderThreat, 0, 100)
	s.Stats.Reform = clamp(s.Stats.Reform+e.Reform, 0, 100)
}

func (s *GameState) advanceAfter(choice Choice) {
	if s.Phase == PhasePrince {
		s.Age = []int{6, 10, 14, 16, 18}[min(s.Turn-1, 4)]
		if s.Turn >= 5 {
			s.coronate()
		}
		return
	}
	if s.Turn%2 == 0 {
		s.Age++
	}
}

func (s *GameState) coronate() {
	s.Phase = PhaseEmperor
	s.Age = max(s.Age, 18)
	s.Stats.Treasury = clamp(50+s.Stats.Legitimacy/3+s.Stats.Learning/5, 35, 120)
	s.Stats.Grain = clamp(48+s.Stats.Charisma/4, 35, 120)
	s.Stats.Populace = clamp(48+s.Stats.Charisma/5+s.Stats.Legitimacy/6, 30, 100)
	s.Stats.Army = clamp(42+s.Stats.Martial/2, 30, 120)
	s.Stats.Diplomacy = clamp(38+s.Stats.Charisma/3+s.Stats.Learning/6, 25, 100)
	s.Stats.Stability = clamp(42+s.Stats.Influence/3+s.Stats.Legitimacy/6, 25, 100)
	s.Stats.BorderThreat = clamp(48-s.Stats.Martial/5, 15, 75)
}

func (s *GameState) applyWorldPressure(domain Domain) {
	if s.Phase != PhaseEmperor {
		return
	}

	s.Stats.Treasury = clamp(s.Stats.Treasury-2, 0, 140)
	s.Stats.Grain = clamp(s.Stats.Grain-1, 0, 140)

	if domain != DomainMilitary {
		s.Stats.BorderThreat = clamp(s.Stats.BorderThreat+2+s.rng.Intn(4), 0, 100)
	}
	if s.Stats.Grain < 25 {
		s.Stats.Populace = clamp(s.Stats.Populace-4, 0, 100)
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
		return &Ending{
			Kind:    EndingDeath,
			Title:   "龙驭宾天",
			Summary: "操劳与暗疾耗尽了你的生命，史官在夜色里合上实录。",
		}
	}
	if s.Phase == PhaseEmperor && (s.Stats.Stability <= 0 || s.Stats.Populace <= 0 || (s.Stats.BorderThreat >= 95 && s.Stats.Army < 80)) {
		return &Ending{
			Kind:    EndingCollapse,
			Title:   "山河失守",
			Summary: "边患、民怨与朝局崩裂同时袭来，王朝在你的御座前倾塌。",
		}
	}
	if s.Phase == PhaseEmperor && s.Turn >= 45 && s.Stats.Stability >= 80 && s.Stats.Populace >= 80 && s.Stats.BorderThreat <= 25 {
		return &Ending{
			Kind:    EndingGolden,
			Title:   "万邦来朝",
			Summary: "仓廪充盈，边境安定，诸国遣使入贡，你开创了被后世反复吟诵的盛世。",
		}
	}
	return nil
}

func (s *GameState) nextScene() *Scene {
	if s.Phase == PhasePrince {
		return princeScene(s.Turn, s.Stats)
	}
	return emperorScene(s)
}

func princeScene(turn int, stats Stats) *Scene {
	scenes := []Scene{
		{
			ID:    "birth-omen",
			Title: "紫宸宫中的啼哭",
			Year:  "皇子元年",
			Mood:  "启蒙",
			Body:  "你出生在风雪初停的清晨。太史说紫气绕阙，贵妃却知道，宫里每一句吉言都可能变成刀锋。",
			Choices: []Choice{
				{ID: "grab-scroll", Text: "抓起案上的竹简", Detail: "让太傅记住你早慧的一面。", Domain: DomainStory, Effects: Effects{Learning: 8, Legitimacy: 2, Health: -1}, Outcome: "你咿呀抓住竹简，满殿笑声里，多了一个“好学皇子”的传闻。"},
				{ID: "smile-consort", Text: "向皇后展露笑容", Detail: "讨得中宫欢心，但母妃心中不安。", Domain: DomainStory, Effects: Effects{Charisma: 7, Influence: 4}, Outcome: "皇后轻抚你的额头，宫人们很快学会了对你多行半礼。"},
				{ID: "cry-loudly", Text: "放声大哭震住众人", Detail: "生命力旺盛，也显得倔强难驯。", Domain: DomainStory, Effects: Effects{Health: 8, Martial: 3, Charisma: -1}, Outcome: "哭声穿过朱门，老内侍低声说：这孩子有帝王的肺腑。"},
			},
		},
		{
			ID:    "study-yard",
			Title: "东宫书院",
			Year:  "六岁",
			Mood:  "养成",
			Body:  "皇子们第一次同席读书。太傅问你：治国先治什么？兄弟们都在等你出错。",
			Choices: []Choice{
				{ID: "answer-people", Text: "答：先安百姓", Detail: "赢得清流赞许。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 4, Legitimacy: 3}, Outcome: "太傅捻须点头，清流大臣开始把你的名字写进密札。"},
				{ID: "answer-army", Text: "答：先强兵甲", Detail: "让武臣另眼相看。", Domain: DomainStory, Effects: Effects{Martial: 8, Influence: 2}, Outcome: "武学师傅当场请你试弓，你拉不开弓，却拉来了将门的好感。"},
				{ID: "answer-father", Text: "答：先顺父皇", Detail: "谨慎稳妥，少惹麻烦。", Domain: DomainStory, Effects: Effects{Influence: 6, Legitimacy: 2, Learning: 2}, Outcome: "父皇听闻后没有评价，但赏下了一方端砚。宫中人都懂沉默的分量。"},
			},
		},
		{
			ID:    "winter-hunt",
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
			ID:    "flood-memorial",
			Title: "南河急报",
			Year:  "十四岁",
			Mood:  "试政",
			Body:  "南河决堤，朝堂争论赈灾银从何处来。父皇将奏章推到你面前，要你试拟朱批。",
			Choices: []Choice{
				{ID: "open-granary", Text: "开仓赈济，严查贪墨", Detail: "仁政与吏治并行。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 5, Reform: 4}, Outcome: "你的朱批被贴到灾区驿站，流民第一次知道京中还有人惦记他们。"},
				{ID: "borrow-merchants", Text: "向皇商借银平灾", Detail: "来钱快，但商人会记账。", Domain: DomainStory, Effects: Effects{Treasury: 6, Influence: 5, Legitimacy: -2}, Outcome: "堤坝保住了，皇商也在你的未来里占了一席。"},
				{ID: "send-army", Text: "调禁军协助筑堤", Detail: "效率高，武臣得势。", Domain: DomainStory, Effects: Effects{Martial: 5, Army: 4, Influence: 3}, Outcome: "铁甲入泥，军令比公文更快。百姓记住了旗号，也记住了你。"},
			},
		},
		{
			ID:    "succession-night",
			Title: "烛影摇红",
			Year:  "十六岁",
			Mood:  "夺嫡",
			Body:  fmt.Sprintf("父皇病重，诸王入宫。你手中有学识 %d、武略 %d、声望 %d。最后一夜，谁先动，谁就可能坐上明日的朝堂。", stats.Learning, stats.Martial, stats.Legitimacy),
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
