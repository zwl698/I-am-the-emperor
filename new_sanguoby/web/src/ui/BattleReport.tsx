import type {CSSProperties} from 'react';
import type {BattleOutcome, City, GameSnapshot, Ruler} from '../api/types';
import {portraitForGeneral, portraitForRuler} from '../game/portraitRegistry';
import {PortraitImage} from './PortraitImage';

type BattleReportProps = {
  outcome: BattleOutcome;
  snapshot: GameSnapshot;
  onClose: () => void;
};

type MapPoint = {
  left: number;
  top: number;
};

const GRID_COLUMNS = 12;
const GRID_ROWS = 9;

export function BattleReport({outcome, snapshot, onClose}: BattleReportProps) {
  const fromCity = snapshot.cities.find((city) => city.id === outcome.fromCityId);
  const targetCity = snapshot.cities.find((city) => city.id === outcome.targetCityId);
  const attacker = snapshot.generals.find((general) => general.id === outcome.generalId);
  const attackerRuler = rulerById(snapshot, outcome.attackerRulerId);
  const defenderRuler = rulerById(snapshot, outcome.defenderRulerId);
  const fromPoint = cityPoint(fromCity);
  const targetPoint = cityPoint(targetCity);
  const routeStyle = fromPoint && targetPoint ? routeLineStyle(fromPoint, targetPoint) : undefined;
  const resultLabel = outcome.captured ? '攻占城池' : outcome.won ? '攻势得手' : '进攻受挫';
  const attackerPortrait = attacker ? portraitForGeneral(attacker) : portraitForRuler(attackerRuler ?? outcome.attackerRulerId);
  const attackerNames = outcome.generalNames?.length ? outcome.generalNames : [outcome.generalName];
  const defenderNames = outcome.defenderGenerals?.length ? outcome.defenderGenerals : ['城防军'];
  const capturedNames = outcome.capturedGenerals?.length ? outcome.capturedGenerals : [];

  return (
    <aside className={`battle-report ${outcome.won ? 'battle-report--won' : 'battle-report--lost'}`} aria-label="战斗军报">
      <div className="battle-report__topline">
        <span>战斗军报</span>
        <button type="button" onClick={onClose}>收起</button>
      </div>

      <div className="battle-report__hero">
        <PortraitImage
          src={attackerPortrait}
          alt={`${outcome.generalName}头像`}
          className="battle-report__portrait"
          fallbackLabel={outcome.generalName}
          loading="eager"
        />
        <div>
          <span className="battle-report__result">{resultLabel}</span>
          <h2>{outcome.fromCityName} → {outcome.targetCityName}</h2>
          <p>{outcome.message}</p>
        </div>
      </div>

      <div className="battle-report__body">
        <div className="battle-report__map" aria-hidden="true">
          {routeStyle ? <span className="battle-report__route" style={routeStyle} /> : null}
          {fromPoint ? (
            <span
              className="battle-report__node battle-report__node--from"
              style={nodeStyle(fromPoint, attackerRuler)}
            >
              起
            </span>
          ) : null}
          {targetPoint ? (
            <span
              className="battle-report__node battle-report__node--target"
              style={nodeStyle(targetPoint, defenderRuler)}
            >
              战
            </span>
          ) : null}
        </div>

        <div className="battle-report__sides">
          <BattleSide
            label="攻方"
            rulerName={outcome.attackerRulerName}
            ruler={attackerRuler}
            cityName={outcome.fromCityName}
            names={attackerNames}
          />
          <BattleSide
            label="守方"
            rulerName={outcome.defenderRulerName}
            ruler={defenderRuler}
            cityName={outcome.targetCityName}
            names={defenderNames}
          />
        </div>
      </div>

      <div className="battle-report__metrics">
        <ReportMetric label="攻势" value={formatNumber(outcome.attackPower)} />
        <ReportMetric label="守势" value={formatNumber(outcome.defensePower)} />
        <ReportMetric label="攻损" value={formatNumber(outcome.attackerLosses)} />
        <ReportMetric label="守损" value={formatNumber(outcome.defenderLosses)} />
        <ReportMetric label="金" value={formatNumber(outcome.money ?? 0)} />
        <ReportMetric label="粮" value={formatNumber(outcome.food ?? 0)} />
        <ReportMetric label="余粮" value={formatNumber(outcome.remainingFood ?? 0)} />
        <ReportMetric label="战势" value={signed(outcome.fieldAdvantage ?? 0)} />
      </div>

      <div className="battle-report__captives">
        <span>俘虏</span>
        <strong>{capturedNames.length ? capturedNames.join('、') : '无'}</strong>
      </div>
    </aside>
  );
}

function BattleSide({
  label,
  rulerName,
  ruler,
  cityName,
  names,
}: {
  label: string;
  rulerName: string;
  ruler?: Ruler;
  cityName: string;
  names: string[];
}) {
  return (
    <section className="battle-report__side">
      <div>
        <span style={{backgroundColor: ruler?.color ?? '#8f866d'}} />
        <strong>{label} · {rulerName}</strong>
      </div>
      <p>{cityName}</p>
      <small>{names.join('、')}</small>
    </section>
  );
}

function ReportMetric({label, value}: {label: string; value: string}) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function rulerById(snapshot: GameSnapshot, rulerId: string): Ruler | undefined {
  return snapshot.rulers.find((ruler) => ruler.id === rulerId);
}

function cityPoint(city?: City): MapPoint | undefined {
  if (!city) {
    return undefined;
  }
  return {
    left: clamp(8 + (city.x / GRID_COLUMNS) * 84, 5, 95),
    top: clamp(10 + (city.y / GRID_ROWS) * 78, 6, 94),
  };
}

function routeLineStyle(from: MapPoint, target: MapPoint): CSSProperties {
  const dx = target.left - from.left;
  const dy = target.top - from.top;
  return {
    left: `${from.left}%`,
    top: `${from.top}%`,
    width: `${Math.hypot(dx, dy)}%`,
    transform: `rotate(${Math.atan2(dy, dx)}rad)`,
  };
}

function nodeStyle(point: MapPoint, ruler?: Ruler): CSSProperties {
  return {
    backgroundColor: ruler?.color ?? '#8f866d',
    left: `${point.left}%`,
    top: `${point.top}%`,
  };
}

function formatNumber(value: number): string {
  if (value >= 10000) {
    return `${(value / 10000).toFixed(1)}万`;
  }
  return `${value}`;
}

function signed(value: number): string {
  return value > 0 ? `+${value}` : `${value}`;
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}
