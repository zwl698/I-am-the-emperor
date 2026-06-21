import {useEffect, useRef, useState} from 'react';
import type {BattleRequest, City, GameSnapshot, General, Ruler} from '../api/types';
import {portraitForGeneral, portraitForRuler} from '../game/portraitRegistry';
import {MusicToggle} from './MusicToggle';
import { PortraitImage } from './PortraitImage';

type HudProps = {
  snapshot: GameSnapshot;
  selectedCity: City;
  onMainMenu: () => void;
  onEndStrategy: () => void;
  onCommand: (commandId: string, generalId: string, targetCityId?: string, targetGeneralId?: string) => void;
  onBattle: (request: BattleRequest) => void;
  busy: boolean;
};

// adjacentCities returns enemy/neutral cities reachable from cityId in one hop.
function adjacentEnemyCities(snapshot: GameSnapshot, cityId: string): City[] {
  return adjacentCities(snapshot, cityId).filter((city) => city.ownerId !== snapshot.playerId);
}

function adjacentFriendlyCities(snapshot: GameSnapshot, cityId: string): City[] {
  return adjacentCities(snapshot, cityId).filter((city) => city.ownerId === snapshot.playerId);
}

function adjacentCities(snapshot: GameSnapshot, cityId: string): City[] {
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
    if (other) {
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
    { id: 'kill', name: '处斩' },
    { id: 'banish', name: '流放' },
    { id: 'largess', name: '赏赐' },
    { id: 'confiscate', name: '没收' },
    { id: 'exchange', name: '交易' },
    { id: 'treat', name: '宴请' },
    { id: 'transportation', name: '输送' },
    { id: 'move', name: '移动' },
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
type ActiveCommand = (typeof COMMAND_GROUPS)[keyof typeof COMMAND_GROUPS][number];

type CommandCost = {
  stamina: number;
  money?: number;
  food?: number;
};

const COMMAND_COSTS: Record<string, CommandCost> = {
  assart: { stamina: 4, money: 50 },
  commerce: { stamina: 4, money: 50 },
  search: { stamina: 4 },
  govern: { stamina: 4, money: 50 },
  inspect: { stamina: 4, money: 50 },
  surrender: { stamina: 0, money: 100 },
  kill: { stamina: 4 },
  banish: { stamina: 4 },
  largess: { stamina: 4, money: 100 },
  confiscate: { stamina: 4 },
  exchange: { stamina: 4 },
  treat: { stamina: 0, money: 100 },
  transportation: { stamina: 4 },
  move: { stamina: 4 },
  alienate: { stamina: 4, money: 50 },
  canvass: { stamina: 4, money: 50 },
  counterespionage: { stamina: 4, money: 50 },
  realienate: { stamina: 4, money: 50 },
  induce: { stamina: 4, money: 50 },
  reconnoitre: { stamina: 4, money: 20 },
  conscription: { stamina: 4, money: 1 },
  distribute: { stamina: 4 },
  depredate: { stamina: 4 },
  battle: { stamina: 4 },
};

const COMMAND_EFFECTS: Record<string, string> = {
  assart: '提升本城农业，受执行武将智力与等级影响。',
  commerce: '提升本城商业，受执行武将智力与等级影响。',
  search: '在本城搜寻财货，高智武将更容易有所获。',
  govern: '提升民忠与防灾，稳住城池根基。',
  inspect: '出巡安民，按武力少量提升民忠。',
  surrender: '整顿归顺事务，消耗金钱换取全城安定。',
  kill: '处置俘虏，民忠下降、防灾略升。',
  banish: '放逐俘虏或本城武将，民忠小幅上升。',
  largess: '赏赐本城武将，显著提升忠诚。',
  confiscate: '没收武将财货，换取金钱但损忠诚与民心。',
  exchange: '调换本城金粮储备，缓解资源失衡。',
  treat: '宴请武将，恢复体力并略提忠诚。',
  transportation: '向相邻己方城池输送金粮。',
  move: '让执行武将移动到相邻己方城池。',
  alienate: '离间周边敌将，降低其忠诚。',
  canvass: '招揽周边敌将，智力越高机会越好。',
  counterespionage: '策反敌探，提升本将忠诚与本城防灾。',
  realienate: '反间布防，提升城中诸将忠诚。',
  induce: '劝降相邻敌城，智力与目标状况会影响成败。',
  reconnoitre: '探明相邻敌城归属，战前判断路线。',
  conscription: '募集兵力，会消耗人口与少量金钱。',
  distribute: '把城内后备兵分配给执行武将。',
  depredate: '掠夺粮草，但会明显损害民忠。',
  battle: '选择相邻敌城发起攻城战。',
};

export function Hud({ snapshot, selectedCity, onMainMenu, onEndStrategy, onCommand, onBattle, busy }: HudProps) {
  const [category, setCategory] = useState<CommandCategory>('内政');
  const [commandId, setCommandId] = useState(COMMAND_GROUPS['内政'][0].id);
  const [generalId, setGeneralId] = useState('');
  const [targetCityId, setTargetCityId] = useState('');
  const [targetGeneralId, setTargetGeneralId] = useState('');
  const [battleGeneralIds, setBattleGeneralIds] = useState<string[]>([]);
  const [battleMoney, setBattleMoney] = useState(0);
  const [battleFood, setBattleFood] = useState(0);
  const [panelOpen, setPanelOpen] = useState(false);
  const rulerByID = new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler]));
  const owner = rulerByID.get(selectedCity.ownerId);
  const generals = snapshot.generals.filter((general) => general.cityId === selectedCity.id);
  const playerGenerals = generals.filter((general) => general.ownerId === snapshot.playerId && !general.captive);
  const captives = generals.filter((general) => general.ownerId === snapshot.playerId && general.captive);
  const player = rulerByID.get(snapshot.playerId);
  const ownerPortrait = portraitForRuler(owner);
  const playable = selectedCity.ownerId === snapshot.playerId;
  const activeCommands = category === '状况' ? [] : COMMAND_GROUPS[category];
  const activeCommand = activeCommands.find((command) => command.id === commandId) ?? activeCommands[0];
  const isBattle = activeCommand?.id === 'battle';
  const selectedBattleGenerals = playerGenerals.filter((general) => battleGeneralIds.includes(general.id));
  const selectedGeneral = isBattle
    ? selectedBattleGenerals[0] ?? playerGenerals[0]
    : playerGenerals.find((general) => general.id === generalId) ?? playerGenerals[0];
  const needsFriendlyTarget = activeCommand?.id === 'move' || activeCommand?.id === 'transportation';
  const needsCaptiveTarget = activeCommand?.id === 'kill';
  const needsGeneralTarget = Boolean(activeCommand && ['kill', 'banish', 'largess', 'confiscate', 'treat'].includes(activeCommand.id));
  const battleTargets = isBattle ? adjacentEnemyCities(snapshot, selectedCity.id) : [];
  const friendlyTargets = needsFriendlyTarget ? adjacentFriendlyCities(snapshot, selectedCity.id) : [];
  const commandTargets = isBattle ? battleTargets : needsFriendlyTarget ? friendlyTargets : [];
  const selectedTarget = commandTargets.find((city) => city.id === targetCityId) ?? commandTargets[0];
  const generalTargets = activeCommand
    ? targetGeneralsForCommand(activeCommand.id, playerGenerals, captives, selectedGeneral?.id)
    : [];
  const selectedTargetGeneral = generalTargets.find((general) => general.id === targetGeneralId) ?? generalTargets[0];
  const commandIssues = commandReadiness({
    activeCommand,
    busy,
    city: selectedCity,
    needsCityTarget: isBattle || needsFriendlyTarget,
    needsGeneralTarget,
    playable,
    selectedGeneral,
    selectedTarget,
    selectedTargetGeneral,
  });
  const battleIssues = isBattle
    ? battleReadiness({
      busy,
      city: selectedCity,
      generals: selectedBattleGenerals,
      money: battleMoney,
      food: battleFood,
      playable,
      selectedTarget,
    })
    : [];
  const effectiveIssues = isBattle ? battleIssues : commandIssues;
  const commandReady = effectiveIssues.length === 0;

  useEffect(() => {
    const firstGeneral = playerGenerals[0]?.id ?? '';
    setGeneralId(firstGeneral);
    setBattleGeneralIds((current) => {
      const available = new Set(playerGenerals.map((general) => general.id));
      const valid = current.filter((id) => available.has(id)).slice(0, 10);
      return valid.length ? valid : firstGeneral ? [firstGeneral] : [];
    });
  }, [selectedCity.id, snapshot.playerId, snapshot.generals]);

  useEffect(() => {
    setBattleMoney(Math.min(selectedCity.money, 50));
    setBattleFood(Math.min(selectedCity.food, Math.max(80, selectedCity.garrison > 0 ? 200 : 120)));
  }, [selectedCity.id, selectedCity.food, selectedCity.garrison, selectedCity.money]);

  useEffect(() => {
    if (category === '状况') {
      return;
    }
    const firstCommand = COMMAND_GROUPS[category][0]?.id ?? '';
    setCommandId(firstCommand);
  }, [category]);

  useEffect(() => {
    if (commandId === 'battle') {
      setTargetCityId(adjacentEnemyCities(snapshot, selectedCity.id)[0]?.id ?? '');
      return;
    }
    if (commandId === 'move' || commandId === 'transportation') {
      setTargetCityId(adjacentFriendlyCities(snapshot, selectedCity.id)[0]?.id ?? '');
      return;
    }
    setTargetCityId('');
  }, [selectedCity.id, commandId, snapshot]);

  useEffect(() => {
    if (!needsGeneralTarget || !activeCommand) {
      setTargetGeneralId('');
      return;
    }
    const firstTarget = targetGeneralsForCommand(activeCommand.id, playerGenerals, captives, selectedGeneral?.id)[0]?.id ?? '';
    setTargetGeneralId(firstTarget);
  }, [activeCommand, captives, needsGeneralTarget, playerGenerals, selectedCity.id, selectedGeneral?.id]);

  return (
    <>
      <header className="topbar">
        <div>
          <h1>三国霸业2026重置版</h1>
          <p>{snapshot.date.year}年 {snapshot.date.month}月 · {player?.name ?? '未定'} 执政</p>
        </div>
        <div className="topbar-actions">
          <MusicToggle />
          <button
            type="button"
            className={`rail-toggle${panelOpen ? ' rail-toggle--open' : ''}`}
            onClick={() => setPanelOpen((open) => !open)}
            aria-controls="status-rail"
            aria-expanded={panelOpen}
          >
            {panelOpen ? '收起军令' : `军令 · ${selectedCity.name}`}
          </button>
          <button type="button" onClick={onMainMenu} disabled={busy}>主菜单</button>
          <button type="button" className="primary" onClick={onEndStrategy} disabled={busy}>策略结束</button>
        </div>
      </header>

      {panelOpen ? <aside id="status-rail" className="status-rail">
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
                    className={(isBattle ? battleGeneralIds.includes(general.id) : selectedGeneral?.id === general.id) ? 'active' : ''}
                    key={general.id}
                    onClick={() => {
                      if (!isBattle) {
                        setGeneralId(general.id);
                        return;
                      }
                      setBattleGeneralIds((current) => toggleBattleGeneral(current, general.id));
                    }}
                  >
                    {general.name}
                    <span>{isBattle ? `兵 ${general.soldiers}` : `体 ${general.stamina}`}</span>
                  </button>
                )) : <p className="muted">此城暂无可行动武将</p>}
              </div>
              {activeCommand ? (
                <CommandPreview
                  command={activeCommand}
                  issues={effectiveIssues}
                  selectedGeneral={selectedGeneral}
                  selectedTarget={selectedTarget}
                  selectedTargetGeneral={selectedTargetGeneral}
                  battleGeneralCount={isBattle ? selectedBattleGenerals.length : 0}
                  battleMoney={isBattle ? battleMoney : 0}
                  battleFood={isBattle ? battleFood : 0}
                />
              ) : null}
              {isBattle ? (
                <BattleSupplyControls
                  city={selectedCity}
                  money={battleMoney}
                  food={battleFood}
                  onMoneyChange={setBattleMoney}
                  onFoodChange={setBattleFood}
                />
              ) : null}
              {isBattle || needsFriendlyTarget ? (
                <div className="battle-targets">
                  <span className="section-label">{isBattle ? '进攻目标' : '目标城池'}</span>
                  {commandTargets.length ? (
                    <>
                      <div className="target-list">
                        {commandTargets.map((target) => (
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
                        disabled={!commandReady}
                        onClick={() => {
                          if (!selectedGeneral || !selectedTarget || !activeCommand) {
                            return;
                          }
                          if (isBattle) {
                            onBattle({
                              cityId: selectedCity.id,
                              targetCityId: selectedTarget.id,
                              generalIds: selectedBattleGenerals.map((general) => general.id),
                              money: battleMoney,
                              food: battleFood,
                            });
                          } else {
                            onCommand(activeCommand.id, selectedGeneral.id, selectedTarget.id);
                          }
                        }}
                      >
                        {isBattle ? `开赴${selectedTarget?.name ?? ''}战场` : `${activeCommand?.name ?? '执行'}至${selectedTarget?.name ?? ''}`}
                      </button>
                    </>
                  ) : (
                    <p className="muted">{isBattle ? '无相邻敌方城池可进攻' : '无相邻己方城池可达'}</p>
                  )}
                </div>
              ) : needsGeneralTarget ? (
                <div className="battle-targets">
                  <span className="section-label">{needsCaptiveTarget ? '处置俘虏' : '目标武将'}</span>
                  {generalTargets.length ? (
                    <>
                      <div className="target-list target-list--generals">
                        {generalTargets.map((target) => (
                          <button
                            type="button"
                            className={selectedTargetGeneral?.id === target.id ? 'active' : ''}
                            key={target.id}
                            onClick={() => setTargetGeneralId(target.id)}
                          >
                            <strong>{target.name}</strong>
                            <span>{target.captive ? '俘虏' : `${target.armsType} · 忠 ${target.loyalty}`}</span>
                          </button>
                        ))}
                      </div>
                      <button
                        type="button"
                        className="primary execute-order"
                        disabled={!commandReady}
                        onClick={() => {
                          if (!selectedGeneral || !selectedTargetGeneral || !activeCommand) {
                            return;
                          }
                          onCommand(activeCommand.id, selectedGeneral.id, undefined, selectedTargetGeneral.id);
                        }}
                      >
                        {activeCommand?.name ?? '执行'}{selectedTargetGeneral?.name ?? ''}
                      </button>
                    </>
                  ) : (
                    <p className="muted">{needsCaptiveTarget ? '此城暂无可处置俘虏' : '此城暂无可选目标武将'}</p>
                  )}
                </div>
              ) : (
                <button
                  type="button"
                  className="primary execute-order"
                  disabled={!commandReady}
                  onClick={() => selectedGeneral && activeCommand && onCommand(activeCommand.id, selectedGeneral.id)}
                >
                  执行{activeCommand?.name ?? '命令'}
                </button>
              )}
              {!playable ? <p className="muted">只能向己方城池下达命令</p> : null}
            </div>
          )}
        </section>
      </aside> : null}

      <EventTicker entries={snapshot.log} />
    </>
  );
}

