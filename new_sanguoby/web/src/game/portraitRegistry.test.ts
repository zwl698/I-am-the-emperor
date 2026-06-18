import { describe, expect, it } from 'vitest';
import { portraitForGeneral, portraitForRuler } from './portraitRegistry';
import type { General, Ruler } from '../api/types';

describe('portraitRegistry', () => {
  it('returns faction portraits for known rulers', () => {
    expect(portraitForRuler('caocao')).toBe('/assets/portraits/generated/caocao.webp');
    expect(portraitForRuler('liubei')).toBe('/assets/portraits/generated/liubei.webp');
    expect(portraitForRuler('sunquan')).toBe('/assets/portraits/generated/sunquan.webp');
    expect(portraitForRuler('dongzhuo')).toBe('/assets/portraits/generated/dongzhuo.webp');
  });

  it('matches legacy ruler records by Chinese name when ids are generated', () => {
    const ruler: Ruler = {
      id: 'ruler-0',
      name: '董卓',
      character: '凶莽',
      color: '#9b2f2f',
    };

    expect(portraitForRuler(ruler)).toBe('/assets/portraits/generated/dongzhuo.webp');
  });

  it('returns generated portraits for known generals', () => {
    const general: General = {
      id: 'xiahou-dun',
      name: '夏侯惇',
      ownerId: 'caocao',
      cityId: 'chenliu',
      level: 6,
      force: 89,
      intellect: 61,
      loyalty: 95,
      stamina: 86,
      soldiers: 850,
      armsType: '步兵',
    };

    expect(portraitForGeneral(general)).toBe('/assets/portraits/generated/xiahou-dun.webp');
  });
});
