import type { RulerOption, ScenarioOption } from '../api/types';
import { portraitForRuler } from '../game/portraitRegistry';

type StartMode = 'main' | 'period' | 'ruler' | 'about';

type StartScreenProps = {
  mode: StartMode;
  scenarios: ScenarioOption[];
  selectedScenario: ScenarioOption | null;
  busy: boolean;
  error: string | null;
  onModeChange: (mode: StartMode) => void;
  onScenarioSelected: (scenario: ScenarioOption) => void;
  onRulerSelected: (ruler: RulerOption) => void;
  onContinue: () => void;
};

export function StartScreen({
  mode,
  scenarios,
  selectedScenario,
  busy,
  error,
  onModeChange,
  onScenarioSelected,
  onRulerSelected,
  onContinue,
}: StartScreenProps) {
  return (
    <main className="start-screen">
      <div className="start-vignette" />
      <section className="start-panel">
        <div className="game-title">
          <span>步步高经典复刻</span>
          <h1>三国霸业</h1>
          <p>新君登基，群雄并起</p>
        </div>

        {mode === 'main' ? (
          <div className="main-menu" aria-label="主菜单">
            <button type="button" className="menu-choice primary" onClick={() => onModeChange('period')} disabled={busy}>
              新君登基
            </button>
            <button type="button" className="menu-choice" onClick={onContinue} disabled={busy}>
              重返沙场
            </button>
            <button type="button" className="menu-choice" onClick={() => onModeChange('about')} disabled={busy}>
              制作群组
            </button>
            <button type="button" className="menu-choice" onClick={() => onModeChange('main')} disabled={busy}>
              卸甲归田
            </button>
          </div>
        ) : null}

        {mode === 'period' ? (
          <div className="period-menu">
            <div className="menu-head">
              <button type="button" onClick={() => onModeChange('main')}>返回</button>
              <strong>选择历史时期</strong>
            </div>
            <div className="period-grid">
              {scenarios.map((scenario) => (
                <button
                  type="button"
                  className="period-card"
                  key={scenario.id}
                  onClick={() => onScenarioSelected(scenario)}
                  disabled={busy}
                >
                  <span>第 {scenario.period} 章</span>
                  <strong>{scenario.name}</strong>
                  <em>{scenario.year}年 · {scenario.rulers.length} 势力 · {scenario.cityMax} 城</em>
                </button>
              ))}
            </div>
          </div>
        ) : null}

        {mode === 'ruler' && selectedScenario ? (
          <div className="ruler-menu">
            <div className="menu-head">
              <button type="button" onClick={() => onModeChange('period')}>返回</button>
              <strong>{selectedScenario.name} · 选择君主</strong>
            </div>
            <div className="ruler-grid">
              {selectedScenario.rulers.map((ruler) => (
                <button
                  type="button"
                  className="ruler-card"
                  key={ruler.id}
                  onClick={() => onRulerSelected(ruler)}
                  disabled={busy}
                >
                  <img src={portraitForRuler(ruler)} alt={`${ruler.name}头像`} className="ruler-portrait" decoding="async" />
                  <span className="ruler-swatch" style={{ backgroundColor: ruler.color }} />
                  <strong>{ruler.name}</strong>
                  <em>{ruler.character} · {ruler.cityCount} 城</em>
                </button>
              ))}
            </div>
          </div>
        ) : null}

        {mode === 'about' ? (
          <div className="about-panel">
            <div className="menu-head">
              <button type="button" onClick={() => onModeChange('main')}>返回</button>
              <strong>制作群组</strong>
            </div>
            <p>原作流程：主菜单、时期选择、君主选择、战略命令、策略结束。</p>
            <p>新版以旧档案资源为数据源，保留旧命令结构并用现代浏览器呈现。</p>
          </div>
        ) : null}

        {error ? <p className="start-error">{error}</p> : null}
      </section>
    </main>
  );
}
