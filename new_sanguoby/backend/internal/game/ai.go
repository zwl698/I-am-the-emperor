package game

import (
	"fmt"
	"math/rand"
	"sort"
)

// ai.go drives the non-player rulers each turn so the campaign feels alive,
// mirroring the legacy "群雄逐鹿" loop where rival warlords develop their cities
// and march on weaker neighbours between the player's strategy phases.
//
// The AI faithfully ports the C source tactic.c logic:
//   - ComputerTactic: per-city decision (interior/diplomacy/armament)
//   - ComputerTacticInterior: 内政 (开垦/招商/搜寻/出巡)
//   - ComputerTacticHarmonize: 协调 (治理/赏赐/没收/移动)
//   - ComputerTacticDiplomatism: 外交 (招降/处斩/离间/招揽/策反/劝降)
//   - ComputerTacticArmament: 军备 (侦察/征兵/分配/掠夺/出征)

const aiAttackArmsAdvantage = 120 // attack only when attacker arms >= defender * 120%

// RunEnemyTurns lets every non-player, non-neutral ruler act once. Each of their
// generals either attacks a clearly weaker adjacent enemy city or develops its
// home city's economy. Returns the number of cities that changed hands.
func (s *GameState) RunEnemyTurns() int {
	captures := 0

	// Stable ruler order keeps turns deterministic for tests/replays.
	rulerIDs := make([]string, 0, len(s.Rulers))
	for _, r := range s.Rulers {
		if r.ID == s.PlayerID || r.ID == "neutral" || r.ID == "" {
			continue
		}
		rulerIDs = append(rulerIDs, r.ID)
	}
	sort.Strings(rulerIDs)

	for _, rulerID := range rulerIDs {
		captures += s.runRulerTurn(rulerID)
	}
	return captures
}

// runRulerTurn acts for a single AI ruler and returns the number of captures.
// Mirrors ComputerTactic from tactic.c: for each city owned by this ruler,
// apply AI benefits (avoidCalamity++, food floor, money balance), then
// choose strategy based on ruler character odds.
func (s *GameState) runRulerTurn(rulerID string) int {
	captures := 0
	rulerName := s.rulerName(rulerID)
	charIdx := s.getRulerCharacter(rulerID)

	// Collect this ruler's actionable generals in a stable order.
	generalIDs := make([]string, 0)
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.OwnerID == rulerID && !g.Captive && g.Stamina >= battleStaminaCost && g.Soldiers > 0 {
			generalIDs = append(generalIDs, g.ID)
		}
	}
	sort.Strings(generalIDs)
	if len(generalIDs) == 0 {
		s.prependLog(fmt.Sprintf("诸侯行动：%s 本月按兵不动，暂无可行动武将。", rulerName))
		return 0
	}

	// Process each city owned by this ruler
	for i := range s.Cities {
		city := &s.Cities[i]
		if city.OwnerID != rulerID {
			continue
		}

		// AI城市福利（对应C源码中ComputerTactic的城池预处理）
		city.AvoidCalamity++
		if city.AvoidCalamity > 100 {
			city.AvoidCalamity = 100
		}
		if city.Food < 100 {
			city.Food = 500
		}
		if city.Money > 10000 {
			city.Money /= 2
		}

		// 策略选择：基于君主性格概率
		rnd := rand.Intn(100)
		if tacticOddsIH[charIdx] > rnd {
			// 内政+协调策略
			s.aiInterior(city, rulerID)
			s.aiHarmonize(city, rulerID)
		} else if tacticOddsD[charIdx] > rnd {
			// 外交策略
			s.aiDiplomatism(city, rulerID)
		} else {
			// 军备策略
			s.aiArmament(city, rulerID)
		}
	}

	// Also let generals act individually for attacks (backward compat with battle system)
	acted := 0
	for _, gid := range generalIDs {
		general := s.findGeneral(gid)
		if general == nil || general.Captive || general.Stamina < battleStaminaCost || general.Soldiers <= 0 {
			continue
		}
		from := s.findCity(general.CityID)
		if from == nil || from.OwnerID != rulerID {
			continue
		}

		target := s.pickAttackTarget(rulerID, from, general)
		if target != nil {
			outcome := s.resolveAttack(from, general, target)
			if outcome.Captured {
				captures++
			}
			acted++
			continue
		}

		// No worthwhile attack: develop the home city instead.
		s.prependLog(fmt.Sprintf("诸侯行动：%s %s", rulerName, s.aiDevelopCity(general, from)))
		acted++
	}
	if acted == 0 {
		s.prependLog(fmt.Sprintf("诸侯行动：%s 本月军令未动。", rulerName))
	}
	return captures
}

