package game

type StrategicState struct {
	Cities   []StrategicCity    `json:"cities"`
	Roads    []StrategicRoad    `json:"roads"`
	Factions []StrategicFaction `json:"factions"`
	Armies   []ArmyGroup        `json:"armies"`
	Logs     []StrategyLog      `json:"logs"`
	Battles  []BattleReport     `json:"battles"`
}

type StrategicCity struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Region      string   `json:"region"`
	OwnerID     string   `json:"ownerId"`
	GovernorID  string   `json:"governorId"`
	X           int      `json:"x"`
	Y           int      `json:"y"`
	Population  int      `json:"population"`
	Commerce    int      `json:"commerce"`
	Agriculture int      `json:"agriculture"`
	Defense     int      `json:"defense"`
	Order       int      `json:"order"`
	Disaster    int      `json:"disaster"`
	Troops      int      `json:"troops"`
	Grain       int      `json:"grain"`
	Gold        int      `json:"gold"`
	Front       bool     `json:"front"`
	Tags        []string `json:"tags"`
}

type StrategicRoad struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Terrain  string `json:"terrain"`
	Risk     int    `json:"risk"`
	Distance int    `json:"distance"`
}

type StrategicFaction struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Ruler         string `json:"ruler"`
	Relation      int    `json:"relation"`
	Threat        int    `json:"threat"`
	Strategy      string `json:"strategy"`
	CapitalCityID string `json:"capitalCityId"`
	Color         string `json:"color"`
	IsPlayer      bool   `json:"isPlayer"`
}

type ArmyGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FactionID   string `json:"factionId"`
	Location    string `json:"location"`
	Target      string `json:"target"`
	CommanderID string `json:"commanderId"`
	Troops      int    `json:"troops"`
	Grain       int    `json:"grain"`
	Morale      int    `json:"morale"`
	Training    int    `json:"training"`
	Siege       int    `json:"siege"`
	Status      string `json:"status"`
}

