package game

import (
	"fmt"
	"math/rand"
)

// battle.go implements the 出征/攻城 (campaign attack) resolution.
//
// The win/lose probability model is a faithful port of the legacy engine's
// FgtCountWon (FgtCount.c): the outcome is decided by comparing total troop
// strength (兵力) and provisions (兵粮) between attacker and defender.
//
//	攻方兵力 > 守方兵力:
//	    攻方 > 守方*2          -> 70% 胜
//	    否则 攻方粮 > 守方粮    -> 60% 胜
//	    否则                  -> 40% 胜
//	攻方兵力 <= 守方兵力:
//	    攻方 < 守方/2          -> 2%  胜
//	    否则 攻方粮 > 守方粮    -> 30% 胜
//	    否则                  -> 10% 胜

var (
	ErrBattleSameOwner    = fmt.Errorf("%w: cannot attack your own city", ErrInvalidCommand)
	ErrBattleNotAdjacent  = fmt.Errorf("%w: target city is not adjacent", ErrInvalidCommand)
	ErrBattleNoSoldiers   = fmt.Errorf("%w: general has no soldiers", ErrGeneralNotReady)
	ErrBattleNoFood       = fmt.Errorf("%w: no provisions assigned", ErrInvalidCommand)
	ErrBattleTargetNoCity = fmt.Errorf("%w: target city not found", ErrInvalidCommand)
)

// BattleOutcome describes the result of one 出征 resolution for the API/UI.
type BattleOutcome struct {
	Won               bool     `json:"won"`
	FromCityID        string   `json:"fromCityId"`
	FromCityName      string   `json:"fromCityName"`
	TargetCityID      string   `json:"targetCityId"`
	TargetCityName    string   `json:"targetCityName"`
	GeneralID         string   `json:"generalId"`
	GeneralName       string   `json:"generalName"`
	GeneralIDs        []string `json:"generalIds"`
	GeneralNames      []string `json:"generalNames"`
	AttackerRulerID   string   `json:"attackerRulerId"`
	AttackerRulerName string   `json:"attackerRulerName"`
	DefenderRulerID   string   `json:"defenderRulerId"`
	DefenderRulerName string   `json:"defenderRulerName"`
	DefenderGenerals  []string `json:"defenderGenerals"`
	Money             int      `json:"money"`
	Food              int      `json:"food"`
	RemainingFood     int      `json:"remainingFood"`
	FieldAdvantage    int      `json:"fieldAdvantage"`
	AttackPower       int      `json:"attackPower"`
	DefensePower      int      `json:"defensePower"`
	AttackerLosses    int      `json:"attackerLosses"`
	DefenderLosses    int      `json:"defenderLosses"`
	Captured          bool     `json:"captured"`
	CapturedGenerals  []string `json:"capturedGenerals"`
	ExperienceGained  int      `json:"experienceGained"`
	LevelUps          []string `json:"levelUps"`
	Message           string   `json:"message"`
}

const battleStaminaCost = 4
const maxBattleGenerals = 10

// battleRand is the source of randomness; overridable in tests.
var defaultBattleRand = rand.Intn
var battleRand = defaultBattleRand

// AdjacentCityIDs returns the set of city IDs reachable from cityID via a single
// route hop (the campaign road network is undirected).
func (s *GameState) AdjacentCityIDs(cityID string) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range s.Routes {
		var other string
		switch cityID {
		case r.From:
			other = r.To
		case r.To:
			other = r.From
		default:
			continue
		}
		if other != "" && !seen[other] {
			seen[other] = true
			out = append(out, other)
		}
	}
	return out
}

func (s *GameState) isAdjacent(fromID, targetID string) bool {
	for _, id := range s.AdjacentCityIDs(fromID) {
		if id == targetID {
			return true
		}
	}
	return false
}

// cityTroops sums every soldier defending a city (garrison + non-captive generals).
func (s *GameState) cityTroops(cityID string) int {
	city := s.findCity(cityID)
	if city == nil {
		return 0
	}
	total := city.Garrison
	for i := range s.Generals {
		if s.Generals[i].CityID == cityID && !s.Generals[i].Captive {
			total += s.Generals[i].Soldiers
		}
	}
	return total
}