// aiInterior 对应 ComputerTacticInterior：内政策略
// 随机执行：开垦(+200农业)/招商(+200商业)/搜寻/出巡(+4民忠,+100人口)
func (s *GameState) aiInterior(city *City, rulerID string) {
	generals := s.generalsInCity(city.ID, rulerID)
	if len(generals) == 0 {
		return
	}

	// Pick one random general to act
	g := generals[rand.Intn(len(generals))]
	if g.Stamina < 4 {
		return
	}
	g.Stamina -= 4

	switch rand.Intn(5) {
	case 0: // 开垦
		gain := 200
		before := city.Farming
		city.Farming = minInt(city.FarmingLimit, city.Farming+gain)
		if before != city.Farming {
			s.prependLog(fmt.Sprintf("诸侯内政：%s 在 %s 开垦，农业 +%d。", g.Name, city.Name, city.Farming-before))
		}
	case 1: // 招商
		gain := 200
		before := city.Commerce
		city.Commerce = minInt(city.CommerceLimit, city.Commerce+gain)
		if before != city.Commerce {
			s.prependLog(fmt.Sprintf("诸侯内政：%s 在 %s 招商，商业 +%d。", g.Name, city.Name, city.Commerce-before))
		}
	case 2: // 搜寻
		if g.Intellect >= 80 {
			bonus := 80 + g.Intellect
			city.Money = minInt(maxResourceValue, city.Money+bonus)
			s.prependLog(fmt.Sprintf("诸侯内政：%s 搜寻 %s，得金%d。", g.Name, city.Name, bonus))
		}
	case 3: // 出巡
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+4)
		city.Population = minInt(city.PopulationLimit, city.Population+100)
		s.prependLog(fmt.Sprintf("诸侯内政：%s 出巡 %s，民忠上升。", g.Name, city.Name))
	case 4: // skip
		g.Stamina += 4 // 恢复体力
	}
}

// aiHarmonize 对应 ComputerTacticHarmonize：协调策略
// 随机执行：治理(恢复状态+4防灾)/赏赐/没收/输送/移动
func (s *GameState) aiHarmonize(city *City, rulerID string) {
	generals := s.generalsInCity(city.ID, rulerID)
	if len(generals) == 0 {
		return
	}

	g := generals[rand.Intn(len(generals))]
	if g.Stamina < 4 {
		return
	}
	g.Stamina -= 4

	switch rand.Intn(7) {
	case 0: // 治理
		city.State = CityStateNormal
		city.AvoidCalamity = minInt(100, city.AvoidCalamity+4)
		s.prependLog(fmt.Sprintf("诸侯协调：%s 治理 %s，防灾上升。", g.Name, city.Name))
	case 1: // 赏赐 - 提升随机武将忠诚
		if len(generals) > 0 {
			target := generals[rand.Intn(len(generals))]
			target.Loyalty = minInt(100, target.Loyalty+8)
			s.prependLog(fmt.Sprintf("诸侯协调：%s 赏赐 %s。", g.Name, target.Name))
		}
	case 2: // 没收 - 获取金钱但降低民忠
		gain := 80 + g.Intellect/2
		city.Money = minInt(maxResourceValue, city.Money+gain)
		city.PeopleDevotion = maxInt(0, city.PeopleDevotion-5)
		s.prependLog(fmt.Sprintf("诸侯协调：%s 没收财货，金+%d。", g.Name, gain))
	case 5: // 移动到相邻友军城
		friendlyCities := s.adjacentCitiesByOwner(city.ID, func(ownerID string) bool {
			return ownerID == rulerID
		})
		if len(friendlyCities) > 0 {
			dest := friendlyCities[rand.Intn(len(friendlyCities))]
			g.CityID = dest.ID
			s.prependLog(fmt.Sprintf("诸侯协调：%s 移动至 %s。", g.Name, dest.Name))
		}
	default:
		g.Stamina += 4 // 恢复
	}
}

