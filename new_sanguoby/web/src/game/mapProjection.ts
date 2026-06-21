export type CityGridPoint = {
  id?: string;
  name?: string;
  x: number;
  y: number;
};

export type ViewportSize = {
  width: number;
  height: number;
};

export type ProjectedPoint = {
  x: number;
  y: number;
};

const LEGACY_GRID_COLUMNS = 12;
const LEGACY_GRID_ROWS = 9;

export function projectCity(city: CityGridPoint, viewport: ViewportSize): ProjectedPoint {
  const paddingX = Math.round(Math.min(viewport.width, viewport.height) / 10);
  const paddingY = Math.round(Math.min(viewport.width, viewport.height) / 10.5);
  const spanX = viewport.width - paddingX * 2;
  const spanY = viewport.height - paddingY * 2;
  const cellX = spanX / LEGACY_GRID_COLUMNS;
  const cellY = spanY / LEGACY_GRID_ROWS;
  const nx = city.x / LEGACY_GRID_COLUMNS;
  const ny = city.y / LEGACY_GRID_ROWS;
  const key = `${city.id ?? ''}:${city.name ?? ''}:${city.x}:${city.y}`;
  const hash = hashString(key);
  const baseX = paddingX + nx * spanX;
  const baseY = paddingY + ny * spanY;
  const rowStagger = (city.y % 2 === 0 ? 0.045 : -0.045) * cellX;
  const terrainFlow = Math.sin((city.x + 1) * 0.82 + (city.y + 2) * 0.47) * cellY * 0.055;
  const jitterX = seededOffset(hash, cellX * 0.11);
  const jitterY = seededOffset(hash >>> 8, cellY * 0.12);

  const x = baseX + rowStagger + jitterX;
  const y = baseY + terrainFlow + jitterY;

  return {
    x: Math.round(clamp(x, paddingX, viewport.width - paddingX)),
    y: Math.round(clamp(y, paddingY, viewport.height - paddingY)),
  };
}

function seededOffset(hash: number, range: number): number {
  const normalized = ((hash % 2001) / 1000) - 1;
  return normalized * range;
}

function hashString(value: string): number {
  let hash = 2166136261;
  for (let i = 0; i < value.length; i++) {
    hash ^= value.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return hash >>> 0;
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}
