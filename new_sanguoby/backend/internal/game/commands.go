package game

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidCommand   = errors.New("invalid command")
	ErrCityNotPlayable  = errors.New("city does not belong to player")
	ErrGeneralNotReady  = errors.New("general cannot execute command")
	ErrInsufficientFund = errors.New("city resources are insufficient")
)

type CommandSpec struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

var CommandSpecs = []CommandSpec{
	{ID: "assart", Name: "开垦", Category: "内政"},
	{ID: "commerce", Name: "招商", Category: "内政"},
	{ID: "search", Name: "搜寻", Category: "内政"},
	{ID: "govern", Name: "治理", Category: "内政"},
	{ID: "inspect", Name: "出巡", Category: "内政"},
	{ID: "surrender", Name: "招降", Category: "内政"},
	{ID: "kill", Name: "处斩", Category: "内政"},
	{ID: "banish", Name: "流放", Category: "内政"},
	{ID: "largess", Name: "赏赐", Category: "内政"},
	{ID: "confiscate", Name: "没收", Category: "内政"},
	{ID: "exchange", Name: "交易", Category: "内政"},
	{ID: "treat", Name: "宴请", Category: "内政"},
	{ID: "transportation", Name: "输送", Category: "内政"},
	{ID: "move", Name: "移动", Category: "内政"},
	{ID: "alienate", Name: "离间", Category: "外交"},
	{ID: "canvass", Name: "招揽", Category: "外交"},
	{ID: "counterespionage", Name: "策反", Category: "外交"},
	{ID: "realienate", Name: "反间", Category: "外交"},
	{ID: "induce", Name: "劝降", Category: "外交"},
	{ID: "reconnoitre", Name: "侦察", Category: "军备"},
	{ID: "conscription", Name: "征兵", Category: "军备"},
	{ID: "distribute", Name: "分配", Category: "军备"},
	{ID: "depredate", Name: "掠夺", Category: "军备"},
	{ID: "battle", Name: "出征", Category: "军备"},
}

func (s *GameState) ApplyCommand(cityID, generalID, commandID string) error {
	return s.ApplyCommandDetailed(cityID, generalID, commandID, "", "")
}

func (s *GameState) ApplyCommandWithTarget(cityID, generalID, commandID, targetCityID string) error {
	return s.ApplyCommandDetailed(cityID, generalID, commandID, targetCityID, "")
}

