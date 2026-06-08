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
  music: {
    ctx: null,
    master: null,
    timers: [],
    enabled: false,
  },
};

const els = {
  startSelected: document.querySelector("#start-selected"),
  continueGame: document.querySelector("#continue-game"),
  musicToggle: document.querySelector("#music-toggle"),
  dynastyGrid: document.querySelector("#dynasty-grid"),
  board: document.querySelector("#game-board"),
  phase: document.querySelector("#phase-label"),
  age: document.querySelector("#age-label"),
  stats: document.querySelector("#stats-list"),
  portrait: document.querySelector("#current-portrait"),
  sceneArt: document.querySelector("#scene-art"),
  comicStrip: document.querySelector("#comic-strip"),
  kicker: document.querySelector("#scene-kicker"),
  title: document.querySelector("#scene-title"),
  body: document.querySelector("#scene-body"),
  choices: document.querySelector("#choice-grid"),
  resolution: document.querySelector("#resolution"),
  currentDynasty: document.querySelector("#current-dynasty"),
  crisis: document.querySelector("#crisis-card"),
  objectives: document.querySelector("#objective-list"),
  provinces: document.querySelector("#province-list"),
  factions: document.querySelector("#faction-list"),
  history: document.querySelector("#history-list"),
  toast: document.querySelector("#toast"),
};

els.startSelected.addEventListener("click", () => createGame());
els.continueGame.addEventListener("click", () => continueGame());
els.musicToggle.addEventListener("click", () => toggleMusic());

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
  renderIdentity();
  renderStats();
  renderScene();
  renderComicStrip();
  renderDynastyStatus();
  renderCrisis();
  renderObjectives();
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

function renderComicStrip() {
  const game = state.game;
  const dynastyIndex = Math.max(0, state.dynasties.findIndex((d) => d.id === game.dynasty.id));
  const panels = [
    {
      className: `dynasty-panel panel-${dynastyIndex}`,
      title: game.dynasty.name,
      caption: game.dynasty.challenge,
    },
    {
      className: game.phase === "emperor" ? "character-panel emperor" : "character-panel prince",
      title: game.phase === "emperor" ? "御座" : "东宫",
      caption: game.phase === "emperor" ? "你的一笔朱批，会让天下震动。" : "少年皇子的一次选择，会在多年后回响。",
    },
    {
      className: "crisis-panel",
      title: game.crisis?.title || "朝局",
      caption: game.crisis ? `烈度 ${game.crisis.severity} · 危机钟 ${game.crisis.clock}/8` : "风暴尚未命名。",
    },
  ];

  els.comicStrip.innerHTML = panels
    .map((panel) => `
      <article class="comic-panel ${panel.className}">
        <span></span>
        <strong>${panel.title}</strong>
        <small>${panel.caption}</small>
      </article>
    `)
    .join("");
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

  const ctx = new AudioContext();
  const master = ctx.createGain();
  master.gain.value = 0.055;
  master.connect(ctx.destination);
  state.music.ctx = ctx;
  state.music.master = master;
  state.music.enabled = true;
  els.musicToggle.textContent = "关闭宫廷乐";
  els.musicToggle.setAttribute("aria-pressed", "true");

  const scale = [261.63, 293.66, 329.63, 392.0, 440.0, 523.25];
  let step = 0;
  const playNote = () => {
    if (!state.music.enabled || !state.music.ctx) return;
    const now = ctx.currentTime;
    const freq = scale[step % scale.length] * (step % 5 === 0 ? 0.5 : 1);
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    const filter = ctx.createBiquadFilter();
    osc.type = step % 3 === 0 ? "triangle" : "sine";
    osc.frequency.setValueAtTime(freq, now);
    filter.type = "lowpass";
    filter.frequency.value = 900;
    gain.gain.setValueAtTime(0, now);
    gain.gain.linearRampToValueAtTime(0.18, now + 0.05);
    gain.gain.exponentialRampToValueAtTime(0.001, now + 1.8);
    osc.connect(filter);
    filter.connect(gain);
    gain.connect(master);
    osc.start(now);
    osc.stop(now + 2);
    step += 1;
  };

  playNote();
  state.music.timers.push(window.setInterval(playNote, 950));
  state.music.timers.push(window.setInterval(() => playDrum(ctx, master), 3800));
  showToast("宫廷乐已开启，可随时关闭。");
}

function playDrum(ctx, master) {
  if (!state.music.enabled) return;
  const now = ctx.currentTime;
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = "sine";
  osc.frequency.setValueAtTime(96, now);
  osc.frequency.exponentialRampToValueAtTime(42, now + 0.32);
  gain.gain.setValueAtTime(0.22, now);
  gain.gain.exponentialRampToValueAtTime(0.001, now + 0.42);
  osc.connect(gain);
  gain.connect(master);
  osc.start(now);
  osc.stop(now + 0.45);
}

function stopMusic() {
  for (const timer of state.music.timers) {
    window.clearInterval(timer);
  }
  state.music.timers = [];
  state.music.enabled = false;
  els.musicToggle.textContent = "开启宫廷乐";
  els.musicToggle.setAttribute("aria-pressed", "false");
  if (state.music.ctx) {
    state.music.ctx.close();
  }
  state.music.ctx = null;
  state.music.master = null;
  showToast("宫廷乐已关闭。");
}

let toastTimer;
function showToast(message) {
  els.toast.textContent = message;
  els.toast.classList.add("show");
  window.clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => els.toast.classList.remove("show"), 2600);
}