function CommandPreview({
  command,
  issues,
  selectedGeneral,
  selectedTarget,
  selectedTargetGeneral,
  battleGeneralCount = 0,
  battleMoney = 0,
  battleFood = 0,
}: {
  command: ActiveCommand;
  issues: string[];
  selectedGeneral?: General;
  selectedTarget?: City;
  selectedTargetGeneral?: General;
  battleGeneralCount?: number;
  battleMoney?: number;
  battleFood?: number;
}) {
  const cost = commandCostFor(command.id);
  const recruits = command.id === 'conscription' && selectedGeneral ? 120 + selectedGeneral.force * 8 : 0;
  return (
    <div className={`command-preview${issues.length ? ' command-preview--blocked' : ''}`}>
      <div className="command-preview__head">
        <span>军令预览</span>
        <strong>{command.name}</strong>
      </div>
      <p>{COMMAND_EFFECTS[command.id] ?? '执行军政命令。'}</p>
      <div className="command-preview__chips">
        <span>体 {cost.stamina}</span>
        {cost.money ? <span>金 {cost.money}</span> : null}
        {cost.food ? <span>粮 {cost.food}</span> : null}
        {command.id === 'battle' ? <span>将 {battleGeneralCount}</span> : null}
        {command.id === 'battle' ? <span>金 {battleMoney}</span> : null}
        {command.id === 'battle' ? <span>粮 {battleFood}</span> : null}
        {recruits ? <span>募 {recruits}</span> : null}
      </div>
      <div className="command-preview__meta">
        <span>将 {selectedGeneral?.name ?? '未选'}</span>
        <span>目标 {commandTargetLabel(command.id, selectedTarget, selectedTargetGeneral)}</span>
      </div>
      <p className={issues.length ? 'command-preview__warning' : 'command-preview__ready'}>
        {issues.length ? issues.join('、') : '军令可下达'}
      </p>
    </div>
  );
}

