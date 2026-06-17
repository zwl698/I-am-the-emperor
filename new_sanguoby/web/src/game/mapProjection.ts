export type CityGridPoint = {
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
  const padding = Math.round(Math.min(viewport.width, viewport.height) / 9);
  const x = padding + (city.x / LEGACY_GRID_COLUMNS) * (viewport.width - padding);
  const y = padding + (city.y / LEGACY_GRID_ROWS) * (viewport.height - padding * 2);

  return {
    x: Math.round(clamp(x, padding, viewport.width - padding)),
    y: Math.round(clamp(y, padding, viewport.height - padding)),
  };
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}
