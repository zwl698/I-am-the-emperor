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
    this.load.svg('city-marker', '/assets/city-marker.svg', { width: 64, height: 64 });
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
    graphics.fillStyle(0x160e0d, 0.16);
    graphics.fillRect(0, 0, width, height);

    graphics.lineStyle(1, 0xf5dc9a, 0.05);
    for (let i = 1; i < 12; i++) {
      const x = (width / 12) * i;
      graphics.lineBetween(x, 0, x, height);
    }
    for (let i = 1; i < 9; i++) {
      const y = (height / 9) * i;
      graphics.lineBetween(0, y, width, y);
    }

    graphics.lineStyle(28, 0x15100f, 0.22);
    graphics.strokeRect(12, 12, width - 24, height - 24);
    graphics.lineStyle(2, 0xffe3a4, 0.28);
    graphics.strokeRect(28, 28, width - 56, height - 56);
  }

  private drawRoutes(width: number, height: number, snapshot: GameSnapshot) {
    const cityByID = new Map(snapshot.cities.map((city) => [city.id, city]));
    const graphics = this.add.graphics();
    const compact = snapshot.cities.length > 24 || width < 760;
    graphics.lineStyle(compact ? 4 : 8, 0x25150e, compact ? 0.16 : 0.26);
    for (const route of snapshot.routes) {
      const from = cityByID.get(route.from);
      const to = cityByID.get(route.to);
      if (!from || !to) {
        continue;
      }
      const start = projectCity(from, { width, height });
      const end = projectCity(to, { width, height });
      graphics.lineBetween(start.x, start.y, end.x, end.y);
    }
    graphics.lineStyle(compact ? 1.5 : 3, 0xf0c978, compact ? 0.42 : 0.68);
    for (const route of snapshot.routes) {
      const from = cityByID.get(route.from);
      const to = cityByID.get(route.to);
      if (!from || !to) {
        continue;
      }
      const start = projectCity(from, { width, height });
      const end = projectCity(to, { width, height });
      graphics.lineBetween(start.x, start.y, end.x, end.y);
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

      const halo = this.add.circle(point.x, point.y, selected ? (compact ? 27 : 37) : (compact ? 21 : 29), tint, selected ? 0.36 : 0.2);
      halo.setStrokeStyle(selected ? 4 : 2, selected ? 0xfff3bd : 0xffe6a8, selected ? 0.86 : 0.42);
      halo.setDepth(selected ? 2 : 0);

      if (selected) {
        const focus = this.add.circle(point.x, point.y, compact ? 35 : 47, 0xfff3bd, 0);
        focus.setStrokeStyle(2, 0x2b1710, 0.7);
        focus.setDepth(1);
      }

      const base = this.add.circle(point.x, point.y + 4, compact ? 13 : 18, 0x1d1410, 0.42);
      base.setScale(1.32, 0.42);
      base.setDepth(selected ? 3 : 1);

      const marker = this.add.image(point.x, point.y, 'city-marker');
      marker.setDisplaySize(selected ? (compact ? 42 : 58) : (compact ? 33 : 48), selected ? (compact ? 42 : 58) : (compact ? 33 : 48));
      marker.setTint(tint);
      marker.setInteractive({ useHandCursor: true });
      marker.setDepth(selected ? 5 : 3);
      marker.on('pointerdown', () => {
        this.selectedCityId = city.id;
        this.onCitySelected(city.id);
        this.renderSnapshot();
      });

      if (city.ownerId !== 'neutral') {
        const banner = this.add.image(point.x + 22, point.y - 22, 'army-banner');
        banner.setDisplaySize(selected ? (compact ? 23 : 34) : (compact ? 18 : 28), selected ? (compact ? 23 : 34) : (compact ? 18 : 28));
        banner.setTint(tint);
        banner.setDepth(selected ? 6 : 4);
      }

      const label = this.add.text(point.x, point.y + 25, labelText(city, ruler), {
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

function labelText(city: City, ruler?: Ruler): string {
  void ruler;
  return city.name;
}

function colorToNumber(color: string): number {
  return Number.parseInt(color.replace('#', ''), 16);
}
