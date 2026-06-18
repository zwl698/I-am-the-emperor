import Phaser from 'phaser';
import type { City, GameSnapshot, Ruler } from '../../api/types';
import { projectCity } from '../../game/mapProjection';

type CampaignSceneOptions = {
  onCitySelected: (cityId: string) => void;
};

export class CampaignScene extends Phaser.Scene {
  private snapshot: GameSnapshot | null = null;
  private onCitySelected: (cityId: string) => void;
  private ready = false;
  private selectedCityId = '';

  constructor(options: CampaignSceneOptions) {
    super('CampaignScene');
    this.onCitySelected = options.onCitySelected;
  }

  preload() {
    this.load.image('campaign-map', '/assets/map/sanguo-campaign-map.png');
    this.load.svg('city-capital', '/assets/city-capital.svg', { width: 72, height: 72 });
    this.load.svg('city-frontier', '/assets/city-frontier.svg', { width: 72, height: 72 });
    this.load.svg('city-fort', '/assets/city-fort.svg', { width: 72, height: 72 });
    this.load.svg('city-port', '/assets/city-port.svg', { width: 72, height: 72 });
    this.load.svg('city-town', '/assets/city-town.svg', { width: 72, height: 72 });
    this.load.svg('army-banner', '/assets/army-banner.svg', { width: 64, height: 64 });
  }

  create() {
    this.ready = true;
    this.scale.on('resize', () => this.renderSnapshot());
    this.renderSnapshot();
  }

  setSnapshot(snapshot: GameSnapshot) {
    this.snapshot = snapshot;
    if (this.ready) {
      this.renderSnapshot();
    }
  }

  setSelectedCity(cityId: string) {
    if (this.selectedCityId === cityId) {
      return;
    }
    this.selectedCityId = cityId;
    if (this.ready) {
      this.renderSnapshot();
    }
  }

  private renderSnapshot() {
    if (!this.snapshot) {
      return;
    }

    const { width, height } = this.scale;
    this.children.removeAll(true);
    this.drawMapBackground(width, height);
    this.drawRoutes(width, height, this.snapshot);
    this.drawCities(width, height, this.snapshot);
  }

  private drawMapBackground(width: number, height: number) {
    const background = this.add.image(width / 2, height / 2, 'campaign-map');
    const scale = Math.max(width / background.width, height / background.height);
    background.setScale(scale);
    background.setDepth(-10);

    const graphics = this.add.graphics();
    graphics.fillStyle(0x160e0d, 0.12);
    graphics.fillRect(0, 0, width, height);

    graphics.fillStyle(0x050303, 0.22);
    graphics.fillCircle(-width * 0.08, height * 1.04, width * 0.62);
    graphics.fillCircle(width * 1.05, -height * 0.12, width * 0.42);

    graphics.lineStyle(28, 0x15100f, 0.18);
    graphics.strokeRect(12, 12, width - 24, height - 24);
    graphics.lineStyle(2, 0xffe3a4, 0.22);
    graphics.strokeRect(28, 28, width - 56, height - 56);
  }

  private drawRoutes(width: number, height: number, snapshot: GameSnapshot) {
    const cityByID = new Map(snapshot.cities.map((city) => [city.id, city]));
    const selectedCity = cityByID.get(this.selectedCityId);
    const graphics = this.add.graphics();
    const compact = snapshot.cities.length > 24 || width < 760;

    for (let index = 0; index < snapshot.routes.length; index++) {
      const route = snapshot.routes[index];
      const from = cityByID.get(route.from);
      const to = cityByID.get(route.to);
      if (!from || !to) {
        continue;
      }
      const start = projectCity(from, { width, height });
      const end = projectCity(to, { width, height });
      const relevant = selectedCity && (route.from === selectedCity.id || route.to === selectedCity.id);
      const routePoints = terrainRoutePoints(start, end, width, height, route.from + route.to + index);
      graphics.lineStyle(
        relevant ? (compact ? 5 : 8) : (compact ? 2 : 4),
        0x25150e,
        relevant ? 0.38 : 0.18,
      );
      strokeRoute(graphics, routePoints);
      graphics.lineStyle(
        relevant ? (compact ? 2 : 3) : (compact ? 1 : 1.5),
        relevant ? 0xffdc8a : 0xe3bd78,
        relevant ? 0.76 : 0.3,
      );
      strokeRoute(graphics, routePoints);
    }
  }

