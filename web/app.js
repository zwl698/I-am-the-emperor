const statMeta = [
  ["legitimacy", "名望"],
  ["health", "健康"],
  ["learning", "学识"],
  ["martial", "武略"],
  ["charisma", "魅力"],
  ["influence", "势力"],
  ["treasury", "国库"],
  ["grain", "粮草"],
  ["populace", "民心"],
  ["army", "军力"],
  ["diplomacy", "邦交"],
  ["stability", "朝稳"],
  ["borderThreat", "边患"],
  ["reform", "新政"],
];

const domainIcon = {
  story: "卷",
  domestic: "民",
  economy: "财",
  military: "兵",
  diplomacy: "使",
  court: "宫",
  reform: "法",
  intrigue: "密",
};

const phaseName = {
  prince: "皇子",
  emperor: "皇帝",
};

const dynastyPanelClass = {
  dayin: "panel-dayin",
  jingyao: "panel-jingyao",
  chengping: "panel-chengping",
  xuanshuo: "panel-xuanshuo",
};

const state = {
  game: null,
  dynasties: [],
  selectedDynasty: "dayin",
  busy: false,
};

const els = {
  startSelected: document.querySelector("#start-selected"),
  continueGame: document.querySelector("#continue-game"),
  dynastyGrid: document.querySelector("#dynasty-grid"),
  board: document.querySelector("#game-board"),
  phase: document.querySelector("#phase-label"),
  age: document.querySelector("#age-label"),
  stats: document.querySelector("#stats-list"),
  portrait: document.querySelector("#current-portrait"),
  sceneArt: document.querySelector("#scene-art"),
  kicker: document.querySelector("#scene-kicker"),
  title: document.querySelector("#scene-title"),
  body: document.querySelector("#scene-body"),
  choices: document.querySelector("#choice-grid"),
  resolution: document.querySelector("#resolution"),
  currentDynasty: document.querySelector("#current-dynasty"),
  crisis: document.querySelector("#crisis-card"),
  provinces: document.querySelector("#province-list"),
  factions: document.querySelector("#faction-list"),
  history: document.querySelector("#history-list"),
  toast: document.querySelector("#toast"),
};

els.startSelected.addEventListener("click", () => createGame());
els.continueGame.addEventListener("click", () => continueGame());

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
    state.game = await readJSON(res);
    localStorage.setItem("emperor-game-id", state.game.id);
    els.resolution.hidden = true;
    renderGame();
    showToast(`${state.game.dynasty.name}开局。史官翻开了第一页。`);
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
    state.game = await readJSON(res);
    state.selectedDynasty = state.game.dynasty.id;
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
  setBusy(true);
  try {
    const res = await fetch(`/api/games/${state.game.id}/choices`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ choiceId }),
    });
    const payload = await readJSON(res);
    state.game = payload.state;
    renderResolution(payload.resolution);
    renderGame();
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

function renderGame() {
  if (!state.game) return;
  els.board.classList.remove("is-empty");
  renderIdentity();
  renderStats();
  renderScene();
  renderDynastyStatus();
  renderCrisis();
  renderProvinces();
  renderFactions();
  renderHistory();
}

function renderIdentity() {
  const game = state.game;
  const portraitClass = game.phase === "emperor" ? "emperor" : "prince";
  els.portrait.className = `portrait-crop ${portraitClass}`;
  els.phase.textContent = `${game.dynasty.name} · ${phaseName[game.phase] || "未知"}`;
  const calendar = game.phase === "emperor" ? `登基${game.reignYear}年 · ${game.season}` : `${game.age} 岁`;
  els.age.textContent = `${calendar} · 第 ${game.turn} 回合`;
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
  els.kicker.textContent = `${scene.year} · ${scene.mood}`;
  els.title.textContent = scene.title;
  els.body.textContent = scene.body;
  els.choices.innerHTML = scene.choices.map(choiceButton).join("");
  document.querySelectorAll("[data-choice]").forEach((button) => {
    button.addEventListener("click", () => choose(button.dataset.choice));
  });
}

function choiceButton(choice) {
  return `
    <button class="choice-card domain-${choice.domain}" type="button" data-choice="${choice.id}">
      <span class="choice-icon">${domainIcon[choice.domain] || "策"}</span>
      <span>
        <strong>${choice.text}</strong>
        <small>${choice.detail}</small>
        <em>${formatEffects(choice.effects)}</em>
      </span>
    </button>
  `;
}

function renderResolution(resolution) {
  if (!resolution) return;
  els.resolution.hidden = false;
  els.resolution.innerHTML = `<strong>朱批已下：</strong>${resolution.summary}<br><small>${formatEffects(resolution.effects)}</small>`;
}

function renderDynastyStatus() {
  const dynasty = state.game.dynasty;
  els.currentDynasty.innerHTML = `
    <div class="panel-title">${dynasty.name}</div>
    <p>${dynasty.background}</p>
    <ul>${dynasty.features.map((feature) => `<li>${feature}</li>`).join("")}</ul>
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

function renderProvinces() {
  els.provinces.innerHTML = (state.game.provinces || [])
    .map((p) => `
      <article class="mini-world-row">
        <strong>${p.name}</strong>
        <span>${p.focus}</span>
        <small>富 ${p.wealth} · 安 ${p.order} · 防 ${p.defense} · 灾 ${p.disaster}</small>
      </article>
    `)
    .join("");
}

function renderFactions() {
  els.factions.innerHTML = (state.game.factions || [])
    .map((faction) => `
      <article class="faction-row">
        <span class="portrait-dot ${faction.portrait}"></span>
        <span>
          <strong>${faction.name}</strong>
          <small>${faction.leader} · ${faction.agenda}</small>
          <em>权势 ${faction.power} · 忠诚 ${faction.loyalty}</em>
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

async function readJSON(res) {
  const payload = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(payload.error || "请求失败");
  return payload;
}

function setBusy(busy) {
  state.busy = busy;
  document.querySelectorAll("button").forEach((button) => {
    button.disabled = busy;
  });
}

let toastTimer;
function showToast(message) {
  els.toast.textContent = message;
  els.toast.classList.add("show");
  window.clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => els.toast.classList.remove("show"), 2600);
}
