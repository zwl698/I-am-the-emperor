import {useEffect, useRef, useState} from 'react';
import type {City, GameSnapshot, General, Ruler} from '../api/types';
import {portraitForGeneral, portraitForRuler} from '../game/portraitRegistry';
import { PortraitImage } from './PortraitImage';

type HudProps = {
  snapshot: GameSnapshot;
  selectedCity: City;
  onMainMenu: () => void;
  onEndStrategy: () => void;
  onCommand: (commandId: string, generalId: string) => void;
  onBattle: (generalId: string, targetCityId: string) => void;
  busy: boolean;
};

// adjacentCities returns enemy/neutral cities reachable from cityId in one hop.
function adjacentEnemyCities(snapshot: GameSnapshot, cityId: string): City[] {
  const cityByID = new Map(snapshot.cities.map((city) => [city.id, city]));
  const neighbours: City[] = [];
  const seen = new Set<string>();
  for (const route of snapshot.routes) {
    let otherId = '';
    if (route.from === cityId) {
      otherId = route.to;
    } else if (route.to === cityId) {
      otherId = route.from;
    } else {
      continue;
    }
    if (!otherId || seen.has(otherId)) {
      continue;
    }
    seen.add(otherId);
    const other = cityByID.get(otherId);
    if (other && other.ownerId !== snapshot.playerId) {
      neighbours.push(other);
    }
  }
  return neighbours;
}

const COMMAND_GROUPS = {
  '内政': [
    { id: 'assart', name: '开垦' },
    { id: 'commerce', name: '招商' },
    { id: 'search', name: '搜寻' },
    { id: 'govern', name: '治理' },
    { id: 'inspect', name: '出巡' },
    { id: 'surrender', name: '招降' },
  ],
  '外交': [
    { id: 'alienate', name: '离间' },
    { id: 'canvass', name: '招揽' },
    { id: 'counterespionage', name: '策反' },
    { id: 'realienate', name: '反间' },
    { id: 'induce', name: '劝降' },
  ],
  '军备': [
    { id: 'reconnoitre', name: '侦察' },
    { id: 'conscription', name: '征兵' },
    { id: 'distribute', name: '分配' },
    { id: 'depredate', name: '掠夺' },
    { id: 'battle', name: '出征' },
  ],
};

type CommandCategory = keyof typeof COMMAND_GROUPS | '状况';