type StrategyLog struct {
	Turn     int    `json:"turn"`
	Season   string `json:"season"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Severity int    `json:"severity"`
}

type BattleReport struct {
	Turn         int      `json:"turn"`
	Season       string   `json:"season"`
	Title        string   `json:"title"`
	CityID       string   `json:"cityId"`
	Attacker     string   `json:"attacker"`
	Defender     string   `json:"defender"`
	Outcome      string   `json:"outcome"`
	AttackerLoss int      `json:"attackerLoss"`
	DefenderLoss int      `json:"defenderLoss"`
	Participants []string `json:"participants"`
	Summary      string   `json:"summary"`
	Factors      []string `json:"factors,omitempty"`
	Severity     int      `json:"severity"`
}

func startingStrategicState(dynastyID string) StrategicState {
	state := StrategicState{
		Cities: []StrategicCity{
			{ID: "capital", Name: "京畿", Region: "中枢", OwnerID: "court", GovernorID: "gu", X: 51, Y: 44, Population: 95, Commerce: 72, Agriculture: 58, Defense: 72, Order: 66, Disaster: 12, Troops: 18000, Grain: 86, Gold: 92, Tags: []string{"都城", "朝堂"}},
			{ID: "luoyang", Name: "洛阳", Region: "中原", OwnerID: "court", GovernorID: "shen", X: 43, Y: 48, Population: 76, Commerce: 66, Agriculture: 62, Defense: 58, Order: 58, Disaster: 16, Troops: 9000, Grain: 70, Gold: 66, Tags: []string{"旧都", "粮道"}},
			{ID: "river-east", Name: "河东", Region: "中原", OwnerID: "court", X: 48, Y: 34, Population: 58, Commerce: 46, Agriculture: 60, Defense: 52, Order: 48, Disaster: 24, Troops: 7000, Grain: 54, Gold: 42, Tags: []string{"渡口"}},
			{ID: "north", Name: "北境", Region: "北疆", OwnerID: "court", GovernorID: "huo", X: 50, Y: 20, Population: 42, Commerce: 32, Agriculture: 38, Defense: 78, Order: 50, Disaster: 28, Troops: 15000, Grain: 48, Gold: 36, Front: true, Tags: []string{"边塞", "马政"}},
			{ID: "snow-ridge", Name: "雪岭", Region: "北疆", OwnerID: "beidi", X: 50, Y: 8, Population: 25, Commerce: 18, Agriculture: 18, Defense: 64, Order: 54, Disaster: 36, Troops: 20000, Grain: 52, Gold: 28, Front: true, Tags: []string{"关隘", "雪道"}},
			{ID: "west", Name: "西陲", Region: "西疆", OwnerID: "court", X: 27, Y: 45, Population: 44, Commerce: 48, Agriculture: 36, Defense: 54, Order: 48, Disaster: 20, Troops: 8000, Grain: 42, Gold: 46, Tags: []string{"商路"}},
			{ID: "jade-pass", Name: "玉门", Region: "西疆", OwnerID: "remnant", X: 16, Y: 41, Population: 30, Commerce: 44, Agriculture: 24, Defense: 60, Order: 42, Disaster: 18, Troops: 12000, Grain: 38, Gold: 36, Front: true, Tags: []string{"关市", "沙路"}},
			{ID: "south", Name: "江南", Region: "江南", OwnerID: "court", GovernorID: "princess", X: 58, Y: 68, Population: 90, Commerce: 82, Agriculture: 80, Defense: 42, Order: 58, Disaster: 18, Troops: 6000, Grain: 90, Gold: 84, Tags: []string{"粮仓", "漕运"}},
			{ID: "canal", Name: "漕都", Region: "江南", OwnerID: "court", X: 51, Y: 59, Population: 68, Commerce: 76, Agriculture: 72, Defense: 46, Order: 54, Disaster: 20, Troops: 5000, Grain: 84, Gold: 70, Tags: []string{"漕运", "府库"}},
			{ID: "east-sea", Name: "东海", Region: "海疆", OwnerID: "haiguo", X: 76, Y: 61, Population: 40, Commerce: 70, Agriculture: 28, Defense: 48, Order: 50, Disaster: 16, Troops: 8000, Grain: 44, Gold: 64, Front: true, Tags: []string{"海贸", "舟师"}},
			{ID: "nanling", Name: "南岭", Region: "南疆", OwnerID: "nanling", X: 45, Y: 84, Population: 38, Commerce: 34, Agriculture: 44, Defense: 58, Order: 56, Disaster: 22, Troops: 11000, Grain: 46, Gold: 30, Front: true, Tags: []string{"山寨", "边贸"}},
			{ID: "bashu", Name: "巴蜀", Region: "西南", OwnerID: "court", X: 27, Y: 74, Population: 62, Commerce: 48, Agriculture: 76, Defense: 62, Order: 56, Disaster: 18, Troops: 7000, Grain: 82, Gold: 44, Tags: []string{"天府", "山道"}},
			{ID: "dockyard", Name: "海舶司", Region: "海疆", OwnerID: "court", X: 69, Y: 72, Population: 45, Commerce: 68, Agriculture: 34, Defense: 44, Order: 52, Disaster: 14, Troops: 5000, Grain: 48, Gold: 66, Tags: []string{"海贸", "舟师"}},
			{ID: "mountain-pass", Name: "剑门", Region: "西南", OwnerID: "court", X: 33, Y: 62, Population: 34, Commerce: 28, Agriculture: 42, Defense: 76, Order: 52, Disaster: 24, Troops: 7000, Grain: 45, Gold: 26, Tags: []string{"关隘", "山道"}},
		},
		Roads: []StrategicRoad{
			{From: "capital", To: "luoyang", Terrain: "官道", Risk: 12, Distance: 2},
			{From: "capital", To: "river-east", Terrain: "官道", Risk: 18, Distance: 2},
			{From: "river-east", To: "north", Terrain: "边道", Risk: 28, Distance: 3},
			{From: "north", To: "snow-ridge", Terrain: "雪道", Risk: 48, Distance: 3},
			{From: "luoyang", To: "west", Terrain: "商路", Risk: 24, Distance: 3},
			{From: "west", To: "jade-pass", Terrain: "沙路", Risk: 42, Distance: 3},
			{From: "luoyang", To: "canal", Terrain: "漕河", Risk: 16, Distance: 2},
			{From: "canal", To: "south", Terrain: "漕河", Risk: 12, Distance: 2},
			{From: "south", To: "dockyard", Terrain: "水路", Risk: 20, Distance: 2},
			{From: "dockyard", To: "east-sea", Terrain: "海路", Risk: 34, Distance: 3},
			{From: "south", To: "nanling", Terrain: "山路", Risk: 36, Distance: 3},
			{From: "west", To: "mountain-pass", Terrain: "山道", Risk: 30, Distance: 2},
			{From: "mountain-pass", To: "bashu", Terrain: "山道", Risk: 26, Distance: 2},
			{From: "bashu", To: "nanling", Terrain: "山道", Risk: 34, Distance: 3},
			{From: "capital", To: "canal", Terrain: "御道", Risk: 10, Distance: 2},
			{From: "river-east", To: "luoyang", Terrain: "官道", Risk: 18, Distance: 2},
			{From: "canal", To: "dockyard", Terrain: "水路", Risk: 18, Distance: 2},
		},
		Factions: []StrategicFaction{
			{ID: "court", Name: "大胤朝廷", Ruler: "皇帝", Relation: 100, Threat: 0, Strategy: "一统山河", CapitalCityID: "capital", Color: "#d7a84f", IsPlayer: true},
			{ID: "beidi", Name: "北狄诸部", Ruler: "阿史那乌勒", Relation: 32, Threat: 72, Strategy: "骑兵侵攻", CapitalCityID: "snow-ridge", Color: "#7ea8d8"},
			{ID: "remnant", Name: "旧朝残部", Ruler: "宇文续", Relation: 24, Threat: 54, Strategy: "复辟骚扰", CapitalCityID: "jade-pass", Color: "#9b7a56"},
			{ID: "rebels", Name: "流寇叛军", Ruler: "赤眉渠帅", Relation: 10, Threat: 48, Strategy: "逐乱而起", CapitalCityID: "river-east", Color: "#9d3c31"},
			{ID: "nanling", Name: "南岭盟寨", Ruler: "火藤大首领", Relation: 42, Threat: 38, Strategy: "观望互市", CapitalCityID: "nanling", Color: "#5f9f74"},
			{ID: "haiguo", Name: "东海诸岛", Ruler: "海国摄政", Relation: 48, Threat: 34, Strategy: "海贸试探", CapitalCityID: "east-sea", Color: "#4f8ea8"},
		},
		Armies: []ArmyGroup{
			{ID: "imperial-guard", Name: "禁军右营", FactionID: "court", Location: "capital", CommanderID: "gu", Troops: 16000, Grain: 70, Morale: 68, Training: 62, Status: "驻防"},
			{ID: "northern-banner", Name: "北府军", FactionID: "court", Location: "north", CommanderID: "huo", Troops: 18000, Grain: 54, Morale: 66, Training: 70, Status: "驻防"},
			{ID: "beidi-vanguard", Name: "黑毡前锋", FactionID: "beidi", Location: "snow-ridge", Troops: 22000, Grain: 58, Morale: 72, Training: 68, Status: "压境"},
			{ID: "jade-remnant", Name: "玉门旧军", FactionID: "remnant", Location: "jade-pass", Troops: 12000, Grain: 42, Morale: 56, Training: 52, Status: "割据"},
			{ID: "nanling-guard", Name: "南岭藤甲", FactionID: "nanling", Location: "nanling", Troops: 11000, Grain: 48, Morale: 60, Training: 58, Status: "观望"},
		},
	}
	applyStrategicDynastySetup(&state, dynastyID)
	return state
}

func applyStrategicDynastySetup(state *StrategicState, dynastyID string) {
	switch dynastyID {
	case "xuanshuo":
		state.adjustFactionThreat("beidi", 14)
		state.adjustCity("north", 0, 0, -8, -6, 14)
		state.adjustArmy("beidi-vanguard", 4000, 8, 6)
	case "chengping":
		state.adjustFactionThreat("rebels", 16)
		state.setCityOwner("river-east", "rebels")
		state.adjustCity("river-east", -8, -10, -6, -16, 22)
		state.Armies = append(state.Armies, ArmyGroup{ID: "red-brow-host", Name: "赤眉流军", FactionID: "rebels", Location: "river-east", Troops: 15000, Grain: 36, Morale: 58, Training: 42, Status: "啸聚"})
	case "jingyao":
		state.adjustFactionThreat("haiguo", 8)
		state.adjustCity("south", 10, 10, 0, 8, -6)
		state.adjustCity("dockyard", 8, 12, 0, 4, -4)
	case "dayin":
		state.adjustFactionThreat("remnant", 10)
		state.adjustCity("west", 0, -4, 4, -4, 4)
	}
}

func (s *GameState) ensureStrategicSystems() {
	if len(s.Strategy.Cities) == 0 || len(s.Strategy.Roads) == 0 || len(s.Strategy.Factions) == 0 || len(s.Strategy.Armies) == 0 {
		s.Strategy = startingStrategicState(s.Dynasty.ID)
	}
}

func (m StrategicState) City(id string) (StrategicCity, bool) {
	for _, city := range m.Cities {
		if city.ID == id {
			return city, true
		}
	}
	return StrategicCity{}, false
}

func (m StrategicState) Faction(id string) (StrategicFaction, bool) {
	for _, faction := range m.Factions {
		if faction.ID == id {
			return faction, true
		}
	}
	return StrategicFaction{}, false
}

func (m StrategicState) Army(id string) (ArmyGroup, bool) {
	for _, army := range m.Armies {
		if army.ID == id {
			return army, true
		}
	}
	return ArmyGroup{}, false
}

func (m StrategicState) Neighbors(cityID string) []string {
	neighbors := []string{}
	for _, road := range m.Roads {
		if road.From == cityID {
			neighbors = append(neighbors, road.To)
		}
		if road.To == cityID {
			neighbors = append(neighbors, road.From)
		}
	}
	return neighbors
}

func (m StrategicState) AreAdjacent(a, b string) bool {
	for _, neighbor := range m.Neighbors(a) {
		if neighbor == b {
			return true
		}
	}
	return false
}

func (m *StrategicState) cityIndex(id string) (int, bool) {
	for i, city := range m.Cities {
		if city.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (m *StrategicState) factionIndex(id string) (int, bool) {
	for i, faction := range m.Factions {
		if faction.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (m *StrategicState) armyIndex(id string) (int, bool) {
	for i, army := range m.Armies {
		if army.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (m *StrategicState) adjustCity(id string, commerce, agriculture, defense, order, disaster int) {
	i, ok := m.cityIndex(id)
	if !ok {
		return
	}
	city := m.Cities[i]
	city.Commerce = clamp(city.Commerce+commerce, 0, 120)
	city.Agriculture = clamp(city.Agriculture+agriculture, 0, 120)
	city.Defense = clamp(city.Defense+defense, 0, 120)
	city.Order = clamp(city.Order+order, 0, 100)
	city.Disaster = clamp(city.Disaster+disaster, 0, 100)
	m.Cities[i] = city
}

func (m *StrategicState) adjustArmy(id string, troops, morale, training int) {
	i, ok := m.armyIndex(id)
	if !ok {
		return
	}
	army := m.Armies[i]
	army.Troops = max(0, army.Troops+troops)
	army.Morale = clamp(army.Morale+morale, 0, 100)
	army.Training = clamp(army.Training+training, 0, 100)
	m.Armies[i] = army
}

func (m *StrategicState) adjustFactionThreat(id string, delta int) {
	i, ok := m.factionIndex(id)
	if !ok {
		return
	}
	m.Factions[i].Threat = clamp(m.Factions[i].Threat+delta, 0, 100)
}

func (m *StrategicState) setCityOwner(cityID, ownerID string) {
	i, ok := m.cityIndex(cityID)
	if !ok {
		return
	}
	m.Cities[i].OwnerID = ownerID
	m.Cities[i].Front = true
}
