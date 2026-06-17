package game

func NewGame(scenarioID, playerID string) *GameState {
	if scenarioID == "" {
		scenarioID = "dongzhuo"
	}
	if playerID == "" {
		playerID = "caocao"
	}

	state := &GameState{
		ScenarioID: scenarioID,
		PlayerID:   playerID,
		Date:       Date{Year: 190, Month: 1},
		Rulers: []Ruler{
			{ID: "caocao", Name: "曹操", Character: "大义", Color: "#9b2f2f"},
			{ID: "liubei", Name: "刘备", Character: "仁德", Color: "#2f7d57"},
			{ID: "sunquan", Name: "孙权", Character: "稳健", Color: "#2e6f9e"},
			{ID: "dongzhuo", Name: "董卓", Character: "狂人", Color: "#7a4d9f"},
			{ID: "neutral", Name: "空城", Character: "无", Color: "#7f7a68"},
		},
		Cities: []City{
			newCity("luoyang", "洛阳", 5, 2, "dongzhuo", 940, 820, 130000, 1600, 1800, 600),
			newCity("chang-an", "长安", 3, 2, "dongzhuo", 900, 760, 115000, 1400, 1500, 700),
			newCity("xuchang", "许昌", 6, 4, "caocao", 800, 400, 90000, 1000, 1200, 0),
			newCity("chenliu", "陈留", 7, 3, "caocao", 720, 360, 76000, 820, 1000, 250),
			newCity("ye", "邺城", 6, 1, "caocao", 860, 500, 96000, 1200, 1100, 300),
			newCity("pingyuan", "平原", 8, 2, "liubei", 650, 340, 61000, 700, 900, 200),
			newCity("xiapi", "下邳", 8, 5, "liubei", 720, 460, 78000, 860, 980, 180),
			newCity("jianye", "建业", 9, 7, "sunquan", 780, 620, 100000, 1500, 1300, 380),
			newCity("wujun", "吴郡", 10, 8, "sunquan", 700, 680, 87000, 1400, 1200, 300),
			newCity("jiangxia", "江夏", 7, 6, "neutral", 660, 420, 62000, 500, 700, 100),
			newCity("chengdu", "成都", 2, 7, "neutral", 920, 540, 110000, 900, 1600, 400),
			newCity("jingzhou", "荆州", 5, 6, "neutral", 840, 580, 97000, 1000, 1400, 450),
		},
		Generals: []General{
			{ID: "cao-cao", Name: "曹操", OwnerID: "caocao", CityID: "xuchang", Level: 8, Force: 72, Intellect: 92, Loyalty: 100, Stamina: 96, Soldiers: 1000, ArmsType: "骑兵"},
			{ID: "xiahou-dun", Name: "夏侯惇", OwnerID: "caocao", CityID: "chenliu", Level: 6, Force: 89, Intellect: 61, Loyalty: 95, Stamina: 86, Soldiers: 850, ArmsType: "步兵"},
			{ID: "liu-bei", Name: "刘备", OwnerID: "liubei", CityID: "pingyuan", Level: 7, Force: 68, Intellect: 78, Loyalty: 100, Stamina: 92, Soldiers: 900, ArmsType: "步兵"},
			{ID: "guan-yu", Name: "关羽", OwnerID: "liubei", CityID: "xiapi", Level: 8, Force: 97, Intellect: 74, Loyalty: 100, Stamina: 90, Soldiers: 980, ArmsType: "骑兵"},
			{ID: "sun-quan", Name: "孙权", OwnerID: "sunquan", CityID: "jianye", Level: 6, Force: 70, Intellect: 82, Loyalty: 100, Stamina: 88, Soldiers: 880, ArmsType: "水军"},
			{ID: "zhou-yu", Name: "周瑜", OwnerID: "sunquan", CityID: "wujun", Level: 8, Force: 79, Intellect: 96, Loyalty: 96, Stamina: 91, Soldiers: 920, ArmsType: "水军"},
			{ID: "dong-zhuo", Name: "董卓", OwnerID: "dongzhuo", CityID: "luoyang", Level: 7, Force: 86, Intellect: 58, Loyalty: 100, Stamina: 90, Soldiers: 1100, ArmsType: "骑兵"},
			{ID: "lv-bu", Name: "吕布", OwnerID: "dongzhuo", CityID: "chang-an", Level: 10, Force: 100, Intellect: 42, Loyalty: 70, Stamina: 96, Soldiers: 1200, ArmsType: "骑兵"},
		},
		Routes: []Route{
			{From: "chang-an", To: "luoyang"},
			{From: "luoyang", To: "ye"},
			{From: "luoyang", To: "xuchang"},
			{From: "xuchang", To: "chenliu"},
			{From: "chenliu", To: "pingyuan"},
			{From: "pingyuan", To: "xiapi"},
			{From: "xuchang", To: "jingzhou"},
			{From: "jingzhou", To: "jiangxia"},
			{From: "jiangxia", To: "jianye"},
			{From: "jianye", To: "wujun"},
			{From: "jingzhou", To: "chengdu"},
		},
		Log: []string{"新君登基，群雄并起。"},
	}
	return state
}

func newCity(id, name string, x, y int, ownerID string, farming, commerce, population, money, food, garrison int) City {
	return City{
		ID:              id,
		Name:            name,
		X:               x,
		Y:               y,
		OwnerID:         ownerID,
		State:           CityStateNormal,
		FarmingLimit:    1000,
		Farming:         farming,
		CommerceLimit:   1000,
		Commerce:        commerce,
		PeopleDevotion:  72,
		AvoidCalamity:   45,
		PopulationLimit: population + 30000,
		Population:      population,
		Money:           money,
		Food:            food,
		Garrison:        garrison,
	}
}
