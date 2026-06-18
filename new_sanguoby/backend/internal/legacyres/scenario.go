package legacyres

import "fmt"

// scenario.go assembles a complete, gameplay-ready snapshot of the legacy
// 董卓 scenario (period 1) by combining decoded resources:
//   - resource 57: city economy records (31 bytes each)
//   - resource 58: city names (GBK)
//   - resource 61: person attribute records (15 bytes each)
//   - resource 62: person names (GBK)
//   - embedded C constants: city map grid coordinates (dCityMapId / C_MAP)
//
// City map coordinates come from the original C source constants in
// `pconst.c` (C_MAP is a 12x9 grid; each non-zero cell holds cityIndex+1).

// CityMapWidth and CityMapHeight match the legacy main campaign map grid.
const (
	CityMapWidth  = 12
	CityMapHeight = 9
)

// cMap is the legacy C_MAP constant from pconst.c: a 12x9 grid (108 cells)
// where a non-zero value is (cityIndex + 1) placed at that grid cell.
var cMap = [CityMapWidth * CityMapHeight]uint8{
	0, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 3,
	0, 0, 4, 0, 0, 5, 0, 6, 7, 0, 8, 0,
	0, 0, 0, 9, 10, 11, 0, 12, 13, 14, 0, 0,
	0, 0, 0, 15, 0, 0, 0, 16, 17, 18, 19, 0,
	0, 0, 20, 0, 0, 0, 21, 22, 0, 23, 24, 0,
	0, 0, 25, 26, 0, 0, 27, 28, 29, 0, 30, 0,
	31, 0, 0, 32, 33, 0, 0, 34, 0, 35, 0, 0,
	0, 0, 0, 0, 0, 36, 0, 37, 0, 38, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

// CityCoord is a city grid position on the campaign map.
type CityCoord struct {
	X int
	Y int
}

// CityCoords returns the grid coordinate for each city index (0-based),
// derived from the legacy C_MAP constant. Index i maps to the cell where
// C_MAP == i+1. Cities not present on the grid get (-1,-1).
func CityCoords(cityCount int) []CityCoord {
	coords := make([]CityCoord, cityCount)
	for i := range coords {
		coords[i] = CityCoord{X: -1, Y: -1}
	}
	for cell, v := range cMap {
		if v == 0 {
			continue
		}
		idx := int(v) - 1
		if idx < 0 || idx >= cityCount {
			continue
		}
		coords[idx] = CityCoord{X: cell % CityMapWidth, Y: cell / CityMapWidth}
	}
	return coords
}

// ScenarioCity is a fully-assembled city ready for the game layer.
type ScenarioCity struct {
	Index           int
	Name            string
	X               int
	Y               int
	Belong          int // ruler person index + 1; 0 = neutral
	SatrapID        int
	FarmingLimit    int
	Farming         int
	CommerceLimit   int
	Commerce        int
	PeopleDevotion  int
	AvoidCalamity   int
	PopulationLimit int
	Population      int
	Money           int
	Food            int
}

// ScenarioPerson is a fully-assembled person ready for the game layer.
type ScenarioPerson struct {
	Index     int
	Name      string
	Level     int
	Force     int
	IQ        int
	Devotion  int
	Character string
	ArmsType  string
	Age       int
	// BelongIndex is the ruler person index this person serves, or -1 if unknown.
	BelongIndex int
}

// Scenario is a complete decoded scenario snapshot.
type Scenario struct {
	Period  int
	Year    int
	Cities  []ScenarioCity
	Persons []ScenarioPerson
}

// LoadScenario decodes the period-1 (董卓) scenario into a gameplay-ready form.
func (a *Archive) LoadScenario() (*Scenario, error) {
	return a.LoadScenarioPeriod(1)
}

func (a *Archive) LoadScenarioPeriod(period uint16) (*Scenario, error) {
	cities, cityNames, err := a.DecodeCitiesWithNamesForPeriod(period)
	if err != nil {
		return nil, err
	}
	persons, personNames, err := a.DecodePersonsWithNamesForPeriod(period)
	if err != nil {
		return nil, err
	}

	coords := CityCoords(len(cities))
	year, err := a.DecodeScenarioYear(period)
	if err != nil {
		return nil, err
	}

	sc := &Scenario{Period: int(period), Year: year}
	for i, c := range cities {
		name := ""
		if i < len(cityNames) {
			name = cityNames[i]
		}
		sc.Cities = append(sc.Cities, ScenarioCity{
			Index:           int(c.Index),
			Name:            name,
			X:               coords[i].X,
			Y:               coords[i].Y,
			Belong:          int(c.Belong),
			SatrapID:        int(c.SatrapID),
			FarmingLimit:    int(c.FarmingLimit),
			Farming:         int(c.Farming),
			CommerceLimit:   int(c.CommerceLimit),
			Commerce:        int(c.Commerce),
			PeopleDevotion:  int(c.PeopleDevotion),
			AvoidCalamity:   int(c.AvoidCalamity),
			PopulationLimit: int(c.PopulationLimit),
			Population:      int(c.Population),
			Money:           int(c.Money),
			Food:            int(c.Food),
		})
	}

	for i, p := range persons {
		name := ""
		if i < len(personNames) {
			name = personNames[i]
		}
		sc.Persons = append(sc.Persons, ScenarioPerson{
			Index:       int(p.Index),
			Name:        name,
			Level:       int(p.Level),
			Force:       int(p.Force),
			IQ:          int(p.IQ),
			Devotion:    int(p.Devotion),
			Character:   CharacterNames[p.Character],
			ArmsType:    ArmsTypeNames[p.ArmsType],
			Age:         int(p.Age),
			BelongIndex: -1,
		})
	}

	return sc, nil
}

func (a *Archive) DecodeScenarioYear(period uint16) (int, error) {
	raw, err := a.Item(57, period)
	if err != nil {
		return 0, fmt.Errorf("scenario year resource: %w", err)
	}
	offset := CityMapCityCount() * CityStructSize
	if len(raw) < offset+2 {
		return 0, fmt.Errorf("%w: scenario year offset outside city resource", ErrShortBuffer)
	}
	return int(raw[offset]) | int(raw[offset+1])<<8, nil
}

func CityMapCityCount() int {
	return 38
}
