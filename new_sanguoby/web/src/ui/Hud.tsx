import { useEffect, useState } from 'react';
import type { City, GameSnapshot, General, Ruler } from '../api/types';
import type { LegacyImageResource, LegacyInventorySummary } from '../game/legacyInventory';
import { portraitForGeneral, portraitForRuler } from '../game/portraitRegistry';

type HudProps = {
  snapshot: GameSnapshot;
  selectedCity: City;
  onMainMenu: () => void;
  onEndStrategy: () => void;
  onCommand: (commandId: string, generalId: string) => void;
  busy: boolean;
  legacySummary: LegacyInventorySummary;
};

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

export function Hud({ snapshot, selectedCity, onMainMenu, onEndStrategy, onCommand, busy, legacySummary }: HudProps) {
  const [category, setCategory] = useState<CommandCategory>('内政');
  const [commandId, setCommandId] = useState(COMMAND_GROUPS['内政'][0].id);
  const [generalId, setGeneralId] = useState('');
  const rulerByID = new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler]));
  const owner = rulerByID.get(selectedCity.ownerId);
  const generals = snapshot.generals.filter((general) => general.cityId === selectedCity.id);
  const playerGenerals = generals.filter((general) => general.ownerId === snapshot.playerId);
  const player = rulerByID.get(snapshot.playerId);
  const ownerPortrait = portraitForRuler(owner);
  const imageResources = legacySummary.imageResources.slice(0, 6);
  const imageArchiveStatus = legacySummary.available
    ? `${legacySummary.presentImageGroups}/${legacySummary.knownImageGroups}`
    : '未连';
  const playable = selectedCity.ownerId === snapshot.playerId;
  const activeCommands = category === '状况' ? [] : COMMAND_GROUPS[category];
  const activeCommand = activeCommands.find((command) => command.id === commandId) ?? activeCommands[0];
  const selectedGeneral = playerGenerals.find((general) => general.id === generalId) ?? playerGenerals[0];

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
            <PortraitImage src={ownerPortrait} label={`${ownerLabel(owner)}头像`} className="owner-portrait" />
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

        <section className="visual-archive">
          <span className="section-label">图像档案</span>
          <div className="visual-stage">
            <img src="/assets/map/sanguo-campaign-map.png" alt="战略地图底图" className="map-preview" />
            <div>
              <strong>{imageArchiveStatus} 组</strong>
              <span>{legacySummary.available ? `${legacySummary.imageResourceItems} 项旧图像` : '现代资产兜底'}</span>
            </div>
          </div>
          {imageResources.length ? (
            <div className="image-resource-grid">
              {imageResources.map((resource) => (
                <LegacyImageChip key={resource.id} resource={resource} />
              ))}
            </div>
          ) : (
            <p className="muted">旧图像资源未连通</p>
          )}
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
          <span className="section-label">旧档案</span>
          <div className="legacy-strip">
            <Metric label="资源" value={legacySummary.available ? legacySummary.totalResources : '未连'} />
            <Metric label="城名" value={legacySummary.cityNames} />
            <Metric label="武将" value={legacySummary.generalScenarios} />
            <Metric label="战场" value={legacySummary.battleMaps} />
          </div>
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
              <button
                type="button"
                className="primary execute-order"
                disabled={busy || !playable || !selectedGeneral || !activeCommand}
                onClick={() => selectedGeneral && activeCommand && onCommand(activeCommand.id, selectedGeneral.id)}
              >
                执行{activeCommand?.name ?? '命令'}
              </button>
              {!playable ? <p className="muted">只能向己方城池下达命令</p> : null}
            </div>
          )}
        </section>
      </aside>

      <footer className="event-log">
        {snapshot.log.map((entry) => <span key={entry}>{entry}</span>)}
      </footer>
    </>
  );
}

function Metric({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function PortraitImage({ src, label, className }: { src: string; label: string; className: string }) {
  const [failed, setFailed] = useState(false);
  if (failed) {
    return (
      <span className={`${className} portrait-fallback`} role="img" aria-label={label}>
        {label.trim().slice(0, 1) || '将'}
      </span>
    );
  }
  return <img src={src} alt={label} className={className} decoding="async" onError={() => setFailed(true)} />;
}

function LegacyImageChip({ resource }: { resource: LegacyImageResource }) {
  return (
    <div className="image-resource-chip">
      <span>{imageGroupName(resource.group)}</span>
      <strong>{resource.label}</strong>
      <em>{resource.itemCount}项</em>
    </div>
  );
}

function GeneralRow({ general }: { general: General }) {
  return (
    <div className="general-row">
      <PortraitImage src={portraitForGeneral(general)} label={`${general.name}头像`} className="general-avatar" />
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

function imageGroupName(group: LegacyImageResource['group']): string {
  switch (group) {
    case 'battle':
      return '战斗';
    case 'campaign':
      return '地图';
    case 'portrait':
      return '头像';
    case 'ui':
      return '界面';
  }
}

function ownerLabel(owner?: Ruler): string {
  if (!owner || owner.id === 'neutral') {
    return '未占领';
  }
  return `${owner.name} · ${owner.character}`;
}
