package game

import "math"

// progression.go implements the legacy general progression systems faithfully
// ported from the C source:
//   - LevelUp (tactic.c): level += 1, capped at maxLevel(20)
//   - FgtGetExp (FightSub.c): exp = sqrt(hurt)/4 - levelDiff + 2
//   - FgtChkAtkEnd (Fight.c): when experience >= FGT_EXPMAX(100), level up and subtract 100
//   - PlcArmsMaxP (PublicFun.c): maxArms = Level*100 + Force*10 + IQ*10

const (
	maxLevel       = 20  // 最大等级 (g_engineConfig.maxLevel default)
	fgtExpMax      = 100 // FGT_EXPMAX: 经验满值触发升级
	killExpSameLvl = 16  // 同级击杀额外经验
	killExpLowLvl  = 24  // 击杀高级武将额外经验
	killExpHighLvl = 8   // 击杀低级武将额外经验

	// 征兵相关 (g_engineConfig defaults in attribute.h)
	armsPerDevotion = 20 // 征兵量与民忠的比例
	armsPerMoney    = 10 // 每个金钱可购买的士兵数
)

// LevelUp raises a general's level by one, capped at maxLevel.
// Mirrors LevelUp() in tactic.c.
func (g *General) LevelUp() {
	g.Level++
	if g.Level > maxLevel {
		g.Level = maxLevel
	}
}

// MaxArms returns the maximum troops a general can command.
// Mirrors PlcArmsMaxP() in PublicFun.c: Level*100 + Force*10 + IQ*10.
func (g *General) MaxArms() int {
	armys := g.Level*100 + g.Force*10 + g.Intellect*10
	if armys > 0xfffe {
		armys = 0xfffe
	}
	return armys
}

// battleExp computes the experience gained from inflicting `hurt` damage.
// Mirrors FgtGetExp() in FightSub.c:
//
//	exp = sqrt(hurt) / 4
//	if attacker is lower level: exp -= levelDiff (can go negative -> clamped 0)
//	else: exp = max(0, exp - levelDiff)
//	exp += 2
func battleExp(hurt, attackerLevel, defenderLevel int) int {
	if hurt < 0 {
		hurt = 0
	}
	exp := int(math.Sqrt(float64(hurt))) / 4
	levelDiff := attackerLevel - defenderLevel
	if levelDiff < 0 {
		// 攻击者等级更低：经验受惩罚较小（C 源码用无符号溢出表示负差，这里直接减绝对值）
		exp -= (-levelDiff)
		if exp < 0 {
			exp = 0
		}
	} else {
		if exp > levelDiff {
			exp -= levelDiff
		} else {
			exp = 0
		}
	}
	exp += 2
	return exp
}

// killBonusExp returns the extra experience for killing an enemy, based on
// the level difference. Mirrors the sType logic in FgtAtkAction (FightSub.c):
//
//	attacker lower level (diff < 0)  -> 24
//	same level (diff == 0)           -> 16
//	attacker higher level (diff > 0) -> 8
func killBonusExp(attackerLevel, defenderLevel int) int {
	diff := attackerLevel - defenderLevel
	switch {
	case diff < 0:
		return killExpLowLvl
	case diff == 0:
		return killExpSameLvl
	default:
		return killExpHighLvl
	}
}

// gainExperience adds experience to a general and handles automatic level-ups.
// Mirrors FgtChkAtkEnd (Fight.c): while experience >= 100, subtract 100 and level up.
// Returns the number of levels gained.
func (g *General) gainExperience(exp int) int {
	if exp <= 0 {
		return 0
	}
	g.Experience += exp
	levels := 0
	for g.Experience >= fgtExpMax && g.Level < maxLevel {
		g.Experience -= fgtExpMax
		g.LevelUp()
		levels++
	}
	// Cap residual experience at maxLevel so it doesn't overflow indefinitely.
	if g.Level >= maxLevel && g.Experience > fgtExpMax {
		g.Experience = fgtExpMax
	}
	return levels
}

// clampSoldiersToMax ensures a general's troops never exceed their command cap.
func (g *General) clampSoldiersToMax() {
	if max := g.MaxArms(); g.Soldiers > max {
		g.Soldiers = max
	}
}

// conscript recruits new troops into the city's reserve (MothballArms).
// Mirrors ConscriptionMake (citycmdc.c):
//
//	arms = PeopleDevotion * armsPerDevotion(20)
//	arms = min(arms, Money * armsPerMoney(10))   // limited by treasury
//	MothballArms += arms
//	Money -= arms / armsPerMoney
//
// Returns the number of troops recruited.
func (s *GameState) conscript(general *General, city *City) int {
	arms := city.PeopleDevotion * armsPerDevotion
	moneyLimit := city.Money * armsPerMoney
	if arms > moneyLimit {
		arms = moneyLimit
	}
	if arms <= 0 {
		return 0
	}
	city.MothballArms += arms
	city.Money -= arms / armsPerMoney
	if city.Money < 0 {
		city.Money = 0
	}
	return arms
}

// distribute transfers reserve troops (MothballArms) to a general up to their
// command cap (MaxArms). Mirrors DistributeMake (citycmdc.c) combined with the
// PlcArmsMax ceiling enforced when allocating troops to a general.
// Returns the number of troops moved.
func (s *GameState) distribute(general *General, city *City) int {
	if city.MothballArms <= 0 {
		return 0
	}
	capacity := general.MaxArms() - general.Soldiers
	if capacity <= 0 {
		return 0
	}
	moved := city.MothballArms
	if moved > capacity {
		moved = capacity
	}
	general.Soldiers += moved
	city.MothballArms -= moved
	return moved
}
