import { useEffect, useMemo, useRef, useState, type CSSProperties } from 'react';
import Phaser from 'phaser';
import type { City, GameSnapshot, Ruler } from '../api/types';
import { projectCity } from '../game/mapProjection';
import { CampaignScene } from './scenes/CampaignScene';

type CampaignMapProps = {
  snapshot: GameSnapshot;
  selectedCityId: string;
  onCitySelected: (cityId: string) => void;
};

export function CampaignMap({ snapshot, selectedCityId, onCitySelected }: CampaignMapProps) {
  const hostRef = useRef<HTMLDivElement | null>(null);
  const sceneRef = useRef<CampaignScene | null>(null);
  const gameRef = useRef<Phaser.Game | null>(null);
  const [viewport, setViewport] = useState({ width: 0, height: 0 });
  const rulerByID = useMemo(() => new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler])), [snapshot.rulers]);
  const cityLabels = useMemo(() => {
    if (!viewport.width || !viewport.height) {
      return [];
    }
    return snapshot.cities.map((city) => ({
      city,
      point: projectCity(city, viewport),
      ruler: rulerByID.get(city.ownerId),
      selected: city.id === selectedCityId,
    }));
  }, [rulerByID, selectedCityId, snapshot.cities, viewport]);

  useEffect(() => {
    if (!hostRef.current || gameRef.current) {
      return;
    }

    const scene = new CampaignScene({ onCitySelected });
    sceneRef.current = scene;
    gameRef.current = new Phaser.Game({
      type: Phaser.AUTO,
      parent: hostRef.current,
      backgroundColor: '#d8c89c',
      scale: {
        mode: Phaser.Scale.RESIZE,
        width: hostRef.current.clientWidth,
        height: hostRef.current.clientHeight,
      },
      render: {
        antialias: true,
        pixelArt: false,
        roundPixels: false,
      },
      scene,
    });

    return () => {
      gameRef.current?.destroy(true);
      gameRef.current = null;
      sceneRef.current = null;
    };
  }, [onCitySelected]);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return;
    }

    const updateSize = () => {
      setViewport({
        width: host.clientWidth,
        height: host.clientHeight,
      });
    };
    updateSize();

    const observer = new ResizeObserver(updateSize);
    observer.observe(host);

    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    sceneRef.current?.setSnapshot(snapshot);
  }, [snapshot]);

  useEffect(() => {
    sceneRef.current?.setSelectedCity(selectedCityId);
  }, [selectedCityId]);

  return (
    <div ref={hostRef} className="campaign-map" aria-label="战略地图">
      <div className="campaign-map__labels" aria-label="城池标注">
        {cityLabels.map(({ city, point, ruler, selected }) => (
          <CityLabel
            key={city.id}
            city={city}
            point={point}
            ruler={ruler}
            selected={selected}
            onSelected={onCitySelected}
          />
        ))}
      </div>
    </div>
  );
}

function CityLabel({
  city,
  point,
  ruler,
  selected,
  onSelected,
}: {
  city: City;
  point: { x: number; y: number };
  ruler?: Ruler;
  selected: boolean;
  onSelected: (cityId: string) => void;
}) {
  const style = {
    left: `${point.x}px`,
    top: `${point.y}px`,
    '--city-owner-color': ruler?.color ?? '#8f8979',
  } as CSSProperties;

  return (
    <button
      type="button"
      className={`map-city-label${selected ? ' map-city-label--selected' : ''}`}
      style={style}
      aria-label={`选择${city.name}`}
      onClick={(event) => {
        event.stopPropagation();
        onSelected(city.id);
      }}
    >
      {city.ownerId !== 'neutral' ? (
        <span className="map-city-label__owner">{rulerSurname(ruler)}</span>
      ) : null}
      <span className="map-city-label__name">{city.name}</span>
    </button>
  );
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
