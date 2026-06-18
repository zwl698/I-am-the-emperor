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
	city := s.findCity(cityID)
	if city == nil {
		return fmt.Errorf("%w: city %s", ErrInvalidCommand, cityID)
	}
	if city.OwnerID != s.PlayerID {
		return ErrCityNotPlayable
	}

	general := s.findGeneral(generalID)
	if general == nil || general.CityID != city.ID || general.OwnerID != s.PlayerID {
		return ErrGeneralNotReady
	}

	spec, ok := commandSpec(commandID)
	if !ok {
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
	case "conscription":
		recruits := 120 + general.Force*8
		city.Population -= recruits * 2
		general.Soldiers += recruits
		s.prependLog(fmt.Sprintf("%s 在 %s 征兵，兵力 +%d。", general.Name, city.Name, recruits))
	case "reconnoitre":
		s.prependLog(fmt.Sprintf("%s 侦察 %s 周边军情。", general.Name, city.Name))
	case "surrender", "alienate", "canvass", "counterespionage", "realienate", "induce":
		s.prependLog(fmt.Sprintf("%s 执行%s，等待对方回应。", general.Name, spec.Name))
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
	case "reconnoitre":
		return commandCostValue{stamina: 10, money: 20}
	case "conscription":
		return commandCostValue{stamina: 12, money: 100, food: 40}
	case "battle":
		return commandCostValue{stamina: 20, food: 100}
	case "assart", "commerce", "govern", "inspect":
		return commandCostValue{stamina: 8, money: 50}
	default:
		return commandCostValue{stamina: 8}
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