function BattleSupplyControls({
  city,
  money,
  food,
  onMoneyChange,
  onFoodChange,
}: {
  city: City;
  money: number;
  food: number;
  onMoneyChange: (value: number) => void;
  onFoodChange: (value: number) => void;
}) {
  return (
    <div className="battle-supplies">
      <span className="section-label">出征军资</span>
      <SupplyInput
        label="金"
        value={money}
        max={city.money}
        onChange={onMoneyChange}
      />
      <SupplyInput
        label="粮"
        value={food}
        max={city.food}
        onChange={onFoodChange}
      />
    </div>
  );
}

function SupplyInput({
  label,
  value,
  max,
  onChange,
}: {
  label: string;
  value: number;
  max: number;
  onChange: (value: number) => void;
}) {
  const safeMax = Math.max(0, max);
  return (
    <label className="supply-input">
      <span>{label}</span>
      <input
        type="range"
        min="0"
        max={safeMax}
        value={Math.min(value, safeMax)}
        onChange={(event) => onChange(clampNumber(Number(event.currentTarget.value), 0, safeMax))}
      />
      <input
        type="number"
        min="0"
        max={safeMax}
        value={Math.min(value, safeMax)}
        onChange={(event) => onChange(clampNumber(Number(event.currentTarget.value), 0, safeMax))}
      />
      <em>/ {safeMax}</em>
    </label>
  );
}

