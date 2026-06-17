const state = {
  game: null,
  dynasties: [],
  selectedDynasty: "dayin",
  busy: false,
  lastAction: null,
  voice: {
    enabled: false,
  },
  music: {
    ctx: null,
    master: null,
    timers: [],
    enabled: false,
    trackId: null,
    step: 0,
  },
};

const els = {
  startSelected: document.querySelector("#start-selected"),
  continueGame: document.querySelector("#continue-game"),
  musicToggle: document.querySelector("#music-toggle"),
  voiceToggle: document.querySelector("#voice-toggle"),
  dynastyGrid: document.querySelector("#dynasty-grid"),
  board: document.querySelector("#game-board"),
  phase: document.querySelector("#phase-label"),
  age: document.querySelector("#age-label"),
  stats: document.querySelector("#stats-list"),
  portrait: document.querySelector("#current-portrait"),
  sceneArt: document.querySelector("#scene-art"),
  comicStrip: document.querySelector("#comic-strip"),
  eventHand: document.querySelector("#event-hand-panel"),
  strategyMap: document.querySelector("#strategy-map-panel"),
  playdesk: document.querySelector("#playdesk-panel"),
  kicker: document.querySelector("#scene-kicker"),
  title: document.querySelector("#scene-title"),
  body: document.querySelector("#scene-body"),
  choices: document.querySelector("#choice-grid"),
  resolution: document.querySelector("#resolution"),
  currentDynasty: document.querySelector("#current-dynasty"),
  crisis: document.querySelector("#crisis-card"),
  events: document.querySelector("#event-list"),
  commandStatus: document.querySelector("#command-status"),
  objectives: document.querySelector("#objective-list"),
  strategy: document.querySelector("#strategy-list"),
  relations: document.querySelector("#relation-list"),
  foreign: document.querySelector("#foreign-list"),
  plots: document.querySelector("#plot-list"),
  opinion: document.querySelector("#opinion-panel"),
  cases: document.querySelector("#case-list"),
  provinces: document.querySelector("#province-list"),
  wars: document.querySelector("#war-list"),
  harem: document.querySelector("#harem-list"),
  heirs: document.querySelector("#heir-list"),
  offices: document.querySelector("#office-list"),
  factions: document.querySelector("#faction-list"),
  court: document.querySelector("#court-list"),
  history: document.querySelector("#history-list"),
  toast: document.querySelector("#toast"),
};

els.startSelected.addEventListener("click", () => createGame());
els.continueGame.addEventListener("click", () => continueGame());
els.musicToggle.addEventListener("click", () => toggleMusic());
els.voiceToggle.addEventListener("click", () => toggleVoice());

boot();

async function boot() {
  renderEmptyStats();
  await loadDynasties();
  await continueGame({ silent: true });
}

async function loadDynasties() {
  try {
    const res = await fetch("/api/dynasties");
    state.dynasties = await readJSON(res);
    state.selectedDynasty = state.dynasties[0]?.id || "dayin";
    renderDynastyChoices();
  } catch (error) {
    showToast(`朝代载入失败：${error.message}`);
  }
}

function renderDynastyChoices() {
  els.dynastyGrid.innerHTML = state.dynasties
    .map((dynasty, index) => `
      <button class="dynasty-option ${dynasty.id === state.selectedDynasty ? "selected" : ""} ${dynastyPanelClass[dynasty.id] || ""}" type="button" data-dynasty="${dynasty.id}" style="--panel-index:${index}">
        <span class="dynasty-art"></span>
        <span class="dynasty-info">
          <strong>${dynasty.name}</strong>
          <em>${dynasty.era}</em>
          <span>${dynasty.challenge}</span>
          <small>${dynasty.features.join(" · ")}</small>
        </span>
      </button>
    `)
    .join("");
  document.querySelectorAll("[data-dynasty]").forEach((button) => {
    button.addEventListener("click", () => {
      state.selectedDynasty = button.dataset.dynasty;
      renderDynastyChoices();
    });
  });
}

async function createGame() {
  setBusy(true);
  try {
    const seed = Date.now();
    const res = await fetch("/api/games", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ seed, dynastyId: state.selectedDynasty }),
    });
    state.game = normalizeGame(await readJSON(res));
    localStorage.setItem("emperor-game-id", state.game.id);
    els.resolution.hidden = true;
    renderGame();
    showToast(`${currentDynasty().name}开局。史官翻开了第一页。`);
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

async function continueGame(options = {}) {
  const id = localStorage.getItem("emperor-game-id");
  if (!id) {
    if (!options.silent) showToast("没有可继续的本地存档。");
    return;
  }
  setBusy(true);
  try {
    const res = await fetch(`/api/games/${id}`);
    state.game = normalizeGame(await readJSON(res));
    state.selectedDynasty = currentDynasty().id;
    renderDynastyChoices();
    renderGame();
    if (!options.silent) showToast("已读档。");
  } catch (error) {
    localStorage.removeItem("emperor-game-id");
    if (!options.silent) showToast("旧存档已失效，请重新开局。");
  } finally {
    setBusy(false);
  }
}

async function choose(choiceId) {
  if (!state.game || state.busy) return;
  const selectedChoice = state.game.scene?.choices?.find((choice) => choice.id === choiceId);
  setBusy(true);
  try {
    const res = await fetch(`/api/games/${state.game.id}/choices`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ choiceId }),
    });
    const payload = await readJSON(res);
    state.game = normalizeGame(payload.state);
    state.lastAction = {
      type: "choice",
      title: selectedChoice?.text || "朝议选择",
      domain: selectedChoice?.domain || "court",
      summary: payload.resolution?.summary || "",
    };
    renderResolution(payload.resolution);
    renderGame();
    speakResolution(payload.resolution);
    pulseCourt();
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

async function issueOrder(kind, target, label) {
  if (!state.game || state.busy || state.game.phase !== "emperor") return;
  setBusy(true);
  try {
    const res = await fetch(`/api/games/${state.game.id}/orders`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind, target }),
    });
    const payload = await readJSON(res);
    state.game = normalizeGame(payload.state);
    state.lastAction = {
      type: "order",
      kind,
      title: label || orderTitle(kind),
      domain: orderDomain(kind),
      summary: payload.resolution?.summary || "",
    };
    renderResolution(payload.resolution);
    renderGame();
    focusPanelForOrder(kind, target);
    speakResolution(payload.resolution);
    pulseCourt();
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