// resolveBattleWon ports FgtCountWon: given strengths and provisions, returns
// whether the attacker wins. randv is a 0..100 roll. The comparisons mirror
// FgtCount.c exactly: legacy stores FGT_WON as 1 and FGT_LOSE as 2, so the
// original `(randv < N) + 1` expressions mean "rolls below N lose".
func resolveBattleWon(attackArms, defendArms, attackFood, defendFood, randv int) bool {
	if attackArms == 0 {
		return false
	}
	if defendArms == 0 {
		return true
	}
	if attackArms > defendArms {
		switch {
		case attackArms>>1 > defendArms:
			return randv >= 30
		case attackFood > defendFood:
			return randv >= 40
		default:
			return randv >= 60
		}
	}
	switch {
	case attackArms < defendArms>>1:
		return randv <= 2
	case attackFood > defendFood:
		return randv <= 30
	default:
		return randv <= 10
	}
}

// ApplyBattle launches an attack from a player-owned city against an adjacent
// city, using the chosen general's army plus the origin garrison as the strike
// force. The outcome follows the legacy probability model.
func (s *GameState) ApplyBattle(fromCityID, generalID, targetCityID string) (*BattleOutcome, error) {
	from := s.findCity(fromCityID)
	if from == nil {
		return nil, fmt.Errorf("%w: city %s", ErrInvalidCommand, fromCityID)
	}
	if from.OwnerID != s.PlayerID {
		return nil, ErrCityNotPlayable
	}

	general := s.findGeneral(generalID)
	if general == nil || general.CityID != fromCityID || general.OwnerID != s.PlayerID || general.Captive {
		return nil, ErrGeneralNotReady
	}
	if general.Stamina < battleStaminaCost {
		return nil, ErrGeneralNotReady
	}
	if general.Soldiers <= 0 {
		return nil, ErrBattleNoSoldiers
	}

	target := s.findCity(targetCityID)
	if target == nil {
		return nil, ErrBattleTargetNoCity
	}
	if target.OwnerID == s.PlayerID {
		return nil, ErrBattleSameOwner
	}
	if !s.isAdjacent(fromCityID, targetCityID) {
		return nil, ErrBattleNotAdjacent
	}

	if target.OwnerID == "neutral" {
		return s.occupyEmptyCity(from, []*General{general}, target, 0, from.Food, from.Food, 0), nil
	}
	return s.resolveAttack(from, general, target), nil
}

// ApplyBattlePlan launches the interactive player battle flow after the UI has
// selected up to ten generals and assigned campaign supplies. The final
// battlefield advantage comes from the tactical mini-map operation.
func (s *GameState) ApplyBattlePlan(fromCityID string, generalIDs []string, targetCityID string, money, food, remainingFood, fieldAdvantage int) (*BattleOutcome, error) {
	from := s.findCity(fromCityID)
	if from == nil {
		return nil, fmt.Errorf("%w: city %s", ErrInvalidCommand, fromCityID)
	}
	if from.OwnerID != s.PlayerID {
		return nil, ErrCityNotPlayable
	}
	if len(generalIDs) == 0 {
		return nil, ErrGeneralNotReady
	}
	if len(generalIDs) > maxBattleGenerals {
		return nil, fmt.Errorf("%w: too many generals", ErrInvalidCommand)
	}
	if money < 0 || food < 0 {
		return nil, fmt.Errorf("%w: negative battle supplies", ErrInvalidCommand)
	}
	if food <= 0 {
		return nil, ErrBattleNoFood
	}
	if remainingFood < 0 || remainingFood > food {
		return nil, fmt.Errorf("%w: invalid remaining food", ErrInvalidCommand)
	}
	if from.Money < money {
		return nil, fmt.Errorf("%w: not enough money", ErrInvalidCommand)
	}
	if from.Food < food {
		return nil, fmt.Errorf("%w: not enough food", ErrInvalidCommand)
	}

	target := s.findCity(targetCityID)
	if target == nil {
		return nil, ErrBattleTargetNoCity
	}
	if target.OwnerID == s.PlayerID {
		return nil, ErrBattleSameOwner
	}
	if !s.isAdjacent(fromCityID, targetCityID) {
		return nil, ErrBattleNotAdjacent
	}

	seen := map[string]bool{}
	generals := make([]*General, 0, len(generalIDs))
	for _, generalID := range generalIDs {
		if seen[generalID] {
			continue
		}
		seen[generalID] = true
		general := s.findGeneral(generalID)
		if general == nil || general.CityID != fromCityID || general.OwnerID != s.PlayerID || general.Captive {
			return nil, ErrGeneralNotReady
		}
		if general.Stamina < battleStaminaCost {
			return nil, ErrGeneralNotReady
		}
		if general.Soldiers <= 0 {
			return nil, ErrBattleNoSoldiers
		}
		generals = append(generals, general)
	}
	if len(generals) == 0 {
		return nil, ErrGeneralNotReady
	}

	from.Money -= money
	from.Food -= food
	if target.OwnerID == "neutral" {
		return s.occupyEmptyCity(from, generals, target, money, food, remainingFood, fieldAdvantage), nil
	}
	return s.resolvePlannedAttack(from, generals, target, money, food, remainingFood, fieldAdvantage), nil
}