function commandReadiness({
  activeCommand,
  busy,
  city,
  needsCityTarget,
  needsGeneralTarget,
  playable,
  selectedGeneral,
  selectedTarget,
  selectedTargetGeneral,
}: {
  activeCommand?: ActiveCommand;
  busy: boolean;
  city: City;
  needsCityTarget: boolean;
  needsGeneralTarget: boolean;
  playable: boolean;
  selectedGeneral?: General;
  selectedTarget?: City;
  selectedTargetGeneral?: General;
}): string[] {
  const issues: string[] = [];
  if (busy) {
    issues.push('军令执行中');
  }
  if (!activeCommand) {
    issues.push('未选命令');
  }
  if (!playable) {
    issues.push('非己方城池');
  }
  if (!selectedGeneral) {
    issues.push('无可行动武将');
  }
  if (!activeCommand || !selectedGeneral) {
    return issues;
  }

  const cost = commandCostFor(activeCommand.id);
  if (selectedGeneral.stamina < cost.stamina) {
    issues.push('体力不足');
  }
  if (city.money < (cost.money ?? 0)) {
    issues.push('金不足');
  }
  if (city.food < (cost.food ?? 0)) {
    issues.push('粮不足');
  }
  if (activeCommand.id === 'conscription') {
    const recruits = 120 + selectedGeneral.force * 8;
    if (city.population < recruits * 2) {
      issues.push('人口不足');
    }
  }
  if (needsCityTarget && !selectedTarget) {
    issues.push('未选目标城');
  }
  if (needsGeneralTarget && !selectedTargetGeneral) {
    issues.push('未选目标武将');
  }
  return issues;
}

