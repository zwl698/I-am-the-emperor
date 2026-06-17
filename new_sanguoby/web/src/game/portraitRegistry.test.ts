import { describe, expect, it } from 'vitest';
import { portraitForGeneral, portraitForRuler } from './portraitRegistry';
import type { General, Ruler } from '../api/types';

describe('portraitRegistry', () => {
  it('returns faction portraits for known rulers', () => {
    expect(portraitForRuler('caocao')).toBe('/assets/portraits/caocao.svg');
    expect(portraitForRuler('liubei')).toBe('/assets/portraits/liubei.svg');
    expect(portraitForRuler('sunquan')).toBe('/assets/portraits/sunquan.svg');
    expect(portraitForRuler('dongzhuo')).toBe('/assets/portraits/dongzhuo.svg');
  });

  it('matches legacy ruler records by Chinese name when ids are generated', () => {
    const ruler: Ruler = {
      id: 'ruler-0',
      name: '董卓',
      character: '凶莽',
      color: '#9b2f2f',
    };

    expect(portraitForRuler(ruler)).toBe('/assets/portraits/dongzhuo.svg');
  });

  it('falls back to the owner portrait for generals without unique art', () => {
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

    expect(portraitForGeneral(general)).toBe('/assets/portraits/caocao.svg');
  });
});