// resolveAttack performs the actual siege math and mutation for any attacker
// (player or AI). Callers are responsible for validating ownership/adjacency
// and ensuring the general can act. The conquered city is transferred to the
// attacking general's owner.
//
// Now includes faithful port of C source FgtCount.c damage formula:
//
//	at = Force * (Level+10) * AtkModulus[armsType]
//	df = IQ * (Level+10) * DfModulus[armsType]
//	hurt = (at/df) * (arms/8) * SubduModu[atkArms][defArms] + 10
func (s *GameState) resolveAttack(from *City, general *General, target *City) *BattleOutcome {
	general.Stamina -= battleStaminaCost

	// Attacker strength: use full damage formula from FgtCount.c
	atkArmsType := armsTypeToInt(general.ArmsType)
	atkForce := float64(general.Force) * float64(general.Level+10) * atkModulus[atkArmsType]
	defArmsType := armsTypeToInt(s.getDefenderArmsType(target.ID))
	defIQ := float64(s.getDefenderIQ(target.ID)) * 10 // simplified: use avg IQ * (Level+10) approx

	attackArms := general.Soldiers + general.Soldiers*general.Force/200
	defendArms := s.cityTroops(target.ID)
	// Defender morale bonus from city devotion (民心向背).
	defendArms += defendArms * target.PeopleDevotion / 200
	defenderRulerID := target.OwnerID
	defenderGenerals := s.activeGeneralNamesInCity(target.ID)

	randv := battleRand(101)
	won := resolveBattleWon(attackArms, defendArms, from.Food, target.Food, randv)

	// Calculate battle damage with subdue matrix if battle occurs
	var rawDamage int
	if won {
		rawDamage = s.calculateBattleDamage(general, target, atkForce, defIQ, atkArmsType, defArmsType)
	} else {
		rawDamage = s.calculateBattleDamage(general, target, atkForce, defIQ, atkArmsType, defArmsType)
		// Attacker loses more when defeated
		rawDamage = rawDamage * 3 / 2
	}

	outcome := &BattleOutcome{
		FromCityID:        from.ID,
		FromCityName:      from.Name,
		TargetCityID:      target.ID,
		TargetCityName:    target.Name,
		GeneralID:         general.ID,
		GeneralName:       general.Name,
		GeneralIDs:        []string{general.ID},
		GeneralNames:      []string{general.Name},
		AttackerRulerID:   general.OwnerID,
		AttackerRulerName: s.rulerName(general.OwnerID),
		DefenderRulerID:   defenderRulerID,
		DefenderRulerName: s.rulerName(defenderRulerID),
		DefenderGenerals:  defenderGenerals,
		Food:              from.Food,
		AttackPower:       attackArms,
		DefensePower:      defendArms,
		Won:               won,
	}

	defenderLevel := s.getDefenderLevel(target.ID)
	if won {
		// Attacker loses a modest share; defender is routed.
		attackerLoss := minInt(general.Soldiers, general.Soldiers*(20+randv/4)/100)
		general.Soldiers -= attackerLoss
		outcome.AttackerLosses = attackerLoss
		outcome.DefenderLosses = s.routDefenders(target.ID)
		// 经验结算：胜利获得 sqrt(伤害)/4 + 击杀奖励
		exp := battleExp(rawDamage, general.Level, defenderLevel)
		exp += killBonusExp(general.Level, defenderLevel)
		s.awardBattleExperience(general, exp, outcome)
		outcome.CapturedGenerals = s.captureCity(target, general)
		outcome.Captured = true
		outcome.Message = fmt.Sprintf("%s军 %s 自 %s 攻克 %s，损兵%d，歼敌%d%s。",
			outcome.AttackerRulerName,
			general.Name,
			from.Name,
			target.Name,
			outcome.AttackerLosses,
			outcome.DefenderLosses,
			capturedSuffix(outcome.CapturedGenerals),
		)
	} else {
		// Attacker is repelled with heavier losses; defender takes light losses.
		attackerLoss := minInt(general.Soldiers, general.Soldiers*(40+randv/3)/100)
		general.Soldiers -= attackerLoss
		outcome.AttackerLosses = attackerLoss
		outcome.DefenderLosses = s.lightDefenderLosses(target.ID, randv)
		// 失败也获得少量经验（基于造成的伤害）
		exp := battleExp(rawDamage, general.Level, defenderLevel)
		s.awardBattleExperience(general, exp, outcome)
		outcome.Message = fmt.Sprintf("%s军 %s 自 %s 进攻 %s 失利，损兵%d，守军损%d。",
			outcome.AttackerRulerName,
			general.Name,
			from.Name,
			target.Name,
			outcome.AttackerLosses,
			outcome.DefenderLosses,
		)
	}

	s.prependLog(outcome.Message)
	return outcome
}