func (s *GameState) ApplyCommandDetailed(cityID, generalID, commandID, targetCityID, targetGeneralID string) error {
	city := s.findCity(cityID)
	if city == nil {
		return fmt.Errorf("%w: city %s", ErrInvalidCommand, cityID)
	}
	if city.OwnerID != s.PlayerID {
		return ErrCityNotPlayable
	}

	general := s.findGeneral(generalID)
	if general == nil || general.CityID != city.ID || general.OwnerID != s.PlayerID || general.Captive {
		return ErrGeneralNotReady
	}

	if _, ok := commandSpec(commandID); !ok {
		return fmt.Errorf("%w: command %s", ErrInvalidCommand, commandID)
	}

	cost := commandCost(commandID)
	if general.Stamina < cost.stamina {
		return ErrGeneralNotReady
	}
	if city.Money < cost.money || city.Food < cost.food {
		return ErrInsufficientFund
	}
	if commandID == "conscription" {
		recruits := 120 + general.Force*8
		if city.Population < recruits*2 {
			return ErrInsufficientFund
		}
	}
	if commandID == "move" || commandID == "transportation" {
		if _, err := s.friendlyAdjacentCity(city, targetCityID); err != nil {
			return err
		}
	}

	var targetGeneral *General
	switch commandID {
	case "kill":
		var err error
		targetGeneral, err = s.captiveTargetGeneral(city.ID, targetGeneralID, s.PlayerID)
		if err != nil {
			return err
		}
	case "banish":
		var err error
		targetGeneral, err = s.banishTargetGeneral(city.ID, targetGeneralID, s.PlayerID)
		if err != nil {
			return err
		}
	case "largess", "confiscate", "treat":
		var err error
		targetGeneral, err = s.activeTargetGeneral(city.ID, targetGeneralID, s.PlayerID, general)
		if err != nil {
			return err
		}
	}

	city.Money -= cost.money
	city.Food -= cost.food
	general.Stamina -= cost.stamina

	switch commandID {
	case "assart":
		gain := 10 + general.Intellect/2 + general.Level*2
		before := city.Farming
		city.Farming = minInt(city.FarmingLimit, city.Farming+gain)
		s.prependLog(fmt.Sprintf("%s 在 %s 开垦，农业 +%d。", general.Name, city.Name, city.Farming-before))
	case "commerce":
		gain := 10 + general.Intellect/2 + general.Level*2
		before := city.Commerce
		city.Commerce = minInt(city.CommerceLimit, city.Commerce+gain)
		s.prependLog(fmt.Sprintf("%s 在 %s 招商，商业 +%d。", general.Name, city.Name, city.Commerce-before))
	case "search":
		bonus := 0
		if general.Intellect >= 80 {
			bonus = 80 + general.Intellect
			city.Money = minInt(maxResourceValue, city.Money+bonus)
		}
		if bonus > 0 {
			s.prependLog(fmt.Sprintf("%s 搜寻 %s，寻得金%d。", general.Name, city.Name, bonus))
		} else {
			s.prependLog(fmt.Sprintf("%s 搜寻 %s，暂无线索。", general.Name, city.Name))
		}
	case "govern":
		devotionGain := 4 + general.Intellect/12
		calamityGain := 3 + general.Intellect/18
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+devotionGain)
		city.AvoidCalamity = minInt(100, city.AvoidCalamity+calamityGain)
		s.prependLog(fmt.Sprintf("%s 治理 %s，民忠与防灾上升。", general.Name, city.Name))
	case "inspect":
		gain := 5 + general.Force/18
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+gain)
		s.prependLog(fmt.Sprintf("%s 出巡 %s，民忠 +%d。", general.Name, city.Name, gain))
	case "surrender":
		s.applySurrender(general, city)
	case "kill":
		targetName := targetGeneral.Name
		targetGeneral.OwnerID = ""
		targetGeneral.CityID = ""
		targetGeneral.Captive = false
		targetGeneral.Soldiers = 0
		targetGeneral.Stamina = 0
		city.PeopleDevotion = maxInt(0, city.PeopleDevotion-4)
		city.AvoidCalamity = minInt(100, city.AvoidCalamity+2)
		s.prependLog(fmt.Sprintf("%s 在 %s 处斩俘虏 %s，民忠下降，防灾略升。", general.Name, city.Name, targetName))
	case "banish":
		targetName := targetGeneral.Name
		destination := s.banishDestination(city.ID, s.PlayerID)
		targetGeneral.OwnerID = "neutral"
		targetGeneral.Captive = false
		targetGeneral.Soldiers = 0
		targetGeneral.Stamina = 0
		targetGeneral.Loyalty = minInt(targetGeneral.Loyalty, 45)
		if targetGeneral.Loyalty <= 0 {
			targetGeneral.Loyalty = 30
		}
		if destination != nil {
			targetGeneral.CityID = destination.ID
		} else {
			targetGeneral.CityID = ""
		}
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+3)
		city.Money = maxInt(0, city.Money-20)
		if destination != nil {
			s.prependLog(fmt.Sprintf("%s 将 %s 流放至 %s，民忠 +3。", general.Name, targetName, destination.Name))
		} else {
			s.prependLog(fmt.Sprintf("%s 将 %s 逐出 %s，民忠 +3。", general.Name, targetName, city.Name))
		}
	case "largess":
		targetGeneral.Loyalty = minInt(100, targetGeneral.Loyalty+8)
		s.prependLog(fmt.Sprintf("%s 赏赐 %s，忠诚提升至 %d。", general.Name, targetGeneral.Name, targetGeneral.Loyalty))
	case "confiscate":
		gain := 80 + general.Intellect/2
		city.Money = minInt(maxResourceValue, city.Money+gain)
		city.PeopleDevotion = maxInt(0, city.PeopleDevotion-5)
		targetGeneral.Loyalty = maxInt(0, targetGeneral.Loyalty-20)
		s.prependLog(fmt.Sprintf("%s 没收 %s 财货，金 +%d，%s 忠诚下降。", general.Name, targetGeneral.Name, gain, targetGeneral.Name))
	case "exchange":
		s.exchangeCityResources(general, city)
	case "treat":
		targetGeneral.Stamina = minInt(100, targetGeneral.Stamina+16)
		targetGeneral.Loyalty = minInt(100, targetGeneral.Loyalty+2)
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+1)
		s.prependLog(fmt.Sprintf("%s 在 %s 宴请 %s，体力恢复，忠诚微升。", general.Name, city.Name, targetGeneral.Name))
	case "transportation":
		target, _ := s.friendlyAdjacentCity(city, targetCityID)
		s.transportSupplies(general, city, target)
	case "move":
		target, _ := s.friendlyAdjacentCity(city, targetCityID)
		general.CityID = target.ID
		s.prependLog(fmt.Sprintf("%s 从 %s 移动至 %s。", general.Name, city.Name, target.Name))
	case "conscription":
		recruits := 120 + general.Force*8
		city.Population -= recruits * 2
		general.Soldiers += recruits
		s.prependLog(fmt.Sprintf("%s 在 %s 征兵，兵力 +%d。", general.Name, city.Name, recruits))
	case "reconnoitre":
		enemyCities := s.adjacentCitiesByOwner(city.ID, func(ownerID string) bool {
			return ownerID != city.OwnerID
		})
		if len(enemyCities) == 0 {
			s.prependLog(fmt.Sprintf("%s 侦察 %s 周边，未见敌踪。", general.Name, city.Name))
		} else {
			s.prependLog(fmt.Sprintf("%s 侦察得报：%s 附近有 %d 座敌城。", general.Name, city.Name, len(enemyCities)))
		}
	case "alienate":
		s.alienateEnemyGeneral(general, city)
	case "canvass":
		s.canvassEnemyGeneral(general, city)
	case "counterespionage":
		general.Loyalty = minInt(100, general.Loyalty+4)
		city.AvoidCalamity = minInt(100, city.AvoidCalamity+4)
		s.prependLog(fmt.Sprintf("%s 在 %s 策反敌探，忠诚与防灾上升。", general.Name, city.Name))
	case "realienate":
		for i := range s.Generals {
			if s.Generals[i].OwnerID == s.PlayerID && s.Generals[i].CityID == city.ID && !s.Generals[i].Captive {
				s.Generals[i].Loyalty = minInt(100, s.Generals[i].Loyalty+3)
			}
		}
		s.prependLog(fmt.Sprintf("%s 在 %s 反间布防，城中诸将忠诚上升。", general.Name, city.Name))
	case "induce":
		s.induceAdjacentCity(general, city)
	case "distribute":
		moved := minInt(city.Garrison, 300)
		if moved > 0 {
			city.Garrison -= moved
			general.Soldiers += moved
		}
		s.prependLog(fmt.Sprintf("%s 在 %s 分配兵力。", general.Name, city.Name))
	case "depredate":
		gain := 50 + general.Force
		city.Food = minInt(maxResourceValue, city.Food+gain)
		city.PeopleDevotion = maxInt(0, city.PeopleDevotion-8)
		s.prependLog(fmt.Sprintf("%s 掠夺 %s，粮食 +%d，民忠下降。", general.Name, city.Name, gain))
	case "battle":
		// 出征需要选择目标城池, 由独立的 ApplyBattle 处理; 此处仅作整军提示。
		s.prependLog(fmt.Sprintf("%s 从 %s 整军待发，请选择进攻目标。", general.Name, city.Name))
	default:
		return fmt.Errorf("%w: command %s", ErrInvalidCommand, commandID)
	}
	return nil
}

