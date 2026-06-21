import {useMemo, useState, type CSSProperties} from 'react';
import type {BattleRequest, City, GameSnapshot, General, Ruler} from '../api/types';
import {portraitForGeneral, portraitForRuler} from '../game/portraitRegistry';
import {PortraitImage} from './PortraitImage';

type BattlefieldOverlayProps = {
  snapshot: GameSnapshot;
  request: BattleRequest;
  busy: boolean;
  onResolve: (request: BattleRequest, fieldAdvantage: number, remainingFood: number) => void;
  onCancel: () => void;
};

type TacticalUnit = {
  id: string;
  generalId?: string;
  name: string;
  side: 'attacker' | 'defender';
  force: number;
  intellect: number;
  soldiers: number;
  hp: number;
  maxHp: number;
  x: number;
  y: number;
  acted: boolean;
  portrait?: string;
  color: string;
};

type Terrain = 'plain' | 'hill' | 'forest' | 'river' | 'gate';

const FIELD_WIDTH = 9;
const FIELD_HEIGHT = 6;

export function BattlefieldOverlay({snapshot, request, busy, onResolve, onCancel}: BattlefieldOverlayProps) {
  const fromCity = snapshot.cities.find((city) => city.id === request.cityId);
  const targetCity = snapshot.cities.find((city) => city.id === request.targetCityId);
  const rulers = new Map(snapshot.rulers.map((ruler) => [ruler.id, ruler]));
  const initialUnits = useMemo(
    () => buildInitialUnits(snapshot, request, rulers),
    [request, rulers, snapshot],
  );
  const [units, setUnits] = useState<TacticalUnit[]>(initialUnits);
  const [selectedUnitId, setSelectedUnitId] = useState(initialUnits.find((unit) => unit.side === 'attacker')?.id ?? '');
  const [day, setDay] = useState(1);
  const [fieldAdvantage, setFieldAdvantage] = useState(0);
  const [usedFood, setUsedFood] = useState(0);
  const [battleLog, setBattleLog] = useState<string[]>(['两军列阵，等待主将军令。']);
  const selectedUnit = units.find((unit) => unit.id === selectedUnitId) ?? units.find((unit) => unit.side === 'attacker');
  const attackers = units.filter((unit) => unit.side === 'attacker');
  const defenders = units.filter((unit) => unit.side === 'defender');
  const remainingFood = Math.max(0, (request.food ?? 0) - usedFood);
  const attackerRuler = rulers.get(snapshot.playerId);
  const defenderRuler = targetCity ? rulers.get(targetCity.ownerId) : undefined;
  const finalAdvantage = finalFieldAdvantage(fieldAdvantage, defenders.length, remainingFood);

  const selectUnit = (unit: TacticalUnit) => {
    setSelectedUnitId(unit.id);
  };

  const pushLog = (entry: string) => {
    setBattleLog((items) => [entry, ...items].slice(0, 8));
  };

  const markActed = (unitId: string) => {
    setUnits((items) => items.map((unit) => unit.id === unitId ? {...unit, acted: true} : unit));
  };

  const handleAdvance = () => {
    if (!selectedUnit || selectedUnit.side !== 'attacker' || selectedUnit.acted) {
      return;
    }
    const target = nearestEnemy(selectedUnit, defenders);
    const next = nextStepToward(selectedUnit, target);
    setUnits((items) => items.map((unit) => unit.id === selectedUnit.id ? {...unit, ...next, acted: true} : unit));
    setFieldAdvantage((value) => clamp(value + 2, -35, 45));
    pushLog(`${selectedUnit.name} 向敌阵推进，占据有利地形。`);
  };

  const handleAttack = () => {
    if (!selectedUnit || selectedUnit.side !== 'attacker' || selectedUnit.acted) {
      return;
    }
    const target = nearestEnemy(selectedUnit, defenders);
    if (!target) {
      return;
    }
    const damage = Math.max(18, Math.round(selectedUnit.force * 0.55 + selectedUnit.soldiers / 80));
    setUnits((items) => {
      const updated = items
        .map((unit) => unit.id === selectedUnit.id ? {...unit, acted: true} : unit)
        .map((unit) => unit.id === target.id ? {...unit, hp: Math.max(0, unit.hp - damage)} : unit);
      return updated.filter((unit) => unit.side === 'attacker' || unit.hp > 0);
    });
    setFieldAdvantage((value) => clamp(value + (target.hp <= damage ? 8 : 4), -35, 45));
    pushLog(target.hp <= damage
      ? `${selectedUnit.name} 击溃 ${target.name}。`
      : `${selectedUnit.name} 攻击 ${target.name}，造成 ${damage} 点伤势。`);
  };

  const handleScheme = () => {
    if (!selectedUnit || selectedUnit.side !== 'attacker' || selectedUnit.acted) {
      return;
    }
    const target = nearestEnemy(selectedUnit, defenders);
    if (!target) {
      return;
    }
    const damage = Math.max(12, Math.round(selectedUnit.intellect * 0.5));
    setUnits((items) => {
      const updated = items
        .map((unit) => unit.id === selectedUnit.id ? {...unit, acted: true} : unit)
        .map((unit) => unit.side === 'defender' && distance(unit, target) <= 2
          ? {...unit, hp: Math.max(0, unit.hp - damage)}
          : unit);
      return updated.filter((unit) => unit.side === 'attacker' || unit.hp > 0);
    });
    setFieldAdvantage((value) => clamp(value + 5, -35, 45));
    pushLog(`${selectedUnit.name} 施计扰乱守军阵脚。`);
  };

  const handleWait = () => {
    if (!selectedUnit || selectedUnit.side !== 'attacker' || selectedUnit.acted) {
      return;
    }
    markActed(selectedUnit.id);
    setFieldAdvantage((value) => clamp(value + 1, -35, 45));
    pushLog(`${selectedUnit.name} 原地整队，稳住军势。`);
  };

  const handleEndTurn = () => {
    const enemyResult = runEnemyTurn(units);
    setUnits(enemyResult.units.map((unit) => unit.side === 'attacker' ? {...unit, acted: false} : unit));
    setUsedFood((value) => value + Math.max(8, attackers.length * 16));
    setDay((value) => value + 1);
    setFieldAdvantage((value) => clamp(value + enemyResult.advantageDelta - 1, -35, 45));
    pushLog(enemyResult.log);
  };

  const handleResolve = () => {
    onResolve(request, finalAdvantage, remainingFood);
  };

  return (
    <section className="battlefield-overlay" aria-label="战场小地图">
      <div className="battlefield-shell">
        <header className="battlefield-header">
          <div>
            <span>出征战场</span>
            <h2>{fromCity?.name ?? '本城'} → {targetCity?.name ?? '敌城'}</h2>
          </div>
          <button type="button" onClick={onCancel} disabled={busy}>撤回军令</button>
        </header>

        <div className="battlefield-layout">
          <aside className="battlefield-side battlefield-side--left">
            <BattleArmyCard
              label="我军"
              ruler={attackerRuler}
              city={fromCity}
              units={attackers}
              reserveFood={remainingFood}
            />
            <BattleArmyCard
              label="守军"
              ruler={defenderRuler}
              city={targetCity}
              units={defenders}
              reserveFood={targetCity?.food ?? 0}
            />
          </aside>

          <div className="battlefield-stage">
            <div className="battlefield-status">
              <span>第 {day} 日</span>
              <span>军粮 {remainingFood}</span>
              <span>战势 {signed(finalAdvantage)}</span>
            </div>
            <div className="battlefield-grid" aria-label="战场格">
              {Array.from({length: FIELD_HEIGHT}).map((_, y) => (
                Array.from({length: FIELD_WIDTH}).map((__, x) => {
                  const unit = units.find((item) => item.x === x && item.y === y);
                  const terrain = terrainAt(x, y);
                  return (
                    <button
                      type="button"
                      key={`${x}-${y}`}
                      className={`battlefield-cell battlefield-cell--${terrain}${unit ? ` battlefield-cell--${unit.side}` : ''}${unit?.id === selectedUnit?.id ? ' battlefield-cell--selected' : ''}`}
                      onClick={() => unit ? selectUnit(unit) : undefined}
                    >
                      {unit ? <span style={{'--unit-color': unit.color} as CSSProperties}>{unit.name.slice(0, 1)}</span> : null}
                    </button>
                  );
                })
              ))}
            </div>
          </div>

          <aside className="battlefield-side battlefield-side--right">
            <SelectedUnitCard unit={selectedUnit} />
            <div className="battlefield-actions">
              <button type="button" onClick={handleAdvance} disabled={busy || !canAct(selectedUnit)}>推进</button>
              <button type="button" onClick={handleAttack} disabled={busy || !canAct(selectedUnit) || !defenders.length}>攻击</button>
              <button type="button" onClick={handleScheme} disabled={busy || !canAct(selectedUnit) || !defenders.length}>计谋</button>
              <button type="button" onClick={handleWait} disabled={busy || !canAct(selectedUnit)}>待机</button>
              <button type="button" onClick={handleEndTurn} disabled={busy}>回合结束</button>
              <button type="button" className="primary" onClick={handleResolve} disabled={busy || !attackers.length}>
                总攻结算
              </button>
            </div>
            <div className="battlefield-log">
              {battleLog.map((entry, index) => <p key={`${entry}-${index}`}>{entry}</p>)}
            </div>
          </aside>
        </div>
      </div>
    </section>
  );
}