function battleReadiness({
  busy,
  city,
  generals,
  money,
  food,
  playable,
  selectedTarget,
}: {
  busy: boolean;
  city: City;
  generals: General[];
  money: number;
  food: number;
  playable: boolean;
  selectedTarget?: City;
}): string[] {
  const issues: string[] = [];
  if (busy) {
    issues.push('军令执行中');
  }
  if (!playable) {
    issues.push('非己方城池');
  }
  if (!generals.length) {
    issues.push('未选出征武将');
  }
  if (generals.some((general) => general.stamina < COMMAND_COSTS.battle.stamina)) {
    issues.push('武将体力不足');
  }
  if (generals.some((general) => general.soldiers <= 0)) {
    issues.push('武将无兵');
  }
  if (!selectedTarget) {
    issues.push('未选目标城');
  }
  if (money < 0 || money > city.money) {
    issues.push('金不足');
  }
  if (food <= 0) {
    issues.push('未配粮草');
  } else if (food > city.food) {
    issues.push('粮不足');
  }
  return issues;
}

function commandCostFor(commandId: string): CommandCost {
  return COMMAND_COSTS[commandId] ?? { stamina: 4 };
}

function commandTargetLabel(commandId: string, selectedTarget?: City, selectedTargetGeneral?: General): string {
  switch (commandId) {
    case 'battle':
      return selectedTarget ? `${selectedTarget.name} · 攻城` : '未选';
    case 'move':
    case 'transportation':
      return selectedTarget?.name ?? '未选';
    case 'kill':
    case 'banish':
    case 'largess':
    case 'confiscate':
    case 'treat':
      return selectedTargetGeneral?.name ?? '未选';
    default:
      return '本城';
  }
}

