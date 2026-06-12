package game

func orderLabel(kind OrderKind) string {
	switch kind {
	case OrderRelief:
		return "御令：赈济"
	case OrderGarrison:
		return "御令：驻防"
	case OrderTax:
		return "御令：督税"
	case OrderInspect:
		return "御令：密查"
	case OrderAppease:
		return "御令：安抚"
	case OrderPurge:
		return "御令：削权"
	case OrderCanal:
		return "御令：修渠"
	case OrderTrade:
		return "御令：互市"
	case OrderMobilize:
		return "御令：动员"
	case OrderCampaign:
		return "御令：出征"
	case OrderFortify:
		return "御令：固边"
	case OrderTruce:
		return "御令：议和"
	case OrderAppoint:
		return "御令：任官"
	case OrderDismiss:
		return "御令：罢官"
	case OrderNameHeir:
		return "御令：册储"
	case OrderFavor:
		return "御令：临幸"
	case OrderMarriage:
		return "御令：联姻"
	case OrderFundProject:
		return "御令：营造"
	case OrderEnactPolicy:
		return "御令：国策"
	case OrderEmbassy:
		return "御令：遣使"
	case OrderTreaty:
		return "御令：盟约"
	case OrderInvestigatePlot:
		return "御令：侦缉"
	case OrderSuppressPlot:
		return "御令：平谋"
	case OrderEducateHeir:
		return "御令：训储"
	case OrderOpenTrial:
		return "御令：明审"
	case OrderClemency:
		return "御令：宽赦"
	case OrderCensorRumor:
		return "御令：禁谣"
	case OrderProclaimVerdict:
		return "御令：宣判"
	default:
		return "御令"
	}
}