export function Hud({ snapshot, selectedCity, onMainMenu, onEndStrategy, onCommand, onBattle, busy }: HudProps) {
  const [category, setCategory] = useState<CommandCategory>('内政');
  const [commandId, setCommandId] = useState(COMMAND_GROUPS['内政'][0].id);
  const [generalId, setGeneralId] = useState('');
  const [targetCityId, setTargetCityId] = useState('');
  const rulerByID = new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler]));
  const owner = rulerByID.get(selectedCity.ownerId);
  const generals = snapshot.generals.filter((general) => general.cityId === selectedCity.id);
  const playerGenerals = generals.filter((general) => general.ownerId === snapshot.playerId);
  const player = rulerByID.get(snapshot.playerId);
  const ownerPortrait = portraitForRuler(owner);
  const playable = selectedCity.ownerId === snapshot.playerId;
  const activeCommands = category === '状况' ? [] : COMMAND_GROUPS[category];
  const activeCommand = activeCommands.find((command) => command.id === commandId) ?? activeCommands[0];
  const selectedGeneral = playerGenerals.find((general) => general.id === generalId) ?? playerGenerals[0];
  const isBattle = activeCommand?.id === 'battle';
  const battleTargets = isBattle ? adjacentEnemyCities(snapshot, selectedCity.id) : [];
  const selectedTarget = battleTargets.find((city) => city.id === targetCityId) ?? battleTargets[0];

  useEffect(() => {
    const firstGeneral = playerGenerals[0]?.id ?? '';
    setGeneralId(firstGeneral);
  }, [selectedCity.id, snapshot.playerId]);

  useEffect(() => {
    if (category === '状况') {
      return;
    }
    const firstCommand = COMMAND_GROUPS[category][0]?.id ?? '';
    setCommandId(firstCommand);
  }, [category]);

  useEffect(() => {
    setTargetCityId(adjacentEnemyCities(snapshot, selectedCity.id)[0]?.id ?? '');
  }, [selectedCity.id, commandId, snapshot]);

  return (
    <>
      <header className="topbar">
        <div>
          <h1>三国霸业</h1>
          <p>{snapshot.date.year}年 {snapshot.date.month}月 · {player?.name ?? '未定'} 执政</p>
        </div>
        <div className="topbar-actions">
          <button type="button" onClick={onMainMenu} disabled={busy}>主菜单</button>
          <button type="button" className="primary" onClick={onEndStrategy} disabled={busy}>策略结束</button>
        </div>
      </header>

      <aside className="status-rail">
        <section>
          <div className="city-hero">
            <PortraitImage
              src={ownerPortrait}
              alt={`${ownerLabel(owner)}头像`}
              className="owner-portrait"
              fallbackLabel={ownerLabel(owner)}
            />
            <div>
              <span className="section-label">所选城池</span>
              <h2>{selectedCity.name}</h2>
              <p className="owner-line">
                <span style={{ backgroundColor: owner?.color ?? '#7f7a68' }} />
                {ownerLabel(owner)}
              </p>
            </div>
          </div>
        </section>

        <section className="stats-grid">
          <Metric label="金" value={selectedCity.money} />
          <Metric label="粮" value={selectedCity.food} />
          <Metric label="农" value={`${selectedCity.farming}/${selectedCity.farmingLimit}`} />
          <Metric label="商" value={`${selectedCity.commerce}/${selectedCity.commerceLimit}`} />
          <Metric label="民忠" value={selectedCity.peopleDevotion} />
          <Metric label="防灾" value={selectedCity.avoidCalamity} />
        </section>

        <section>
          <span className="section-label">驻守武将</span>
          <div className="general-list">
            {generals.length ? generals.map((general) => (
              <GeneralRow key={general.id} general={general} />
            )) : <p className="muted">暂无武将驻守</p>}
          </div>
        </section>

        <section>
          <span className="section-label">军政命令</span>
          <div className="command-tabs">
            {(['内政', '外交', '军备', '状况'] as CommandCategory[]).map((item) => (
              <button
                type="button"
                className={category === item ? 'active' : ''}
                key={item}
                onClick={() => setCategory(item)}
              >
                {item}
              </button>
            ))}
          </div>
          {category === '状况' ? (
            <div className="city-status">
              <Metric label="人口" value={selectedCity.population} />
              <Metric label="上限" value={selectedCity.populationLimit} />
              <Metric label="后备" value={selectedCity.garrison} />
              <Metric label="状态" value={selectedCity.state === 'famine' ? '饥荒' : '正常'} />
            </div>
          ) : (
            <div className="command-panel">
              <div className="order-list">
                {activeCommands.map((command) => (
                  <button
                    type="button"
                    className={commandId === command.id ? 'active' : ''}
                    key={command.id}
                    onClick={() => setCommandId(command.id)}
                  >
                    {command.name}
                  </button>
                ))}
              </div>
              <div className="executor-list">
                {playerGenerals.length ? playerGenerals.map((general) => (
                  <button
                    type="button"
                    className={selectedGeneral?.id === general.id ? 'active' : ''}
                    key={general.id}
                    onClick={() => setGeneralId(general.id)}
                  >
                    {general.name}
                    <span>体 {general.stamina}</span>
                  </button>
                )) : <p className="muted">此城暂无可行动武将</p>}
              </div>
              {isBattle ? (
                <div className="battle-targets">
                  <span className="section-label">进攻目标</span>
                  {battleTargets.length ? (
                    <>
                      <div className="target-list">
                        {battleTargets.map((target) => (
                          <button
                            type="button"
                            className={selectedTarget?.id === target.id ? 'active' : ''}
                            key={target.id}
                            onClick={() => setTargetCityId(target.id)}
                          >
                            {target.name}
                            <span>{ownerLabel(rulerByID.get(target.ownerId))}</span>
                          </button>
                        ))}
                      </div>
                      <button
                        type="button"
                        className="primary execute-order"
                        disabled={busy || !playable || !selectedGeneral || !selectedTarget}
                        onClick={() => selectedGeneral && selectedTarget && onBattle(selectedGeneral.id, selectedTarget.id)}
                      >
                        进攻{selectedTarget?.name ?? ''}
                      </button>
                    </>
                  ) : (
                    <p className="muted">无相邻敌方城池可进攻</p>
                  )}
                </div>
              ) : (
                <button
                  type="button"
                  className="primary execute-order"
                  disabled={busy || !playable || !selectedGeneral || !activeCommand}
                  onClick={() => selectedGeneral && activeCommand && onCommand(activeCommand.id, selectedGeneral.id)}
                >
                  执行{activeCommand?.name ?? '命令'}
                </button>
              )}
              {!playable ? <p className="muted">只能向己方城池下达命令</p> : null}
            </div>
          )}
        </section>
      </aside>

      <EventTicker entries={snapshot.log} />
    </>
  );
}