function EventTicker({ entries }: { entries: string[] }) {
  const previousEntries = useRef<string[] | null>(null);
  const [queue, setQueue] = useState<string[]>([]);
  const [current, setCurrent] = useState('');
  const [serial, setSerial] = useState(0);
  const [reportEntries, setReportEntries] = useState<string[]>([]);
  const [reportOpen, setReportOpen] = useState(false);

  useEffect(() => {
    if (previousEntries.current === null) {
      previousEntries.current = entries;
      return;
    }
    const additions = findNewLogEntries(entries, previousEntries.current);
    previousEntries.current = entries;
    if (additions.length) {
      const chronological = additions.reverse();
      setQueue((items) => [...items, ...chronological]);
      if (chronological.length > 1 || chronological.some((entry) => entry.startsWith('诸侯行动：'))) {
        setReportEntries(chronological);
        setReportOpen(true);
      }
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
    <>
      {reportOpen && reportEntries.length ? (
        <aside className="campaign-report" aria-label="本月军情">
          <div className="campaign-report__head">
            <strong>本月军情</strong>
            <button type="button" onClick={() => setReportOpen(false)}>收起</button>
          </div>
          <ol>
            {reportEntries.map((entry, index) => (
              <li key={`${entry}-${index}`}>{entry}</li>
            ))}
          </ol>
        </aside>
      ) : null}
      <footer className={`event-log ${current ? 'event-log--active' : ''}`} aria-live="polite">
        <span className="event-log__label">战报</span>
        {current ? (
          <span key={serial} className="event-log__message">{current}</span>
        ) : (
          <span className="event-log__idle">静候军令</span>
        )}
      </footer>
    </>
  );
}

function findNewLogEntries(current: string[], previous: string[]): string[] {
  if (!current.length) {
    return [];
  }
  if (!previous.length) {
    return current.slice(0, Math.min(current.length, 24));
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
  if (matchIndex === 0) {
    return [];
  }
  const previousHeadIndex = current.indexOf(previous[0]);
  if (previousHeadIndex > 0) {
    return current.slice(0, previousHeadIndex);
  }
  return current.slice(0, Math.min(current.length, 24));
}

function Metric({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function targetGeneralsForCommand(commandId: string, playerGenerals: General[], captives: General[], executorId?: string): General[] {
  switch (commandId) {
    case 'kill':
      return captives;
    case 'banish':
      return [...captives, ...playerGenerals.filter((general) => general.id !== executorId)];
    case 'largess':
    case 'confiscate':
    case 'treat':
      return playerGenerals;
    default:
      return [];
  }
}

function toggleBattleGeneral(current: string[], generalId: string): string[] {
  if (current.includes(generalId)) {
    return current.filter((id) => id !== generalId);
  }
  return [...current, generalId].slice(0, 10);
}

function clampNumber(value: number, minValue: number, maxValue: number): number {
  if (Number.isNaN(value)) {
    return minValue;
  }
  return Math.max(minValue, Math.min(maxValue, Math.round(value)));
}

function GeneralRow({ general }: { general: General }) {
  return (
    <div className={`general-row${general.captive ? ' general-row--captive' : ''}`}>
      <PortraitImage
        src={portraitForGeneral(general)}
        alt={`${general.name}头像`}
        className="general-avatar"
        fallbackLabel={general.name}
      />
      <div className="general-name">
        <strong>{general.name}</strong>
        <span>{general.captive ? '俘虏' : general.armsType} · Lv.{general.level}</span>
      </div>
      <span>武 {general.force}</span>
      <span>智 {general.intellect}</span>
      <span>{general.captive ? '俘' : `兵 ${general.soldiers}`}</span>
    </div>
  );
}

function ownerLabel(owner?: Ruler): string {
  if (!owner || owner.id === 'neutral') {
    return '未占领';
  }
  return `${owner.name} · ${owner.character}`;
}
