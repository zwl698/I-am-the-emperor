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
  military: "兵",
  diplomacy: "使",
  court: "宫",
};

const phaseName = {
  prince: "皇子",
  emperor: "皇帝",
};

const state = {
  game: null,
  busy: false,
};

const els = {
  newGame: document.querySelector("#new-game"),
  continueGame: document.querySelector("#continue-game"),
  board: document.querySelector("#game-board"),
  phase: document.querySelector("#phase-label"),
  age: document.querySelector("#age-label"),
  stats: document.querySelector("#stats-list"),
  kicker: document.querySelector("#scene-kicker"),
  title: document.querySelector("#scene-title"),
  body: document.querySelector("#scene-body"),
  choices: document.querySelector("#choice-grid"),
  resolution: document.querySelector("#resolution"),
  history: document.querySelector("#history-list"),
  toast: document.querySelector("#toast"),
};

els.newGame.addEventListener("click", () => createGame());
els.continueGame.addEventListener("click", () => continueGame());

renderEmptyStats();
continueGame({ silent: true });

async function createGame() {
  setBusy(true);
  try {
    const seed = Date.now();
    const res = await fetch("/api/games", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ seed }),
    });
    state.game = await readJSON(res);
    localStorage.setItem("emperor-game-id", state.game.id);
    els.resolution.hidden = true;
    renderGame();
    showToast("新局已开。太史翻开了第一页。");
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
  renderHistory();
}

function renderIdentity() {
  const game = state.game;
  els.phase.textContent = phaseName[game.phase] || "未知";
  els.age.textContent = `${game.age} 岁 · 第 ${game.turn} 回合`;
}

function renderEmptyStats() {
  els.stats.innerHTML = statMeta
    .map(([, label]) => statRow(label, 0, false))
    .join("");
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
  if (game.ending) {
    els.kicker.textContent = "结局";
    els.title.textContent = game.ending.title;
    els.body.textContent = game.ending.summary;
    els.choices.innerHTML = `
      <article class="ending-card">
        <p>你的王朝走到了这一页。可以重开一局，尝试另一条帝王路。</p>
      </article>
    `;
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
    <button class="choice-card" type="button" data-choice="${choice.id}">
      <span class="choice-icon">${domainIcon[choice.domain] || "策"}</span>
      <span>
        <h3>${choice.text}</h3>
        <p>${choice.detail}</p>
        <span class="choice-effects">${formatEffects(choice.effects)}</span>
      </span>
    </button>
  `;
}

function renderResolution(resolution) {
  if (!resolution) return;
  els.resolution.hidden = false;
  els.resolution.innerHTML = `<strong>朱批已下：</strong>${resolution.summary}<br><small>${formatEffects(resolution.effects)}</small>`;
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
