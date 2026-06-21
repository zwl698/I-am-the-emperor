package game

import (
	"fmt"
	"math/rand"
)

const (
	maxResourceValue = 30000
	staminaRenew     = 4
	staminaTreat     = 50 // 宴请恢复体力

	// 城市状态常量对应C源码 STATE_NORMAL/STATE_FAMINE 等
)

// 君主性格内政/外交概率表 (KingTacticOddsIH/KingTacticOddsD)
// 索引: 0=冒进 1=狂人 2=奸诈 3=大义 4=和平
var tacticOddsIH = [5]int{10, 20, 30, 40, 50} // 内政+协调概率
var tacticOddsD = [5]int{20, 40, 70, 70, 80}  // 外交概率

// 武将性格离间概率 (ALIENATE_*)
// 索引: 0=卤莽 1=怕死 2=贪财 3=大志 4=忠义
var alienateOdds = [5]int{50, 30, 40, 30, 5}

// 武将性格招揽概率 (CANVASS_*)
var canvassOdds = [5]int{15, 40, 30, 20, 5}

// 武将性格策反概率 (COUNTERESPIONAGE_*)
var counterespionageOdds = [5]int{30, 10, 20, 60, 5}

// 武将性格招降概率 (SURRENDER_*)
var surrenderOdds = [5]int{15, 60, 30, 20, 5}

// 兵种相克系数矩阵 SubduModu[attacker][defender]
// 兵种顺序: 0骑兵 1步兵 2弓兵 3水军 4极兵 5玄兵
var subdueModu = [6][6]float64{
	{1.0, 1.2, 0.8, 1.0, 0.7, 1.3}, // 骑兵
	{0.8, 1.0, 1.2, 1.0, 0.6, 1.2}, // 步兵
	{1.2, 0.8, 1.0, 1.0, 1.1, 1.2}, // 弓兵
	{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}, // 水军
	{1.1, 1.3, 0.9, 1.0, 1.0, 1.5}, // 极兵
	{0.6, 0.6, 0.6, 0.6, 0.6, 0.6}, // 玄兵
}

// 兵种攻击系数 AtkModulus
var atkModulus = [6]float64{1.0, 0.8, 0.9, 0.8, 1.3, 0.4}

// 兵种防御系数 DfModulus
var dfModulus = [6]float64{0.7, 1.2, 1.0, 1.1, 1.2, 0.6}

// EndStrategyPhase runs the full end-of-turn sequence triggered by the player's
// "策略结束": rival warlords act, then the calendar/economy is settled, then the
// victory state is evaluated. This is the entry point used by the server.
func (s *GameState) EndStrategyPhase() {
	currentDate := formatDate(s.Date)
	s.prependLog(currentDate + " 策略结束，诸侯开始行动。")
	s.RunEnemyTurns()
	s.AdvanceMonth()
	s.evaluateVictory()
}

// AdvanceMonth settles the calendar and city economy for one month.
// Mirrors the C source ConditionUpdate -> CitiesUpDataDate -> RandEvents -> EventStateDeal chain.
func (s *GameState) AdvanceMonth() {
	s.Date.Month++
	if s.Date.Month > 12 {
		s.Date.Year++
		s.Date.Month = 1
		// 年度更新：武将年龄增长
		s.personYearlyUpdate()
	}

	// 1. 体力恢复 (所有武将)，并强制兵力不超过统兵上限 (PlcArmsMax)
	for i := range s.Generals {
		s.Generals[i].Stamina = minInt(100, s.Generals[i].Stamina+staminaRenew)
		s.Generals[i].clampSoldiersToMax()
	}

	// 2. 收粮（6月和10月，对应 HarvestryFood）
	if s.Date.Month == 6 || s.Date.Month == 10 {
		for i := range s.Cities {
			city := &s.Cities[i]
			if city.OwnerID == "" {
				continue
			}
			city.Food = minInt(maxResourceValue, city.Food+city.Farming/4)
		}
	}

	// 3. 收税（每3个月，对应 RevenueMoney）
	if s.Date.Month%3 == 0 {
		for i := range s.Cities {
			city := &s.Cities[i]
			if city.OwnerID == "" {
				continue
			}
			if city.Money < maxResourceValue {
				city.Money += city.Commerce / 2
			}
		}
	}

	// 4. 城市逐月更新 (CitiesUpDataDate 核心循环)
	for i := range s.Cities {
		city := &s.Cities[i]
		if city.OwnerID == "" {
			continue
		}

		// 防灾每月递减（每3个月随机减1-4）
		if s.Date.Month%3 == 0 {
			rnd := 1 + rand.Intn(4)
			if city.AvoidCalamity > rnd {
				city.AvoidCalamity -= rnd
			}
		}

		// 金钱上限检查
		if city.Money > maxResourceValue {
			city.Money = maxResourceValue
		}

		// 人口增长
		city.Population = minInt(city.PopulationLimit, city.Population+50)

		// 粮草消耗（考虑城市状态对兵力的影响）
		s.consumeMonthlyFoodWithState(city)

		// 俘虏归顺检查：原归属与当前城市归属相同的俘虏自动归顺
		s.checkCaptiveAllegiance(city)
	}

	// 5. 随机事件 (RandEvents)
	s.processRandomEvents()

	// 6. 事件状态处理 (EventStateDeal) - 各状态对城市的持续影响
	s.processEventStates()

	// 7. 设置太守 (SetCitySatrap - 每个城最高智力者为太守)
	s.setCitySatraps()

	s.prependLog(formatDate(s.Date) + " 政令已结算，诸将体力恢复。")
}

