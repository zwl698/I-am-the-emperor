import type { General, Ruler, RulerOption } from '../api/types';

const GENERATED_ROOT = '/assets/portraits/generated';
const portraitPath = (name: string) => `${GENERATED_ROOT}/${name}.webp`;

export const DEFAULT_PORTRAIT = portraitPath('neutral');

const PORTRAITS = {
  caocao: portraitPath('caocao'),
  liubei: portraitPath('liubei'),
  sunquan: portraitPath('sunquan'),
  dongzhuo: portraitPath('dongzhuo'),
  neutral: DEFAULT_PORTRAIT,
  yuanshao: portraitPath('yuanshao'),
  yuanshu: portraitPath('yuanshu'),
  sunjian: portraitPath('sunjian'),
  sunce: portraitPath('sunce'),
  mateng: portraitPath('mateng'),
  taoqian: portraitPath('taoqian'),
  liuyan: portraitPath('liuyan'),
  liubiao: portraitPath('liubiao'),
  gongsunzan: portraitPath('gongsunzan'),
  gongsundu: portraitPath('gongsundu'),
  zhanglu: portraitPath('zhanglu'),
  kongrong: portraitPath('kongrong'),
  liudai: portraitPath('liudai'),
  zhangyang: portraitPath('zhangyang'),
  hanfu: portraitPath('hanfu'),
  wangkuang: portraitPath('wangkuang'),
  wanglang: portraitPath('wanglang'),
  yanbaihu: portraitPath('yanbaihu'),
  lvbu: portraitPath('lvbu'),
  liuzhang: portraitPath('liuzhang'),
  zhangxiu: portraitPath('zhangxiu'),
  xiahouDun: portraitPath('xiahou-dun'),
  xiahouYuan: portraitPath('xiahou-yuan'),
  dianwei: portraitPath('dianwei'),
  xuchu: portraitPath('xuchu'),
  simayi: portraitPath('simayi'),
  xunyu: portraitPath('xunyu'),
  guojia: portraitPath('guojia'),
  guanyu: portraitPath('guanyu'),
  zhangfei: portraitPath('zhangfei'),
  zhaoyun: portraitPath('zhaoyun'),
  zhugeliang: portraitPath('zhugeliang'),
  machao: portraitPath('machao'),
  huangzhong: portraitPath('huangzhong'),
  weiyan: portraitPath('weiyan'),
  zhouyu: portraitPath('zhouyu'),
  luxun: portraitPath('luxun'),
  lvmeng: portraitPath('lvmeng'),
  ganning: portraitPath('ganning'),
  taishici: portraitPath('taishici'),
  huanggai: portraitPath('huanggai'),
  chengpu: portraitPath('chengpu'),
  jiaxu: portraitPath('jiaxu'),
  huaxiong: portraitPath('huaxiong'),
};

const RULER_PORTRAITS: Record<string, string> = {
  caocao: PORTRAITS.caocao,
  liubei: PORTRAITS.liubei,
  sunquan: PORTRAITS.sunquan,
  dongzhuo: PORTRAITS.dongzhuo,
  neutral: PORTRAITS.neutral,
};

const GENERAL_PORTRAITS: Record<string, string> = {
  'cao-cao': PORTRAITS.caocao,
  'liu-bei': PORTRAITS.liubei,
  'sun-quan': PORTRAITS.sunquan,
  'dong-zhuo': PORTRAITS.dongzhuo,
  'xiahou-dun': PORTRAITS.xiahouDun,
  'guan-yu': PORTRAITS.guanyu,
  'zhou-yu': PORTRAITS.zhouyu,
  'lv-bu': PORTRAITS.lvbu,
};

