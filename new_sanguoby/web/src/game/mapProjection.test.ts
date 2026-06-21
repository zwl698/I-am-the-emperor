import {describe, expect, it} from 'vitest';
import {projectCity} from './mapProjection';

describe('projectCity', () => {
  it('projects legacy 12x9 city coordinates into a bounded geographic map position', () => {
    const point = projectCity({ name: '长安', x: 6, y: 4 }, { width: 1200, height: 720 });

    expect(point.x).toBeGreaterThan(560);
    expect(point.x).toBeLessThan(680);
    expect(point.y).toBeGreaterThan(285);
    expect(point.y).toBeLessThan(395);
    expect(point).not.toEqual({ x: 640, y: 329 });
  });

  it('keeps edge cities inside the visual padding', () => {
    const point = projectCity({ x: 0, y: 8 }, { width: 1200, height: 720 });

    expect(point.x).toBeGreaterThanOrEqual(72);
    expect(point.y).toBeGreaterThanOrEqual(72);
    expect(point.x).toBeLessThanOrEqual(1128);
    expect(point.y).toBeLessThanOrEqual(648);
  });

  it('preserves east-west and north-south ordering from the legacy map', () => {
    const west = projectCity({ name: '西凉', x: 1, y: 0 }, { width: 1200, height: 720 });
    const east = projectCity({ name: '北平', x: 9, y: 0 }, { width: 1200, height: 720 });
    const north = projectCity({ name: '平原', x: 8, y: 2 }, { width: 1200, height: 720 });
    const south = projectCity({ name: '建业', x: 9, y: 7 }, { width: 1200, height: 720 });

    expect(west.x).toBeLessThan(east.x);
    expect(north.y).toBeLessThan(south.y);
  });

  it('stably separates cities that share the same old grid cell', () => {
    const first = projectCity({ name: '长安', x: 6, y: 4 }, { width: 1200, height: 720 });
    const second = projectCity({ name: '洛阳', x: 6, y: 4 }, { width: 1200, height: 720 });
    const repeat = projectCity({ name: '长安', x: 6, y: 4 }, { width: 1200, height: 720 });

    expect(first).toEqual(repeat);
    expect(first).not.toEqual(second);
  });
});
