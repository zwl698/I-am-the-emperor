import Phaser from 'phaser';
import type { City, GameSnapshot, Ruler } from '../../api/types';
import { projectCity } from '../../game/mapProjection';

type CampaignSceneOptions = {
  onCitySelected: (cityId: string) => void;
};

type FactionShape = 'hex' | 'shield' | 'diamond' | 'square' | 'circle' | 'banner' | 'octagon' | 'triangle';
type FactionPattern = 'slash' | 'cross' | 'bars' | 'dot' | 'chevron' | 'split';

type FactionStyle = {
  primary: number;
  secondary: number;
  dark: number;
  light: number;
  icon: number;
  textHex: string;
  shape: FactionShape;
  pattern: FactionPattern;
};

type FactionPreset = {
  primary: string;
  secondary: string;
  dark: string;
  light: string;
  icon: string;
  shape: FactionShape;
  pattern: FactionPattern;
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
      const style = factionStyleFor(ruler, city.ownerId);
      const selected = city.id === this.selectedCityId;
      const role = cityRole(city);
      const markerKey = cityMarkerKey(role);
      const markerSize = markerDisplaySize(role, compact, selected);

      const halo = this.add.circle(point.x, point.y, selected ? markerSize * 0.9 : markerSize * 0.7, style.primary, selected ? 0.34 : 0.2);
      halo.setStrokeStyle(selected ? 4 : 3, selected ? style.light : style.primary, selected ? 0.96 : 0.82);
      halo.setDepth(selected ? 2 : 0);

      if (selected) {
        const focus = this.add.circle(point.x, point.y, markerSize * 0.86, 0xfff3bd, 0);
        focus.setStrokeStyle(2, 0x2b1710, 0.7);
        focus.setDepth(1);
      }

      drawFactionFrame(this, point.x, point.y, markerSize * (selected ? 1.34 : 1.18), style, selected);

      const base = this.add.circle(point.x, point.y + markerSize * 0.08, markerSize * 0.34, 0x1d1410, 0.42);
      base.setScale(1.32, 0.42);
      base.setDepth(selected ? 3 : 1);

      const marker = this.add.image(point.x, point.y, markerKey);
      marker.setDisplaySize(markerSize, markerSize);
      marker.setTint(style.icon);
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
        banner.setTint(style.secondary);
        banner.setDepth(selected ? 6 : 4);

        const surname = rulerSurname(ruler);
        const badgeX = point.x + markerSize * 0.43;
        const badgeY = point.y - markerSize * 0.66;
        drawOwnerBadgePlate(this, badgeX, badgeY, markerSize * (surname.length > 1 ? 0.72 : 0.56), markerSize * 0.36, style, selected);
        const ownerBadge = this.add.text(badgeX, badgeY, surname, {
          fontFamily: '"Noto Serif SC", "Songti SC", serif',
          fontSize: selected ? (compact ? '12px' : '15px') : (compact ? '10px' : '13px'),
          fontStyle: 'bold',
          color: style.textHex,
          padding: { x: 0, y: 0 },
        });
        ownerBadge.setOrigin(0.5, 0.5);
        ownerBadge.setShadow(0, 2, 'rgba(0,0,0,0.62)', 3, true, true);
        ownerBadge.setDepth(selected ? 9 : 7);
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

const FACTION_PRESETS: FactionPreset[] = [
  { primary: '#d93832', secondary: '#ff9b4a', dark: '#35100d', light: '#ffe0ad', icon: '#ffe4bc', shape: 'shield', pattern: 'chevron' },
  { primary: '#11a36b', secondary: '#42dcc0', dark: '#083026', light: '#c9ffe9', icon: '#d8fff1', shape: 'hex', pattern: 'cross' },
  { primary: '#336dff', secondary: '#6ee2ff', dark: '#0b1d4a', light: '#d9e7ff', icon: '#dfeaff', shape: 'diamond', pattern: 'slash' },
  { primary: '#9c54ff', secondary: '#ff71d9', dark: '#281044', light: '#ecd9ff', icon: '#f0ddff', shape: 'octagon', pattern: 'dot' },
  { primary: '#d29322', secondary: '#ffe15c', dark: '#3b2307', light: '#fff0b4', icon: '#fff1c5', shape: 'square', pattern: 'bars' },
  { primary: '#05aab8', secondary: '#7afff1', dark: '#06343a', light: '#cffff8', icon: '#dcfffb', shape: 'banner', pattern: 'split' },
  { primary: '#d9498f', secondary: '#ff99c9', dark: '#3a1028', light: '#ffd7ec', icon: '#ffe0f0', shape: 'circle', pattern: 'dot' },
  { primary: '#74a737', secondary: '#c7ee61', dark: '#1e2d0d', light: '#e8ffc1', icon: '#f0ffd2', shape: 'triangle', pattern: 'chevron' },
  { primary: '#c7652d', secondary: '#ffbd67', dark: '#351606', light: '#ffe0b8', icon: '#ffe7ca', shape: 'shield', pattern: 'bars' },
  { primary: '#5864ff', secondary: '#9ab6ff', dark: '#15164a', light: '#e0e5ff', icon: '#e7ebff', shape: 'diamond', pattern: 'cross' },
  { primary: '#c84f3f', secondary: '#ffd064', dark: '#35110b', light: '#ffd9c9', icon: '#ffe3d8', shape: 'hex', pattern: 'slash' },
  { primary: '#2aa36f', secondary: '#9cf065', dark: '#0a2d20', light: '#d5ffcc', icon: '#e6ffdd', shape: 'banner', pattern: 'chevron' },
  { primary: '#c64fbd', secondary: '#ffc3f1', dark: '#331034', light: '#ffe2fb', icon: '#ffe8fb', shape: 'octagon', pattern: 'split' },
  { primary: '#6fad35', secondary: '#ffe76a', dark: '#213008', light: '#eeffc0', icon: '#f5ffd6', shape: 'circle', pattern: 'cross' },
  { primary: '#0089b8', secondary: '#80d5ff', dark: '#06283b', light: '#d4f0ff', icon: '#e0f5ff', shape: 'square', pattern: 'slash' },
  { primary: '#c8a42c', secondary: '#fff49a', dark: '#352907', light: '#fff4bf', icon: '#fff7d4', shape: 'triangle', pattern: 'bars' },
  { primary: '#884bd6', secondary: '#caa2ff', dark: '#231041', light: '#eadbff', icon: '#f1e5ff', shape: 'shield', pattern: 'dot' },
  { primary: '#1fa88d', secondary: '#7bf0d7', dark: '#08332d', light: '#d2fff5', icon: '#e0fff9', shape: 'hex', pattern: 'split' },
  { primary: '#d84a68', secondary: '#ffad73', dark: '#38121b', light: '#ffdce1', icon: '#ffe6ea', shape: 'banner', pattern: 'cross' },
  { primary: '#42a846', secondary: '#8ee9a5', dark: '#0f2d13', light: '#d8ffd9', icon: '#e6ffe8', shape: 'diamond', pattern: 'chevron' },
];

const NEUTRAL_STYLE: FactionStyle = {
  primary: 0x8f8979,
  secondary: 0xd0c4a4,
  dark: 0x2d2923,
  light: 0xe8dfc6,
  icon: 0xefe2bf,
  textHex: '#fff4d4',
  shape: 'circle',
  pattern: 'bars',
};

const FACTION_STYLE_BY_NAME: Record<string, number> = {
  '董卓': 0,
  '曹操': 1,
  '袁绍': 2,
  '袁术': 3,
  '孙坚': 4,
  '马腾': 5,
  '陶谦': 6,
  '孔融': 7,
  '刘岱': 8,
  '张杨': 9,
  '韩馥': 10,
  '王匡': 11,
  '刘备': 12,
  '刘表': 13,
  '刘焉': 14,
  '公孙瓒': 15,
  '公孙度': 16,
  '张鲁': 17,
  '孙权': 4,
};

function factionStyleFor(ruler?: Ruler, ownerID?: string): FactionStyle {
  if (!ruler || ownerID === 'neutral' || ruler.id === 'neutral') {
    return NEUTRAL_STYLE;
  }
  const preset = FACTION_PRESETS[factionStyleIndex(ruler) % FACTION_PRESETS.length];
  return {
    primary: colorToNumber(preset.primary),
    secondary: colorToNumber(preset.secondary),
    dark: colorToNumber(preset.dark),
    light: colorToNumber(preset.light),
    icon: colorToNumber(preset.icon),
    textHex: '#fff8df',
    shape: preset.shape,
    pattern: preset.pattern,
  };
}

function factionStyleIndex(ruler: Ruler): number {
  for (const [name, index] of Object.entries(FACTION_STYLE_BY_NAME)) {
    if (ruler.name.startsWith(name)) {
      return index;
    }
  }
  return hashString(`${ruler.id}:${ruler.name}`) % FACTION_PRESETS.length;
}

function drawFactionFrame(scene: Phaser.Scene, x: number, y: number, size: number, style: FactionStyle, selected: boolean) {
  const shadow = scene.add.graphics();
  shadow.fillStyle(0x080504, selected ? 0.58 : 0.46);
  fillFactionShape(shadow, x + 1, y + size * 0.06, size * 1.08, style.shape);
  shadow.setDepth(selected ? 3.4 : 1.4);

  const frame = scene.add.graphics();
  frame.fillStyle(style.dark, selected ? 0.96 : 0.88);
  fillFactionShape(frame, x, y, size * 1.08, style.shape);
  frame.fillStyle(style.primary, selected ? 0.92 : 0.82);
  fillFactionShape(frame, x, y, size * 0.94, style.shape);
  drawFactionPattern(frame, x, y, size, style, selected);
  frame.lineStyle(selected ? 4 : 3, style.light, selected ? 0.98 : 0.82);
  strokeFactionShape(frame, x, y, size * 1.08, style.shape);
  frame.lineStyle(1.5, style.dark, 0.72);
  strokeFactionShape(frame, x, y, size * 0.88, style.shape);
  frame.setDepth(selected ? 4 : 2);
}

function drawFactionPattern(graphics: Phaser.GameObjects.Graphics, x: number, y: number, size: number, style: FactionStyle, selected: boolean) {
  const lineWidth = Math.max(2, size * 0.045);
  graphics.lineStyle(lineWidth, style.light, selected ? 0.68 : 0.52);
  switch (style.pattern) {
    case 'slash':
      graphics.lineBetween(x - size * 0.28, y + size * 0.24, x + size * 0.28, y - size * 0.24);
      break;
    case 'cross':
      graphics.lineBetween(x - size * 0.26, y, x + size * 0.26, y);
      graphics.lineBetween(x, y - size * 0.26, x, y + size * 0.26);
      break;
    case 'bars':
      graphics.lineBetween(x - size * 0.28, y - size * 0.12, x + size * 0.28, y - size * 0.12);
      graphics.lineBetween(x - size * 0.28, y + size * 0.13, x + size * 0.28, y + size * 0.13);
      break;
    case 'dot':
      graphics.fillStyle(style.secondary, selected ? 0.86 : 0.7);
      graphics.fillCircle(x, y, size * 0.11);
      graphics.fillCircle(x - size * 0.23, y + size * 0.08, size * 0.055);
      graphics.fillCircle(x + size * 0.23, y - size * 0.08, size * 0.055);
      break;
    case 'chevron':
      graphics.lineBetween(x - size * 0.26, y + size * 0.12, x, y - size * 0.16);
      graphics.lineBetween(x, y - size * 0.16, x + size * 0.26, y + size * 0.12);
      break;
    case 'split':
      graphics.lineBetween(x, y - size * 0.31, x, y + size * 0.31);
      graphics.lineStyle(lineWidth * 0.74, style.secondary, selected ? 0.72 : 0.56);
      graphics.lineBetween(x - size * 0.22, y, x + size * 0.22, y);
      break;
  }
}

function drawOwnerBadgePlate(scene: Phaser.Scene, x: number, y: number, width: number, height: number, style: FactionStyle, selected: boolean) {
  const badge = scene.add.graphics();
  badge.fillStyle(style.dark, 0.95);
  badge.fillRoundedRect(x - width / 2 - 2, y - height / 2 + 2, width + 4, height + 4, 5);
  badge.fillStyle(style.primary, selected ? 0.98 : 0.9);
  badge.fillRoundedRect(x - width / 2, y - height / 2, width, height, 5);
  badge.lineStyle(selected ? 2 : 1.5, style.light, 0.92);
  badge.strokeRoundedRect(x - width / 2, y - height / 2, width, height, 5);
  badge.setDepth(selected ? 8 : 6);
}

function fillFactionShape(graphics: Phaser.GameObjects.Graphics, x: number, y: number, size: number, shape: FactionShape) {
  if (shape === 'circle') {
    graphics.fillCircle(x, y, size / 2);
    return;
  }
  if (shape === 'square') {
    graphics.fillRoundedRect(x - size / 2, y - size / 2, size, size, size * 0.16);
    return;
  }
  graphics.fillPoints(factionShapePoints(x, y, size, shape), true);
}

function strokeFactionShape(graphics: Phaser.GameObjects.Graphics, x: number, y: number, size: number, shape: FactionShape) {
  if (shape === 'circle') {
    graphics.strokeCircle(x, y, size / 2);
    return;
  }
  if (shape === 'square') {
    graphics.strokeRoundedRect(x - size / 2, y - size / 2, size, size, size * 0.16);
    return;
  }
  graphics.strokePoints(factionShapePoints(x, y, size, shape), true, true);
}

function factionShapePoints(x: number, y: number, size: number, shape: Exclude<FactionShape, 'circle' | 'square'>): Phaser.Math.Vector2[] {
  switch (shape) {
    case 'shield':
      return scaledPoints(x, y, size, [[0, -0.5], [0.46, -0.33], [0.4, 0.16], [0, 0.52], [-0.4, 0.16], [-0.46, -0.33]]);
    case 'diamond':
      return scaledPoints(x, y, size, [[0, -0.54], [0.54, 0], [0, 0.54], [-0.54, 0]]);
    case 'banner':
      return scaledPoints(x, y, size, [[-0.5, -0.42], [0.32, -0.42], [0.5, -0.21], [0.34, 0], [0.5, 0.21], [0.32, 0.42], [-0.5, 0.42]]);
    case 'triangle':
      return scaledPoints(x, y, size, [[0, -0.56], [0.52, 0.42], [-0.52, 0.42]]);
    case 'octagon':
      return regularPolygonPoints(x, y, size * 0.5, 8, Math.PI / 8);
    case 'hex':
    default:
      return regularPolygonPoints(x, y, size * 0.52, 6, Math.PI / 6);
  }
}

function regularPolygonPoints(x: number, y: number, radius: number, sides: number, angleOffset: number): Phaser.Math.Vector2[] {
  const points: Phaser.Math.Vector2[] = [];
  for (let i = 0; i < sides; i++) {
    const angle = -Math.PI / 2 + angleOffset + (Math.PI * 2 * i) / sides;
    points.push(new Phaser.Math.Vector2(x + Math.cos(angle) * radius, y + Math.sin(angle) * radius));
  }
  return points;
}

function scaledPoints(x: number, y: number, size: number, points: number[][]): Phaser.Math.Vector2[] {
  return points.map(([px, py]) => new Phaser.Math.Vector2(x + px * size, y + py * size));
}

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