// consumeMonthlyFoodWithState 计算城市每月粮草消耗，考虑城市状态对兵力的影响
// 对应 CitiesUpDataDate 中的粮耗计算
func (s *GameState) consumeMonthlyFoodWithState(city *City) {
	totalSoldiers := city.MothballArms + city.Garrison
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID && !s.Generals[i].Captive {
			as := s.Generals[i].Soldiers
			// 城市状态影响实际兵力
			switch city.State {
			case CityStateDrought, CityStateFlood:
				as = as - as/4 // 旱灾/水灾减25%兵力
			case CityStateRebellion:
				as /= 2 // 暴动减半兵力
			}
			totalSoldiers += as
		}
	}

	upkeep := totalSoldiers / 50
	if upkeep <= 0 {
		return
	}
	if city.Food > upkeep {
		city.Food -= upkeep
		return
	}

	// 饥荒：粮食耗尽
	city.Food = 0
	city.State = CityStateFamine
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID && !s.Generals[i].Captive {
			s.Generals[i].Soldiers /= 2
		}
	}
}

// checkCaptiveAllegiance 检查俘虏中是否有原属该城君主的人，自动归顺
func (s *GameState) checkCaptiveAllegiance(city *City) {
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.CityID == city.ID && g.Captive {
			// 如果俘虏的原归属与当前城主一致，自动归顺
			// OldBelong 在我们的模型中通过记录实现
			if g.OwnerID == city.OwnerID {
				g.Captive = false
			}
		}
	}
}

// processRandomEvents 对应 C 源码 RandEvents
// 正常城市有概率遭遇灾害，已有灾害的城市有概率恢复
func (s *GameState) processRandomEvents() {
	for i := range s.Cities {
		city := &s.Cities[i]
		if city.OwnerID == "" {
			continue
		}

		rnd := rand.Intn(100)
		switch city.State {
		case CityStateNormal:
			// 正常城市：随机值 > 防灾值时可能遭灾
			if rnd > city.AvoidCalamity {
				disaster := rand.Intn(5)
				switch disaster {
				case 0:
					city.State = CityStateDrought
					s.prependLog(fmt.Sprintf("%s 遭遇旱灾！", city.Name))
				case 1:
					city.State = CityStateFlood
					s.prependLog(fmt.Sprintf("%s 遭遇水灾！", city.Name))
				case 2:
					// 暴动需要额外判断民忠
					if rand.Intn(100) > city.PeopleDevotion {
						city.State = CityStateRebellion
						s.prependLog(fmt.Sprintf("%s 发生暴动！", city.Name))
					}
					// 3,4: 无灾害
				}
			}
		case CityStateFamine:
			// 饥荒：有粮则恢复正常
			if city.Food > 0 {
				city.State = CityStateNormal
			}
		case CityStateDrought, CityStateFlood:
			// 旱灾/水灾：防灾值够高时恢复
			if rnd < city.AvoidCalamity {
				oldState := city.State
				city.State = CityStateNormal
				s.prependLog(fmt.Sprintf("%s %s已恢复。", city.Name, oldState))
			}
		case CityStateRebellion:
			// 暴动：民忠够高时恢复
			if rnd < city.PeopleDevotion {
				city.State = CityStateNormal
				s.prependLog(fmt.Sprintf("%s 暴动已平息。", city.Name))
			}
		}
	}
}

