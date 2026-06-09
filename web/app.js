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

const sceneAssetNames = [
  "birth-chamber",
  "east-palace-study",
  "winter-hunt",
  "flood-levee",
  "succession-hall",
  "throne-court",
  "granary-relief",
  "tax-office",
  "frontier-fortress",
  "envoy-pass",
  "reform-archive",
  "secret-tribunal",
  "banquet-hall",
  "jiangnan-canal",
  "northern-battlefield",
  "desert-market",
  "imperial-garden",
  "rain-corridor",
  "ancestral-temple",
  "ministry-office",
  "dockyard-fleet",
  "drill-ground",
  "rebel-village",
  "silk-market",
  "mountain-monastery",
  "exam-hall",
  "map-room",
  "palace-dawn",
  "diplomatic-tent",
  "festival-night",
];

const portraitAssetNames = [
  "infant-prince",
  "teen-prince",
  "young-emperor",
  "elder-emperor",
  "stern-tutor",
  "frontier-general",
  "finance-minister",
  "grand-princess",
  "noble-consort",
  "young-empress",
  "queen-dowager",
  "palace-maid",
  "eunuch-spymaster",
  "scholar-official",
  "reformist-official",
  "corrupt-magistrate",
  "merchant-leader",
  "foreign-envoy",
  "nomad-khan",
  "monk-strategist",
  "female-diplomat",
  "guard-captain",
  "rebel-leader",
  "river-engineer",
  "imperial-physician",
  "astrologer",
  "poet",
  "court-painter",
  "farmer-representative",
  "masked-assassin",
];

const sceneGalleryFallback = assetPaths("/assets/scenes/scene", sceneAssetNames);
const portraitGalleryFallback = assetPaths("/assets/portraits/portrait", portraitAssetNames);

const portraitIndexByRole = {
  infant: 0,
  prince: 1,
  emperor: 2,
  elder: 3,
  tutor: 4,
  general: 5,
  minister: 6,
  consort: 7,
  empress: 9,
  dowager: 10,
  maid: 11,
  spy: 12,
  scholar: 13,
  reformer: 14,
  corrupt: 15,
  merchant: 16,
  envoy: 17,
  khan: 18,
  monk: 19,
  diplomat: 20,
  guard: 21,
  rebel: 22,
  engineer: 23,
  physician: 24,
  astrologer: 25,
  poet: 26,
  painter: 27,
  farmer: 28,
  assassin: 29,
};

const provinceOrders = [
  { kind: "relief", label: "赈", title: "赈济：降灾情、涨民心，耗粮银" },
  { kind: "garrison", label: "驻", title: "驻防：升防务、压边患，耗军费" },
  { kind: "tax", label: "税", title: "督税：涨国库，伤地方秩序" },
  { kind: "canal", label: "渠", title: "修渠：长期富庶与新政，耗国库" },
  { kind: "trade", label: "市", title: "互市：涨财政外交，略增边患" },
  { kind: "inspect", label: "查", title: "密查：升秩序，抓胥吏" },
];

const factionOrders = [
  { kind: "appease", label: "安", title: "安抚：涨忠诚，耗银与威权" },
  { kind: "purge", label: "削", title: "削权：降权势，激化党争" },
  { kind: "inspect", label: "查", title: "密查：压权势，损稳定" },
];

const musicTracks = [
  { id: "birth", name: "雪宫摇篮", root: 196, scale: [0, 3, 5, 7, 10, 12], tempo: 72, wave: "sine", drum: false },
  { id: "study", name: "东宫书声", root: 220, scale: [0, 2, 5, 7, 9, 12], tempo: 84, wave: "triangle", drum: false },
  { id: "hunt", name: "雪猎急弦", root: 174, scale: [0, 3, 5, 7, 10, 12], tempo: 120, wave: "sawtooth", drum: true },
  { id: "flood", name: "南河雨鼓", root: 164, scale: [0, 2, 3, 7, 8, 12], tempo: 92, wave: "triangle", drum: true },
  { id: "succession", name: "烛影夺嫡", root: 146, scale: [0, 1, 5, 7, 8, 12], tempo: 96, wave: "sine", drum: true },
  { id: "court", name: "太和晨钟", root: 196, scale: [0, 2, 4, 7, 9, 12], tempo: 82, wave: "triangle", drum: true },
  { id: "people", name: "粥棚烟火", root: 185, scale: [0, 2, 5, 7, 9, 12], tempo: 76, wave: "sine", drum: false },
  { id: "treasury", name: "银库算盘", root: 207, scale: [0, 2, 4, 6, 9, 12], tempo: 102, wave: "square", drum: false },
  { id: "war", name: "边塞战鼓", root: 130, scale: [0, 3, 5, 7, 10, 12], tempo: 128, wave: "sawtooth", drum: true },
  { id: "envoy", name: "驼铃万里", root: 174, scale: [0, 2, 5, 7, 11, 12], tempo: 88, wave: "triangle", drum: false },
  { id: "intrigue", name: "密档夜灯", root: 155, scale: [0, 1, 5, 6, 8, 12], tempo: 90, wave: "sine", drum: true },
  { id: "festival", name: "万邦烟火", root: 247, scale: [0, 2, 4, 7, 9, 14], tempo: 112, wave: "triangle", drum: true },
];

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
  kicker: document.querySelector("#scene-kicker"),
  title: document.querySelector("#scene-title"),
  body: document.querySelector("#scene-body"),
  choices: document.querySelector("#choice-grid"),
  resolution: document.querySelector("#resolution"),
  currentDynasty: document.querySelector("#current-dynasty"),
  crisis: document.querySelector("#crisis-card"),
  commandStatus: document.querySelector("#command-status"),
  objectives: document.querySelector("#objective-list"),
  provinces: document.querySelector("#province-list"),
  factions: document.querySelector("#faction-list"),
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
      title: label || orderTitle(kind),
      domain: orderDomain(kind),
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

