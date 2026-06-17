import type { General, Ruler } from '../api/types';

const DEFAULT_PORTRAIT = '/assets/portraits/neutral.svg';

const RULER_PORTRAITS: Record<string, string> = {
  caocao: '/assets/portraits/caocao.svg',
  liubei: '/assets/portraits/liubei.svg',
  sunquan: '/assets/portraits/sunquan.svg',
  dongzhuo: '/assets/portraits/dongzhuo.svg',
  neutral: DEFAULT_PORTRAIT,
};

const GENERAL_PORTRAITS: Record<string, string> = {
  'cao-cao': RULER_PORTRAITS.caocao,
  'liu-bei': RULER_PORTRAITS.liubei,
  'sun-quan': RULER_PORTRAITS.sunquan,
  'dong-zhuo': RULER_PORTRAITS.dongzhuo,
};

const NAME_PORTRAITS: Array<[string, string]> = [
  ['曹操', RULER_PORTRAITS.caocao],
  ['刘备', RULER_PORTRAITS.liubei],
  ['孙权', RULER_PORTRAITS.sunquan],
  ['董卓', RULER_PORTRAITS.dongzhuo],
  ['吕布', RULER_PORTRAITS.dongzhuo],
  ['袁绍', RULER_PORTRAITS.sunquan],
  ['袁术', RULER_PORTRAITS.dongzhuo],
  ['马腾', RULER_PORTRAITS.liubei],
  ['陶谦', RULER_PORTRAITS.liubei],
  ['刘焉', RULER_PORTRAITS.liubei],
  ['刘表', RULER_PORTRAITS.liubei],
  ['张鲁', RULER_PORTRAITS.sunquan],
  ['孔融', RULER_PORTRAITS.sunquan],
];

export function portraitForRuler(ruler?: Ruler | string): string {
  if (!ruler) {
    return DEFAULT_PORTRAIT;
  }
  if (typeof ruler === 'string') {
    return RULER_PORTRAITS[ruler] ?? portraitByName(ruler) ?? DEFAULT_PORTRAIT;
  }
  return RULER_PORTRAITS[ruler.id] ?? portraitByName(ruler.name) ?? portraitByColor(ruler.color);
}

export function portraitForGeneral(general: General): string {
  return GENERAL_PORTRAITS[general.id] ?? portraitByName(general.name) ?? portraitForRuler(general.ownerId);
}

function portraitByName(name: string): string | undefined {
  return NAME_PORTRAITS.find(([knownName]) => name.includes(knownName))?.[1];
}

function portraitByColor(color: string): string {
  const hue = Number.parseInt(color.replace('#', '').slice(0, 2), 16);
  if (hue > 150) {
    return RULER_PORTRAITS.caocao;
  }
  if (hue > 100) {
    return RULER_PORTRAITS.dongzhuo;
  }
  if (hue > 60) {
    return RULER_PORTRAITS.liubei;
  }
  return RULER_PORTRAITS.sunquan;
}
