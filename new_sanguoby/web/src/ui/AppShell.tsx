import {useCallback, useEffect, useMemo, useState} from 'react';
import {advanceMonth, applyCommand, createGame, getCurrentGame, getScenarios, launchBattle} from '../api/client';
import type {BattleOutcome, BattleRequest, City, GameSnapshot, RulerOption, ScenarioOption} from '../api/types';
import {useStrategyMusicAutoStart} from '../audio/useStrategyMusic';
import {CampaignMap} from '../phaser/CampaignMap';
import {BattlefieldOverlay} from './BattlefieldOverlay';
import {BattleReport} from './BattleReport';
import {Hud} from './Hud';
import {StartScreen} from './StartScreen';

type AppMode = 'main' | 'period' | 'ruler' | 'about' | 'game';

export function AppShell() {
  useStrategyMusicAutoStart();
  const [snapshot, setSnapshot] = useState<GameSnapshot | null>(null);
  const [scenarios, setScenarios] = useState<ScenarioOption[]>([]);
  const [selectedScenario, setSelectedScenario] = useState<ScenarioOption | null>(null);
  const [selectedCityId, setSelectedCityId] = useState('');
  const [mode, setMode] = useState<AppMode>('main');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastBattle, setLastBattle] = useState<BattleOutcome | null>(null);
  const [pendingBattle, setPendingBattle] = useState<BattleRequest | null>(null);

  const loadBootstrap = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const scenarioList = await getScenarios();
      setScenarios(scenarioList.scenarios);
      setSelectedScenario(scenarioList.scenarios[0] ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '读取旧档案失败');
    } finally {
      setBusy(false);
    }
  }, []);

  useEffect(() => {
    void loadBootstrap();
  }, [loadBootstrap]);

  const selectedCity = useMemo<City | null>(() => {
    if (!snapshot) {
      return null;
    }
    return snapshot.cities.find((city) => city.id === selectedCityId) ?? snapshot.cities[0] ?? null;
  }, [selectedCityId, snapshot]);

  const enterGame = useCallback((next: GameSnapshot, preferredCityId = '') => {
    const preferredCity = preferredCityId
      ? next.cities.find((city) => city.id === preferredCityId)
      : null;
    setSnapshot(next);
    setSelectedCityId(preferredCity?.id ?? next.cities.find((city) => city.ownerId === next.playerId)?.id ?? next.cities[0]?.id ?? '');
    setMode('game');
  }, []);

  const handleScenarioSelected = useCallback((scenario: ScenarioOption) => {
    setLastBattle(null);
    setPendingBattle(null);
    setSelectedScenario(scenario);
    setMode('ruler');
  }, []);

  const handleRulerSelected = useCallback(async (ruler: RulerOption) => {
    if (!selectedScenario) {
      return;
    }
    setBusy(true);
    setError(null);
    setLastBattle(null);
    setPendingBattle(null);
    try {
      enterGame(await createGame({ scenarioId: selectedScenario.id, playerId: ruler.id }));
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建新游戏失败');
    } finally {
      setBusy(false);
    }
  }, [enterGame, selectedScenario]);

  const handleContinue = useCallback(async () => {
    setBusy(true);
    setError(null);
    setLastBattle(null);
    setPendingBattle(null);
    try {
      enterGame(await getCurrentGame());
    } catch (err) {
      setError(err instanceof Error ? err.message : '读取战局失败');
    } finally {
      setBusy(false);
    }
  }, [enterGame]);

  const handleAdvanceMonth = useCallback(async () => {
    setBusy(true);
    setError(null);
    setLastBattle(null);
    setPendingBattle(null);
    try {
      enterGame(await advanceMonth(), selectedCity?.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : '推进月份失败');
    } finally {
      setBusy(false);
    }
  }, [enterGame, selectedCity]);

  const handleCommand = useCallback(async (commandId: string, generalId: string, targetCityId?: string, targetGeneralId?: string) => {
    if (!selectedCity) {
      return;
    }
    setBusy(true);
    setError(null);
    setLastBattle(null);
    setPendingBattle(null);
    try {
      enterGame(
        await applyCommand({ cityId: selectedCity.id, generalId, commandId, targetCityId, targetGeneralId }),
        commandId === 'move' && targetCityId ? targetCityId : selectedCity.id,
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : '命令执行失败');
    } finally {
      setBusy(false);
    }
  }, [enterGame, selectedCity]);

  const handleBattle = useCallback((request: BattleRequest) => {
    setError(null);
    setLastBattle(null);
    setPendingBattle(request);
  }, []);

  const handleBattleResolve = useCallback(async (request: BattleRequest, fieldAdvantage: number, remainingFood: number) => {
    if (!snapshot) {
      return;
    }
    setBusy(true);
    setError(null);
    setLastBattle(null);
    try {
      const result = await launchBattle({ ...request, fieldAdvantage, remainingFood });
      const followCityId = result.outcome.captured ? result.outcome.targetCityId : request.cityId;
      setPendingBattle(null);
      enterGame(result.snapshot, followCityId);
      setLastBattle(result.outcome);
    } catch (err) {
      setError(err instanceof Error ? err.message : '出征失败');
    } finally {
      setBusy(false);
    }
  }, [enterGame, snapshot]);

  if (mode !== 'game') {
    return (
      <StartScreen
        mode={mode}
        scenarios={scenarios}
        selectedScenario={selectedScenario}
        busy={busy}
        error={error}
        onModeChange={setMode}
        onScenarioSelected={handleScenarioSelected}
        onRulerSelected={handleRulerSelected}
        onContinue={handleContinue}
      />
    );
  }

  if (!snapshot || !selectedCity) {
    return (
      <main className="loading-screen">
        <h1>三国霸业2026重置版</h1>
        <p>{error ?? '正在整军备战...'}</p>
        {error ? <button type="button" onClick={loadBootstrap}>重试</button> : null}
      </main>
    );
  }

  return (
    <main className="app-shell">
      <CampaignMap snapshot={snapshot} selectedCityId={selectedCity.id} onCitySelected={setSelectedCityId} />
      <Hud
        snapshot={snapshot}
        selectedCity={selectedCity}
        onMainMenu={() => {
          setLastBattle(null);
          setPendingBattle(null);
          setMode('main');
        }}
        onEndStrategy={handleAdvanceMonth}
        onCommand={handleCommand}
        onBattle={handleBattle}
        busy={busy}
      />
      {lastBattle ? (
        <BattleReport outcome={lastBattle} snapshot={snapshot} onClose={() => setLastBattle(null)} />
      ) : null}
      {pendingBattle ? (
        <BattlefieldOverlay
          snapshot={snapshot}
          request={pendingBattle}
          busy={busy}
          onResolve={handleBattleResolve}
          onCancel={() => setPendingBattle(null)}
        />
      ) : null}
      {error ? <div role="alert" className="error-toast">{error}</div> : null}
    </main>
  );
}