async function issueAction(kind, target, mode, label) {
  if (!state.game || state.busy || state.game.phase !== "emperor") return;
  setBusy(true);
  try {
    const res = await fetch(`/api/games/${state.game.id}/actions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind, target, mode }),
    });
    const payload = await readJSON(res);
    state.game = normalizeGame(payload.state);
    state.lastAction = {
      type: "action",
      kind,
      mode,
      title: label || orderTitle(mode),
      domain: orderDomain(mode),
      summary: payload.resolution?.summary || "",
    };
    renderResolution(payload.resolution);
    renderGame();
    focusPanelForAction(kind, target, mode);
    speakResolution(payload.resolution);
    pulseCourt();
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

async function resolveCrisis(choiceId) {
  if (!state.game || state.busy) return;
  setBusy(true);
  try {
    const res = await fetch(`/api/games/${state.game.id}/crisis-choice`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ choiceId }),
    });
    const payload = await readJSON(res);
    state.game = normalizeGame(payload.state);
    state.lastAction = {
      type: "crisis",
      title: "圣裁",
      domain: "intrigue",
      summary: payload.resolution?.summary || "",
    };
    renderResolution(payload.resolution);
    renderGame();
    pulseCourt();
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

function renderGame() {
  if (!state.game) return;
  els.board.classList.remove("is-empty");
  // 进入正式游戏后隐藏开局界面（landing + 朝代选择），避免开局图与游戏面板堆叠造成的割裂感
  document.body.classList.add("game-active");
  // 标记当前阶段，供 CSS 区分皇子（剧情叠图）与皇帝（多面板流式）布局
  document.body.classList.remove("phase-prince", "phase-emperor");
  document.body.classList.add(state.game.phase === "emperor" ? "phase-emperor" : "phase-prince");
  renderIdentity();
  renderStats();
  renderScene();
  renderComicStrip();
  renderExternalPanelsIfReady();
  renderDynastyStatus();
  renderCrisis();
  renderCommands();
  renderObjectives();
  renderProvinces();
  renderWars();
  renderFactions();
  renderCourt();
  renderHistory();
  attachOrderButtons();
  attachActionButtons();
  attachFocusButtons();
  syncMusicToScene();
}

function renderIdentity() {
  const game = state.game;
  const dynasty = currentDynasty();
  const portraitClass = game.phase === "emperor" ? "emperor" : "prince";
  els.portrait.className = `portrait-crop ${portraitClass}`;
  els.portrait.style.backgroundImage = `url('${identityPortrait()}')`;
  els.phase.textContent = `${dynasty.name} · ${phaseName[game.phase] || "未知"}`;
  const calendar = game.phase === "emperor" ? `登基${game.reignYear || 1}年 · ${game.season || "春"}` : `${game.age || 1} 岁`;
  const commandText = game.phase === "emperor" ? ` · 御令剩 ${game.command ?? 0}` : "";
  els.age.textContent = `${calendar} · 第 ${game.turn} 回合${commandText}`;

  // Emperor status panel (traits + health + abdication)
  const statusPanel = document.getElementById("emperor-status-panel");
  if (!statusPanel) return;
  if (game.phase !== "emperor") {
    statusPanel.innerHTML = "";
    statusPanel.hidden = true;
    return;
  }
  statusPanel.hidden = false;
  renderEmperorStatus(game, statusPanel);
}

function renderEmptyStats() {
  els.stats.innerHTML = statMeta.map(([, label]) => statRow(label, 0, false)).join("");
}

function renderStats() {
  const stats = state.game.stats;
  els.stats.innerHTML = statMeta
    .filter(([key]) => shouldShowStat(key, stats))
    .map(([key, label]) => {
      const value = stats[key] ?? 0;
      const danger = key === "borderThreat" ? value >= 70 : value <= 25;
      return statRow(label, value, danger, key === "borderThreat");
    })
    .join("");
}

function shouldShowStat(key, stats) {
  if (["treasury", "grain", "populace", "army", "diplomacy", "stability", "borderThreat", "reform"].includes(key)) {
    return state.game.phase === "emperor" || (stats[key] ?? 0) > 0;
  }
  return true;
}

function statRow(label, value, danger, invert = false) {
  const width = Math.max(0, Math.min(100, value));
  const fillClass = danger ? "stat-fill danger" : "stat-fill";
  const shown = invert ? `${value} 越低越好` : value;
  return `
    <div class="stat-row">
      <div class="stat-head"><span>${label}</span><span>${shown}</span></div>
      <div class="stat-track"><div class="${fillClass}" style="width:${width}%"></div></div>
    </div>
  `;
}

function renderScene() {
  const game = state.game;
  if (game.scene?.art) {
    els.sceneArt.style.backgroundImage = `url('${game.scene.art}')`;
  }
  if (game.ending) {
    els.kicker.textContent = "结局";
    els.title.textContent = game.ending.title;
    els.body.textContent = game.ending.summary;
    els.choices.innerHTML = `<article class="ending-card"><p>你的王朝走到了这一页。换一个朝代，或换一种帝王性格，再开一局。</p></article>`;
    return;
  }

  const scene = game.scene;
  if (!scene) {
    els.kicker.textContent = "待诏";
    els.title.textContent = "朝堂暂歇";
    els.body.textContent = "史官还没有拿到下一页奏章。可以重新开局，或继续等待后端返回新场景。";
    els.choices.innerHTML = "";
    return;
  }
  els.kicker.textContent = `${scene.year} · ${scene.mood}`;
  els.title.textContent = scene.title;
  els.body.textContent = scene.body;
  els.choices.innerHTML = (scene.choices || []).map(choiceButton).join("");
  document.querySelectorAll("[data-choice]").forEach((button) => {
    button.addEventListener("click", () => choose(button.dataset.choice));
  });
}

function renderComicStrip() {
  const game = state.game;
  const action = state.lastAction;
  const panels = typeof dynamicComicPanels === "function" ? dynamicComicPanels(game, action) : [];

  els.comicStrip.innerHTML = panels
    .map((panel) => `
      <article class="comic-panel ${panel.className}">
        <span style="background-image:url('${panel.image}')"></span>
        <strong>${panel.title}</strong>
        <small>${panel.caption}</small>
      </article>
    `)
    .join("");
}

function choiceButton(choice) {
  const princeTrait = princeTraitForChoice(choice.id);
  const traitHint = princeTrait ? `<span class="choice-trait-hint">→ 获得「${princeTrait.name}」${princeTrait.desc ? "：" + princeTrait.desc : ""}</span>` : "";
  const domainLabel = choice.domain === "story" ? "成长抉择" : "大议题 · 推进一季";
  return `
    <button class="choice-card domain-${choice.domain}" type="button" data-choice="${choice.id}">
      <span class="choice-icon">${domainIcon[choice.domain] || "策"}</span>
      <span>
        <strong>${choice.text}</strong>
        <small>${choice.detail}</small>
        <em>${formatEffects(choice.effects)}</em>
        ${traitHint}
        <b>${domainLabel}</b>
      </span>
    </button>
  `;
}

function princeTraitForChoice(choiceId) {
  const map = {
    "grab-scroll":      { name: "好学", desc: "新政+25%" },
    "smile-consort":    { name: "亲和", desc: "邦交+15%" },
    "cry-loudly":       { name: "铁腕", desc: "势力+20%" },
    "answer-people":    { name: "仁厚", desc: "民政+20%" },
    "answer-army":      { name: "好大喜功", desc: "军务+15%" },
    "answer-father":    { name: "远见", desc: "新政+15%" },
    "mount-again":      { name: "好大喜功", desc: "军务+15%" },
    "protect-servant":  { name: "仁厚", desc: "民政+20%" },
    "accuse-brother":   { name: "多疑", desc: "暗线+20%" },
    "open-granary":     { name: "仁厚", desc: "民政+20%" },
    "borrow-merchants": { name: "节俭", desc: "国库消耗-20%" },
    "send-army":        { name: "好大喜功", desc: "军务+15%" },
    "secure-edict":     { name: "远见", desc: "新政+15%" },
    "control-guards":   { name: "铁腕", desc: "势力+20%" },
    "appeal-clans":     { name: "亲和", desc: "邦交+15%" },
  };
  return map[choiceId] || null;
}

function renderResolution(resolution) {
  if (!resolution) return;
  els.resolution.hidden = false;
  els.resolution.innerHTML = `<strong>朱批已下：</strong>${resolution.summary}<br><small>${formatEffects(resolution.effects)}</small>`;
}

function renderDynastyStatus() {
  const dynasty = currentDynasty();
  els.currentDynasty.innerHTML = `
    <div class="panel-title">${dynasty.name}</div>
    <p>${dynasty.background}</p>
    <ul>${(dynasty.features || []).map((feature) => `<li>${feature}</li>`).join("")}</ul>
  `;
}

function renderCrisis() {
  const crisis = state.game.crisis;
  els.crisis.innerHTML = `
    <div class="panel-title">${crisis.title}</div>
    <p>${crisis.summary}</p>
    <div class="danger-clock">
      <span style="width:${Math.min(100, crisis.severity)}%"></span>
    </div>
    <small>烈度 ${crisis.severity} · 危机钟 ${crisis.clock}/8</small>
  `;
}

// ──────────────────────────────────────────────
// 皇帝状态面板（特质 + 健康 + 禅位压力）
// ──────────────────────────────────────────────

function renderEmperorStatus(game, panel) {
  const traits = game.emperorTraits || [];
  const cond = game.condition || {};
  const health = game.stats.health ?? 100;
  const maxHealth = cond.maxHealth ?? 100;
  const healthPct = Math.min(100, Math.round((health / maxHealth) * 100));
  const maxHealthPct = Math.min(100, maxHealth);
  const trendClass = { "强健": "trend-strong", "平稳": "trend-stable", "衰退": "trend-decline", "危急": "trend-critical" }[cond.healthTrend] || "";
  const trendIcon = { "强健": "💪", "平稳": "😌", "衰退": "⚠️", "危急": "🚨" }[cond.healthTrend] || "";
  const abdicationRisk = cond.abdicationRisk ?? 0;
  const abdicationClass = abdicationRisk >= 60 ? "abdication-critical" : abdicationRisk >= 30 ? "abdication-warning" : "abdication-safe";

  // Predicted health next year
  const nextYearHealth = Math.max(0, health - (cond.decayRate ?? 0));
  const nextYearMax = maxHealthByAgePrediction(game.age + 1);
  const willDecline = nextYearHealth < health || nextYearMax < maxHealth;

  let html = "";

  // ── Traits section ──
  if (traits.length > 0) {
    html += `
      <div class="emperor-traits-block">
        <div class="panel-title">帝德</div>
        <div class="trait-tags-row">
          ${traits.map((t) => `
            <span class="trait-tag trait-hover" data-trait-id="${t.id || t.name}">
              ${t.name}
              <span class="trait-tooltip">
                <strong>${t.name}</strong>
                <em>${t.description || ""}</em>
                <span class="trait-tooltip-source">来源：${t.source || "成长"}${t.acquiredAge ? ` · ${t.acquiredAge}岁` : ""}</span>
                <span class="trait-tooltip-effects">${traitEffectDescription(t.id || t.name)}</span>
              </span>
            </span>
          `).join("")}
        </div>
      </div>
    `;
  }

  // ── Health section ──
  if (cond.healthTrend) {
    html += `
      <div class="emperor-health-block">
        <div class="panel-title">龙体 <span class="health-trend-badge ${trendClass}">${trendIcon}${cond.healthTrend}</span></div>
        <div class="health-bar-container">
          <div class="health-bar-track">
            <div class="health-bar-max" style="width:${maxHealthPct}%"></div>
            <div class="health-bar-current ${trendClass}" style="width:${healthPct}%"></div>
            ${willDecline ? `<div class="health-bar-predicted" style="width:${Math.min(100, Math.round((Math.min(nextYearHealth, nextYearMax) / 100) * 100))}%"></div>` : ""}
          </div>
          <div class="health-bar-labels">
            <span>${health}/${maxHealth}</span>
            <span>上限${maxHealth}</span>
          </div>
        </div>
        <div class="health-details">
          <small>年衰 ${cond.decayRate ?? 0}${willDecline ? ` → 预计明年 ${Math.min(nextYearHealth, nextYearMax)}` : ""}</small>
          ${cond.lastIllness ? `<small class="illness-note">最近：${cond.lastIllness} (第${cond.illnessTurn}回合) · 累计发病 ${cond.illnessCount ?? 0} 次</small>` : ""}
        </div>
      </div>
    `;
  }

  // ── Abdication pressure section ──
  if (abdicationRisk > 0) {
    html += `
      <div class="abdication-block ${abdicationClass}">
        <div class="panel-title">禅位压力</div>
        <div class="abdication-gauge">
          <svg viewBox="0 0 120 60" class="abdication-svg">
            <path d="M 10 55 A 50 50 0 0 1 110 55" fill="none" stroke="rgba(86,30,13,0.15)" stroke-width="8" stroke-linecap="round" />
            <path d="M 10 55 A 50 50 0 0 1 110 55" fill="none" stroke="${abdicationGaugeColor(abdicationRisk)}" stroke-width="8" stroke-linecap="round" stroke-dasharray="${abdicationRisk * 1.57} 157" class="abdication-gauge-fill" />
            <text x="60" y="48" text-anchor="middle" class="abdication-gauge-text">${abdicationRisk}</text>
          </svg>
        </div>
        <small class="abdication-desc">${abdicationDescription(abdicationRisk, game)}</small>
      </div>
    `;
  }

  panel.innerHTML = html;
}

function traitEffectDescription(traitId) {
  const map = {
    benevolent: "民政+20% · 军务-10% · 每年健康+1",
    suspicious: "暗线+20% · 朝堂-10% · 危机钟+1",
    ambitious: "新政+20% · 国库消耗+15%",
    vainglorious: "军务+15% · 国库-15% · 每年健康-1",
    frugal: "国库消耗-20% · 宫廷-10% · 每年健康+1",
    ruthless: "势力+20% · 民心-15% · 危机钟+1",
    scholarly: "新政+25% · 学识+15%",
    charismatic: "邦交+15% · 魅力+10% · 危机钟-1",
    paranoid: "势力+15% · 每年健康-1",
    visionary: "新政+15% · 朝堂-10%",
    仁厚: "民政+20% · 军务-10% · 每年健康+1",
    多疑: "暗线+20% · 朝堂-10% · 危机钟+1",
    雄才: "新政+20% · 国库消耗+15%",
    好大喜功: "军务+15% · 国库-15% · 每年健康-1",
    节俭: "国库消耗-20% · 宫廷-10% · 每年健康+1",
    铁腕: "势力+20% · 民心-15% · 危机钟+1",
    好学: "新政+25% · 学识+15%",
    亲和: "邦交+15% · 魅力+10% · 危机钟-1",
    偏执: "势力+15% · 每年健康-1",
    远见: "新政+15% · 朝堂-10%",
  };
  return map[traitId] || "未知效果";
}

function maxHealthByAgePrediction(age) {
  if (age < 30) return 100;
  if (age < 40) return 90;
  if (age < 50) return 78;
  if (age < 60) return 64;
  if (age < 70) return 48;
  return 30;
}

function abdicationGaugeColor(risk) {
  if (risk >= 70) return "#d93a2f";
  if (risk >= 50) return "#e87d2d";
  if (risk >= 30) return "#f6a33d";
  return "#4ecdc4";
}

function abdicationDescription(risk, game) {
  const heirs = game.heirs || [];
  const namedHeir = heirs.find((h) => game.succession?.namedHeirId === h.id);
  if (risk >= 70) return `朝臣已在暗议内禅。${namedHeir ? `储君${namedHeir.name}的灯比御书房亮得更晚。` : "但尚无储君，诸王蠢蠢欲动。"}`;
  if (risk >= 50) return "朝中开始有人低声议论禅位，东宫的人比以前更勤了。";
  if (risk >= 30) return "有风声说龙体需要调养，几位老臣开始多看东宫几眼。";
  return "朝局尚稳，暂无禅位议论。";
}

function renderCommands() {
  const game = state.game;
  if (game.phase !== "emperor") {
    els.commandStatus.innerHTML = "皇子阶段先积累性格、名望与盟友。登基后会开启每季度多道御令、地方治理和派系压制。";
    return;
  }
  const command = game.command ?? 0;
  els.commandStatus.innerHTML = `
    <strong>${command} 道御令可用</strong>
    <span>先玩上方御前操作台或右侧系统面板处理具体战局、案卷、任官，再点朝会大议题推进季度。</span>
  `;
}

function renderObjectives() {
  els.objectives.innerHTML = (state.game.objectives || [])
    .map((objective) => {
      const percent = Math.min(100, Math.round((objective.progress / objective.target) * 100));
      return `
        <article class="objective-row ${objective.completed ? "completed" : ""}">
          <div>
            <strong>${objective.title}</strong>
            <small>${objective.description}</small>
          </div>
          <div class="objective-track"><span style="width:${percent}%"></span></div>
          <em>${objective.progress}/${objective.target} · ${objective.reward}</em>
        </article>
      `;
    })
    .join("");
}

function renderProvinces() {
  const canOrder = state.game.phase === "emperor";
  els.provinces.innerHTML = (state.game.provinces || [])
    .map((p) => `
      <article class="mini-world-row">
        <div class="row-head">
          <strong>${p.name}</strong>
          <small>${provinceTemperature(p)}</small>
        </div>
        <span>${p.focus}</span>
        <small>富 ${p.wealth} · 安 ${p.order} · 防 ${p.defense} · 灾 ${p.disaster}</small>
        ${canOrder ? orderButtons(provinceOrders, p.id, p.name) : ""}
      </article>
    `)
    .join("");
}

function renderWars() {
  const wars = state.game.wars || [];
  if (wars.length === 0) {
    els.wars.innerHTML = `<article class="mini-world-row"><strong>边境暂宁</strong><span>暂无外战，但边患仍会随局势变化。</span></article>`;
    return;
  }
  const canOrder = state.game.phase === "emperor";
  els.wars.innerHTML = wars
    .map((war) => {
      const progress = Math.min(100, war.progress ?? 0);
      return `
        <article class="war-row" data-war-id="${escapeAttr(war.id)}">
          <div class="row-head">
            <strong>${war.name}</strong>
            <small>${war.stage}</small>
          </div>
          <span>${war.enemy} · ${war.front} · 历时 ${war.duration} 季</span>
          <div class="war-meter"><b>战果</b><i style="width:${progress}%"></i><em>${progress}/100</em></div>
          <small>敌势 ${war.threat} · 粮道 ${war.supply} · 士气 ${war.morale}</small>
          ${canOrder ? strategyMapJumpButton(war) : ""}
          ${canOrder ? orderButtons(warOrders, war.id, war.name) : ""}
        </article>
      `;
    })
    .join("");
}

function renderFactions() {
  const canOrder = state.game.phase === "emperor";
  els.factions.innerHTML = (state.game.factions || [])
    .map((faction) => `
      <article class="faction-row">
        <span class="portrait-dot" style="background-image:url('${portraitForFaction(faction)}')"></span>
        <span>
          <strong>${faction.name}</strong>
          <small>${faction.leader} · ${faction.agenda}</small>
          <em>权势 ${faction.power} · 忠诚 ${faction.loyalty}</em>
          ${canOrder ? orderButtons(factionOrders, faction.id, faction.name) : ""}
        </span>
      </article>
    `)
    .join("");
}

function renderCourt() {
  els.court.innerHTML = (state.game.court || [])
    .map((minister) => `
      <article class="court-row">
        <span class="portrait-dot" style="background-image:url('${portraitForMinister(minister)}')"></span>
        <span>
          <strong>${minister.name}</strong>
          <small>${minister.role} · ${minister.trait}</small>
          <em>忠 ${minister.loyalty} · 才 ${minister.ability} · 野 ${minister.ambition} · 廉 ${minister.integrity} · 压 ${minister.stress}</em>
        </span>
      </article>
    `)
    .join("");
}

function renderHistory() {
  const history = state.game.history || [];
  if (history.length === 0) {
    els.history.innerHTML = `<li class="history-item">史官蘸墨以待。</li>`;
    return;
  }
  els.history.innerHTML = history
    .slice()
    .reverse()
    .slice(0, 8)
    .map((entry) => `
      <li class="history-item">
        <strong>${entry.age} 岁 · ${phaseName[entry.phase] || entry.phase}</strong>
        ${entry.choice}<br />
        <span>${entry.summary}</span>
      </li>
    `)
    .join("");
}

function formatEffects(effects = {}) {
  const names = Object.fromEntries(statMeta);
  const text = Object.entries(effects)
    .filter(([, value]) => value !== 0)
    .map(([key, value]) => `${names[key] || key}${value > 0 ? "+" : ""}${value}`)
    .join("、");
  return text || "无直接变化";
}

if (typeof window !== "undefined") {
  window.formatEffects = formatEffects;
}

function sceneGallery() {
  const gallery = state.game?.assets?.sceneGallery;
  return Array.isArray(gallery) && gallery.length >= 30 ? gallery : sceneGalleryFallback;
}

function portraitGallery() {
  const gallery = state.game?.assets?.portraitGallery;
  return Array.isArray(gallery) && gallery.length >= 30 ? gallery : portraitGalleryFallback;
}

function sceneAt(index) {
  const gallery = sceneGallery();
  return gallery[((index % gallery.length) + gallery.length) % gallery.length];
}

function portraitAt(index) {
  const gallery = portraitGallery();
  return gallery[((index % gallery.length) + gallery.length) % gallery.length];
}

function identityPortrait() {
  const game = state.game;
  if (!game) return portraitAt(0);
  if (game.phase === "emperor") return portraitAt(portraitIndexByRole.emperor);
  if (game.age <= 6) return portraitAt(portraitIndexByRole.infant);
  return portraitAt(portraitIndexByRole.prince);
}

function portraitForFaction(faction) {
  const map = {
    tutor: "tutor",
    general: "general",
    minister: "minister",
    consort: "consort",
    scholar: "scholar",
    merchant: "merchant",
    border: "general",
    clan: "consort",
  };
  const role = map[faction.portrait] || map[faction.id] || "scholar";
  return portraitAt(portraitIndexByRole[role] ?? portraitIndexByRole.scholar);
}

function portraitForMinister(minister) {
  const map = {
    太傅: "tutor",
    大将军: "general",
    户部尚书: "minister",
    长公主: "consort",
  };
  const role = map[minister.role] || minister.portrait || "scholar";
  return portraitAt(portraitIndexByRole[role] ?? portraitIndexByRole.scholar);
}

function domainSceneArt(domain) {
  const map = {
    story: 1,
    domestic: 6,
    economy: 7,
    military: 14,
    diplomacy: 28,
    court: 5,
    reform: 10,
    intrigue: 11,
  };
  return sceneAt(map[domain] ?? 5);
}

function crisisSceneArt() {
  const crisis = state.game?.crisis;
  if (!crisis) return sceneAt(5);
  if (crisis.clock >= 6 || crisis.severity >= 80) return sceneAt(22);
  if ((state.game?.stats?.borderThreat ?? 0) >= 70) return sceneAt(14);
  if ((state.game?.stats?.stability ?? 0) >= 82) return sceneAt(29);
  return sceneAt(18);
}

function orderButtons(orders, target, targetName) {
  const disabled = (state.game.command ?? 0) <= 0 ? `disabled data-state-disabled="true"` : "";
  return `
    <div class="order-buttons" aria-label="${targetName}御令">
      ${orders
        .map(
          (order) =>
            `<button type="button" ${disabled} data-order-kind="${order.kind}" data-order-target="${target}" data-order-label="${order.title}" title="${order.title}">${order.label}</button>`,
        )
        .join("")}
    </div>
  `;
}

function strategyMapJumpButton(war) {
  const target = strategicTargetForWar(war.id);
  return `
    <div class="order-buttons flow-buttons">
      <button class="module-jump" type="button" data-focus-panel="strategy-map-panel" data-focus-target="${escapeAttr(target)}" data-focus-label="战略地图 · ${escapeAttr(war.name)}" title="进入战略地图查看${escapeAttr(war.name)}">入图</button>
    </div>
  `;
}

function attachOrderButtons() {
  document.querySelectorAll("[data-order-kind]").forEach((button) => {
    button.addEventListener("click", () => {
      issueOrder(button.dataset.orderKind, button.dataset.orderTarget, button.dataset.orderLabel);
    });
  });
}

function attachFocusButtons() {
  document.querySelectorAll("[data-focus-panel]").forEach((button) => {
    button.addEventListener("click", () => {
      if (button.dataset.actionKind || button.dataset.orderKind) return;
      focusPanel(button.dataset.focusPanel, button.dataset.focusTarget, button.dataset.focusLabel);
    });
  });
}

function attachActionButtons() {
  document.querySelectorAll("[data-action-kind]").forEach((button) => {
    button.addEventListener("click", () => {
      issueAction(button.dataset.actionKind, button.dataset.actionTarget, button.dataset.actionMode, button.dataset.actionLabel);
    });
  });
}

function focusPanelForOrder(kind, target) {
  if (["mobilize", "campaign", "fortify", "truce"].includes(kind)) {
    return focusPanel("strategy-map-panel", strategicTargetForWar(target), "战略地图");
  }
  if (["embassy", "treaty"].includes(kind)) return focusPanel("foreign-list", target, "外邦诸国");
  if (["investigate_plot", "suppress_plot"].includes(kind)) return focusPanel("plot-list", target, "密谋暗线");
  if (["open_trial", "clemency", "censor_rumor", "proclaim_verdict"].includes(kind)) return focusPanel("case-list", target, "刑狱案卷");
  if (["appoint", "dismiss"].includes(kind)) return focusPanel("office-list", firstTargetPart(target), "官职任免");
  if (["name_heir", "educate_heir"].includes(kind)) return focusPanel("heir-list", firstTargetPart(target), "储位继承");
  if (["favor_consort", "marriage_alliance"].includes(kind)) return focusPanel("harem-list", target, "后宫势力");
  if (["fund_project", "enact_policy"].includes(kind)) return focusPanel("strategy-list", target, "长期国策");
  if (["relief", "garrison", "tax", "canal", "trade"].includes(kind)) return focusPanel("province-list", target, "四方省势");
  if (["appease", "purge"].includes(kind) || (kind === "inspect" && hasCollectionItem("factions", target))) {
    return focusPanel("faction-list", target, "朝堂派系");
  }
  if (kind === "inspect") return focusPanel("province-list", target, "四方省势");
  return false;
}

function focusPanelForAction(kind, target, mode) {
  if (["city_develop", "army_command", "siege_command", "governor_assign", "war_tactic"].includes(kind)) {
    return focusPanel("strategy-map-panel", focusTargetForStrategicAction(kind, target, mode), "战略地图");
  }
  if (kind === "map_allocation") {
    const cityTarget = hasStrategyCity(target) ? target : "";
    return cityTarget ? focusPanel("strategy-map-panel", cityTarget, "战略地图") : focusPanel("province-list", target, "四方省势");
  }
  if (kind === "trial_move") return focusPanel("case-list", target, "刑狱案卷");
  if (kind === "office_assign") return focusPanel("office-list", firstTargetPart(target), "官职任免");
  if (kind === "envoy_mission") return focusPanel("foreign-list", target, "外邦诸国");
  if (kind === "heir_lesson") return focusPanel("heir-list", target, "储位继承");
  if (kind === "talent_recruit") return focusPanel("talent-list", target, "天下人才谱");
  return false;
}

function focusTargetForStrategicAction(kind, target, mode) {
  if (kind === "war_tactic") return strategicTargetForWar(target);
  const [primary, secondary] = splitTarget(target);
  if (kind === "city_develop" || kind === "governor_assign") return primary;
  if (kind === "siege_command") return secondary || primary;
  if (kind === "army_command" && ["march", "assault", "besiege"].includes(mode)) return secondary || primary;
  return primary;
}

function focusPanel(panelID, targetID = "", label = "") {
  const panel = document.getElementById(panelID);
  if (!panel) return false;
  clearModuleFocus();
  panel.classList.add("module-focus");
  const target = findFocusTarget(panel, targetID);
  if (target) target.classList.add("module-focus-target");
  const scrollTarget = target || panel;
  if (typeof scrollTarget.scrollIntoView === "function") {
    scrollTarget.scrollIntoView({ behavior: "smooth", block: "center", inline: "center" });
  }
  window.clearTimeout(state.focusTimer);
  state.focusTimer = window.setTimeout(clearModuleFocus, 1800);
  showToast(`已切到${label || focusPanelLabel(panelID)}。`);
  return true;
}

function clearModuleFocus() {
  document.querySelectorAll(".module-focus, .module-focus-target").forEach((node) => {
    node.classList.remove("module-focus", "module-focus-target");
  });
}

function findFocusTarget(panel, targetID = "") {
  if (!targetID) return null;
  const selectors = focusTargetSelectors(targetID).join(",");
  if (!selectors) return null;
  return panel.querySelector(selectors) || document.querySelector(selectors);
}

function focusTargetSelectors(targetID) {
  const value = selectorValue(targetID);
  return [
    `[data-city-id="${value}"]`,
    `[data-army-id="${value}"]`,
    `[data-strategy-target="${value}"]`,
    `[data-war-id="${value}"]`,
    `[data-province-id="${value}"]`,
    `[data-case-id="${value}"]`,
    `[data-office-id="${value}"]`,
    `[data-foreign-id="${value}"]`,
    `[data-heir-id="${value}"]`,
  ];
}

function focusPanelLabel(panelID) {
  const labels = {
    "strategy-map-panel": "战略地图",
    "playdesk-panel": "御前操作台",
    "event-hand-panel": "事件手牌",
    "province-list": "四方省势",
    "war-list": "军机外战",
    "foreign-list": "外邦诸国",
    "plot-list": "密谋暗线",
    "case-list": "刑狱案卷",
    "office-list": "官职任免",
    "heir-list": "储位继承",
    "harem-list": "后宫势力",
    "strategy-list": "长期国策",
    "faction-list": "朝堂派系",
    "talent-list": "天下人才谱",
  };
  return labels[panelID] || "相关模块";
}

function strategicTargetForWar(warID = "") {
  const map = {
    "snow-ridge": "snow-ridge",
    "western-oath": "jade-pass",
    "river-bandits": "river-east",
    "jade-pass": "jade-pass",
    north: "snow-ridge",
  };
  return map[warID] || warID || "snow-ridge";
}

function hasStrategyCity(cityID) {
  return hasCollectionItem("strategy.cities", cityID);
}

function hasCollectionItem(collectionPath, id) {
  const list =
    collectionPath === "strategy.cities"
      ? state.game?.strategy?.cities
      : state.game?.[collectionPath];
  return Array.isArray(list) && list.some((item) => item.id === id);
}

function firstTargetPart(target = "") {
  return splitTarget(target)[0];
}

function splitTarget(target = "") {
  return String(target || "").split(":");
}

function escapeAttr(value = "") {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function selectorValue(value = "") {
  return String(value ?? "").replaceAll("\\", "\\\\").replaceAll('"', '\\"');
}

function renderExternalPanelsIfReady() {
  if (typeof window.renderExternalPanels !== "function") return;
  window.renderExternalPanels(state.game, els, { portraitAt, portraitIndexByRole });
}

function provinceTemperature(p) {
  if (p.disaster >= 60) return "灾情急";
  if (p.order <= 35) return "民变险";
  if (p.defense <= 35) return "防务弱";
  if (p.wealth >= 75) return "富庶";
  return "可治";
}

function orderTitle(kind) {
  return [
    ...provinceOrders,
    ...factionOrders,
    ...warOrders,
    ...consortOrders,
    ...heirOrders,
    ...heirTrainingOrders,
    ...officeOrders,
    ...projectOrders,
    ...policyOrders,
    ...foreignOrders,
    ...plotOrders,
    ...justiceOrders,
  ].find((order) => order.kind === kind)?.title || "御令";
}

function orderDomain(kind) {
  const map = {
    relief: "domestic",
    garrison: "military",
    tax: "economy",
    inspect: "intrigue",
    appease: "court",
    purge: "intrigue",
    canal: "reform",
    trade: "diplomacy",
    mobilize: "military",
    campaign: "military",
    fortify: "military",
    truce: "diplomacy",
    appoint: "court",
    dismiss: "intrigue",
    name_heir: "court",
    favor_consort: "court",
    marriage_alliance: "diplomacy",
    fund_project: "reform",
    enact_policy: "court",
    embassy: "diplomacy",
    treaty: "diplomacy",
    investigate_plot: "intrigue",
    suppress_plot: "intrigue",
    educate_heir: "court",
    open_trial: "intrigue",
    clemency: "court",
    censor_rumor: "intrigue",
    proclaim_verdict: "court",
  };
  return map[kind] || "court";
}

function normalizeGame(game) {
  if (!game) return game;
  const dynastyBase = findDynasty(game.dynasty?.id || state.selectedDynasty) || defaultDynasty();
  const dynasty = { ...dynastyBase, ...(game.dynasty || {}) };
  return {
    ...game,
    dynasty,
    assets: {
      hero: "/assets/palace-hero.png",
      dynasties: "/assets/dynasty-scroll.png",
      characters: "/assets/characters.png",
      sceneGallery: sceneGalleryFallback,
      portraitGallery: portraitGalleryFallback,
      ...(game.assets || {}),
    },
    command: game.command ?? 0, reignYear: game.phase === "emperor" ? game.reignYear || 1 : game.reignYear || 0, season: game.season || "春",
    crisis: game.crisis || {
      title: "朝局未明",
      severity: 40,
      clock: 2,
      summary: "旧存档缺少新版危机数据，已按默认朝局继续。",
    },
    objectives: game.objectives || [],
    factions: game.factions || [],
    provinces: game.provinces || [],
    court: game.court || [], talentPool: game.talentPool || [],
    harem: game.harem || [],
    heirs: game.heirs || [],
    succession: game.succession || {
      namedHeirId: "",
      stability: 0,
      dispute: 0,
      maternalClanPower: 0,
      lastSuccessionMove: "旧存档缺少储位数据。",
    },
    offices: game.offices || [],
    projects: game.projects || [],
    policies: game.policies || [],
    relations: game.relations || [],
    foreignStates: game.foreignStates || [],
    plots: game.plots || [],
    legalCases: game.legalCases || [],
    publicOpinion: game.publicOpinion || {
      popular: 50,
      elite: 45,
      rumor: 40,
      fear: 25,
      justice: 45,
      lastEdict: "旧存档缺少舆论数据，法司按默认风声续局。",
    },
    emperorTraits: game.emperorTraits || [],
    condition: game.condition || {
      healthTrend: "强健",
      maxHealth: 100,
      decayRate: 0,
      lastIllness: "",
      illnessTurn: 0,
      illnessCount: 0,
      abdicationRisk: 0,
    },
    wars: game.wars || [],
    strategy: game.strategy || { cities: [], roads: [], factions: [], armies: [], logs: [], battles: [] },
    recentEvents: game.recentEvents || [],
    eventLog: game.eventLog || [],
    eventHand: game.eventHand || [],
    history: game.history || [],
  };
}

function currentDynasty() {
  return state.game?.dynasty || findDynasty(state.selectedDynasty) || defaultDynasty();
}

function findDynasty(id) {
  return state.dynasties.find((dynasty) => dynasty.id === id);
}

function defaultDynasty() {
  return state.dynasties[0] || {
    id: "dayin",
    name: "大胤",
    era: "开国元年",
    background: "旧都新定，功臣拥兵，百废待兴。",
    features: ["开国功臣强势", "国库充实但朝制未稳"],
    challenge: "用刀剑打下天下后，如何让刀剑回鞘。",
    asset: "/assets/dynasty-scroll.png",
    palette: "ember",
  };
}

async function readJSON(res) {
  const payload = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(payload.error || "请求失败");
  return payload;
}

function setBusy(busy) {
  state.busy = busy;
  document.querySelectorAll("button").forEach((button) => {
    button.disabled = busy || button.dataset.stateDisabled === "true";
  });
}

function pulseCourt() {
  document.body.classList.remove("court-pulse");
  window.requestAnimationFrame(() => {
    document.body.classList.add("court-pulse");
    window.setTimeout(() => document.body.classList.remove("court-pulse"), 760);
  });
}

async function toggleMusic() {
  if (state.music.enabled) {
    stopMusic();
    return;
  }
  await startMusic();
}

async function startMusic() {
  const AudioContext = window.AudioContext || window.webkitAudioContext;
  if (!AudioContext) {
    showToast("当前浏览器不支持 Web Audio。");
    return;
  }

  const ctx = state.music.ctx || new AudioContext();
  if (ctx.state === "suspended") await ctx.resume();
  if (!state.music.master) {
    const master = ctx.createGain();
    master.gain.value = 0.06;
    master.connect(ctx.destination);
    state.music.master = master;
  }
  state.music.ctx = ctx;
  state.music.enabled = true;
  els.musicToggle.setAttribute("aria-pressed", "true");
  syncMusicToScene(true);
  showToast(`宫廷乐已开启：${currentMusicTrack().name}`);
}

function syncMusicToScene(force = false) {
  if (!state.music.enabled || !state.music.ctx || !state.music.master) return;
  const track = currentMusicTrack();
  if (!force && state.music.trackId === track.id) return;
  clearMusicTimers();
  state.music.trackId = track.id;
  state.music.step = 0;
  els.musicToggle.textContent = `乐：${track.name}`;
  playMusicNote(track);
  const beatMs = Math.round(60000 / track.tempo);
  state.music.timers.push(window.setInterval(() => playMusicNote(track), beatMs));
  if (track.drum) {
    state.music.timers.push(window.setInterval(() => playDrum(track), beatMs * 4));
  }
}

function currentMusicTrack() {
  const game = state.game;
  if (!game) return musicTracks[5];
  if (game.ending?.kind === "golden_age") return musicTracks[11];
  if (game.ending) return musicTracks[10];
  if (state.lastAction?.domain) {
    const byDomain = {
      domestic: "people",
      economy: "treasury",
      military: "war",
      diplomacy: "envoy",
      reform: "study",
      intrigue: "intrigue",
      court: "court",
    };
    const id = byDomain[state.lastAction.domain];
    const track = musicTracks.find((item) => item.id === id);
    if (track) return track;
  }
  if (game.phase === "prince") {
    const byScene = {
      "birth-omen": "birth",
      "study-yard": "study",
      "winter-hunt": "hunt",
      "flood-memorial": "flood",
      "succession-night": "succession",
    };
    return musicTracks.find((item) => item.id === byScene[game.scene?.id]) || musicTracks[0];
  }
  if ((game.stats?.borderThreat ?? 0) >= 70) return musicTracks[8];
  if ((game.crisis?.severity ?? 0) >= 72) return musicTracks[10];
  if ((game.stats?.stability ?? 0) >= 82 && (game.stats?.populace ?? 0) >= 82) return musicTracks[11];
  return musicTracks[5];
}

function playMusicNote(track) {
  if (!state.music.enabled || !state.music.ctx || !state.music.master) return;
  const ctx = state.music.ctx;
  const now = ctx.currentTime;
  const step = state.music.step++;
  const interval = track.scale[step % track.scale.length];
  const octave = step % 8 === 0 ? 0.5 : step % 5 === 0 ? 2 : 1;
  const freq = track.root * Math.pow(2, interval / 12) * octave;
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  const filter = ctx.createBiquadFilter();
  osc.type = track.wave;
  osc.frequency.setValueAtTime(freq, now);
  filter.type = "lowpass";
  filter.frequency.value = track.id === "war" ? 1200 : 880;
  gain.gain.setValueAtTime(0.0001, now);
  gain.gain.exponentialRampToValueAtTime(track.id === "intrigue" ? 0.11 : 0.18, now + 0.05);
  gain.gain.exponentialRampToValueAtTime(0.001, now + 1.65);
  osc.connect(filter);
  filter.connect(gain);
  gain.connect(state.music.master);
  osc.start(now);
  osc.stop(now + 1.8);
}

function playDrum(track) {
  if (!state.music.enabled) return;
  const ctx = state.music.ctx;
  const master = state.music.master;
  const now = ctx.currentTime;
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = "sine";
  osc.frequency.setValueAtTime(track.id === "war" ? 112 : 92, now);
  osc.frequency.exponentialRampToValueAtTime(42, now + 0.32);
  gain.gain.setValueAtTime(track.id === "war" ? 0.28 : 0.18, now);
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.42);
  osc.connect(gain);
  gain.connect(master);
  osc.start(now);
  osc.stop(now + 0.45);
}

function stopMusic() {
  clearMusicTimers();
  state.music.enabled = false;
  state.music.trackId = null;
  els.musicToggle.textContent = "开启宫廷乐";
  els.musicToggle.setAttribute("aria-pressed", "false");
  if (state.music.ctx) {
    state.music.ctx.close();
  }
  state.music.ctx = null;
  state.music.master = null;
  showToast("宫廷乐已关闭。");
}

function clearMusicTimers() {
  for (const timer of state.music.timers) {
    window.clearInterval(timer);
  }
  state.music.timers = [];
}

function toggleVoice() {
  state.voice.enabled = !state.voice.enabled;
  els.voiceToggle.textContent = state.voice.enabled ? "关闭配音" : "开启配音";
  els.voiceToggle.setAttribute("aria-pressed", String(state.voice.enabled));
  if (state.voice.enabled) {
    speakText("配音已开启。朱批、御令和危机会由史官念出。");
    showToast("配音已开启，可随时关闭。");
  } else {
    window.speechSynthesis?.cancel();
    showToast("配音已关闭。");
  }
}

function speakResolution(resolution) {
  if (!state.voice.enabled || !resolution?.summary) return;
  speakText(`朱批已下。${resolution.summary}`);
}

function speakText(text) {
  const synth = window.speechSynthesis;
  if (!synth) {
    showToast("当前浏览器不支持配音。");
    return;
  }
  synth.cancel();
  const utterance = new SpeechSynthesisUtterance(text.slice(0, 180));
  utterance.lang = "zh-CN";
  utterance.rate = 0.92;
  utterance.pitch = 0.88;
  const voices = synth.getVoices();
  utterance.voice = voices.find((voice) => voice.lang.toLowerCase().startsWith("zh")) || null;
  synth.speak(utterance);
}

let toastTimer;
function showToast(message) {
  els.toast.textContent = message;
  els.toast.classList.add("show");
  window.clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => els.toast.classList.remove("show"), 2600);
}