// processEventStates 对应 C 源码 EventStateDeal
// 各种状态每回合对城市属性的影响
func (s *GameState) processEventStates() {
	for i := range s.Cities {
		city := &s.Cities[i]

		switch city.State {
		case CityStateNormal:
			continue
		case CityStateFamine:
			// 饥荒：商业-5%，民忠-5%，后备兵-50%，人口-25%
			city.Commerce -= city.Commerce / 20
			city.PeopleDevotion -= city.PeopleDevotion / 20
			city.MothballArms /= 2
			city.Population -= city.Population / 4
		case CityStateDrought:
			// 旱灾：粮食-5%，后备兵-25%，人口-25%
			city.Food -= city.Food / 20
			city.MothballArms -= city.MothballArms / 4
			city.Population -= city.Population / 4
		case CityStateFlood:
			// 水灾：粮食-5%，商业-10%，金钱-10%，后备兵-25%，人口-25%
			city.Food -= city.Food / 20
			city.Commerce -= city.Commerce / 10
			city.Money -= city.Money / 10
			city.MothballArms -= city.MothballArms / 4
			city.Population -= city.Population / 4
		case CityStateRebellion:
			// 暴动：粮食-5%，商业-5%，金钱-5%，民忠-10%，后备兵-50%
			city.Food -= city.Food / 20
			city.Commerce -= city.Commerce / 20
			city.Money -= city.Money / 20
			city.PeopleDevotion -= city.PeopleDevotion / 10
			city.MothballArms /= 2
		}

		// 所有非正常状态都会导致农业下降
		if city.State != CityStateNormal {
			city.Farming -= city.Farming / 20
		}

		// 属性下界保护
		if city.PeopleDevotion < 0 {
			city.PeopleDevotion = 0
		}
		if city.Commerce < 0 {
			city.Commerce = 0
		}
		if city.Money < 0 {
			city.Money = 0
		}
		if city.Food < 0 {
			city.Food = 0
		}
		if city.Population < 0 {
			city.Population = 0
		}
	}
}

// personYearlyUpdate 对应 PersonUpDatadate
// 每年1月武将年龄+1
func (s *GameState) personYearlyUpdate() {
	for i := range s.Generals {
		s.Generals[i].Age++
	}
}

// setCitySatraps 对应 SetCitySatrap
// 设置每个城中最高智力且归属该城君主的武将为太守
func (s *GameState) setCitySatraps() {
	for i := range s.Cities {
		city := &s.Cities[i]
		if city.OwnerID == "" {
			city.SatrapID = ""
			continue
		}

		bestIQ := -1
		bestID := ""
		for j := range s.Generals {
			g := &s.Generals[j]
			if g.CityID == city.ID && g.OwnerID == city.OwnerID && !g.Captive {
				if g.Intellect > bestIQ {
					bestIQ = g.Intellect
					bestID = g.ID
				}
			}
		}
		city.SatrapID = bestID
	}
}

// evaluateVictory checks whether the campaign has reached an end state and logs
// the result. The legacy game ends when one ruler holds every owned city.
func (s *GameState) evaluateVictory() {
	owners := map[string]bool{}
	playerCities := 0
	for i := range s.Cities {
		owner := s.Cities[i].OwnerID
		if owner == "" || owner == "neutral" {
			continue
		}
		owners[owner] = true
		if owner == s.PlayerID {
			playerCities++
		}
	}

	switch {
	case playerCities == 0:
		s.prependLog("大势已去，主公基业尽失！")
	case len(owners) == 1 && owners[s.PlayerID]:
		s.prependLog("天下归一，主公成就霸业！")
	}
}

// consumeMonthlyFood is the simplified food consumption (kept for backward compat).
func (s *GameState) consumeMonthlyFood(city *City) {
	totalSoldiers := city.Garrison
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID && !s.Generals[i].Captive {
			totalSoldiers += s.Generals[i].Soldiers
		}
	}

	upkeep := totalSoldiers / 50
	if upkeep <= 0 {
		return
	}
	if city.Food > upkeep {
		city.Food -= upkeep
		return
	}

	city.Food = 0
	city.State = CityStateFamine
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID && !s.Generals[i].Captive {
			s.Generals[i].Soldiers /= 2
		}
	}
}

func formatDate(date Date) string {
	return itoa(date.Year) + "年" + itoa(date.Month) + "月"
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt is defined in commands.go (package-level)

// armsTypeToInt converts arms type string to index for subdue matrix lookup
func armsTypeToInt(armsType string) int {
	switch armsType {
	case "骑兵":
		return 0
	case "步兵":
		return 1
	case "弓箭兵":
		return 2
	case "水军":
		return 3
	case "极兵":
		return 4
	case "玄兵":
		return 5
	default:
		return 1
	}
}

// getRulerCharacter returns the character index (0-4) for a ruler's character string
func (s *GameState) getRulerCharacter(rulerID string) int {
	for _, r := range s.Rulers {
		if r.ID == rulerID {
			switch r.Character {
			case "冒进", "卤莽":
				return 0
			case "狂人", "怕死":
				return 1
			case "奸诈", "贪财":
				return 2
			case "大义", "大志":
				return 3
			case "和平", "忠义":
				return 4
			default:
				return 2
			}
		}
	}
	return 2 // default 奸诈
}
