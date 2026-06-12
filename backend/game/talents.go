package game

import (
	"fmt"
	"strings"
)

type talentTemplate struct {
	Key         string
	Name        string
	Inspiration string
	Origin      string
	School      string
	Role        string
	Trait       string
	Specialty   Domain
	Loyalty     int
	Ability     int
	Ambition    int
	Integrity   int
	Portrait    string
}

type talentVariant struct {
	Key       string
	Suffix    string
	Role      string
	Trait     string
	Loyalty   int
	Ability   int
	Ambition  int
	Integrity int
	Stress    int
}

func startingTalentPool(dynastyID string) []Minister {
	pool := make([]Minister, 0, len(historicalTalentTemplates())*len(talentVariants()))
	for templateIndex, template := range historicalTalentTemplates() {
		for variantIndex, variant := range talentVariants() {
			pool = append(pool, buildTalent(template, variant, dynastyID, templateIndex, variantIndex))
		}
	}
	return pool
}

func (s *GameState) ensureTalentPool() {
	if len(s.TalentPool) == 0 {
		s.TalentPool = removeCourtTalents(startingTalentPool(s.Dynasty.ID), s.Court)
	}
}

func (s *GameState) recruitTalent(talentID string) (Effects, string, error) {
	s.ensureTalentPool()
	index, ok := s.findTalentIndex(talentID)
	if !ok {
		if _, exists := s.findMinisterIndex(talentID); exists {
			return Effects{}, "", fmt.Errorf("talent %q is already in court", talentID)
		}
		return Effects{}, "", fmt.Errorf("unknown talent target %q", talentID)
	}
	talent := s.TalentPool[index]
	talent.Loyalty = clamp(talent.Loyalty+6, 0, 100)
	talent.Stress = clamp(talent.Stress+4, 0, 100)
	s.Court = append(s.Court, talent)
	s.TalentPool = append(s.TalentPool[:index], s.TalentPool[index+1:]...)
	s.adjustFactionByID(factionForSpecialty(talent.Specialty), 2, 3)
	effects := talentRecruitEffects(talent.Specialty, talent)
	summary := fmt.Sprintf("你下诏征辟%s入朝，%s出身%s，取法%s。%s进入群臣班底，可参与任官、太守与军务调度。",
		talent.Name, talent.Trait, talent.Origin, talent.Inspiration, talent.Name)
	return effects, summary, nil
}

func (s *GameState) findTalentIndex(id string) (int, bool) {
	for i, talent := range s.TalentPool {
		if talent.ID == id {
			return i, true
		}
	}
	return 0, false
}

func removeCourtTalents(pool []Minister, court []Minister) []Minister {
	taken := map[string]bool{}
	for _, minister := range court {
		taken[minister.ID] = true
	}
	filtered := pool[:0]
	for _, talent := range pool {
		if !taken[talent.ID] {
			filtered = append(filtered, talent)
		}
	}
	return filtered
}

func buildTalent(template talentTemplate, variant talentVariant, dynastyID string, templateIndex, variantIndex int) Minister {
	dynastyLoyalty := 0
	dynastyAmbition := 0
	switch dynastyID {
	case "dayin":
		if template.Specialty == DomainMilitary {
			dynastyLoyalty += 5
			dynastyAmbition += 4
		}
	case "jingyao":
		if template.Specialty == DomainEconomy || template.Specialty == DomainDiplomacy {
			dynastyLoyalty += 4
		}
	case "chengping":
		if template.Specialty == DomainReform || template.Specialty == DomainDomestic {
			dynastyAmbition += 5
		}
	case "xuanshuo":
		if template.Specialty == DomainMilitary || template.Specialty == DomainDiplomacy {
			dynastyLoyalty += 4
		}
	}
	name := template.Name + variant.Suffix
	role := template.Role
	if variant.Role != "" {
		role = variant.Role
	}
	trait := template.Trait
	if variant.Trait != "" {
		trait = variant.Trait
	}
	id := fmt.Sprintf("talent-%s-%s", template.Key, variant.Key)
	return Minister{
		ID:          id,
		Name:        name,
		Role:        role,
		Trait:       trait,
		Loyalty:     clamp(template.Loyalty+variant.Loyalty+dynastyLoyalty+(templateIndex+variantIndex)%5-2, 15, 96),
		Ability:     clamp(template.Ability+variant.Ability+(templateIndex*3+variantIndex)%7-3, 28, 99),
		Ambition:    clamp(template.Ambition+variant.Ambition+dynastyAmbition+(templateIndex+variantIndex*2)%9-4, 8, 96),
		Integrity:   clamp(template.Integrity+variant.Integrity+(templateIndex*5+variantIndex)%7-3, 10, 99),
		Stress:      clamp(variant.Stress+(templateIndex+variantIndex)%12, 0, 45),
		Portrait:    template.Portrait,
		Specialty:   template.Specialty,
		Origin:      template.Origin,
		Inspiration: template.Inspiration,
		School:      template.School,
	}
}

