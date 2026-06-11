package game

import "fmt"

func princeScene(turn int, state *GameState) *Scene {
	scenes := []Scene{
		{
			ID:    "birth-omen",
			Title: "紫宸宫中的啼哭",
			Year:  state.Dynasty.Era,
			Mood:  "启蒙",
			Art:   sceneArt(state, 0),
			Body:  fmt.Sprintf("你出生在%s的风雪清晨。%s 宫中每一句吉言都可能变成刀锋。", state.Dynasty.Name, state.Dynasty.Background),
			Choices: []Choice{
				{ID: "grab-scroll", Text: "抓起案上的竹简", Detail: "让太傅记住你早慧的一面。", Domain: DomainStory, Effects: Effects{Learning: 8, Legitimacy: 2, Health: -1}, Outcome: "你咿呀抓住竹简，满殿笑声里，多了一个“好学皇子”的传闻。"},
				{ID: "smile-consort", Text: "向皇后展露笑容", Detail: "讨得中宫欢心，但母妃心中不安。", Domain: DomainStory, Effects: Effects{Charisma: 7, Influence: 4}, Outcome: "皇后轻抚你的额头，宫人们很快学会了对你多行半礼。"},
				{ID: "cry-loudly", Text: "放声大哭震住众人", Detail: "生命力旺盛，也显得倔强难驯。", Domain: DomainStory, Effects: Effects{Health: 8, Martial: 3, Charisma: -1}, Outcome: "哭声穿过朱门，老内侍低声说：这孩子有帝王的肺腑。"},
			},
		},
		{
			ID: "study-yard", Title: "东宫书院", Year: "六岁", Mood: "养成", Art: sceneArt(state, 1),
			Body: "皇子们第一次同席读书。太傅问你：治国先治什么？兄弟们都在等你出错。",
			Choices: []Choice{
				{ID: "answer-people", Text: "答：先安百姓", Detail: "赢得清流赞许。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 4, Legitimacy: 3}, Outcome: "太傅捻须点头，清流大臣开始把你的名字写进密札。"},
				{ID: "answer-army", Text: "答：先强兵甲", Detail: "让武臣另眼相看。", Domain: DomainStory, Effects: Effects{Martial: 8, Influence: 2}, Outcome: "武学师傅当场请你试弓，你拉不开弓，却拉来了将门的好感。"},
				{ID: "answer-father", Text: "答：先顺父皇", Detail: "谨慎稳妥，少惹麻烦。", Domain: DomainStory, Effects: Effects{Influence: 6, Legitimacy: 2, Learning: 2}, Outcome: "父皇听闻后没有评价，但赏下了一方端砚。宫中人都懂沉默的分量。"},
			},
		},
		{
			ID: "winter-hunt", Title: "皇家冬狩", Year: "十岁", Mood: "锋芒", Art: sceneArt(state, 2),
			Body: "猎场上，三皇子故意惊马。你摔在雪里，侍卫们一瞬间不敢动。",
			Choices: []Choice{
				{ID: "mount-again", Text: "忍痛重新上马", Detail: "以勇气换取威望。", Domain: DomainStory, Effects: Effects{Martial: 9, Legitimacy: 4, Health: -4}, Outcome: "你带伤上马，雪原上响起军士的喝彩。三皇子的笑容僵住了。"},
				{ID: "protect-servant", Text: "先扶起被撞倒的小内侍", Detail: "仁名会悄悄传开。", Domain: DomainStory, Effects: Effects{Charisma: 8, Populace: 2, Legitimacy: 1}, Outcome: "一个小小内侍救不了天下，却能让天下相信你会低头看人。"},
				{ID: "accuse-brother", Text: "当众指认三皇子", Detail: "直接开战，风险很高。", Domain: DomainStory, Effects: Effects{Influence: 8, Martial: 3, Stability: -3}, Outcome: "猎场瞬间安静。你赢得了一批拥护者，也让夺嫡提早见血。"},
			},
		},
		{
			ID: "flood-memorial", Title: "南河急报", Year: "十四岁", Mood: "试政", Art: sceneArt(state, 3),
			Body: "南河决堤，朝堂争论赈灾银从何处来。父皇将奏章推到你面前，要你试拟朱批。",
			Choices: []Choice{
				{ID: "open-granary", Text: "开仓赈济，严查贪墨", Detail: "仁政与吏治并行。", Domain: DomainStory, Effects: Effects{Learning: 7, Charisma: 5, Reform: 4}, Outcome: "你的朱批被贴到灾区驿站，流民第一次知道京中还有人惦记他们。"},
				{ID: "borrow-merchants", Text: "向皇商借银平灾", Detail: "来钱快，但商人会记账。", Domain: DomainStory, Effects: Effects{Treasury: 6, Influence: 5, Legitimacy: -2}, Outcome: "堤坝保住了，皇商也在你的未来里占了一席。"},
				{ID: "send-army", Text: "调禁军协助筑堤", Detail: "效率高，武臣得势。", Domain: DomainStory, Effects: Effects{Martial: 5, Army: 4, Influence: 3}, Outcome: "铁甲入泥，军令比公文更快。百姓记住了旗号，也记住了你。"},
			},
		},
		{
			ID: "succession-night", Title: "烛影摇红", Year: "十六岁", Mood: "夺嫡", Art: sceneArt(state, 4),
			Body: fmt.Sprintf("父皇病重，诸王入宫。你手中有学识 %d、武略 %d、声望 %d。最后一夜，谁先动，谁就可能坐上明日的朝堂。", state.Stats.Learning, state.Stats.Martial, state.Stats.Legitimacy),
			Choices: []Choice{
				{ID: "secure-edict", Text: "请太傅与中书共同护诏", Detail: "以制度和名分夺位。", Domain: DomainCourt, Effects: Effects{Legitimacy: 10, Stability: 5, Influence: 4}, Outcome: "玉玺落印，群臣跪伏。你没有拔剑，却让刀兵失去了名义。"},
				{ID: "control-guards", Text: "联络禁军封锁宫门", Detail: "以速度换取皇位。", Domain: DomainMilitary, Effects: Effects{Martial: 8, Influence: 8, Stability: -4}, Outcome: "宫门在夜色中合拢。天亮时，反对你的人已经错过了进宫的时辰。"},
				{ID: "appeal-clans", Text: "向宗室许诺共治", Detail: "稳住人心，但埋下掣肘。", Domain: DomainDiplomacy, Effects: Effects{Charisma: 7, Legitimacy: 6, Reform: -2}, Outcome: "宗室长者替你说了第一句话。你得到了皇位，也得到了许多双盯着你的眼睛。"},
			},
		},
	}
	return cloneScene(scenes[min(turn, len(scenes)-1)])
}

func emperorScene(s *GameState) *Scene {
	year := fmt.Sprintf("登基%d年 · %s", max(1, s.ReignYear), s.Season)
	return &Scene{
		ID:    fmt.Sprintf("court-%d", s.Turn),
		Title: "太和朝议",
		Year:  year,
		Mood:  emperorMood(s.Stats),
		Art:   emperorSceneArt(s),
		Body:  crisisLine(s) + " 六部、边军、宗室、清流与商帮都在等你落子。选择不是单纯加减数值，派系与省份会记住你的每一道旨意。",
		Choices: []Choice{
			{ID: fmt.Sprintf("relief-%d", s.Turn), Text: "户部开仓，巡抚赈济", Detail: "民政线：压灾情、保民心，但消耗粮银。", Domain: DomainDomestic, Effects: Effects{Treasury: -10, Grain: -8, Populace: 12, Stability: 5, Legitimacy: 2}, Outcome: "粥棚沿官道铺开，流民队伍短了，户部账册却重得像铁。"},
			{ID: fmt.Sprintf("tax-%d", s.Turn), Text: "重估田亩，整顿盐铁", Detail: "财政线：增加国库，激怒豪强商帮。", Domain: DomainEconomy, Effects: Effects{Treasury: 18, Populace: -4, Stability: -4, Reform: 3}, Outcome: "银车入库，盐引重发。商帮笑着谢恩，袖中却攥紧了旧账。"},
			{ID: fmt.Sprintf("train-%d", s.Turn), Text: "拨银练兵，轮戍边镇", Detail: "军务线：提升军力、压边患，耗费国库。", Domain: DomainMilitary, Effects: Effects{Treasury: -12, Army: 15, BorderThreat: -13, Martial: 1}, Outcome: "军营号角重新响亮，边将呈上的地图终于少了几个红圈。"},
			{ID: fmt.Sprintf("envoy-%d", s.Turn), Text: "遣使联姻，分化诸邦", Detail: "外交线：缓冲战争，稳住宗室和外邦。", Domain: DomainDiplomacy, Effects: Effects{Diplomacy: 13, BorderThreat: -8, Treasury: -4, Stability: 1}, Outcome: "使团携金册出关，远方可汗收下礼物，也收起了一半刀锋。"},
			{ID: fmt.Sprintf("reform-%d", s.Turn), Text: "设考成法，裁撤冗官", Detail: "新法线：长期强国，短期引爆党争。", Domain: DomainReform, Effects: Effects{Reform: 12, Treasury: 5, Stability: -7, Populace: 3, Legitimacy: 2}, Outcome: "新法像春雷落进官场。有人称颂清明，也有人开始深夜串门。"},
			{ID: fmt.Sprintf("spy-%d", s.Turn), Text: "开缇骑密档，夜审朋党", Detail: "暗线：削派系权势，但损声望与稳定。", Domain: DomainIntrigue, Effects: Effects{Influence: 7, Stability: -6, Legitimacy: -4, Health: -2}, Outcome: "宫门后的灯亮到三更，第二天朝会上少了几张熟悉的脸。"},
			{ID: fmt.Sprintf("banquet-%d", s.Turn), Text: "大宴群臣，粉饰太平", Detail: "宫廷线：短期安抚派系，长期损害国本。", Domain: DomainCourt, Effects: Effects{Treasury: -9, Grain: -5, Populace: -7, Stability: -5, BorderThreat: 5, Health: -2}, Outcome: "钟鼓响彻宫城，杯盏遮住了奏章。散席后，问题仍在殿外等你。"},
		},
	}
}

func sceneArt(s *GameState, index int) string {
	if s == nil || len(s.Assets.SceneGallery) == 0 {
		return "/assets/palace-hero.png"
	}
	index %= len(s.Assets.SceneGallery)
	if index < 0 {
		index += len(s.Assets.SceneGallery)
	}
	return s.Assets.SceneGallery[index]
}

func emperorSceneArt(s *GameState) string {
	switch {
	case s.Stats.BorderThreat >= 70:
		return sceneArt(s, 14)
	case s.Crisis.Severity >= 75:
		return sceneArt(s, 22)
	case s.Stats.Reform >= 65:
		return sceneArt(s, 10)
	case s.Stats.Diplomacy >= 72:
		return sceneArt(s, 28)
	case s.Stats.Stability >= 82 && s.Stats.Populace >= 82:
		return sceneArt(s, 29)
	default:
		return sceneArt(s, 5+s.Turn+s.ReignYear)
	}
}