function BattleArmyCard({
  label,
  ruler,
  city,
  units,
  reserveFood,
}: {
  label: string;
  ruler?: Ruler;
  city?: City;
  units: TacticalUnit[];
  reserveFood: number;
}) {
  return (
    <section className="battle-army-card">
      <div>
        <PortraitImage
          src={portraitForRuler(ruler)}
          alt={`${ruler?.name ?? label}头像`}
          className="battle-army-card__portrait"
          fallbackLabel={ruler?.name ?? label}
        />
        <div>
          <span>{label}</span>
          <strong>{ruler?.name ?? '空城'}</strong>
          <p>{city?.name ?? '未知'} · 粮 {reserveFood}</p>
        </div>
      </div>
      <small>{units.length ? units.map((unit) => unit.name).join('、') : '无可战之兵'}</small>
    </section>
  );
}

function SelectedUnitCard({unit}: {unit?: TacticalUnit}) {
  if (!unit) {
    return <div className="selected-unit-card"><p>未选武将</p></div>;
  }
  return (
    <div className="selected-unit-card">
      <PortraitImage
        src={unit.portrait ?? portraitForRuler()}
        alt={`${unit.name}头像`}
        className="selected-unit-card__portrait"
        fallbackLabel={unit.name}
      />
      <div>
        <span>{unit.side === 'attacker' ? '我军武将' : '守军武将'}</span>
        <strong>{unit.name}</strong>
        <p>武 {unit.force} · 智 {unit.intellect} · 兵 {unit.soldiers}</p>
        <meter min="0" max={unit.maxHp} value={unit.hp} />
      </div>
    </div>
  );
}