// awardBattleExperience grants experience to a general and records any level-ups
// in the outcome. Logs a notice when the general levels up.
func (s *GameState) awardBattleExperience(general *General, exp int, outcome *BattleOutcome) {
	outcome.ExperienceGained += exp
	levels := general.gainExperience(exp)
	if levels > 0 {
		outcome.LevelUps = append(outcome.LevelUps, general.Name)
		s.prependLog(fmt.Sprintf("%s 经验充足，晋升至 %d 级！", general.Name, general.Level))
	}
}

// awardPlannedExperience distributes battle experience across a group of
// attacking generals, using enemy losses as the equivalent damage value.
// When `killed` is true, each general also receives a kill bonus.
func (s *GameState) awardPlannedExperience(generals []*General, enemyLosses, defenderLevel int, killed bool, outcome *BattleOutcome) {
	if len(generals) == 0 {
		return
	}
	// 伤害按参战武将均摊，避免人多时单将经验过高
	share := enemyLosses / len(generals)
	for _, g := range generals {
		exp := battleExp(share, g.Level, defenderLevel)
		if killed {
			exp += killBonusExp(g.Level, defenderLevel)
		}
		s.awardBattleExperience(g, exp, outcome)
	}
}

// getDefenderLevel returns the level of the lead defender (highest level) in a city.
func (s *GameState) getDefenderLevel(cityID string) int {
	best := 1
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.CityID == cityID && !g.Captive && g.Soldiers > 0 {
			if g.Level > best {
				best = g.Level
			}
		}
	}
	return best
}

