import {useCallback, useEffect, useMemo, useState} from 'react';
import {advanceMonth, createGame, getCurrentGame, getLegacyResources} from '../api/client';
import type {City, GameSnapshot, LegacyResources} from '../api/types';
import {summarizeLegacyInventory} from '../game/legacyInventory';
import {CampaignMap} from '../phaser/CampaignMap';
import {Hud} from './Hud';

export function AppShell() {
  const [snapshot, setSnapshot] = useState<GameSnapshot | null>(null);
  const [legacyResources, setLegacyResources] = useState<LegacyResources | null>(null);
  const [selectedCityId, setSelectedCityId] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadCurrent = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const [next, legacy] = await Promise.all([
        getCurrentGame(),
        getLegacyResources().catch(() => null),
      ]);
      setSnapshot(next);
      setLegacyResources(legacy);
      setSelectedCityId(next.cities.find((city) => city.ownerId === next.playerId)?.id ?? next.cities[0]?.id ?? '');
    } catch (err) {
      setError(err instanceof Error ? err.message : '读取游戏状态失败');
    } finally {
      setBusy(false);
    }
  }, []);

  useEffect(() => {
    void loadCurrent();
  }, [loadCurrent]);

  const selectedCity = useMemo<City | null>(() => {
    if (!snapshot) {
      return null;
    }
    return snapshot.cities.find((city) => city.id === selectedCityId) ?? snapshot.cities[0] ?? null;
  }, [selectedCityId, snapshot]);

  const legacySummary = useMemo(() => summarizeLegacyInventory(legacyResources), [legacyResources]);

  const handleNewGame = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const next = await createGame({ scenarioId: 'dongzhuo', playerId: '' });
      setSnapshot(next);
      setSelectedCityId(
        next.cities.find((city) => city.ownerId === next.playerId)?.id ?? next.cities[0]?.id ?? '',
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建新游戏失败');
    } finally {
      setBusy(false);
    }
  }, []);

  const handleAdvanceMonth = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      setSnapshot(await advanceMonth());
    } catch (err) {
      setError(err instanceof Error ? err.message : '推进月份失败');
    } finally {
      setBusy(false);
    }
  }, []);

  if (!snapshot || !selectedCity) {
    return (
      <main className="loading-screen">
        <h1>三国霸业</h1>
        <p>{error ?? '正在整军备战...'}</p>
        {error ? <button type="button" onClick={loadCurrent}>重试</button> : null}
      </main>
    );
  }

  return (
    <main className="app-shell">
      <CampaignMap snapshot={snapshot} onCitySelected={setSelectedCityId} />
      <Hud
        snapshot={snapshot}
        selectedCity={selectedCity}
        onNewGame={handleNewGame}
        onAdvanceMonth={handleAdvanceMonth}
        busy={busy}
        legacySummary={legacySummary}
      />
      {error ? <div role="alert" className="error-toast">{error}</div> : null}
    </main>
  );
}
