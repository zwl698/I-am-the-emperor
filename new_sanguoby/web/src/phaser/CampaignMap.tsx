import { useEffect, useRef } from 'react';
import Phaser from 'phaser';
import type { GameSnapshot } from '../api/types';
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
    sceneRef.current?.setSnapshot(snapshot);
  }, [snapshot]);

  useEffect(() => {
    sceneRef.current?.setSelectedCity(selectedCityId);
  }, [selectedCityId]);

  return <div ref={hostRef} className="campaign-map" aria-label="战略地图" />;
}
