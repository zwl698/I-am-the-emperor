package game

import (
	"errors"
	"fmt"
	"strings"
)

type ActionKind string

const (
	ActionMapAllocation ActionKind = "map_allocation"
	ActionWarTactic     ActionKind = "war_tactic"
	ActionTrialMove     ActionKind = "trial_move"
	ActionOfficeAssign  ActionKind = "office_assign"
	ActionHeirLesson    ActionKind = "heir_lesson"
	ActionEnvoyMission  ActionKind = "envoy_mission"
)

type ActionRequest struct {
	Kind   ActionKind `json:"kind"`
	Target string     `json:"target"`
	Mode   string     `json:"mode"`
	Amount int        `json:"amount,omitempty"`
}

type ActionDefinition struct {
	Kind        ActionKind `json:"kind"`
	Mode        string     `json:"mode"`
	Label       string     `json:"label"`
	Panel       string     `json:"panel"`
	Domain      Domain     `json:"domain"`
	Description string     `json:"description"`
}

func ActionCatalog() []ActionDefinition {
	return []ActionDefinition{
		{Kind: ActionMapAllocation, Mode: "relief", Label: "赈灾开仓", Panel: "山河调度", Domain: DomainDomestic, Description: "压低灾害与民变风险，消耗国库和粮草。"},
		{Kind: ActionMapAllocation, Mode: "garrison", Label: "驻军固防", Panel: "山河调度", Domain: DomainMilitary, Description: "提升地方防务，缓解边患。"},
		{Kind: ActionMapAllocation, Mode: "tax", Label: "清丈征税", Panel: "山河调度", Domain: DomainEconomy, Description: "快速补财政，牺牲地方秩序。"},
		{Kind: ActionMapAllocation, Mode: "inspect", Label: "巡按暗访", Panel: "山河调度", Domain: DomainIntrigue, Description: "查地方弊政，稳秩序并增皇权。"},
		{Kind: ActionMapAllocation, Mode: "canal", Label: "开渠转运", Panel: "山河调度", Domain: DomainReform, Description: "投资长期粮道，改善灾害恢复。"},
		{Kind: ActionMapAllocation, Mode: "trade", Label: "开放互市", Panel: "山河调度", Domain: DomainDiplomacy, Description: "增强财政与外交，也会引入边境试探。"},
		{Kind: ActionWarTactic, Mode: "mobilize", Label: "拨粮练兵", Panel: "兵棋沙盘", Domain: DomainMilitary, Description: "提高补给士气，准备后续攻势。"},
		{Kind: ActionWarTactic, Mode: "campaign", Label: "出塞决战", Panel: "兵棋沙盘", Domain: DomainMilitary, Description: "推进战役进度，消耗军粮。"},
		{Kind: ActionWarTactic, Mode: "fortify", Label: "筑垒固边", Panel: "兵棋沙盘", Domain: DomainMilitary, Description: "降低敌势并强化北境防线。"},
		{Kind: ActionWarTactic, Mode: "truce", Label: "设帐议和", Panel: "兵棋沙盘", Domain: DomainDiplomacy, Description: "用外交换取边患暂缓。"},
		{Kind: ActionTrialMove, Mode: "open_trial", Label: "开堂会审", Panel: "三司会审", Domain: DomainIntrigue, Description: "公开审案并推进刑狱线索。"},
		{Kind: ActionTrialMove, Mode: "clemency", Label: "恩旨缓刑", Panel: "三司会审", Domain: DomainCourt, Description: "用仁政换舆论，但可能纵容余党。"},
		{Kind: ActionTrialMove, Mode: "censor_rumor", Label: "禁谣清议", Panel: "三司会审", Domain: DomainIntrigue, Description: "压制流言，增加言路压力。"},
		{Kind: ActionTrialMove, Mode: "proclaim_verdict", Label: "明诏定案", Panel: "三司会审", Domain: DomainReform, Description: "用判词建立规则与威信。"},
		{Kind: ActionOfficeAssign, Mode: "appoint", Label: "任命官员", Panel: "六部任免", Domain: DomainCourt, Description: "把群臣能力转化为部门权威。"},
		{Kind: ActionOfficeAssign, Mode: "dismiss", Label: "罢免官员", Panel: "六部任免", Domain: DomainCourt, Description: "清理错配职位，震动派系。"},
		{Kind: ActionHeirLesson, Mode: "study", Label: "经筵讲学", Panel: "东宫培养", Domain: DomainCourt, Description: "提升皇嗣才学和正统形象。"},
		{Kind: ActionHeirLesson, Mode: "martial", Label: "骑射校猎", Panel: "东宫培养", Domain: DomainMilitary, Description: "训练胆略，争取武臣背书。"},
		{Kind: ActionHeirLesson, Mode: "ritual", Label: "太庙习礼", Panel: "东宫培养", Domain: DomainCourt, Description: "压低储位争议，稳宗室观感。"},
		{Kind: ActionEnvoyMission, Mode: "embassy", Label: "遣使修好", Panel: "鸿胪外交", Domain: DomainDiplomacy, Description: "提升邦交关系并打探虚实。"},
		{Kind: ActionEnvoyMission, Mode: "treaty", Label: "缔结盟约", Panel: "鸿胪外交", Domain: DomainDiplomacy, Description: "以资源和承诺换长期边境缓冲。"},
	}
}