func talentVariants() []talentVariant {
	return []talentVariant{
		{Key: "court", Suffix: "·待诏", Role: "翰林待诏", Trait: "稳健", Loyalty: 6, Ability: 0, Ambition: -5, Integrity: 4, Stress: 8},
		{Key: "frontier", Suffix: "·边策", Role: "边镇参议", Trait: "敢断", Loyalty: 1, Ability: 3, Ambition: 4, Integrity: -2, Stress: 12},
		{Key: "reform", Suffix: "·新法", Role: "制置推官", Trait: "变法", Loyalty: -2, Ability: 5, Ambition: 7, Integrity: 1, Stress: 15},
		{Key: "envoy", Suffix: "·远使", Role: "鸿胪客卿", Trait: "通译", Loyalty: 0, Ability: 2, Ambition: 1, Integrity: 0, Stress: 10},
	}
}

func talentRecruitEffects(specialty Domain, talent Minister) Effects {
	bonus := max(1, talent.Ability/35)
	switch specialty {
	case DomainDomestic:
		return Effects{Populace: bonus + 1, Stability: 1}
	case DomainEconomy:
		return Effects{Treasury: bonus + 1, Reform: 1}
	case DomainMilitary:
		return Effects{Army: bonus + 1, BorderThreat: -1, Martial: 1}
	case DomainDiplomacy:
		return Effects{Diplomacy: bonus + 1, BorderThreat: -1}
	case DomainReform:
		return Effects{Reform: bonus + 1, Stability: -1}
	case DomainIntrigue:
		return Effects{Influence: bonus + 1, Stability: -1}
	case DomainCourt:
		return Effects{Influence: 1, Stability: bonus}
	default:
		return Effects{Influence: bonus}
	}
}

func factionForSpecialty(specialty Domain) string {
	switch specialty {
	case DomainMilitary:
		return "border"
	case DomainEconomy, DomainDiplomacy:
		return "merchant"
	case DomainCourt:
		return "clan"
	default:
		return "scholar"
	}
}

