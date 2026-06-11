package game

import "math/rand"

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

type OrderKind string

const (
	OrderRelief   OrderKind = "relief"
	OrderGarrison OrderKind = "garrison"
	OrderTax      OrderKind = "tax"
	OrderInspect  OrderKind = "inspect"
	OrderAppease  OrderKind = "appease"
	OrderPurge    OrderKind = "purge"
	OrderCanal    OrderKind = "canal"
	OrderTrade    OrderKind = "trade"
	OrderMobilize OrderKind = "mobilize"
	OrderCampaign OrderKind = "campaign"
	OrderFortify  OrderKind = "fortify"
	OrderTruce    OrderKind = "truce"
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
	Hero            string   `json:"hero"`
	Dynasties       string   `json:"dynasties"`
	Characters      string   `json:"characters"`
	SceneGallery    []string `json:"sceneGallery"`
	PortraitGallery []string `json:"portraitGallery"`
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
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Trait     string `json:"trait"`
	Loyalty   int    `json:"loyalty"`
	Ability   int    `json:"ability"`
	Ambition  int    `json:"ambition"`
	Integrity int    `json:"integrity"`
	Stress    int    `json:"stress"`
	Portrait  string `json:"portrait"`
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

type WarCampaign struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Enemy    string `json:"enemy"`
	Front    string `json:"front"`
	Stage    string `json:"stage"`
	Threat   int    `json:"threat"`
	Supply   int    `json:"supply"`
	Morale   int    `json:"morale"`
	Progress int    `json:"progress"`
	Duration int    `json:"duration"`
}

type OrderRequest struct {
	Kind   OrderKind `json:"kind"`
	Target string    `json:"target"`
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
	Command    int            `json:"command"`
	Stats      Stats          `json:"stats"`
	Factions   []Faction      `json:"factions"`
	Court      []Minister     `json:"court"`
	Provinces  []Province     `json:"provinces"`
	Wars       []WarCampaign  `json:"wars"`
	Crisis     Crisis         `json:"crisis"`
	Objectives []Objective    `json:"objectives"`
	Scene      *Scene         `json:"scene,omitempty"`
	Ending     *Ending        `json:"ending,omitempty"`
	History    []HistoryEntry `json:"history"`

	rng *rand.Rand
}