func (s *GameState) ApplyAction(req ActionRequest) (*Resolution, error) {
	if s == nil {
		return nil, errors.New("game state is nil")
	}
	order, err := req.toOrderRequest()
	if err != nil {
		return nil, err
	}
	return s.ApplyOrder(order)
}

func (req ActionRequest) toOrderRequest() (OrderRequest, error) {
	req.Kind = ActionKind(strings.TrimSpace(string(req.Kind)))
	req.Target = strings.TrimSpace(req.Target)
	req.Mode = strings.TrimSpace(req.Mode)
	if req.Kind == "" {
		return OrderRequest{}, errors.New("missing action kind")
	}
	if req.Target == "" {
		return OrderRequest{}, errors.New("missing action target")
	}

	switch req.Kind {
	case ActionMapAllocation:
		return actionOrderFromMode(req.Target, req.Mode, map[string]OrderKind{
			"relief":   OrderRelief,
			"garrison": OrderGarrison,
			"tax":      OrderTax,
			"inspect":  OrderInspect,
			"canal":    OrderCanal,
			"trade":    OrderTrade,
		})
	case ActionWarTactic:
		return actionOrderFromMode(req.Target, req.Mode, map[string]OrderKind{
			"mobilize": OrderMobilize,
			"campaign": OrderCampaign,
			"fortify":  OrderFortify,
			"truce":    OrderTruce,
		})
	case ActionTrialMove:
		return actionOrderFromMode(req.Target, req.Mode, map[string]OrderKind{
			"open_trial":       OrderOpenTrial,
			"clemency":         OrderClemency,
			"censor_rumor":     OrderCensorRumor,
			"proclaim_verdict": OrderProclaimVerdict,
		})
	case ActionOfficeAssign:
		mode := req.Mode
		if mode == "" {
			mode = "appoint"
		}
		return actionOrderFromMode(req.Target, mode, map[string]OrderKind{
			"appoint": OrderAppoint,
			"dismiss": OrderDismiss,
		})
	case ActionHeirLesson:
		mode := req.Mode
		if mode == "" {
			mode = "study"
		}
		return OrderRequest{Kind: OrderEducateHeir, Target: req.Target + ":" + mode}, nil
	case ActionEnvoyMission:
		return actionOrderFromMode(req.Target, req.Mode, map[string]OrderKind{
			"embassy": OrderEmbassy,
			"treaty":  OrderTreaty,
		})
	default:
		return OrderRequest{}, fmt.Errorf("unknown action kind %q", req.Kind)
	}
}

func actionOrderFromMode(target, mode string, modes map[string]OrderKind) (OrderRequest, error) {
	if mode == "" && len(modes) == 1 {
		for _, kind := range modes {
			return OrderRequest{Kind: kind, Target: target}, nil
		}
	}
	kind, ok := modes[mode]
	if !ok {
		return OrderRequest{}, fmt.Errorf("unknown action mode %q", mode)
	}
	return OrderRequest{Kind: kind, Target: target}, nil
}