  private drawCities(width: number, height: number, snapshot: GameSnapshot) {
    const rulerByID = new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler]));
    const compact = snapshot.cities.length > 24 || width < 760;
    for (const city of snapshot.cities) {
      const point = projectCity(city, { width, height });
      const ruler = rulerByID.get(city.ownerId);
      const tint = colorToNumber(ruler?.color ?? '#7f7a68');
      const selected = city.id === this.selectedCityId;
      const role = cityRole(city);
      const markerKey = cityMarkerKey(role);
      const markerSize = markerDisplaySize(role, compact, selected);

      const halo = this.add.circle(point.x, point.y, selected ? markerSize * 0.78 : markerSize * 0.6, tint, selected ? 0.44 : 0.28);
      halo.setStrokeStyle(selected ? 4 : 3, selected ? 0xfff3bd : tint, selected ? 0.94 : 0.78);
      halo.setDepth(selected ? 2 : 0);

      if (selected) {
        const focus = this.add.circle(point.x, point.y, markerSize * 0.86, 0xfff3bd, 0);
        focus.setStrokeStyle(2, 0x2b1710, 0.7);
        focus.setDepth(1);
      }

      const base = this.add.circle(point.x, point.y + markerSize * 0.08, markerSize * 0.34, 0x1d1410, 0.42);
      base.setScale(1.32, 0.42);
      base.setDepth(selected ? 3 : 1);

      const marker = this.add.image(point.x, point.y, markerKey);
      marker.setDisplaySize(markerSize, markerSize);
      marker.setTint(tint);
      marker.setInteractive({ useHandCursor: true });
      marker.setDepth(selected ? 5 : 3);
      marker.on('pointerdown', () => {
        this.selectedCityId = city.id;
        this.onCitySelected(city.id);
        this.renderSnapshot();
      });

      if (city.ownerId !== 'neutral') {
        const banner = this.add.image(point.x + markerSize * 0.42, point.y - markerSize * 0.42, 'army-banner');
        banner.setDisplaySize(selected ? markerSize * 0.58 : markerSize * 0.46, selected ? markerSize * 0.58 : markerSize * 0.46);
        banner.setTint(tint);
        banner.setDepth(selected ? 6 : 4);

        const surname = rulerSurname(ruler);
        const ownerBadge = this.add.text(point.x + markerSize * 0.43, point.y - markerSize * 0.66, surname, {
          fontFamily: '"Noto Serif SC", "Songti SC", serif',
          fontSize: selected ? (compact ? '12px' : '15px') : (compact ? '10px' : '13px'),
          fontStyle: 'bold',
          color: '#fff6d8',
          backgroundColor: ruler?.color ?? '#5c4a32',
          padding: { x: compact ? 4 : 5, y: 2 },
        });
        ownerBadge.setOrigin(0.5, 0.5);
        ownerBadge.setShadow(0, 2, 'rgba(0,0,0,0.62)', 3, true, true);
        ownerBadge.setDepth(selected ? 8 : 6);
      }

      const label = this.add.text(point.x, point.y + markerSize * 0.52, labelText(city, ruler), {
        fontFamily: '"Noto Serif SC", "Songti SC", serif',
        fontSize: selected ? (compact ? '13px' : '17px') : (compact ? '12px' : '15px'),
        color: selected ? '#fff8d8' : '#fff1c7',
        backgroundColor: selected ? 'rgba(73, 31, 20, 0.9)' : 'rgba(29, 18, 13, 0.72)',
        padding: { x: selected ? (compact ? 6 : 9) : (compact ? 4 : 7), y: compact ? 2 : 4 },
      });
      label.setOrigin(0.5, 0);
      label.setShadow(0, 2, 'rgba(0,0,0,0.6)', 4, true, true);
      label.setDepth(selected ? 7 : 4);
    }
  }
}

type Point = {
  x: number;
  y: number;
};

function labelText(city: City, ruler?: Ruler): string {
  void ruler;
  return city.name;
}

function terrainRoutePoints(start: Point, end: Point, width: number, height: number, seed: string): Point[] {
  const dx = end.x - start.x;
  const dy = end.y - start.y;
  const distance = Math.max(1, Math.hypot(dx, dy));
  const perpendicular = { x: -dy / distance, y: dx / distance };
  const hash = hashString(seed);
  const direction = hash % 2 === 0 ? 1 : -1;
  const bend = direction * clamp(distance * 0.26, 24, Math.min(width, height) * 0.18);

  const first = shortenPoint(start, end, 18);
  const last = shortenPoint(end, start, 18);
  const controlA = terrainControlPoint(first, last, perpendicular, bend, width, height, 0.34, hash);
  const controlB = terrainControlPoint(first, last, perpendicular, bend * 0.58, width, height, 0.68, hash >> 3);
  const points: Point[] = [];
  const steps = distance > 520 ? 18 : distance > 280 ? 14 : 10;

  for (let i = 0; i <= steps; i++) {
    const t = i / steps;
    points.push(cubicBezier(first, controlA, controlB, last, t));
  }
  return points;
}