func (s *GameState) resolvePlannedAttack(from *City, generals []*General, target *City, money, food, remainingFood, fieldAdvantage int) *BattleOutcome {
	for _, general := range generals {
		general.Stamina -= battleStaminaCost
	}

	attackArms := plannedAttackPower(generals)
	if money > 0 {
		attackArms += minInt(money*4, maxInt(120, attackArms/3))
	}
	fieldAdvantage = clampInt(fieldAdvantage, -35, 45)
	if fieldAdvantage != 0 {
		attackArms += attackArms * fieldAdvantage / 100
	}
	attackArms = maxInt(0, attackArms)

	defendArms := s.cityTroops(target.ID)
	defendArms += defendArms * target.PeopleDevotion / 200
	defenderRulerID := target.OwnerID
	defenderGenerals := s.activeGeneralNamesInCity(target.ID)
	generalIDs := battleGeneralIDs(generals)
	generalNames := battleGeneralNames(generals)
	lead := generals[0]

	randv := battleRand(101)
	won := resolveBattleWon(attackArms, defendArms, remainingFood, target.Food, randv)

	outcome := &BattleOutcome{
		FromCityID:        from.ID,
		FromCityName:      from.Name,
		TargetCityID:      target.ID,
		TargetCityName:    target.Name,
		GeneralID:         lead.ID,
		GeneralName:       lead.Name,
		GeneralIDs:        generalIDs,
		GeneralNames:      generalNames,
		AttackerRulerID:   lead.OwnerID,
		AttackerRulerName: s.rulerName(lead.OwnerID),
		DefenderRulerID:   defenderRulerID,
		DefenderRulerName: s.rulerName(defenderRulerID),
		DefenderGenerals:  defenderGenerals,
		Money:             money,
		Food:              food,
		RemainingFood:     remainingFood,
		FieldAdvantage:    fieldAdvantage,
		AttackPower:       attackArms,
		DefensePower:      defendArms,
		Won:               won,
	}

	totalSoldiers := totalGeneralSoldiers(generals)
	forceLabel := battleForceLabel(generalNames)
	defenderLevel := s.getDefenderLevel(target.ID)
	if won {
		attackerLoss := minInt(totalSoldiers, totalSoldiers*(20+randv/4)/100)
		outcome.AttackerLosses = applyGeneralLosses(generals, attackerLoss)
		outcome.DefenderLosses = s.routDefenders(target.ID)
		// 经验结算：以歼敌数作为等效伤害，每位参战武将分得经验+击杀奖励
		s.awardPlannedExperience(generals, outcome.DefenderLosses, defenderLevel, true, outcome)
		outcome.CapturedGenerals = s.captureCityWithAttackers(target, lead.OwnerID, generals)
		outcome.Captured = true
		outcome.Message = fmt.Sprintf("%s军 %s 自 %s 携金%d粮%d 攻克 %s，损兵%d，歼敌%d%s。",
			outcome.AttackerRulerName,
			forceLabel,
			from.Name,
			money,
			food,
			target.Name,
			outcome.AttackerLosses,
			outcome.DefenderLosses,
			capturedSuffix(outcome.CapturedGenerals),
		)
	} else {
		attackerLoss := minInt(totalSoldiers, totalSoldiers*(40+randv/3)/100)
		outcome.AttackerLosses = applyGeneralLosses(generals, attackerLoss)
		outcome.DefenderLosses = s.lightDefenderLosses(target.ID, randv)
		// 失败时按守军损失分得少量经验，无击杀奖励
		s.awardPlannedExperience(generals, outcome.DefenderLosses, defenderLevel, false, outcome)
		outcome.Message = fmt.Sprintf("%s军 %s 自 %s 进攻 %s 失利，耗金%d粮%d，损兵%d，守军损%d。",
			outcome.AttackerRulerName,
			forceLabel,
			from.Name,
			target.Name,
			money,
			food,
			outcome.AttackerLosses,
			outcome.DefenderLosses,
		)
	}

	applyBattleCityAftermath(target, remainingFood)
	s.prependLog(outcome.Message)
	return outcome
}