type commandCostValue struct {
	stamina int
	money   int
	food    int
}

func commandCost(commandID string) commandCostValue {
	switch commandID {
	case "search", "kill", "banish", "exchange", "transportation", "move", "distribute", "depredate", "battle":
		return commandCostValue{stamina: 4}
	case "conscription":
		return commandCostValue{stamina: 4, money: 1}
	case "reconnoitre":
		return commandCostValue{stamina: 4, money: 20}
	case "surrender", "treat":
		return commandCostValue{stamina: 0, money: 100}
	case "assart", "commerce", "govern", "inspect", "alienate", "canvass", "counterespionage", "realienate", "induce":
		return commandCostValue{stamina: 4, money: 50}
	case "largess":
		return commandCostValue{stamina: 4, money: 100}
	default:
		return commandCostValue{stamina: 4}
	}
}

func commandSpec(commandID string) (CommandSpec, bool) {
	for _, spec := range CommandSpecs {
		if spec.ID == commandID {
			return spec, true
		}
	}
	return CommandSpec{}, false
}

func (s *GameState) findCity(id string) *City {
	for i := range s.Cities {
		if s.Cities[i].ID == id {
			return &s.Cities[i]
		}
	}
	return nil
}

func (s *GameState) findGeneral(id string) *General {
	for i := range s.Generals {
		if s.Generals[i].ID == id {
			return &s.Generals[i]
		}
	}
	return nil
}

func (s *GameState) prependLog(entry string) {
	s.Log = append([]string{entry}, s.Log...)
	if len(s.Log) > 8 {
		s.Log = s.Log[:8]
	}
}

