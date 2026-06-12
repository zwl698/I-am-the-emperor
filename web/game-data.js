var statMeta = [
  ["legitimacy", "名望"],
  ["health", "健康"],
  ["learning", "学识"],
  ["martial", "武略"],
  ["charisma", "魅力"],
  ["influence", "势力"],
  ["treasury", "国库"],
  ["grain", "粮草"],
  ["populace", "民心"],
  ["army", "军力"],
  ["diplomacy", "邦交"],
  ["stability", "朝稳"],
  ["borderThreat", "边患"],
  ["reform", "新政"],
];

var domainIcon = {
  story: "卷",
  domestic: "民",
  economy: "财",
  military: "兵",
  diplomacy: "使",
  court: "宫",
  reform: "法",
  intrigue: "密",
};

var phaseName = {
  prince: "皇子",
  emperor: "皇帝",
};

var sceneAssetNames = [
  "birth-chamber",
  "east-palace-study",
  "winter-hunt",
  "flood-levee",
  "succession-hall",
  "throne-court",
  "granary-relief",
  "tax-office",
  "frontier-fortress",
  "envoy-pass",
  "reform-archive",
  "secret-tribunal",
  "banquet-hall",
  "jiangnan-canal",
  "northern-battlefield",
  "desert-market",
  "imperial-garden",
  "rain-corridor",
  "ancestral-temple",
  "ministry-office",
  "dockyard-fleet",
  "drill-ground",
  "rebel-village",
  "silk-market",
  "mountain-monastery",
  "exam-hall",
  "map-room",
  "palace-dawn",
  "diplomatic-tent",
  "festival-night",
];

var portraitAssetNames = [
  "infant-prince",
  "teen-prince",
  "young-emperor",
  "elder-emperor",
  "stern-tutor",
  "frontier-general",
  "finance-minister",
  "grand-princess",
  "noble-consort",
  "young-empress",
  "queen-dowager",
  "palace-maid",
  "eunuch-spymaster",
  "scholar-official",
  "reformist-official",
  "corrupt-magistrate",
  "merchant-leader",
  "foreign-envoy",
  "nomad-khan",
  "monk-strategist",
  "female-diplomat",
  "guard-captain",
  "rebel-leader",
  "river-engineer",
  "imperial-physician",
  "astrologer",
  "poet",
  "court-painter",
  "farmer-representative",
  "masked-assassin",
];

var sceneGalleryFallback = assetPaths("/assets/scenes/scene", sceneAssetNames);
var portraitGalleryFallback = assetPaths("/assets/portraits/portrait", portraitAssetNames);

var portraitIndexByRole = {
  infant: 0,
  prince: 1,
  emperor: 2,
  elder: 3,
  tutor: 4,
  general: 5,
  minister: 6,
  consort: 7,
  empress: 9,
  dowager: 10,
  maid: 11,
  spy: 12,
  scholar: 13,
  reformer: 14,
  corrupt: 15,
  merchant: 16,
  envoy: 17,
  khan: 18,
  monk: 19,
  diplomat: 20,
  guard: 21,
  rebel: 22,
  engineer: 23,
  physician: 24,
  astrologer: 25,
  poet: 26,
  painter: 27,
  farmer: 28,
  assassin: 29,
};

var provinceOrders = [
  { kind: "relief", label: "赈", title: "赈济：降灾情、涨民心，耗粮银" },
  { kind: "garrison", label: "驻", title: "驻防：升防务、压边患，耗军费" },
  { kind: "tax", label: "税", title: "督税：涨国库，伤地方秩序" },
  { kind: "canal", label: "渠", title: "修渠：长期富庶与新政，耗国库" },
  { kind: "trade", label: "市", title: "互市：涨财政外交，略增边患" },
  { kind: "inspect", label: "查", title: "密查：升秩序，抓胥吏" },
];

var factionOrders = [
  { kind: "appease", label: "安", title: "安抚：涨忠诚，耗银与威权" },
  { kind: "purge", label: "削", title: "削权：降权势，激化党争" },
  { kind: "inspect", label: "查", title: "密查：压权势，损稳定" },
];

var warOrders = [
  { kind: "mobilize", label: "动", title: "动员：增粮道士气，耗粮银" },
  { kind: "campaign", label: "征", title: "出征：推进战役，降低敌势，损兵粮" },
  { kind: "fortify", label: "固", title: "固边：筑堡屯粮，压低威胁" },
  { kind: "truce", label: "和", title: "议和：外交缓战，武臣不满" },
];

