package game

import (
	"fmt"
	"sort"

	"new_sanguoby/backend/internal/legacyres"
)

// rulerColors assigns distinct campaign colors to rulers in a stable order.
var rulerColors = []string{
	"#9b2f2f", "#2f7d57", "#2e6f9e", "#7a4d9f", "#b07c2f",
	"#3f8f8f", "#a8418a", "#557a2f", "#8a5a2f", "#4a5fae",
	"#9e5b3a", "#3a8f5f", "#7f3f6f", "#5f7f3f", "#2f6f7f",
	"#8f6f2f", "#6f2f8f", "#2f8f6f", "#8f2f4f", "#4f8f2f",
}

// NewGameFromArchive builds a campaign GameState from the legacy dat.lib
// archive at the given path. It replaces the authored seed with the real
// 董卓 scenario: actual city names, economy, ruler factions, and generals.
//
// If the archive cannot be loaded, an error is returned so the caller can fall
// back to the authored seed.
func NewGameFromArchive(archivePath, scenarioID, playerID string) (*GameState, error) {
	archive, err := legacyres.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open legacy archive: %w", err)
	}
	period := periodFromScenarioID(scenarioID)
	scenario, err := archive.LoadScenarioPeriod(uint16(period))
	if err != nil {
		return nil, fmt.Errorf("load legacy scenario: %w", err)
	}
	return buildGameFromScenario(scenario, playerID), nil
}

// buildGameFromScenario converts a decoded legacy scenario into a GameState.
func buildGameFromScenario(sc *legacyres.Scenario, playerID string) *GameState {
	// 1. Determine which person indices are rulers: a ruler is any person whose
	//    (index+1) appears as a city Belong value.
	rulerSet := map[int]bool{}
	for _, c := range sc.Cities {
		if c.Belong > 0 {
			rulerSet[c.Belong-1] = true
		}
	}
	rulerIndices := make([]int, 0, len(rulerSet))
	for idx := range rulerSet {
		rulerIndices = append(rulerIndices, idx)
	}
	sort.Ints(rulerIndices)

	// 2. Build rulers from the corresponding persons.
	rulers := make([]Ruler, 0, len(rulerIndices)+1)
	rulerID := map[int]string{}
	for i, idx := range rulerIndices {
		var p *legacyres.ScenarioPerson
		if idx >= 0 && idx < len(sc.Persons) {
			p = &sc.Persons[idx]
		}
		name := fmt.Sprintf("君主%d", idx)
		character := "未知"
		if p != nil {
			name = p.Name
			character = p.Character
		}
		id := fmt.Sprintf("ruler-%d", idx)
		rulerID[idx] = id
		rulers = append(rulers, Ruler{
			ID:        id,
			Name:      name,
			Character: character,
			Color:     rulerColors[i%len(rulerColors)],
		})
	}
	rulers = append(rulers, Ruler{ID: "neutral", Name: "空城", Character: "无", Color: "#7f7a68"})

	// 3. Build cities.
	cityIDByIndex := map[int]string{}
	cities := make([]City, 0, len(sc.Cities))
	for _, c := range sc.Cities {
		ownerID := "neutral"
		if c.Belong > 0 {
			if id, ok := rulerID[c.Belong-1]; ok {
				ownerID = id
			}
		}
		id := fmt.Sprintf("city-%d", c.Index)
		cityIDByIndex[c.Index] = id
		cities = append(cities, City{
			ID:              id,
			Name:            c.Name,
			X:               c.X,
			Y:               c.Y,
			OwnerID:         ownerID,
			State:           CityStateNormal,
			FarmingLimit:    c.FarmingLimit,
			Farming:         c.Farming,
			CommerceLimit:   c.CommerceLimit,
			Commerce:        c.Commerce,
			PeopleDevotion:  c.PeopleDevotion,
			AvoidCalamity:   c.AvoidCalamity,
			PopulationLimit: c.PopulationLimit,
			Population:      c.Population,
			Money:           c.Money,
			Food:            c.Food,
			Garrison:        0,
		})
	}

	// 4. Pick a capital city per ruler (first city they own).
	capitalByRuler := map[int]string{}
	for _, c := range sc.Cities {
		if c.Belong > 0 {
			rulerIdx := c.Belong - 1
			if _, ok := capitalByRuler[rulerIdx]; !ok {
				capitalByRuler[rulerIdx] = cityIDByIndex[c.Index]
			}
		}
	}

	// 5. Place city satraps where the legacy scenario names them, then ensure
	//    every ruler exists in their capital. The compact resource does not keep
	//    full person queues, but this restores the old flow: most owned cities
	//    have a usable officer for player commands.
	generals := make([]General, 0, len(rulerIndices)+len(sc.Cities))
	usedPersons := map[int]bool{}
	addGeneral := func(personIndex int, ownerID string, cityID string) {
		if personIndex < 0 || personIndex >= len(sc.Persons) || usedPersons[personIndex] {
			return
		}
		p := sc.Persons[personIndex]
		level := p.Level
		if level <= 0 {
			level = 1
		}
		generals = append(generals, General{
			ID:        fmt.Sprintf("gen-%d", p.Index),
			Name:      p.Name,
			OwnerID:   ownerID,
			CityID:    cityID,
			Level:     level,
			Force:     p.Force,
			Intellect: p.IQ,
			Loyalty:   p.Devotion,
			Stamina:   100,
			Soldiers:  initialSoldiers(p.Force),
			ArmsType:  p.ArmsType,
		})
		usedPersons[personIndex] = true
	}

	for _, c := range sc.Cities {
		if c.Belong <= 0 {
			continue
		}
		ownerID := rulerID[c.Belong-1]
		cityID := cityIDByIndex[c.Index]
		addGeneral(c.SatrapID, ownerID, cityID)
	}

	for _, idx := range rulerIndices {
		cityID, ok := capitalByRuler[idx]
		if !ok {
			continue
		}
		addGeneral(idx, rulerID[idx], cityID)
	}

	// 6. Build routes between adjacent cities on the grid (4-neighbour).
	routes := buildGridRoutes(sc.Cities, cityIDByIndex)

	// 7. Resolve player ruler: default to the first ruler if not specified/found.
	resolvedPlayer := ""
	if playerID != "" {
		for _, r := range rulers {
			if r.ID == playerID {
				resolvedPlayer = playerID
				break
			}
		}
	}
	if resolvedPlayer == "" && len(rulers) > 0 {
		resolvedPlayer = rulers[0].ID
	}

	return &GameState{
		ScenarioID: fmt.Sprintf("period-%d", sc.Period),
		PlayerID:   resolvedPlayer,
		Date:       Date{Year: sc.Year, Month: 1},
		Rulers:     rulers,
		Cities:     cities,
		Generals:   generals,
		Routes:     routes,
		Log:        []string{fmt.Sprintf("%d年，群雄并起，董卓弄权。", sc.Year)},
	}
}