func (s *GameState) friendlyAdjacentCity(city *City, targetCityID string) (*City, error) {
	if targetCityID == "" {
		return nil, fmt.Errorf("%w: target city required", ErrInvalidCommand)
	}
	target := s.findCity(targetCityID)
	if target == nil {
		return nil, fmt.Errorf("%w: target city %s", ErrInvalidCommand, targetCityID)
	}
	if target.ID == city.ID {
		return nil, fmt.Errorf("%w: target city is the origin", ErrInvalidCommand)
	}
	if target.OwnerID != city.OwnerID {
		return nil, fmt.Errorf("%w: target city is not friendly", ErrInvalidCommand)
	}
	if !s.isAdjacent(city.ID, target.ID) {
		return nil, fmt.Errorf("%w: target city is not adjacent", ErrInvalidCommand)
	}
	return target, nil
}

func (s *GameState) cityGeneral(cityID, generalID string) *General {
	if generalID == "" {
		return nil
	}
	general := s.findGeneral(generalID)
	if general == nil || general.CityID != cityID {
		return nil
	}
	return general
}

func (s *GameState) firstCaptiveInCity(cityID, ownerID string) *General {
	for i := range s.Generals {
		general := &s.Generals[i]
		if general.CityID == cityID && general.OwnerID == ownerID && general.Captive {
			return general
		}
	}
	return nil
}

func (s *GameState) captiveTargetGeneral(cityID, targetGeneralID, ownerID string) (*General, error) {
	if targetGeneralID == "" {
		if target := s.firstCaptiveInCity(cityID, ownerID); target != nil {
			return target, nil
		}
		return nil, fmt.Errorf("%w: captive target required", ErrInvalidCommand)
	}
	target := s.cityGeneral(cityID, targetGeneralID)
	if target == nil || target.OwnerID != ownerID || !target.Captive {
		return nil, fmt.Errorf("%w: captive target %s", ErrInvalidCommand, targetGeneralID)
	}
	return target, nil
}

func (s *GameState) banishTargetGeneral(cityID, targetGeneralID, ownerID string) (*General, error) {
	if targetGeneralID == "" {
		if target := s.firstCaptiveInCity(cityID, ownerID); target != nil {
			return target, nil
		}
		return nil, fmt.Errorf("%w: general target required", ErrInvalidCommand)
	}
	target := s.cityGeneral(cityID, targetGeneralID)
	if target == nil || target.OwnerID != ownerID {
		return nil, fmt.Errorf("%w: general target %s", ErrInvalidCommand, targetGeneralID)
	}
	return target, nil
}

func (s *GameState) activeTargetGeneral(cityID, targetGeneralID, ownerID string, fallback *General) (*General, error) {
	if targetGeneralID == "" {
		return fallback, nil
	}
	target := s.cityGeneral(cityID, targetGeneralID)
	if target == nil || target.OwnerID != ownerID || target.Captive {
		return nil, fmt.Errorf("%w: general target %s", ErrInvalidCommand, targetGeneralID)
	}
	return target, nil
}

func (s *GameState) banishDestination(originCityID, playerID string) *City {
	for i := range s.Cities {
		city := &s.Cities[i]
		if city.ID != originCityID && city.OwnerID == "neutral" {
			return city
		}
	}
	for i := range s.Cities {
		city := &s.Cities[i]
		if city.ID != originCityID && city.OwnerID != playerID {
			return city
		}
	}
	return nil
}

func (s *GameState) adjacentCitiesByOwner(cityID string, match func(ownerID string) bool) []*City {
	var cities []*City
	for _, id := range s.AdjacentCityIDs(cityID) {
		city := s.findCity(id)
		if city != nil && match(city.OwnerID) {
			cities = append(cities, city)
		}
	}
	return cities
}

func (s *GameState) exchangeCityResources(general *General, city *City) {
	if city.Food < city.Population/120 && city.Money >= 500 {
		food := minInt(400, city.Money/5)
		city.Money -= food * 5
		city.Food = minInt(maxResourceValue, city.Food+food)
		s.prependLog(fmt.Sprintf("%s 在 %s 买入粮%d。", general.Name, city.Name, food))
		return
	}
	if city.Food >= 200 {
		food := minInt(300, city.Food/5)
		city.Food -= food
		city.Money = minInt(maxResourceValue, city.Money+food*2)
		s.prependLog(fmt.Sprintf("%s 在 %s 卖粮%d，得金%d。", general.Name, city.Name, food, food*2))
		return
	}
	s.prependLog(fmt.Sprintf("%s 在 %s 交易，无可调度资源。", general.Name, city.Name))
}