function buildInitialUnits(snapshot: GameSnapshot, request: BattleRequest, rulers: Map<string, Ruler>): TacticalUnit[] {
  const generalIds = request.generalIds?.length ? request.generalIds : request.generalId ? [request.generalId] : [];
  const attackers = generalIds
    .map((id) => snapshot.generals.find((general) => general.id === id))
    .filter((general): general is General => Boolean(general))
    .slice(0, 10);
  const target = snapshot.cities.find((city) => city.id === request.targetCityId);
  const defenders = snapshot.generals
    .filter((general) => general.cityId === request.targetCityId && !general.captive)
    .sort((a, b) => b.soldiers - a.soldiers)
    .slice(0, 6);
  const attackerColor = rulers.get(snapshot.playerId)?.color ?? '#d93832';
  const defenderColor = rulers.get(target?.ownerId ?? '')?.color ?? '#8f8979';
  const units: TacticalUnit[] = [];

  attackers.forEach((general, index) => {
    units.push(unitFromGeneral(general, 'attacker', 0, clamp(index, 0, FIELD_HEIGHT - 1), attackerColor));
  });
  if (defenders.length) {
    defenders.forEach((general, index) => {
      units.push(unitFromGeneral(general, 'defender', FIELD_WIDTH - 1, clamp(index, 0, FIELD_HEIGHT - 1), defenderColor));
    });
  } else {
    const hp = Math.max(90, Math.round((target?.garrison ?? 500) / 8));
    units.push({
      id: 'defender-garrison',
      name: '城防军',
      side: 'defender',
      force: 45,
      intellect: 35,
      soldiers: target?.garrison ?? 500,
      hp,
      maxHp: hp,
      x: FIELD_WIDTH - 1,
      y: Math.floor(FIELD_HEIGHT / 2),
      acted: false,
      color: defenderColor,
    });
  }
  return units;
}

