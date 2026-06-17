import type { City, GameSnapshot, General, Ruler } from '../api/types';
import type { LegacyInventorySummary } from '../game/legacyInventory';
import { portraitForGeneral, portraitForRuler } from '../game/portraitRegistry';

type HudProps = {
  snapshot: GameSnapshot;
  selectedCity: City;
  onNewGame: () => void;
  onAdvanceMonth: () => void;
  busy: boolean;
  legacySummary: LegacyInventorySummary;
};

export function Hud({ snapshot, selectedCity, onNewGame, onAdvanceMonth, busy, legacySummary }: HudProps) {
  const rulerByID = new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler]));
  const owner = rulerByID.get(selectedCity.ownerId);
  const generals = snapshot.generals.filter((general) => general.cityId === selectedCity.id);
  const player = rulerByID.get(snapshot.playerId);
  const ownerPortrait = portraitForRuler(owner);

  return (
    <>
      <header className="topbar">
        <div>
          <h1>三国霸业</h1>
          <p>{snapshot.date.year}年 {snapshot.date.month}月 · {player?.name ?? '未定'} 执政</p>
        </div>
        <div className="topbar-actions">
          <button type="button" onClick={onNewGame} disabled={busy}>新君登基</button>
          <button type="button" className="primary" onClick={onAdvanceMonth} disabled={busy}>推进一月</button>
        </div>
      </header>

      <aside className="status-rail">
        <section>
          <div className="city-hero">
            <img src={ownerPortrait} alt="" className="owner-portrait" />
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
          <div className="command-grid">
            <button type="button">内政</button>
            <button type="button">外交</button>
            <button type="button">军备</button>
            <button type="button">情报</button>
          </div>
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

function GeneralRow({ general }: { general: General }) {
  return (
    <div className="general-row">
      <img src={portraitForGeneral(general)} alt="" className="general-avatar" />
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