function terrainControlPoint(
  start: Point,
  end: Point,
  perpendicular: Point,
  bend: number,
  width: number,
  height: number,
  t: number,
  hash: number,
): Point {
  const base = lerpPoint(start, end, t);
  const bias = terrainBias(base, width, height, hash);
  return clampPoint({
    x: base.x + perpendicular.x * bend + bias.x,
    y: base.y + perpendicular.y * bend + bias.y,
  }, width, height);
}

function terrainBias(point: Point, width: number, height: number, hash: number): Point {
  const x = point.x / width;
  const y = point.y / height;
  const bias = { x: 0, y: 0 };

  // West and south-west mountains: routes descend into valleys instead of
  // cutting straight across ridges.
  if (x < 0.34 && y > 0.25) {
    bias.x += width * 0.025;
    bias.y -= height * 0.02;
  }
  // Central river belt: nudge roads along the watercourse seen on the map.
  if (x > 0.32 && x < 0.78 && y > 0.28 && y < 0.72) {
    bias.x += Math.sin((x + hash * 0.001) * Math.PI * 2) * width * 0.025;
    bias.y += Math.cos((x + 0.15) * Math.PI * 2) * height * 0.035;
  }
  // Eastern coast: lean paths slightly north/south along the coastline.
  if (x > 0.76) {
    bias.y += y < 0.5 ? height * 0.025 : -height * 0.018;
  }
  return bias;
}

function strokeRoute(graphics: Phaser.GameObjects.Graphics, points: Point[]) {
  for (let i = 1; i < points.length; i++) {
    const previous = points[i - 1];
    const current = points[i];
    graphics.lineBetween(previous.x, previous.y, current.x, current.y);
  }
}

function cubicBezier(a: Point, b: Point, c: Point, d: Point, t: number): Point {
  const mt = 1 - t;
  return {
    x: mt * mt * mt * a.x + 3 * mt * mt * t * b.x + 3 * mt * t * t * c.x + t * t * t * d.x,
    y: mt * mt * mt * a.y + 3 * mt * mt * t * b.y + 3 * mt * t * t * c.y + t * t * t * d.y,
  };
}

function lerpPoint(a: Point, b: Point, t: number): Point {
  return {
    x: a.x + (b.x - a.x) * t,
    y: a.y + (b.y - a.y) * t,
  };
}

function shortenPoint(from: Point, to: Point, amount: number): Point {
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const distance = Math.max(1, Math.hypot(dx, dy));
  return {
    x: from.x + (dx / distance) * amount,
    y: from.y + (dy / distance) * amount,
  };
}

function clampPoint(point: Point, width: number, height: number): Point {
  return {
    x: clamp(point.x, 24, width - 24),
    y: clamp(point.y, 24, height - 24),
  };
}

function hashString(value: string): number {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash * 31 + value.charCodeAt(i)) >>> 0;
  }
  return hash;
}

function cityRole(city: City): 'capital' | 'frontier' | 'fort' | 'port' | 'town' {
  const capitals = new Set(['洛阳', '长安', '许昌', '邺', '成都', '建业', '襄阳']);
  const frontier = new Set(['西凉', '安定', '汉中', '巴郡', '云南', '武陵', '零陵', '桂阳']);
  const forts = new Set(['平原', '南皮', '晋阳', '天水', '宛城', '寿春']);
  const ports = new Set(['江夏', '江陵', '庐江', '长沙', '吴', '会稽']);
  if (capitals.has(city.name)) {
    return 'capital';
  }
  if (ports.has(city.name)) {
    return 'port';
  }
  if (frontier.has(city.name)) {
    return 'frontier';
  }
  if (forts.has(city.name)) {
    return 'fort';
  }
  return 'town';
}

function cityMarkerKey(role: ReturnType<typeof cityRole>): string {
  return `city-${role}`;
}

function markerDisplaySize(role: ReturnType<typeof cityRole>, compact: boolean, selected: boolean): number {
  const base = role === 'capital' ? 54 : role === 'frontier' ? 48 : role === 'fort' ? 46 : role === 'port' ? 45 : 42;
  const compactScale = compact ? 0.72 : 1;
  const selectedScale = selected ? 1.24 : 1;
  return Math.round(base * compactScale * selectedScale);
}

function rulerSurname(ruler?: Ruler): string {
  if (!ruler || ruler.id === 'neutral' || !ruler.name) {
    return '无';
  }
  if (ruler.name.startsWith('公孙')) {
    return '公孙';
  }
  return ruler.name.slice(0, 1);
}

function colorToNumber(color: string): number {
  return Number.parseInt(color.replace('#', ''), 16);
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}