function unitFromGeneral(general: General, side: TacticalUnit['side'], x: number, y: number, color: string): TacticalUnit {
  const hp = Math.max(80, Math.round(general.soldiers / 8));
  return {
    id: `${side}-${general.id}`,
    generalId: general.id,
    name: general.name,
    side,
    force: general.force,
    intellect: general.intellect,
    soldiers: general.soldiers,
    hp,
    maxHp: hp,
    x,
    y,
    acted: false,
    portrait: portraitForGeneral(general),
    color,
  };
}

function runEnemyTurn(units: TacticalUnit[]): {units: TacticalUnit[]; advantageDelta: number; log: string} {
  const attackers = units.filter((unit) => unit.side === 'attacker');
  const defenders = units.filter((unit) => unit.side === 'defender');
  if (!attackers.length || !defenders.length) {
    return {units, advantageDelta: 0, log: '两军短暂停顿，战场尘土未落。'};
  }
  const defender = defenders[0];
  const target = nearestEnemy(defender, attackers);
  if (!target) {
    return {units, advantageDelta: 0, log: '守军谨慎观望。'};
  }
  if (distance(defender, target) <= 2) {
    const damage = Math.max(10, Math.round(defender.force * 0.4 + defender.soldiers / 120));
    return {
      units: units.map((unit) => unit.id === target.id ? {...unit, hp: Math.max(1, unit.hp - damage)} : unit),
      advantageDelta: -3,
      log: `${defender.name} 反击 ${target.name}，我军阵脚受挫。`,
    };
  }
  return {
    units: units.map((unit) => unit.id === defender.id ? {...unit, x: Math.max(0, unit.x - 1)} : unit),
    advantageDelta: -1,
    log: `${defender.name} 率守军压上。`,
  };
}

function nearestEnemy(unit: TacticalUnit, enemies: TacticalUnit[]): TacticalUnit | undefined {
  return enemies.slice().sort((a, b) => distance(unit, a) - distance(unit, b))[0];
}

function nextStepToward(unit: TacticalUnit, target?: TacticalUnit): Pick<TacticalUnit, 'x' | 'y'> {
  if (!target) {
    return {x: clamp(unit.x + 1, 0, FIELD_WIDTH - 1), y: unit.y};
  }
  return {
    x: clamp(unit.x + Math.sign(target.x - unit.x), 0, FIELD_WIDTH - 1),
    y: clamp(unit.y + Math.sign(target.y - unit.y), 0, FIELD_HEIGHT - 1),
  };
}

function terrainAt(x: number, y: number): Terrain {
  if (x === FIELD_WIDTH - 1 && y >= 2 && y <= 3) {
    return 'gate';
  }
  if ((x === 3 || x === 4) && y >= 1 && y <= 4) {
    return 'river';
  }
  if ((x + y) % 7 === 0) {
    return 'hill';
  }
  if ((x * 2 + y) % 6 === 0) {
    return 'forest';
  }
  return 'plain';
}

function canAct(unit?: TacticalUnit): boolean {
  return Boolean(unit && unit.side === 'attacker' && !unit.acted && unit.hp > 0);
}

function distance(a: TacticalUnit, b: TacticalUnit): number {
  return Math.abs(a.x - b.x) + Math.abs(a.y - b.y);
}

function finalFieldAdvantage(fieldAdvantage: number, defenderCount: number, remainingFood: number): number {
  const defeatedBonus = defenderCount === 0 ? 18 : 0;
  const supplyPenalty = remainingFood <= 0 ? -15 : 0;
  return clamp(fieldAdvantage + defeatedBonus + supplyPenalty, -35, 45);
}

function signed(value: number): string {
  return value > 0 ? `+${value}` : `${value}`;
}

function clamp(value: number, minValue: number, maxValue: number): number {
  return Math.max(minValue, Math.min(maxValue, value));
}
