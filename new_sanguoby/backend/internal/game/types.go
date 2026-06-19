package game

import "fmt"

type CityState string

const (
	CityStateNormal CityState = "normal"
	CityStateFamine CityState = "famine"
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
	Garrison        int       `json:"garrison"`
}

type General struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerID   string `json:"ownerId"`
	CityID    string `json:"cityId"`
	Level     int    `json:"level"`
	Force     int    `json:"force"`
	Intellect int    `json:"intellect"`
	Loyalty   int    `json:"loyalty"`
	Stamina   int    `json:"stamina"`
	Soldiers  int    `json:"soldiers"`
	ArmsType  string `json:"armsType"`
	Captive   bool   `json:"captive"`
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
