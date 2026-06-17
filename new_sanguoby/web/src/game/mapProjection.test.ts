import {describe, expect, it} from 'vitest';
import {projectCity} from './mapProjection';

describe('projectCity', () => {
  it('projects legacy 12x9 city coordinates into a bounded viewport', () => {
    const point = projectCity({ x: 6, y: 4 }, { width: 1200, height: 720 });

    // padding = min(1200,720)/9 = 80
    // x = 80 + (6/12)*(1200-80) = 640
    // y = 80 + (4/9)*(720-160) ≈ 329
    expect(point.x).toBe(640);
    expect(point.y).toBe(329);
  });

  it('keeps edge cities inside the visual padding', () => {
    const point = projectCity({ x: 0, y: 8 }, { width: 1200, height: 720 });

    // x clamps to left padding 80; y = 80 + (8/9)*560 ≈ 578
    expect(point.x).toBe(80);
    expect(point.y).toBe(578);
  });
});