function renderGame() {
  if (!state.game) return;
  els.board.classList.remove("is-empty");
  renderIdentity();
  renderStats();
  renderScene();
  renderComicStrip();
  renderDynastyStatus();
  renderCrisis();
  renderCommands();
  renderObjectives();
  renderProvinces();
  renderFactions();
  renderHistory();
  attachOrderButtons();
  syncMusicToScene();
}

function renderIdentity() {
  const game = state.game;
  const dynasty = currentDynasty();
  const portraitClass = game.phase === "emperor" ? "emperor" : "prince";
  els.portrait.className = `portrait-crop ${portraitClass}`;
  els.portrait.style.backgroundImage = `url('${identityPortrait()}')`;
  els.phase.textContent = `${dynasty.name} · ${phaseName[game.phase] || "未知"}`;
  const calendar = game.phase === "emperor" ? `登基${game.reignYear}年 · ${game.season}` : `${game.age} 岁`;
  const commandText = game.phase === "emperor" ? ` · 御令剩 ${game.command ?? 0}` : "";
  els.age.textContent = `${calendar} · 第 ${game.turn} 回合${commandText}`;
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
  const dynasty = currentDynasty();
  const action = state.lastAction;
  const panels = [
    {
      className: "scene-panel",
      image: game.scene?.art || sceneGallery()[0],
      title: dynasty.name,
      caption: game.scene?.mood ? `${game.scene.year} · ${game.scene.mood}` : dynasty.challenge,
    },
    {
      className: "character-panel",
      image: identityPortrait(),
      title: game.phase === "emperor" ? "御座落笔" : "东宫心性",
      caption: game.phase === "emperor" ? "你的一笔朱批，会让天下震动。" : "少年皇子的一次选择，会在多年后回响。",
    },
    {
      className: `action-panel domain-${action?.domain || "court"}`,
      image: domainSceneArt(action?.domain || game.scene?.choices?.[0]?.domain),
      title: action?.title || "命运未落子",
      caption: action?.summary || "点击剧情选择，或登基后下御令，漫画会随之推进。",
    },
    {
      className: "crisis-panel",
      image: crisisSceneArt(),
      title: game.crisis?.title || "朝局",
      caption: game.crisis ? `烈度 ${game.crisis.severity} · 危机钟 ${game.crisis.clock}/8` : "风暴尚未命名。",
    },
  ];

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
  return `
    <button class="choice-card domain-${choice.domain}" type="button" data-choice="${choice.id}">
      <span class="choice-icon">${domainIcon[choice.domain] || "策"}</span>
      <span>
        <strong>${choice.text}</strong>
        <small>${choice.detail}</small>
        <em>${formatEffects(choice.effects)}</em>
        <b>大议题 · 推进一季</b>
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

function renderCommands() {
  const game = state.game;
  if (game.phase !== "emperor") {
    els.commandStatus.innerHTML = "皇子阶段先积累性格、名望与盟友。登基后会开启每季度多道御令、地方治理和派系压制。";
    return;
  }
  const command = game.command ?? 0;
  els.commandStatus.innerHTML = `
    <strong>${command} 道御令可用</strong>
    <span>先用御令处理具体省份/派系，再点中央大议题推进季度。盛世终局至少需要 72 个大回合。</span>
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

function assetPaths(prefix, names) {
  return names.map((name, index) => `${prefix}-${String(index + 1).padStart(2, "0")}-${name}.png`);
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
  const disabled = (state.game.command ?? 0) <= 0 ? "disabled" : "";
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

function attachOrderButtons() {
  document.querySelectorAll("[data-order-kind]").forEach((button) => {
    button.addEventListener("click", () => {
      issueOrder(button.dataset.orderKind, button.dataset.orderTarget, button.dataset.orderLabel);
    });
  });
}

function provinceTemperature(p) {
  if (p.disaster >= 60) return "灾情急";
  if (p.order <= 35) return "民变险";
  if (p.defense <= 35) return "防务弱";
  if (p.wealth >= 75) return "富庶";
  return "可治";
}

function orderTitle(kind) {
  return [...provinceOrders, ...factionOrders].find((order) => order.kind === kind)?.title || "御令";
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
    command: game.command ?? 0,
    crisis: game.crisis || {
      title: "朝局未明",
      severity: 40,
      clock: 2,
      summary: "旧存档缺少新版危机数据，已按默认朝局继续。",
    },
    objectives: game.objectives || [],
    factions: game.factions || [],
    provinces: game.provinces || [],
    court: game.court || [],
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