func (s *GameState) occupyEmptyCity(from *City, generals []*General, target *City, money, food, remainingFood, fieldAdvantage int) *BattleOutcome {
	lead := generals[0]
	for _, general := range generals {
		general.CityID = target.ID
		general.Captive = false
	}
	target.OwnerID = lead.OwnerID
	generalIDs := battleGeneralIDs(generals)
	generalNames := battleGeneralNames(generals)
	fieldAdvantage = clampInt(fieldAdvantage, -35, 45)
	outcome := &BattleOutcome{
		Won:               true,
		FromCityID:        from.ID,
		FromCityName:      from.Name,
		TargetCityID:      target.ID,
		TargetCityName:    target.Name,
		GeneralID:         lead.ID,
		GeneralName:       lead.Name,
		GeneralIDs:        generalIDs,
		GeneralNames:      generalNames,
		AttackerRulerID:   lead.OwnerID,
		AttackerRulerName: s.rulerName(lead.OwnerID),
		DefenderRulerID:   "neutral",
		DefenderRulerName: s.rulerName("neutral"),
		Money:             money,
		Food:              food,
		RemainingFood:     remainingFood,
		FieldAdvantage:    fieldAdvantage,
		AttackPower:       plannedAttackPower(generals),
		Captured:          true,
		Message: fmt.Sprintf("%s军 %s 自 %s 入驻空城 %s。",
			s.rulerName(lead.OwnerID),
			battleForceLabel(generalNames),
			from.Name,
			target.Name,
		),
	}
	s.prependLog(outcome.Message)
	return outcome
}

func (s *GameState) rulerName(ownerID string) string {
	if ownerID == "" {
		return "无主"
	}
	for _, ruler := range s.Rulers {
		if ruler.ID == ownerID {
			if ruler.Name != "" {
				return ruler.Name
			}
			return ruler.ID
		}
	}
	if ownerID == "neutral" {
		return "空城"
	}
	return ownerID
}

func (s *GameState) activeGeneralNamesInCity(cityID string) []string {
	names := make([]string, 0)
	for i := range s.Generals {
		general := &s.Generals[i]
		if general.CityID == cityID && !general.Captive {
			names = append(names, general.Name)
		}
	}
	return names
}

func capturedSuffix(generals []string) string {
	switch len(generals) {
	case 0:
		return ""
	case 1:
		return "，俘虏" + generals[0]
	default:
		return fmt.Sprintf("，俘虏%s等%d将", generals[0], len(generals))
	}
}

func plannedAttackPower(generals []*General) int {
	total := 0
	for _, general := range generals {
		total += general.Soldiers + general.Soldiers*general.Force/200
	}
	return total
}

func totalGeneralSoldiers(generals []*General) int {
	total := 0
	for _, general := range generals {
		total += general.Soldiers
	}
	return total
}

func applyGeneralLosses(generals []*General, requestedLoss int) int {
	total := totalGeneralSoldiers(generals)
	if total <= 0 || requestedLoss <= 0 {
		return 0
	}
	remaining := minInt(total, requestedLoss)
	applied := 0
	for i, general := range generals {
		loss := requestedLoss * general.Soldiers / total
		if i == len(generals)-1 {
			loss = remaining
		}
		loss = minInt(general.Soldiers, loss)
		general.Soldiers -= loss
		applied += loss
		remaining -= loss
		if remaining <= 0 {
			break
		}
	}
	return applied
}

func battleGeneralIDs(generals []*General) []string {
	out := make([]string, 0, len(generals))
	for _, general := range generals {
		out = append(out, general.ID)
	}
	return out
}

func battleGeneralNames(generals []*General) []string {
	out := make([]string, 0, len(generals))
	for _, general := range generals {
		out = append(out, general.Name)
	}
	return out
}

func battleForceLabel(names []string) string {
	switch len(names) {
	case 0:
		return "无名军"
	case 1:
		return names[0]
	default:
		return fmt.Sprintf("%s等%d将", names[0], len(names))
	}
}

// routDefenders wipes most of a routed city's defenders and returns the losses.
func (s *GameState) routDefenders(cityID string) int {
	city := s.findCity(cityID)
	losses := 0
	if city != nil {
		losses += city.Garrison
		city.Garrison = 0
	}
	for i := range s.Generals {
		if s.Generals[i].CityID == cityID && !s.Generals[i].Captive {
			losses += s.Generals[i].Soldiers
			s.Generals[i].Soldiers = 0
		}
	}
	return losses
}