func historicalTalentTemplates() []talentTemplate {
	raw := []talentTemplate{
		{"zhang-liang", "张子房", "张良", "中国楚汉", "黄老谋略", "谋臣", "帷幄", DomainCourt, 68, 92, 42, 82, "scholar"},
		{"xiao-he", "萧酂侯", "萧何", "中国楚汉", "律令后勤", "户曹", "持重", DomainEconomy, 72, 88, 36, 86, "minister"},
		{"han-xin", "韩淮阴", "韩信", "中国楚汉", "兵形势", "都督", "奇兵", DomainMilitary, 46, 96, 86, 42, "general"},
		{"chen-ping", "陈曲逆", "陈平", "中国楚汉", "权变谋略", "中书令", "机变", DomainIntrigue, 54, 88, 62, 48, "scholar"},
		{"zhuge-liang", "诸葛武侯", "诸葛亮", "中国三国", "法度屯田", "丞相参军", "谨密", DomainReform, 78, 95, 38, 92, "scholar"},
		{"cao-cao", "曹孟德", "曹操", "中国三国", "军政一体", "行军司马", "雄略", DomainMilitary, 44, 93, 92, 44, "general"},
		{"sun-quan", "孙仲谋", "孙权", "中国三国", "江海经营", "都督府属", "持衡", DomainDiplomacy, 58, 84, 78, 55, "minister"},
		{"zhou-yu", "周公瑾", "周瑜", "中国三国", "水战火攻", "水军都监", "英发", DomainMilitary, 62, 91, 66, 66, "general"},
		{"guan-zhong", "管夷吾", "管仲", "中国春秋", "富国强兵", "计相", "经国", DomainEconomy, 64, 94, 58, 72, "minister"},
		{"shang-yang", "商君", "商鞅", "中国战国", "法家变法", "制置使", "峻法", DomainReform, 42, 93, 74, 64, "minister"},
		{"li-si", "李通古", "李斯", "中国秦汉", "郡县文书", "廷尉属", "苛细", DomainReform, 36, 90, 86, 34, "minister"},
		{"wei-qing", "卫长平", "卫青", "中国汉武", "骑兵远征", "骠骑参军", "沉勇", DomainMilitary, 74, 88, 48, 78, "general"},
		{"huo-qubing", "霍冠军", "霍去病", "中国汉武", "奔袭战", "轻骑都尉", "锐进", DomainMilitary, 62, 94, 70, 66, "general"},
		{"sima-qian", "司马太史", "司马迁", "中国汉代", "史学观察", "起居郎", "通古", DomainCourt, 60, 86, 34, 88, "scholar"},
		{"wang-anshi", "王荆公", "王安石", "中国宋代", "财政新法", "参知政事", "执拗", DomainReform, 52, 92, 68, 82, "minister"},
		{"sima-guang", "司马温公", "司马光", "中国宋代", "保守史鉴", "谏议大夫", "谨厚", DomainCourt, 70, 86, 40, 92, "scholar"},
		{"yue-fei", "岳武穆", "岳飞", "中国宋代", "精忠军纪", "统制官", "忠烈", DomainMilitary, 88, 90, 42, 96, "general"},
		{"zheng-he", "郑三宝", "郑和", "中国明代", "远航外交", "市舶使", "远略", DomainDiplomacy, 76, 86, 54, 74, "envoy"},
		{"yu-qian", "于忠肃", "于谦", "中国明代", "京师保卫", "兵部侍郎", "刚烈", DomainMilitary, 86, 88, 36, 94, "general"},
		{"zhang-juzheng", "张太岳", "张居正", "中国明代", "考成财政", "内阁学士", "整饬", DomainReform, 58, 94, 78, 72, "minister"},
		{"hai-rui", "海刚峰", "海瑞", "中国明代", "清廉监察", "御史", "清峻", DomainIntrigue, 72, 80, 26, 99, "scholar"},
		{"lin-zexu", "林少穆", "林则徐", "中国清代", "禁烟海防", "钦差参赞", "严毅", DomainDiplomacy, 74, 88, 48, 94, "minister"},
		{"zeng-guofan", "曾涤生", "曾国藩", "中国清代", "团练理学", "营务大臣", "坚忍", DomainMilitary, 66, 90, 68, 84, "general"},
		{"li-hongzhang", "李少荃", "李鸿章", "中国清代", "洋务外交", "北洋参议", "权衡", DomainDiplomacy, 42, 89, 82, 52, "envoy"},
		{"ban-chao", "班定远", "班超", "中国汉代", "西域经营", "都护府属", "胆略", DomainDiplomacy, 70, 87, 62, 72, "envoy"},
		{"di-renjie", "狄梁公", "狄仁杰", "中国唐代", "断狱举贤", "大理寺丞", "明断", DomainIntrigue, 76, 88, 38, 90, "scholar"},
		{"wei-zheng", "魏郑公", "魏征", "中国唐代", "直谏纳谏", "谏议大夫", "犯颜", DomainCourt, 68, 84, 34, 96, "scholar"},
		{"wu-zetian", "武曌", "武则天", "中国唐代", "权力组织", "内廷学士", "御衡", DomainCourt, 40, 92, 96, 50, "consort"},
		{"li-shimin", "李世民", "李世民", "中国唐代", "军政纳谏", "天策参谋", "英断", DomainMilitary, 56, 96, 82, 72, "general"},
		{"liu-bowen", "刘伯温", "刘基", "中国明初", "天文军谋", "军师祭酒", "玄算", DomainIntrigue, 58, 90, 52, 78, "scholar"},
		{"caesar", "凯撒", "尤利乌斯·凯撒", "罗马", "军团政治", "归化将军", "雄辩", DomainMilitary, 36, 95, 96, 38, "general"},
		{"augustus", "奥古斯都", "屋大维", "罗马", "元首制度", "法政客卿", "制衡", DomainCourt, 58, 91, 86, 62, "minister"},
		{"cicero", "西塞罗", "西塞罗", "罗马", "共和演说", "廷辩博士", "雄辩", DomainCourt, 52, 88, 62, 70, "scholar"},
		{"hannibal", "汉尼拔", "汉尼拔", "迦太基", "迂回决战", "外籍都尉", "险绝", DomainMilitary, 46, 96, 78, 58, "general"},
		{"alexander", "亚历山大", "亚历山大大帝", "马其顿", "远征融合", "外藩统军", "万里", DomainMilitary, 40, 97, 98, 44, "general"},
		{"pericles", "伯里克利", "伯里克利", "希腊", "公民财政", "议政客卿", "雅量", DomainDiplomacy, 60, 88, 68, 76, "envoy"},
		{"themistocles", "地米斯托克利", "地米斯托克利", "希腊", "海军战略", "水师客卿", "海略", DomainMilitary, 42, 90, 82, 48, "general"},
		{"cleopatra", "克娄巴特拉", "克娄巴特拉", "埃及", "宫廷外交", "外邦女官", "魅谋", DomainDiplomacy, 38, 88, 92, 42, "consort"},
		{"hammurabi", "汉谟拉比", "汉谟拉比", "巴比伦", "成文法典", "律令博士", "法衡", DomainReform, 58, 91, 64, 76, "minister"},
		{"cyrus", "居鲁士", "居鲁士大帝", "波斯", "宽制帝国", "藩政客卿", "怀远", DomainDiplomacy, 62, 90, 72, 76, "envoy"},
		{"darius", "大流士", "大流士一世", "波斯", "行省税制", "度支客卿", "整序", DomainEconomy, 54, 89, 78, 62, "minister"},
		{"ashoka", "阿育王", "阿育王", "印度孔雀", "仁政法敕", "德政博士", "悔悟", DomainDomestic, 72, 86, 46, 90, "scholar"},
		{"chanakya", "考底利耶", "考底利耶", "印度孔雀", "政略财政", "权谋博士", "冷算", DomainIntrigue, 44, 94, 74, 48, "scholar"},
		{"akbar", "阿克巴", "阿克巴大帝", "莫卧儿", "宗教宽容", "藩部议臣", "包容", DomainDiplomacy, 68, 90, 72, 74, "envoy"},
		{"saladin", "萨拉丁", "萨拉丁", "阿尤布", "骑士仁义", "西域统军", "仁勇", DomainMilitary, 76, 90, 58, 88, "general"},
		{"suleiman", "苏莱曼", "苏莱曼大帝", "奥斯曼", "律法扩张", "外法客卿", "宏制", DomainReform, 58, 92, 82, 70, "minister"},
		{"mehmed", "穆罕默德二世", "穆罕默德二世", "奥斯曼", "攻城炮兵", "炮垒客卿", "破城", DomainMilitary, 42, 92, 88, 48, "general"},
		{"genghis", "铁木真", "成吉思汗", "蒙古", "骑射组织", "归义万户", "草原", DomainMilitary, 36, 97, 96, 38, "general"},
		{"kublai", "忽必烈", "忽必烈", "蒙古元代", "多元帝国", "藩院客卿", "兼容", DomainDiplomacy, 54, 90, 82, 58, "envoy"},
		{"napoleon", "拿破仑", "拿破仑", "法国", "军团法典", "火器都尉", "疾雷", DomainMilitary, 34, 98, 98, 36, "general"},
		{"richelieu", "黎塞留", "黎塞留", "法国", "中央集权", "枢密客卿", "铁腕", DomainIntrigue, 42, 91, 84, 48, "minister"},
		{"louis-xiv", "路易十四", "路易十四", "法国", "宫廷集权", "礼制客卿", "威仪", DomainCourt, 38, 86, 94, 36, "consort"},
		{"bismarck", "俾斯麦", "俾斯麦", "普鲁士", "铁血外交", "外务参赞", "铁血", DomainDiplomacy, 40, 94, 86, 50, "envoy"},
		{"frederick", "腓特烈", "腓特烈大帝", "普鲁士", "军政启蒙", "练兵客卿", "严整", DomainMilitary, 46, 92, 84, 62, "general"},
		{"elizabeth", "伊丽莎白", "伊丽莎白一世", "英格兰", "海权均势", "海邦女官", "权衡", DomainDiplomacy, 58, 90, 82, 68, "consort"},
		{"churchill", "丘吉尔", "温斯顿·丘吉尔", "英国", "战时动员", "军机顾问", "鼓舞", DomainMilitary, 52, 88, 78, 56, "minister"},
		{"washington", "华盛顿", "乔治·华盛顿", "美国", "军政节制", "归化统领", "自持", DomainMilitary, 86, 88, 34, 96, "general"},
		{"franklin", "富兰克林", "本杰明·富兰克林", "美国", "科学外交", "博物客卿", "机智", DomainDiplomacy, 66, 90, 48, 82, "envoy"},
		{"lincoln", "林肯", "亚伯拉罕·林肯", "美国", "联邦民权", "法政博士", "仁毅", DomainDomestic, 80, 90, 54, 94, "scholar"},
		{"hamilton", "汉密尔顿", "亚历山大·汉密尔顿", "美国", "财政银行", "度支博士", "锐算", DomainEconomy, 48, 92, 84, 58, "minister"},
		{"mandela", "曼德拉", "纳尔逊·曼德拉", "南非", "和解政治", "和议客卿", "宽忍", DomainDomestic, 84, 88, 40, 96, "scholar"},
		{"mansa-musa", "曼萨穆萨", "曼萨·穆萨", "马里", "黄金商路", "互市客卿", "富远", DomainEconomy, 62, 86, 64, 70, "merchant"},
		{"shaka", "沙卡", "沙卡祖鲁", "祖鲁", "步兵改革", "步阵教头", "猛进", DomainMilitary, 40, 90, 86, 42, "general"},
		{"pachacuti", "帕查库提", "帕查库提", "印加", "道路仓储", "山道客卿", "筑路", DomainDomestic, 62, 88, 76, 68, "minister"},
		{"montezuma", "蒙特祖马", "蒙特祖马一世", "阿兹特克", "贡赋联盟", "贡赋客卿", "祭政", DomainEconomy, 42, 84, 82, 42, "minister"},
		{"sejong", "世宗", "朝鲜世宗", "朝鲜", "文字农政", "经筵博士", "爱民", DomainDomestic, 78, 90, 38, 90, "scholar"},
		{"yi-sunsin", "李舜臣", "李舜臣", "朝鲜", "水师防御", "舟师统制", "死守", DomainMilitary, 82, 92, 36, 96, "general"},
		{"tokugawa", "德川家康", "德川家康", "日本", "幕府秩序", "藩政客卿", "忍成", DomainCourt, 50, 90, 84, 58, "minister"},
		{"oda", "织田信长", "织田信长", "日本", "火器革新", "铁炮教头", "破旧", DomainReform, 34, 92, 96, 34, "general"},
		{"meiji", "明治", "明治天皇", "日本", "维新国家", "维新顾问", "开化", DomainReform, 58, 88, 78, 68, "minister"},
		{"peter", "彼得", "彼得大帝", "俄罗斯", "海军西化", "船政客卿", "开海", DomainReform, 38, 92, 94, 38, "minister"},
		{"catherine", "叶卡捷琳娜", "叶卡捷琳娜二世", "俄罗斯", "贵族行政", "北国女官", "扩张", DomainDiplomacy, 42, 90, 94, 42, "consort"},
		{"lenin", "列宁", "列宁", "俄罗斯", "革命组织", "密议客卿", "组织", DomainIntrigue, 28, 92, 96, 30, "scholar"},
		{"marx", "马克思", "卡尔·马克思", "德国", "政治经济", "太学博士", "批判", DomainReform, 34, 91, 70, 58, "scholar"},
		{"adam-smith", "亚当斯密", "亚当·斯密", "苏格兰", "市场财政", "商税博士", "观市", DomainEconomy, 54, 90, 42, 78, "merchant"},
		{"keynes", "凯恩斯", "约翰·凯恩斯", "英国", "宏观财政", "度支博士", "调控", DomainEconomy, 50, 91, 56, 68, "minister"},
		{"darwin", "达尔文", "查尔斯·达尔文", "英国", "博物演化", "格物博士", "慎察", DomainReform, 58, 88, 30, 86, "scholar"},
		{"newton", "牛顿", "艾萨克·牛顿", "英国", "数学天文", "钦天博士", "精微", DomainReform, 46, 96, 52, 76, "scholar"},
		{"galileo", "伽利略", "伽利略", "意大利", "实验天文", "观象博士", "求证", DomainReform, 44, 90, 58, 72, "scholar"},
		{"curie", "居里", "玛丽·居里", "波兰法国", "实验科学", "药物格物", "坚毅", DomainDomestic, 70, 92, 34, 92, "physician"},
		{"turing", "图灵", "艾伦·图灵", "英国", "密码计算", "机巧博士", "破译", DomainIntrigue, 56, 94, 38, 84, "scholar"},
		{"tesla", "特斯拉", "尼古拉·特斯拉", "塞尔维亚", "电机工程", "工部客卿", "奇想", DomainReform, 42, 91, 62, 66, "scholar"},
		{"ibn-khaldun", "伊本赫勒敦", "伊本·赫勒敦", "北非", "社会史观", "史政博士", "洞察", DomainCourt, 58, 90, 44, 84, "scholar"},
		{"ibn-sina", "伊本西那", "伊本·西那", "波斯", "医学哲学", "太医博士", "医哲", DomainDomestic, 72, 90, 34, 86, "physician"},
		{"machiavelli", "马基雅维利", "马基雅维利", "意大利", "现实政略", "枢密参议", "冷眼", DomainIntrigue, 30, 90, 78, 36, "scholar"},
		{"leonardo", "达芬奇", "列奥纳多·达芬奇", "意大利", "工程艺术", "工绘博士", "全才", DomainReform, 50, 94, 44, 74, "painter"},
		{"joan", "贞德", "圣女贞德", "法国", "信念动员", "义勇校尉", "炽信", DomainMilitary, 78, 86, 42, 92, "general"},
		{"nightingale", "南丁格尔", "弗洛伦斯·南丁格尔", "英国", "军医护理", "医营女官", "仁护", DomainDomestic, 82, 88, 28, 96, "physician"},
		{"rosa", "罗莎", "罗莎·帕克斯", "美国", "民权抗争", "民声客卿", "不屈", DomainDomestic, 84, 82, 34, 98, "scholar"},
		{"hypatia", "希帕提娅", "希帕提娅", "亚历山大里亚", "数学哲学", "太学女博士", "明思", DomainReform, 62, 90, 32, 88, "scholar"},
		{"florence-medici", "美第奇", "洛伦佐·美第奇", "佛罗伦萨", "金融艺术", "商馆客卿", "赞助", DomainEconomy, 48, 86, 78, 54, "merchant"},
		{"qin-shihuang", "秦始皇", "秦始皇", "中国秦代", "一统郡县", "法政参议", "一统", DomainReform, 32, 94, 98, 34, "minister"},
	}
	for i := range raw {
		if raw[i].Portrait == "" {
			raw[i].Portrait = portraitForTalent(raw[i].Specialty)
		}
		raw[i].Key = strings.TrimSpace(raw[i].Key)
	}
	return raw
}

func portraitForTalent(specialty Domain) string {
	switch specialty {
	case DomainMilitary:
		return "general"
	case DomainEconomy:
		return "merchant"
	case DomainDiplomacy:
		return "envoy"
	case DomainDomestic:
		return "scholar"
	default:
		return "scholar"
	}
}