var consortOrders = [
  { kind: "favor_consort", label: "宠", title: "临幸：提升宠爱与外戚影响，可能增加储位争议" },
  { kind: "marriage_alliance", label: "姻", title: "联姻：稳外戚与邦交，耗国库并抬高母族" },
];

var heirOrders = [
  { kind: "name_heir", label: "储", title: "册储：指定继承人，提升拥护但可能激化争议" },
];

var heirTrainingOrders = [
  { kind: "educate_heir", focus: "study", label: "经", title: "经史：提升资质与文治名望" },
  { kind: "educate_heir", focus: "drill", label: "射", title: "骑射：提升资质与野心，培养武略" },
  { kind: "educate_heir", focus: "rites", label: "礼", title: "礼法：提升拥护与储位稳定" },
];

var officeOrders = [
  { kind: "appoint", label: "任", title: "任官：指派臣子掌官署，消耗御令" },
  { kind: "dismiss", label: "罢", title: "罢官：清空官位，震慑群臣但制造空转" },
];

var projectOrders = [
  { kind: "fund_project", label: "营", title: "营造：投入银粮人手推进多年国策工程" },
];

var policyOrders = [
  { kind: "enact_policy", label: "策", title: "国策：启用或暂罢常驻政策，每季自动生效" },
];

var foreignOrders = [
  { kind: "embassy", label: "使", title: "遣使：改善关系、降低威胁，耗国库" },
  { kind: "treaty", label: "盟", title: "盟约：关系足够时签订长期贡贸盟约" },
];

var plotOrders = [
  { kind: "investigate_plot", label: "侦", title: "侦缉：降低隐秘和进度，可能暴露阴谋" },
  { kind: "suppress_plot", label: "平", title: "平谋：阴谋暴露后可直接结案" },
];

var justiceOrders = [
  { kind: "open_trial", label: "审", title: "明审：公开审理案件，提升法度但牵动派系" },
  { kind: "clemency", label: "赦", title: "宽赦：从轻发落，稳局面但伤清议" },
  { kind: "censor_rumor", label: "禁", title: "禁谣：压低流言和热度，但增加畏惧并损名望" },
  { kind: "proclaim_verdict", label: "宣", title: "宣判：已结案件榜示天下，转化为民望与士论" },
];

var musicTracks = [
  { id: "birth", name: "雪宫摇篮", root: 196, scale: [0, 3, 5, 7, 10, 12], tempo: 72, wave: "sine", drum: false },
  { id: "study", name: "东宫书声", root: 220, scale: [0, 2, 5, 7, 9, 12], tempo: 84, wave: "triangle", drum: false },
  { id: "hunt", name: "雪猎急弦", root: 174, scale: [0, 3, 5, 7, 10, 12], tempo: 120, wave: "sawtooth", drum: true },
  { id: "flood", name: "南河雨鼓", root: 164, scale: [0, 2, 3, 7, 8, 12], tempo: 92, wave: "triangle", drum: true },
  { id: "succession", name: "烛影夺嫡", root: 146, scale: [0, 1, 5, 7, 8, 12], tempo: 96, wave: "sine", drum: true },
  { id: "court", name: "太和晨钟", root: 196, scale: [0, 2, 4, 7, 9, 12], tempo: 82, wave: "triangle", drum: true },
  { id: "people", name: "粥棚烟火", root: 185, scale: [0, 2, 5, 7, 9, 12], tempo: 76, wave: "sine", drum: false },
  { id: "treasury", name: "银库算盘", root: 207, scale: [0, 2, 4, 6, 9, 12], tempo: 102, wave: "square", drum: false },
  { id: "war", name: "边塞战鼓", root: 130, scale: [0, 3, 5, 7, 10, 12], tempo: 128, wave: "sawtooth", drum: true },
  { id: "envoy", name: "驼铃万里", root: 174, scale: [0, 2, 5, 7, 11, 12], tempo: 88, wave: "triangle", drum: false },
  { id: "intrigue", name: "密档夜灯", root: 155, scale: [0, 1, 5, 6, 8, 12], tempo: 90, wave: "sine", drum: true },
  { id: "festival", name: "万邦烟火", root: 247, scale: [0, 2, 4, 7, 9, 14], tempo: 112, wave: "triangle", drum: true },
];

var dynastyPanelClass = {
  dayin: "panel-dayin",
  jingyao: "panel-jingyao",
  chengping: "panel-chengping",
  xuanshuo: "panel-xuanshuo",
};

function assetPaths(prefix, names) {
  return names.map((name, index) => `${prefix}-${String(index + 1).padStart(2, "0")}-${name}.png`);
}