// lightDefenderLosses applies a small casualty to defenders after a failed siege.
func (s *GameState) lightDefenderLosses(cityID string, randv int) int {
	losses := 0
	for i := range s.Generals {
		if s.Generals[i].CityID == cityID && !s.Generals[i].Captive {
			loss := s.Generals[i].Soldiers * (5 + randv/10) / 100
			s.Generals[i].Soldiers -= loss
			losses += loss
		}
	}
	return losses
}

// captureCity transfers a city to the attacker, moves the conquering general in,
// and captures the defeated officers.
func (s *GameState) captureCity(target *City, general *General) []string {
	return s.captureCityWithAttackers(target, general.OwnerID, []*General{general})
}

func (s *GameState) captureCityWithAttackers(target *City, ownerID string, attackers []*General) []string {
	capturedGenerals := []string{}
	attackerIDs := map[string]bool{}
	target.OwnerID = ownerID
	for _, attacker := range attackers {
		attackerIDs[attacker.ID] = true
		attacker.CityID = target.ID
		attacker.Captive = false
	}
	for i := range s.Generals {
		defender := &s.Generals[i]
		if defender.CityID != target.ID || attackerIDs[defender.ID] || defender.OwnerID == ownerID {
			continue
		}
		defender.OwnerID = ownerID
		defender.Captive = true
		defender.Soldiers = 0
		defender.Stamina = 0
		defender.Loyalty = minInt(defender.Loyalty, 45)
		if defender.Loyalty <= 0 {
			defender.Loyalty = 30
		}
		capturedGenerals = append(capturedGenerals, defender.Name)
	}
	// Defending generals that survived (none after a rout) would defect; here the
	// city's leaderless remnants surrender, leaving the garrison emptied.
	target.Garrison = 0
	return capturedGenerals
}

func applyBattleCityAftermath(city *City, attackerRemainingFood int) {
	city.Farming -= city.Farming / 20
	city.Commerce -= city.Commerce / 20
	city.Money -= city.Money / 20
	city.PeopleDevotion -= city.PeopleDevotion / 10
	city.Food = maxInt(0, city.Food+attackerRemainingFood)
}

func clampInt(value, minValue, maxValue int) int {
	return maxInt(minValue, minInt(maxValue, value))
}

// getDefenderArmsType returns the primary arms type of defenders in a city
func (s *GameState) getDefenderArmsType(cityID string) string {
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.CityID == cityID && !g.Captive && g.Soldiers > 0 {
			return g.ArmsType
		}
	}
	return "步兵" // default
}

// getDefenderIQ returns the average IQ of defenders in a city
func (s *GameState) getDefenderIQ(cityID string) int {
	totalIQ := 0
	count := 0
	for i := range s.Generals {
		g := &s.Generals[i]
		if g.CityID == cityID && !g.Captive && g.Soldiers > 0 {
			totalIQ += g.Intellect
			count++
		}
	}
	if count == 0 {
		return 50
	}
	return totalIQ / count
}

// calculateBattleDamage ports the FgtCount.c CountAtkHurt formula:
//
//	hurt = (at/df) * (arms/8) * SubduModu[atkArms][defArms] + 10
//
// This gives the raw damage before applying loss ratios.
func (s *GameState) calculateBattleDamage(attacker *General, target *City, atkForce, defDF float64, atkArmsType, defArmsType int) int {
	if defDF <= 0 {
		defDF = 1.0 // avoid divide by zero
	}
	rawHurt := (atkForce / defDF) * (float64(attacker.Soldiers) / 8.0)
	// Apply subdue matrix (兵种相克)
	subdue := subdueModu[atkArmsType][defArmsType]
	rawHurt *= subdue
	rawHurt += 10 // +10 prevents stalemate when both sides have tiny armies
	return int(rawHurt)
}
