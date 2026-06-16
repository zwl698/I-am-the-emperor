package game

import "fmt"

// ──────────────────────────────────────────────
// 皇帝性格特质系统
// ──────────────────────────────────────────────

type TraitID string

const (
	TraitBenevolent   TraitID = "benevolent"   // 仁厚：民政+20%，军务-10%
	TraitSuspicious   TraitID = "suspicious"   // 多疑：暗线+20%，朝稳-10%
	TraitAmbitious    TraitID = "ambitious"    // 雄才：新政+20%，国库消耗+15%
	TraitVainglorious TraitID = "vainglorious" // 好大喜功：军务+15%，国库-15%
	TraitFrugal       TraitID = "frugal"       // 节俭：国库消耗-20%，魅力-10%
	TraitRuthless     TraitID = "ruthless"     // 铁腕：势力+20%，民心-15%
	TraitScholarly    TraitID = "scholarly"    // 好学：学识+15%，新政+10%
	TraitCharismatic  TraitID = "charismatic"  // 亲和：邦交+15%，魅力+10%
	TraitParanoid     TraitID = "paranoid"     // 偏执：势力+15%，健康-10%
	TraitVisionary    TraitID = "visionary"    // 远见：新政+15%，朝稳-10%
)

type EmperorTrait struct {
	ID          TraitID `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	AcquiredAge int     `json:"acquiredAge"` // 获得该特质的年龄/年份
	Source      string  `json:"source"`      // 获得来源（哪个选择触发的）
}

// traitModifiers defines how each trait scales effects by domain.
// Returns a multiplier (1.0 = neutral).
func traitEffectMultiplier(traitID TraitID, domain Domain) float64 {
	switch traitID {
	case TraitBenevolent:
		if domain == DomainDomestic {
			return 1.20
		}
		if domain == DomainMilitary {
			return 0.90
		}
	case TraitSuspicious:
		if domain == DomainIntrigue {
			return 1.20
		}
		if domain == DomainCourt {
			return 0.90
		}
	case TraitAmbitious:
		if domain == DomainReform {
			return 1.20
		}
	case TraitVainglorious:
		if domain == DomainMilitary {
			return 1.15
		}
		if domain == DomainEconomy {
			return 0.85
		}
	case TraitFrugal:
		if domain == DomainEconomy {
			return 1.20
		}
		if domain == DomainCourt {
			return 0.90
		}
	case TraitRuthless:
		if domain == DomainIntrigue {
			return 1.20
		}
		if domain == DomainDomestic {
			return 0.85
		}
	case TraitScholarly:
		if domain == DomainReform {
			return 1.25
		}
	case TraitCharismatic:
		if domain == DomainDiplomacy {
			return 1.15
		}
	case TraitParanoid:
		if domain == DomainIntrigue {
			return 1.15
		}
	case TraitVisionary:
		if domain == DomainReform {
			return 1.15
		}
		if domain == DomainCourt {
			return 0.90
		}
	}
	return 1.0
}

// traitHealthModifier returns health change per year from traits.
func traitHealthModifier(traitID TraitID) int {
	switch traitID {
	case TraitParanoid:
		return -1 // 偏执伤身
	case TraitBenevolent:
		return 1 // 仁者寿
	case TraitVainglorious:
		return -1 // 好大喜功劳身
	case TraitFrugal:
		return 1 // 节制养生
	default:
		return 0
	}
}

// traitCrisisModifier returns crisis clock change modifier from traits.
func traitCrisisModifier(traitID TraitID) int {
	switch traitID {
	case TraitRuthless:
		return 1 // 铁腕激化危机
	case TraitBenevolent:
		return -1 // 仁厚缓解危机
	case TraitSuspicious:
		return 1 // 多疑制造不安
	case TraitCharismatic:
		return -1 // 亲和稳定人心
	default:
		return 0
	}
}

// ──────────────────────────────────────────────
// 皇帝健康与衰老系统
// ──────────────────────────────────────────────

// EmperorCondition represents the emperor's physical/mental condition.
type EmperorCondition struct {
	HealthTrend    string `json:"healthTrend"`    // "强健"/"平稳"/"衰退"/"危急"
	MaxHealth      int    `json:"maxHealth"`      // 当前健康上限（随衰老降低）
	DecayRate      int    `json:"decayRate"`      // 每年健康自然衰退值
	LastIllness    string `json:"lastIllness"`    // 最近一次疾病描述
	IllnessTurn    int    `json:"illnessTurn"`    // 最近一次发病回合
	IllnessCount   int    `json:"illnessCount"`   // 累计发病次数
	AbdicationRisk int    `json:"abdicationRisk"` // 禅位压力（0-100）
}

// healthDecayByAge returns natural health decay rate per year based on age.
func healthDecayByAge(age int) int {
	switch {
	case age < 30:
		return 0
	case age < 40:
		return 1
	case age < 50:
		return 2
	case age < 60:
		return 4
	case age < 70:
		return 6
	default:
		return 9
	}
}

// maxHealthByAge returns the health ceiling based on age.
func maxHealthByAge(age int) int {
	switch {
	case age < 30:
		return 100
	case age < 40:
		return 90
	case age < 50:
		return 78
	case age < 60:
		return 64
	case age < 70:
		return 48
	default:
		return 30
	}
}

// healthTrendLabel describes the current health trajectory.
func healthTrendLabel(health, maxHealth int) string {
	ratio := float64(health) / float64(maxHealth)
	switch {
	case ratio >= 0.8:
		return "强健"
	case ratio >= 0.5:
		return "平稳"
	case ratio >= 0.25:
		return "衰退"
	default:
		return "危急"
	}
}

// ──────────────────────────────────────────────
// 核心计算函数
// ──────────────────────────────────────────────

// computeTraitEffects applies all emperor traits to scale an effect set by domain.
func (s *GameState) computeTraitEffects(effects Effects, domain Domain) Effects {
	result := effects
	multiplier := 1.0
	for _, trait := range s.EmperorTraits {
		multiplier *= traitEffectMultiplier(trait.ID, domain)
	}
	// Only scale the "resource" effects, not personal stats like Health/Legitimacy
	result.Treasury = scaleInt(result.Treasury, multiplier)
	result.Grain = scaleInt(result.Grain, multiplier)
	result.Populace = scaleInt(result.Populace, multiplier)
	result.Army = scaleInt(result.Army, multiplier)
	result.Diplomacy = scaleInt(result.Diplomacy, multiplier)
	result.Stability = scaleInt(result.Stability, multiplier)
	result.BorderThreat = scaleInt(result.BorderThreat, multiplier)
	result.Reform = scaleInt(result.Reform, multiplier)
	result.Influence = scaleInt(result.Influence, multiplier)
	return result
}

// traitCrisisClockDelta computes total crisis clock modifier from traits.
func (s *GameState) traitCrisisClockDelta() int {
	delta := 0
	for _, trait := range s.EmperorTraits {
		delta += traitCrisisModifier(trait.ID)
	}
	return delta
}

// traitHealthDelta computes total health modifier from traits per year.
func (s *GameState) traitHealthDelta() int {
	delta := 0
	for _, trait := range s.EmperorTraits {
		delta += traitHealthModifier(trait.ID)
	}
	return delta
}

// applyAging applies aging effects when a new year begins.
// Called from advanceCalendar or similar.
func (s *GameState) applyAging() {
	if s.Phase != PhaseEmperor {
		return
	}

	decay := healthDecayByAge(s.Age)
	decay += s.traitHealthDelta()
	if decay < 0 {
		decay = 0
	}

	// Natural health decay
	s.Stats.Health = clamp(s.Stats.Health-decay, 0, 100)

	// Update max health
	maxH := maxHealthByAge(s.Age)
	s.Condition.MaxHealth = maxH
	s.Condition.DecayRate = decay

	// Cap health to max
	if s.Stats.Health > maxH {
		s.Stats.Health = maxH
	}

	// Update trend label
	s.Condition.HealthTrend = healthTrendLabel(s.Stats.Health, maxH)

	// Age-related illness: chance increases with age
	illnessChance := 0
	if s.Age >= 50 {
		illnessChance = (s.Age - 45) * 2
	}
	if s.Stats.Health < 40 {
		illnessChance += 20
	}
	s.ensureRNG()
	if illnessChance > 0 && s.rng.Intn(100) < illnessChance {
		s.triggerIllness()
	}

	// Update abdication risk based on age and health
	s.Condition.AbdicationRisk = clamp(
		(s.Age-55)*3+(100-s.Stats.Health)/2+len(s.Heirs)*5,
		0, 100,
	)
}

// triggerIllness generates an illness event for the aging emperor.
func (s *GameState) triggerIllness() {
	illnesses := []struct {
		name   string
		effect Effects
	}{
		{"风寒缠身", Effects{Health: -6, Stability: -2}},
		{"旧伤复发", Effects{Health: -8, Martial: -2}},
		{"心悸气短", Effects{Health: -10, Influence: -3}},
		{"目眩头痛", Effects{Health: -7, Learning: -2}},
		{"脾胃不和", Effects{Health: -5, Charisma: -2}},
	}

	s.ensureRNG()
	ill := illnesses[s.rng.Intn(len(illnesses))]
	s.applyEffects(ill.effect)
	s.Condition.LastIllness = ill.name
	s.Condition.IllnessTurn = s.Turn
	s.Condition.IllnessCount++

	s.History = append(s.History, HistoryEntry{
		Turn:    s.Turn,
		Age:     s.Age,
		Phase:   s.Phase,
		Choice:  "龙体欠安",
		Summary: fmt.Sprintf("皇帝%s，御医入诊。", ill.name),
		Effects: ill.effect,
	})
}

// AddTrait adds a new emperor trait, preventing duplicates.
func (s *GameState) AddTrait(id TraitID, source string) {
	for _, t := range s.EmperorTraits {
		if t.ID == id {
			return // Already has this trait
		}
	}
	name, desc := traitNameDesc(id)
	s.EmperorTraits = append(s.EmperorTraits, EmperorTrait{
		ID:          id,
		Name:        name,
		Description: desc,
		AcquiredAge: s.Age,
		Source:      source,
	})
}

// traitNameDesc returns the display name and description for a trait.
func traitNameDesc(id TraitID) (string, string) {
	switch id {
	case TraitBenevolent:
		return "仁厚", "民政选项效果+20%，军务选项效果-10%，每年健康+1"
	case TraitSuspicious:
		return "多疑", "暗线选项效果+20%，朝堂选项效果-10%，危机钟+1"
	case TraitAmbitious:
		return "雄才", "新政选项效果+20%，国库消耗+15%"
	case TraitVainglorious:
		return "好大喜功", "军务选项效果+15%，国库选项效果-15%，每年健康-1"
	case TraitFrugal:
		return "节俭", "国库消耗-20%，宫廷选项效果-10%，每年健康+1"
	case TraitRuthless:
		return "铁腕", "势力选项效果+20%，民心选项效果-15%，危机钟+1"
	case TraitScholarly:
		return "好学", "学识+15%，新政选项效果+10%"
	case TraitCharismatic:
		return "亲和", "邦交选项效果+15%，魅力+10%，危机钟-1"
	case TraitParanoid:
		return "偏执", "势力选项效果+15%，每年健康-1"
	case TraitVisionary:
		return "远见", "新政选项效果+15%，朝堂选项效果-10%"
	default:
		return string(id), "未知特质"
	}
}

// scaleInt scales an integer by a multiplier, rounding towards zero.
func scaleInt(value int, multiplier float64) int {
	if value == 0 || multiplier == 1.0 {
		return value
	}
	result := int(float64(value) * multiplier)
	// Preserve direction: don't let scaling flip the sign
	if value > 0 && result < 0 {
		return 0
	}
	if value < 0 && result > 0 {
		return 0
	}
	return result
}

// ──────────────────────────────────────────────
// 初始化与查询
// ──────────────────────────────────────────────

// initEmperorCondition sets up the initial condition for a new emperor.
func (s *GameState) initEmperorCondition() {
	s.Condition = EmperorCondition{
		HealthTrend:    "强健",
		MaxHealth:      100,
		DecayRate:      0,
		AbdicationRisk: 0,
	}
}

// princeChoiceTrait maps a prince-phase choice ID to a trait.
// Each story beat grants one trait reflecting the player's character development.
var princeChoiceTrait = map[string]TraitID{
	// birth-omen: 紫宸宫中的啼哭
	"grab-scroll":   TraitScholarly,   // 抓起竹简 → 好学
	"smile-consort": TraitCharismatic, // 向皇后展露笑容 → 亲和
	"cry-loudly":    TraitRuthless,    // 放声大哭 → 铁腕

	// study-yard: 东宫书院
	"answer-people": TraitBenevolent,   // 答：先安百姓 → 仁厚
	"answer-army":   TraitVainglorious, // 答：先强兵甲 → 好大喜功
	"answer-father": TraitVisionary,    // 答：先顺父皇 → 远见

	// winter-hunt: 皇家冬狩
	"mount-again":     TraitVainglorious, // 忍痛重新上马 → 好大喜功
	"protect-servant": TraitBenevolent,   // 先扶起被撞倒的小内侍 → 仁厚
	"accuse-brother":  TraitSuspicious,   // 当众指认三皇子 → 多疑

	// flood-memorial: 南河急报
	"open-granary":     TraitBenevolent,   // 开仓赈济 → 仁厚
	"borrow-merchants": TraitFrugal,       // 向皇商借银 → 节俭
	"send-army":        TraitVainglorious, // 调禁军协助 → 好大喜功

	// succession-night: 烛影摇红
	"secure-edict":   TraitVisionary,   // 请太傅与中书共同护诏 → 远见
	"control-guards": TraitRuthless,    // 联络禁军封锁宫门 → 铁腕
	"appeal-clans":   TraitCharismatic, // 向宗室许诺共治 → 亲和
}

// princeChoiceSource returns a human-readable source description for a prince choice.
var princeChoiceSource = map[string]string{
	"grab-scroll":      "幼年抓起竹简",
	"smile-consort":    "幼年展露笑容",
	"cry-loudly":       "幼年放声大哭",
	"answer-people":    "六岁答：先安百姓",
	"answer-army":      "六岁答：先强兵甲",
	"answer-father":    "六岁答：先顺父皇",
	"mount-again":      "十岁忍痛上马",
	"protect-servant":  "十岁扶起内侍",
	"accuse-brother":   "十岁当众指认",
	"open-granary":     "十四岁开仓赈济",
	"borrow-merchants": "十四岁向皇商借银",
	"send-army":        "十四岁调禁军筑堤",
	"secure-edict":     "十六岁请太傅护诏",
	"control-guards":   "十六岁联络禁军",
	"appeal-clans":     "十六岁向宗室许诺",
}

// grantTraitForPrinceChoice grants a trait when the player makes a prince-phase choice.
func (s *GameState) grantTraitForPrinceChoice(choiceID string) {
	traitID, ok := princeChoiceTrait[choiceID]
	if !ok {
		return
	}
	source := princeChoiceSource[choiceID]
	if source == "" {
		source = "皇子成长"
	}
	s.AddTrait(traitID, source)
}

// assignInitialTraits gives the emperor starting traits based on dynasty.
// Prince-phase traits have already been granted by grantTraitForPrinceChoice
// during the prince phase; this function only adds the dynasty base trait
// and situational traits (e.g. health-based).
func (s *GameState) assignInitialTraits() {
	// Base trait from dynasty
	switch s.Dynasty.ID {
	case "dayin":
		s.AddTrait(TraitAmbitious, "大殷基业")
	case "tang":
		s.AddTrait(TraitCharismatic, "李唐风骨")
	case "song":
		s.AddTrait(TraitScholarly, "赵宋文脉")
	case "ming":
		s.AddTrait(TraitSuspicious, "朱明祖训")
	default:
		s.AddTrait(TraitBenevolent, "开国之君")
	}

	// Situational trait: frail body breeds paranoia
	if s.Stats.Health < 50 {
		s.AddTrait(TraitParanoid, "体弱多疑")
	}

	// Situational trait: extreme martial focus
	if s.Stats.Martial >= 65 {
		s.AddTrait(TraitVainglorious, "尚武之性")
	}
}

// HasTrait checks if the emperor has a specific trait.
func (s *GameState) HasTrait(id TraitID) bool {
	for _, t := range s.EmperorTraits {
		if t.ID == id {
			return true
		}
	}
	return false
}

// TraitSummary returns a short description of the emperor's trait combination.
func (s *GameState) TraitSummary() string {
	if len(s.EmperorTraits) == 0 {
		return "尚未显露性格"
	}
	result := ""
	for _, t := range s.EmperorTraits {
		if result != "" {
			result += "、"
		}
		result += t.Name
	}
	return result
}