// aiDiplomatism 对应 ComputerTacticDiplomatism：外交策略
// 先处理俘虏（招降/处斩），再对敌将执行离间/招揽/策反/劝降
func (s *GameState) aiDiplomatism(city *City, rulerID string) {
	generals := s.generalsInCity(city.ID, rulerID)
	captives := s.captivesInCity(city.ID, rulerID)
	if len(generals) == 0 {
		return
	}

	g := generals[rand.Intn(len(generals))]
	if g.Stamina < 4 {
		return
	}
	g.Stamina -= 4

	// 有俘虏时优先处理
	if len(captives) > 0 && rand.Intn(8) == 0 {
		captive := captives[rand.Intn(len(captives))]
		// 招降：直接归顺
		captive.Captive = false
		captive.Loyalty = minInt(100, 50+g.Intellect/10)
		s.prependLog(fmt.Sprintf("诸侯外交：%s 招降 %s 成功。", g.Name, captive.Name))
		return
	}
	if len(captives) > 0 && rand.Intn(8) == 1 {
		captive := captives[rand.Intn(len(captives))]
		// 处斩
		s.prependLog(fmt.Sprintf("诸侯外交：%s 处斩 %s。", g.Name, captive.Name))
		// Remove captive from game
		captive.OwnerID = ""
		captive.CityID = ""
		captive.Soldiers = 0
		city.PeopleDevotion = maxInt(0, city.PeopleDevotion-4)
		city.AvoidCalamity = minInt(100, city.AvoidCalamity+2)
		return
	}

	// 外交指令需要相邻敌城的敌将
	target := s.findAdjacentEnemyGeneral(city)
	if target == nil {
		g.Stamina += 4
		return
	}

	switch rand.Intn(8) {
	case 3: // 离间 - 降低忠诚，受目标性格影响
		loss := alienateOdds[target.Character]
		loss = loss + g.Intellect/20
		target.Loyalty = maxInt(0, target.Loyalty-loss)
		s.prependLog(fmt.Sprintf("诸侯外交：%s 离间 %s，忠诚-%d。", g.Name, target.Name, loss))
	case 4: // 招揽 - 高智力+低忠诚可成功
		score := g.Intellect + (100 - target.Loyalty)
		baseProb := canvassOdds[target.Character]
		if score >= 120 || rand.Intn(100) < baseProb {
			target.OwnerID = city.OwnerID
			target.CityID = city.ID
			target.Captive = false
			target.Loyalty = minInt(100, 70+g.Intellect/10)
			s.prependLog(fmt.Sprintf("诸侯外交：%s 招揽 %s 成功！", g.Name, target.Name))
		} else {
			target.Loyalty = maxInt(0, target.Loyalty-4)
			s.prependLog(fmt.Sprintf("诸侯外交：%s 游说 %s，忠诚动摇。", g.Name, target.Name))
		}
	case 5: // 策反 - 对敌方太守
		target.Loyalty = minInt(100, target.Loyalty+4)
		city.AvoidCalamity = minInt(100, city.AvoidCalamity+4)
		s.prependLog(fmt.Sprintf("诸侯外交：%s 策反敌探。", g.Name))
	case 6: // 劝降 - 劝降相邻弱城
		enemyCities := s.adjacentCitiesByOwner(city.ID, func(ownerID string) bool {
			return ownerID != city.OwnerID && ownerID != "neutral"
		})
		for _, ec := range enemyCities {
			if ec.PeopleDevotion+ec.AvoidCalamity < g.Intellect {
				ec.OwnerID = city.OwnerID
				ec.PeopleDevotion = maxInt(0, ec.PeopleDevotion-8)
				ec.AvoidCalamity = maxInt(0, ec.AvoidCalamity-4)
				// 城中武将归顺
				for i := range s.Generals {
					if s.Generals[i].CityID == ec.ID {
						s.Generals[i].OwnerID = city.OwnerID
						s.Generals[i].Captive = false
						s.Generals[i].Loyalty = minInt(100, 60+g.Intellect/12)
					}
				}
				s.prependLog(fmt.Sprintf("诸侯外交：%s 劝降 %s 成功！", g.Name, ec.Name))
				return
			}
		}
		s.prependLog(fmt.Sprintf("诸侯外交：%s 劝降邻城，守军观望。", g.Name))
	default:
		g.Stamina += 4
	}
}