func (s *GameState) transportSupplies(general *General, from, target *City) {
	food := minInt(500, from.Food/5)
	money := minInt(500, from.Money/5)
	arms := minInt(300, from.Garrison/5)
	if food == 0 && money == 0 && arms == 0 {
		s.prependLog(fmt.Sprintf("%s 整理输送队，但 %s 暂无余粮余金。", general.Name, from.Name))
		return
	}
	from.Food -= food
	from.Money -= money
	from.Garrison -= arms
	target.Food = minInt(maxResourceValue, target.Food+food)
	target.Money = minInt(maxResourceValue, target.Money+money)
	target.Garrison += arms
	s.prependLog(fmt.Sprintf("%s 从 %s 输送至 %s：粮%d 金%d 后备%d。", general.Name, from.Name, target.Name, food, money, arms))
}

func (s *GameState) applySurrender(general *General, city *City) {
	targets := s.adjacentCitiesByOwner(city.ID, func(ownerID string) bool {
		return ownerID != city.OwnerID && ownerID != "neutral"
	})
	for _, target := range targets {
		defenders := s.cityTroops(target.ID)
		if defenders <= general.Intellect*8 {
			s.transferCityPeacefully(target, city.OwnerID, 65+general.Intellect/10)
			target.PeopleDevotion = maxInt(0, target.PeopleDevotion-12)
			target.AvoidCalamity = maxInt(0, target.AvoidCalamity-4)
			s.prependLog(fmt.Sprintf("%s 招降成功，%s 归顺。", general.Name, target.Name))
			return
		}
	}
	s.prependLog(fmt.Sprintf("%s 招降邻城，敌军暂不动摇。", general.Name))
}

func (s *GameState) alienateEnemyGeneral(general *General, city *City) {
	target := s.findAdjacentEnemyGeneral(city)
	if target == nil {
		s.prependLog(fmt.Sprintf("%s 离间无门，%s 附近未见敌将。", general.Name, city.Name))
		return
	}
	loss := 6 + general.Intellect/20
	target.Loyalty = maxInt(0, target.Loyalty-loss)
	s.prependLog(fmt.Sprintf("%s 离间 %s，敌将忠诚 -%d。", general.Name, target.Name, loss))
}

func (s *GameState) canvassEnemyGeneral(general *General, city *City) {
	target := s.findAdjacentEnemyGeneral(city)
	if target == nil {
		s.prependLog(fmt.Sprintf("%s 招揽无门，%s 附近未见敌将。", general.Name, city.Name))
		return
	}
	score := general.Intellect + (100 - target.Loyalty)
	if score >= 120 {
		target.OwnerID = city.OwnerID
		target.CityID = city.ID
		target.Captive = false
		target.Loyalty = minInt(100, 70+general.Intellect/10)
		s.prependLog(fmt.Sprintf("%s 招揽成功，%s 来投。", general.Name, target.Name))
		return
	}
	target.Loyalty = maxInt(0, target.Loyalty-4)
	s.prependLog(fmt.Sprintf("%s 游说 %s，虽未归顺，忠诚动摇。", general.Name, target.Name))
}

func (s *GameState) induceAdjacentCity(general *General, city *City) {
	targets := s.adjacentCitiesByOwner(city.ID, func(ownerID string) bool {
		return ownerID != city.OwnerID && ownerID != "neutral"
	})
	for _, target := range targets {
		if target.PeopleDevotion+target.AvoidCalamity < general.Intellect {
			s.transferCityPeacefully(target, city.OwnerID, 60+general.Intellect/12)
			target.PeopleDevotion = maxInt(0, target.PeopleDevotion-8)
			s.prependLog(fmt.Sprintf("%s 劝降成功，%s 易帜。", general.Name, target.Name))
			return
		}
	}
	s.prependLog(fmt.Sprintf("%s 劝降邻城，守军仍旧观望。", general.Name))
}

func (s *GameState) findAdjacentEnemyGeneral(city *City) *General {
	adjacent := map[string]bool{}
	for _, id := range s.AdjacentCityIDs(city.ID) {
		adjacent[id] = true
	}
	var target *General
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.Captive || g.OwnerID == city.OwnerID || !adjacent[g.CityID] {
			continue
		}
		if target == nil || g.Loyalty < target.Loyalty {
			target = g
		}
	}
	return target
}

func (s *GameState) transferCityPeacefully(city *City, ownerID string, loyalty int) {
	city.OwnerID = ownerID
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID {
			s.Generals[i].OwnerID = ownerID
			s.Generals[i].Captive = false
			s.Generals[i].Loyalty = minInt(100, maxInt(s.Generals[i].Loyalty, loyalty))
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