const NAME_PORTRAITS: Array<[string, string]> = [
  ['夏侯惇', PORTRAITS.xiahouDun],
  ['夏侯渊', PORTRAITS.xiahouYuan],
  ['司马懿', PORTRAITS.simayi],
  ['诸葛亮', PORTRAITS.zhugeliang],
  ['太史慈', PORTRAITS.taishici],
  ['公孙瓒', PORTRAITS.gongsunzan],
  ['公孙度', PORTRAITS.gongsundu],
  ['严白虎', PORTRAITS.yanbaihu],
  ['曹操', PORTRAITS.caocao],
  ['曹丕', PORTRAITS.caocao],
  ['曹仁', PORTRAITS.xiahouDun],
  ['曹洪', PORTRAITS.xiahouYuan],
  ['典韦', PORTRAITS.dianwei],
  ['许褚', PORTRAITS.xuchu],
  ['荀彧', PORTRAITS.xunyu],
  ['荀攸', PORTRAITS.xunyu],
  ['郭嘉', PORTRAITS.guojia],
  ['程昱', PORTRAITS.guojia],
  ['张辽', PORTRAITS.lvbu],
  ['徐晃', PORTRAITS.xiahouDun],
  ['张郃', PORTRAITS.yuanshao],
  ['刘备', PORTRAITS.liubei],
  ['刘禅', PORTRAITS.liubei],
  ['刘焉', PORTRAITS.liuyan],
  ['刘璋', PORTRAITS.liuzhang],
  ['刘表', PORTRAITS.liubiao],
  ['刘岱', PORTRAITS.liudai],
  ['关羽', PORTRAITS.guanyu],
  ['张飞', PORTRAITS.zhangfei],
  ['赵云', PORTRAITS.zhaoyun],
  ['马超', PORTRAITS.machao],
  ['黄忠', PORTRAITS.huangzhong],
  ['魏延', PORTRAITS.weiyan],
  ['庞统', PORTRAITS.zhugeliang],
  ['法正', PORTRAITS.xunyu],
  ['孙权', PORTRAITS.sunquan],
  ['孙坚', PORTRAITS.sunjian],
  ['孙策', PORTRAITS.sunce],
  ['周瑜', PORTRAITS.zhouyu],
  ['陆逊', PORTRAITS.luxun],
  ['吕蒙', PORTRAITS.lvmeng],
  ['甘宁', PORTRAITS.ganning],
  ['黄盖', PORTRAITS.huanggai],
  ['程普', PORTRAITS.chengpu],
  ['鲁肃', PORTRAITS.luxun],
  ['董卓', PORTRAITS.dongzhuo],
  ['吕布', PORTRAITS.lvbu],
  ['华雄', PORTRAITS.huaxiong],
  ['李儒', PORTRAITS.jiaxu],
  ['贾诩', PORTRAITS.jiaxu],
  ['袁绍', PORTRAITS.yuanshao],
  ['袁术', PORTRAITS.yuanshu],
  ['颜良', PORTRAITS.yuanshao],
  ['文丑', PORTRAITS.yuanshu],
  ['马腾', PORTRAITS.mateng],
  ['韩遂', PORTRAITS.mateng],
  ['陶谦', PORTRAITS.taoqian],
  ['张鲁', PORTRAITS.zhanglu],
  ['孔融', PORTRAITS.kongrong],
  ['张杨', PORTRAITS.zhangyang],
  ['韩馥', PORTRAITS.hanfu],
  ['王匡', PORTRAITS.wangkuang],
  ['王朗', PORTRAITS.wanglang],
  ['张绣', PORTRAITS.zhangxiu],
];

export function portraitForRuler(ruler?: Ruler | RulerOption | string): string {
  if (!ruler) {
    return DEFAULT_PORTRAIT;
  }
  if (typeof ruler === 'string') {
    return RULER_PORTRAITS[ruler] ?? portraitByName(ruler) ?? DEFAULT_PORTRAIT;
  }
  return portraitByName(ruler.name) ?? RULER_PORTRAITS[ruler.id] ?? portraitByColor(ruler.color);
}

export function portraitForGeneral(general: General): string {
  return portraitByName(general.name) ?? GENERAL_PORTRAITS[general.id] ?? portraitForRuler(general.ownerId);
}

function portraitByName(name: string): string | undefined {
  return NAME_PORTRAITS.find(([knownName]) => name.includes(knownName))?.[1];
}

function portraitByColor(color: string): string {
  const hex = color.replace('#', '');
  if (hex.length < 6) {
    return DEFAULT_PORTRAIT;
  }
  const red = Number.parseInt(hex.slice(0, 2), 16);
  const green = Number.parseInt(hex.slice(2, 4), 16);
  const blue = Number.parseInt(hex.slice(4, 6), 16);
  if ([red, green, blue].some(Number.isNaN)) {
    return DEFAULT_PORTRAIT;
  }
  if (green >= red && green >= blue) {
    return PORTRAITS.liubei;
  }
  if (blue >= red && blue >= green) {
    return red > 95 ? PORTRAITS.dongzhuo : PORTRAITS.sunquan;
  }
  return red > 150 && green > 110 ? PORTRAITS.yuanshao : PORTRAITS.caocao;
}