// aiArmament 对应 ComputerTacticArmament：军备策略
// 征兵/分配/掠夺/侦察/出征
func (s *GameState) aiArmament(city *City, rulerID string) {
	generals := s.generalsInCity(city.ID, rulerID)
	if len(generals) == 0 {
		return
	}

	// Use first general as primary for recruitment
	pidx := rand.Intn(len(generals))
	pgeneral := generals[pidx]

	switch rand.Intn(9) {
	case 1, 2, 3, 4, 5: // 征兵 (概率5/9)
		if pgeneral.Stamina < 4 {
			return
		}
		pgeneral.Stamina -= 4
		recruits := 120 + pgeneral.Force*4
		maxRecruits := pgeneral.Level*100 + (pgeneral.Force+pgeneral.Intellect)*10
		if recruits > maxRecruits {
			recruits = maxRecruits
		}
		if city.Population >= recruits*2 {
			city.Population -= recruits * 2
			pgeneral.Soldiers += recruits
			city.MothballArms += recruits / 2
			s.prependLog(fmt.Sprintf("诸侯军备：%s 在 %s 征兵 +%d。", pgeneral.Name, city.Name, recruits))
		}
	case 6: // 掠夺
		if len(generals) == 0 {
			return
		}
		g := generals[rand.Intn(len(generals))]
		if g.Stamina < 4 {
			return
		}
		g.Stamina -= 4
		gain := 50 + g.Force
		city.Food = minInt(maxResourceValue, city.Food+gain)
		city.PeopleDevotion = maxInt(0, city.PeopleDevotion-8)
		s.prependLog(fmt.Sprintf("诸侯军备：%s 掠夺 %s，粮+%d。", g.Name, city.Name, gain))
	case 7: // 出征 - 需要足够武将(>=4)且首将兵力>=1000
		if len(generals) < 4 {
			return
		}
		if pgeneral.Soldiers < 1000 {
			return
		}
		// Sort by soldiers descending
		sort.Slice(generals, func(i, j int) bool {
			return generals[i].Soldiers > generals[j].Soldiers
		})
		enemyCities := s.adjacentCitiesByOwner(city.ID, func(ownerID string) bool {
			return ownerID != rulerID && ownerID != "neutral"
		})
		if len(enemyCities) == 0 || rand.Intn(len(enemyCities)*2) >= len(enemyCities) {
			return
		}
		target := enemyCities[rand.Intn(len(enemyCities))]
		outcome := s.resolveAttack(city, pgeneral, target)
		if outcome.Captured {
			s.prependLog(fmt.Sprintf("诸侯军备：%s 自 %s 攻克 %s！", pgeneral.Name, city.Name, target.Name))
		} else {
			s.prependLog(fmt.Sprintf("诸侯军备：%s 自 %s 进攻 %s 失利。", pgeneral.Name, city.Name, target.Name))
		}
	default:
		// 侦察等不需要额外处理
	}
}

// pickAttackTarget returns the best adjacent enemy city to attack, or nil if no
// target is clearly weaker than the attacking force.
func (s *GameState) pickAttackTarget(rulerID string, from *City, general *General) *City {
	attackArms := general.Soldiers + general.Soldiers*general.Force/200

	neighbourIDs := s.AdjacentCityIDs(from.ID)
	sort.Strings(neighbourIDs)

	var best *City
	bestMargin := 0
	for _, nbID := range neighbourIDs {
		nb := s.findCity(nbID)
		if nb == nil || nb.OwnerID == rulerID {
			continue
		}
		defendArms := s.cityTroops(nbID)
		defendArms += defendArms * nb.PeopleDevotion / 200
		// Require a meaningful arms advantage before committing to a siege.
		if attackArms*100 < defendArms*aiAttackArmsAdvantage {
			continue
		}
		margin := attackArms - defendArms
		if best == nil || margin > bestMargin {
			best = nb
			bestMargin = margin
		}
	}
	return best
}

// aiDevelopCity invests an AI general's action into the home city's economy,
// reusing the same growth shape as the player's 内政 commands.
func (s *GameState) aiDevelopCity(general *General, city *City) string {
	general.Stamina = maxInt(0, general.Stamina-4)
	gain := 10 + general.Intellect/2 + general.Level*2

	switch {
	case city.Farming < city.FarmingLimit:
		before := city.Farming
		city.Farming = minInt(city.FarmingLimit, city.Farming+gain)
		return fmt.Sprintf("令 %s 在 %s 开垦，农业 +%d。", general.Name, city.Name, city.Farming-before)
	case city.Commerce < city.CommerceLimit:
		before := city.Commerce
		city.Commerce = minInt(city.CommerceLimit, city.Commerce+gain)
		return fmt.Sprintf("令 %s 在 %s 招商，商业 +%d。", general.Name, city.Name, city.Commerce-before)
	case city.PeopleDevotion < 100:
		before := city.PeopleDevotion
		city.PeopleDevotion = minInt(100, city.PeopleDevotion+4+general.Intellect/12)
		return fmt.Sprintf("令 %s 巡抚 %s，民忠 +%d。", general.Name, city.Name, city.PeopleDevotion-before)
	default:
		// Fully developed: reinforce the garrison from population.
		if city.Population > 4000 {
			recruits := 100 + general.Force*4
			city.Population -= recruits * 2
			city.MothballArms += recruits
			return fmt.Sprintf("令 %s 在 %s 募兵入城，后备 +%d。", general.Name, city.Name, recruits)
		}
	}
	return fmt.Sprintf("令 %s 驻守 %s，城中暂不动员。", general.Name, city.Name)
}

// Helper: get generals in a city belonging to a specific ruler (non-captive)
func (s *GameState) generalsInCity(cityID, rulerID string) []*General {
	var result []*General
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.CityID == cityID && g.OwnerID == rulerID && !g.Captive {
			result = append(result, g)
		}
	}
	return result
}

// Helper: get captives in a city belonging to a specific ruler
func (s *GameState) captivesInCity(cityID, rulerID string) []*General {
	var result []*General
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.CityID == cityID && g.OwnerID == rulerID && g.Captive {
			result = append(result, g)
		}
	}
	return result
}
