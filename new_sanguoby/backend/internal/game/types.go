package game

import "fmt"

type CityState string

const (
	CityStateNormal    CityState = "normal"    // 正常
	CityStateFamine    CityState = "famine"    // 饥荒
	CityStateDrought   CityState = "drought"   // 旱灾
	CityStateFlood     CityState = "flood"     // 水灾
	CityStateRebellion CityState = "rebellion" // 暴动
)

type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

type Ruler struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Character string `json:"character"`
	Color     string `json:"color"`
}

type City struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	X               int       `json:"x"`
	Y               int       `json:"y"`
	OwnerID         string    `json:"ownerId"`
	SatrapID        string    `json:"satrapId"` // 太守武将ID
	State           CityState `json:"state"`
	FarmingLimit    int       `json:"farmingLimit"`
	Farming         int       `json:"farming"`
	CommerceLimit   int       `json:"commerceLimit"`
	Commerce        int       `json:"commerce"`
	PeopleDevotion  int       `json:"peopleDevotion"`
	AvoidCalamity   int       `json:"avoidCalamity"`
	PopulationLimit int       `json:"populationLimit"`
	Population      int       `json:"population"`
	Money           int       `json:"money"`
	Food            int       `json:"food"`
	MothballArms    int       `json:"mothballArms"` // 后备兵力
	Garrison        int       `json:"garrison"`
}

type General struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	OwnerID    string `json:"ownerId"`
	CityID     string `json:"cityId"`
	Level      int    `json:"level"`
	Force      int    `json:"force"`
	Intellect  int    `json:"intellect"`
	Loyalty    int    `json:"loyalty"`
	Character  int    `json:"character"` // 0卤莽 1怕死 2贪财 3大志 4忠义
	Experience int    `json:"experience"`
	Stamina    int    `json:"stamina"`
	Soldiers   int    `json:"soldiers"`
	ArmsType   string `json:"armsType"` // 骑兵/步兵/弓箭兵/水军/极兵/玄兵
	Equip      [2]int `json:"equip"`    // 装备道具ID
	Age        int    `json:"age"`
	Captive    bool   `json:"captive"`
}

type Route struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GameState struct {
	ScenarioID string    `json:"scenarioId"`
	PlayerID   string    `json:"playerId"`
	Date       Date      `json:"date"`
	Rulers     []Ruler   `json:"rulers"`
	Cities     []City    `json:"cities"`
	Generals   []General `json:"generals"`
	Routes     []Route   `json:"routes"`
	Log        []string  `json:"log"`
}

func (s *GameState) CityByID(id string) *City {
	for i := range s.Cities {
		if s.Cities[i].ID == id {
			return &s.Cities[i]
		}
	}
	panic(fmt.Sprintf("city %q not found", id))
}

func (s *GameState) GeneralByID(id string) *General {
	for i := range s.Generals {
		if s.Generals[i].ID == id {
			return &s.Generals[i]
		}
	}
	panic(fmt.Sprintf("general %q not found", id))
}

func (s *GameState) PlayerRuler() Ruler {
	for _, ruler := range s.Rulers {
		if ruler.ID == s.PlayerID {
			return ruler
		}
	}
	return Ruler{}
}