function EventTicker({ entries }: { entries: string[] }) {
  const previousEntries = useRef<string[] | null>(null);
  const [queue, setQueue] = useState<string[]>([]);
  const [current, setCurrent] = useState('');
  const [serial, setSerial] = useState(0);

  useEffect(() => {
    if (previousEntries.current === null) {
      previousEntries.current = entries;
      return;
    }
    const additions = findNewLogEntries(entries, previousEntries.current);
    previousEntries.current = entries;
    if (additions.length) {
      setQueue((items) => [...items, ...additions.reverse()]);
    }
  }, [entries]);

  useEffect(() => {
    if (current || !queue.length) {
      return;
    }
    const [next, ...rest] = queue;
    setCurrent(next);
    setSerial((value) => value + 1);
    setQueue(rest);
  }, [current, queue]);

  useEffect(() => {
    if (!current) {
      return;
    }
    const timer = window.setTimeout(() => setCurrent(''), 3200);
    return () => window.clearTimeout(timer);
  }, [current]);

  return (
    <footer className={`event-log ${current ? 'event-log--active' : ''}`} aria-live="polite">
      <span className="event-log__label">战报</span>
      {current ? (
        <span key={serial} className="event-log__message">{current}</span>
      ) : (
        <span className="event-log__idle">静候军令</span>
      )}
    </footer>
  );
}

function findNewLogEntries(current: string[], previous: string[]): string[] {
  if (!current.length) {
    return [];
  }
  if (!previous.length) {
    return current.slice(0, 1);
  }
  const matchIndex = current.findIndex((_, index) => {
    if (index + previous.length > current.length) {
      return false;
    }
    return previous.every((entry, previousIndex) => current[index + previousIndex] === entry);
  });
  if (matchIndex > 0) {
    return current.slice(0, matchIndex);
  }
  return matchIndex === 0 ? [] : current.slice(0, 1);
}

function Metric({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function GeneralRow({ general }: { general: General }) {
  return (
    <div className="general-row">
      <PortraitImage
        src={portraitForGeneral(general)}
        alt={`${general.name}头像`}
        className="general-avatar"
        fallbackLabel={general.name}
      />
      <div className="general-name">
        <strong>{general.name}</strong>
        <span>{general.armsType} · Lv.{general.level}</span>
      </div>
      <span>武 {general.force}</span>
      <span>智 {general.intellect}</span>
      <span>兵 {general.soldiers}</span>
    </div>
  );
}

function ownerLabel(owner?: Ruler): string {
  if (!owner || owner.id === 'neutral') {
    return '未占领';
  }
  return `${owner.name} · ${owner.character}`;
}