func periodFromScenarioID(scenarioID string) int {
	switch scenarioID {
	case "period-2", "caocao-rise":
		return 2
	case "period-3", "chibi":
		return 3
	case "period-4", "three-kingdoms":
		return 4
	default:
		return 1
	}
}

// initialSoldiers derives a faithful starting troop count for a ruler general
// from their martial force (higher 武力 commands a larger initial army).
func initialSoldiers(force int) int {
	soldiers := 500 + force*10
	if soldiers > 2000 {
		soldiers = 2000
	}
	return soldiers
}

// buildGridRoutes connects cities that are orthogonally adjacent on the campaign
// grid, producing a clean road network for the strategic map.
func buildGridRoutes(cities []legacyres.ScenarioCity, idByIndex map[int]string) []Route {
	type cell struct{ x, y int }
	occupied := map[cell]int{}
	for _, c := range cities {
		if c.X >= 0 && c.Y >= 0 {
			occupied[cell{c.X, c.Y}] = c.Index
		}
	}

	seen := map[string]bool{}
	routes := make([]Route, 0)
	addRoute := func(a, b int) {
		ka := idByIndex[a]
		kb := idByIndex[b]
		key := ka + "|" + kb
		if a > b {
			key = kb + "|" + ka
		}
		if seen[key] {
			return
		}
		seen[key] = true
		routes = append(routes, Route{From: ka, To: kb})
	}

	for _, c := range cities {
		if c.X < 0 || c.Y < 0 {
			continue
		}
		// Search outward along the 4 cardinal directions to the nearest city,
		// so the road network stays connected even with sparse grids.
		dirs := []cell{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
		for _, d := range dirs {
			for step := 1; step < legacyres.CityMapWidth; step++ {
				nx, ny := c.X+d.x*step, c.Y+d.y*step
				if nx < 0 || ny < 0 || nx >= legacyres.CityMapWidth || ny >= legacyres.CityMapHeight {
					break
				}
				if other, ok := occupied[cell{nx, ny}]; ok {
					addRoute(c.Index, other)
					break
				}
			}
		}
	}
	return routes
}
