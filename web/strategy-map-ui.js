(function () {
  function renderStrategyMap(game, target) {
    if (!target) return;
    if (!game || game.phase !== "emperor") {
      target.innerHTML = `<section class="strategy-map locked"><strong>战略地图</strong><span>登基后开启城池、道路、军团与对外战争地图。</span></section>`;
      return;
    }
    const strategy = normalizedStrategy(game.strategy);
    if (!strategy.cities.length) {
      target.innerHTML = `<section class="strategy-map"><div class="strategy-map-head"><strong>战略地图</strong><small>暂无地图数据。</small></div></section>`;
      return;
    }
    const cityByID = Object.fromEntries(strategy.cities.map((city) => [city.id, city]));
    const courtCities = strategy.cities.filter((city) => city.ownerId === "court").length;
    const fronts = strategy.cities.filter((city) => city.front).length;
    target.innerHTML = `
      <section class="strategy-map">
        <div class="strategy-map-head">
          <strong>山河舆图</strong>
          <small>${strategy.cities.length} 城 · 朝廷据 ${courtCities} · 前线 ${fronts} · ${strategy.armies.length} 军团</small>
        </div>
        <div class="strategy-board" aria-label="山河战略地图">
          ${terrainCanvas(strategy, cityByID)}
          <div class="strategy-board-overlay">
            ${strategy.cities.map((city) => cityNode(city, strategy, game)).join("")}
            ${strategy.armies.map((army) => armyPiece(army, cityByID, strategy, game)).join("")}
          </div>
          ${mapLegend(strategy)}
        </div>
        <div class="strategy-command-deck">
          ${cityCommandDeck(strategy, game)}
          ${armyCommandDeck(strategy, game)}
          ${battleReports(strategy)}
          ${strategyLogs(strategy)}
        </div>
      </section>
    `;
  }

  // 形象化地图底图：用 SVG 画出河流、山脉、海岸与州域分块，再叠加道路网，
  // 让画面像一张真正的山河舆图，而不是抽象网格。
  function terrainCanvas(strategy, cityByID) {
    const roads = (strategy.roads || [])
      .map((road) => roadPath(road, cityByID))
      .filter(Boolean)
      .join("");
    const labels = regionLabels()
      .map((r) => `<text class="map-region-label" x="${r.x}" y="${r.y}">${safe(r.name)}</text>`)
      .join("");
    return `
      <svg class="strategy-terrain" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
        <defs>
          <radialGradient id="mapPlain" cx="48%" cy="46%" r="62%">
            <stop offset="0%" stop-color="#7d6233" stop-opacity="0.55" />
            <stop offset="55%" stop-color="#5b4a2a" stop-opacity="0.42" />
            <stop offset="100%" stop-color="#33301f" stop-opacity="0.5" />
          </radialGradient>
          <linearGradient id="mapSea" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#1f4f5e" stop-opacity="0.85" />
            <stop offset="100%" stop-color="#15323d" stop-opacity="0.9" />
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="100" height="100" fill="url(#mapPlain)" />
        <!-- 东南海域 -->
        <path d="M82 38 Q92 50 90 70 Q88 92 70 100 L100 100 L100 30 Z" fill="url(#mapSea)" />
        <path d="M64 96 Q74 86 82 70 Q88 56 84 42" class="map-coast" />
        <!-- 北疆雪山 -->
        <path d="M30 2 L46 16 L62 4 L74 18 L58 22 L50 14 L40 22 Z" class="map-mountain map-snow" />
        <path d="M44 18 L52 10 L60 20 Z" class="map-mountain map-snow" />
        <!-- 西陲沙岭 -->
        <path d="M6 30 L20 38 L14 50 L4 46 Z" class="map-mountain map-desert" />
        <path d="M12 44 L24 48 L20 60 L8 56 Z" class="map-mountain map-desert" />
        <!-- 西南群山 -->
        <path d="M22 60 L36 64 L32 82 L20 78 Z" class="map-mountain" />
        <path d="M30 70 L42 74 L40 90 L28 86 Z" class="map-mountain" />
        <!-- 大河：北境—河东—洛阳—漕都—江南—入海 -->
        <path d="M50 14 Q46 28 44 36 Q42 46 48 56 Q54 64 58 70 Q66 76 78 64" class="map-river" />
        <!-- 漕运支流 -->
        <path d="M44 48 Q40 56 33 62" class="map-river map-river-thin" />
        ${labels}
        ${roads}
      </svg>
    `;
  }

  function regionLabels() {
    return [
      { name: "北疆", x: 50, y: 6 },
      { name: "中原", x: 44, y: 40 },
      { name: "西疆", x: 14, y: 34 },
      { name: "西南", x: 26, y: 70 },
      { name: "江南", x: 56, y: 74 },
      { name: "海疆", x: 82, y: 58 },
      { name: "南疆", x: 45, y: 92 },
    ];
  }

  function roadPath(road, cityByID) {
    const from = cityByID[road.from];
    const to = cityByID[road.to];
    if (!from || !to) return "";
    const mx = (from.x + to.x) / 2;
    const my = (from.y + to.y) / 2;
    // 道路弯一点，更像山河之间的路径
    const nx = -(to.y - from.y);
    const ny = to.x - from.x;
    const len = Math.sqrt(nx * nx + ny * ny) || 1;
    const bend = 4;
    const cx = mx + (nx / len) * bend;
    const cy = my + (ny / len) * bend;
    return `<path class="map-road risk-${riskClass(road.risk)}" d="M${from.x} ${from.y} Q${cx} ${cy} ${to.x} ${to.y}"><title>${safeAttr(road.terrain)} · 风险 ${clamp(road.risk)}</title></path>`;
  }

  function mapLegend(strategy) {
    const factions = (strategy.factions || []).slice(0, 6);
    return `
      <div class="strategy-legend" aria-hidden="true">
        ${factions
          .map(
            (faction) => `<span><i style="--owner:${safeAttr(faction.color || "#b74a38")}"></i>${safe(faction.name)}</span>`,
          )
          .join("")}
      </div>
    `;
  }

  function cityIcon(city) {
    const tags = city.tags || [];
    if (tags.includes("都城")) return "城";
    if (tags.some((t) => t.includes("海") || t.includes("舟"))) return "港";
    if (tags.some((t) => t.includes("关") || t.includes("山"))) return "关";
    if (tags.some((t) => t.includes("粮") || t.includes("仓") || t.includes("漕"))) return "仓";
    if (tags.some((t) => t.includes("市") || t.includes("商"))) return "市";
    return "城";
  }

  function cityNode(city, strategy, game) {
    const faction = strategy.factions.find((item) => item.id === city.ownerId);
    const style = `left:${city.x}%;top:${city.y}%;--owner:${safeAttr(faction?.color || "#b74a38")}`;
    const isCourt = city.ownerId === "court";
    const noCommand = (game.command ?? 0) <= 0;
    const developMode = city.disaster >= 40 ? "relief" : "fortify";
    const developLabel = city.disaster >= 40 ? "赈" : "筑";
    const developTitle = city.disaster >= 40 ? "开仓赈灾" : "筑城修垒";
    const disabled = noCommand || !isCourt ? `disabled data-state-disabled="true"` : "";
    const alert = city.disaster >= 40 ? "alert" : "";
    return `
      <article class="strategy-city owner-${safeAttr(city.ownerId)} ${city.front ? "front" : ""} ${alert}" style="${style}" data-city-id="${safeAttr(city.id)}" data-strategy-target="${safeAttr(city.id)}" title="${safe(city.name)} · ${safe(faction?.name || city.ownerId)}">
        <span class="city-marker">${safe(cityIcon(city))}</span>
        <span class="city-card">
          <b>${safe(city.name)}</b>
          <small>兵${shortNumber(city.troops)} · 粮${city.grain} · 防${city.defense}</small>
          <em>${safe(faction?.name || city.ownerId)}</em>
          <button type="button" ${disabled} data-action-kind="city_develop" data-action-mode="${safeAttr(developMode)}" data-action-target="${safeAttr(city.id)}" data-action-label="${safeAttr(developTitle)}" title="${noCommand ? "御令已尽，下季再行" : developTitle}">${safe(developLabel)}</button>
        </span>
      </article>
    `;
  }

  function armyPiece(army, cityByID, strategy, game) {
    const city = cityByID[army.location];
    if (!city) return "";
    const faction = strategy.factions.find((item) => item.id === army.factionId);
    const offset = army.factionId === "court" ? -5 : 5;
    const style = `left:${clamp(city.x + offset)}%;top:${clamp(city.y + 9)}%;--owner:${safeAttr(faction?.color || "#b74a38")}`;
    const action = armyPrimaryAction(army, strategy);
    const noCommand = (game.command ?? 0) <= 0;
    const disabled = noCommand || army.factionId !== "court" || !action.target ? `disabled data-state-disabled="true"` : "";
    const hint = army.factionId !== "court" ? `${army.name} · ${army.status}` : noCommand ? "御令已尽，下季再行" : action.label;
    return `
      <button class="strategy-army owner-${safeAttr(army.factionId)}" type="button" ${disabled} style="${style}" data-army-id="${safeAttr(army.id)}" data-strategy-target="${safeAttr(army.id)}" data-action-kind="army_command" data-action-mode="${safeAttr(action.mode)}" data-action-target="${safeAttr(action.target)}" data-action-label="${safeAttr(action.label)}" title="${safeAttr(hint)}">
        <b>${safe(army.name.slice(0, 2))}</b>
        <small>${shortNumber(army.troops)}</small>
      </button>
    `;
  }

  function cityCommandDeck(strategy, game) {
    const city = strategy.cities.find((item) => item.ownerId === "court" && item.front) || strategy.cities.find((item) => item.ownerId === "court");
    if (!city) return `<article class="strategy-command-card"><h3>城池经营</h3><p>暂无可治理城池。</p></article>`;
    const noCommand = (game.command ?? 0) <= 0;
    const disabled = noCommand ? `disabled data-state-disabled="true"` : "";
    return `
      <article class="strategy-command-card" data-strategy-target="${safeAttr(city.id)}">
        <h3>城池经营 · ${safe(city.name)}</h3>
        <p>民${stat(city.order)} · 灾${stat(city.disaster)} · 农${stat(city.agriculture)} · 商${stat(city.commerce)}</p>
        ${noCommand ? `<p class="strategy-command-hint">御令已尽，推进一季后可再下城池军令。</p>` : ""}
        <div class="strategy-actions">
          ${strategyAction("city_develop", "farm", city.id, "垦田", "垦田积粮", disabled)}
          ${strategyAction("city_develop", "market", city.id, "修市", "修市开榷", disabled)}
          ${strategyAction("city_develop", "fortify", city.id, "筑城", "筑城修垒", disabled)}
          ${strategyAction("city_develop", "relief", city.id, "赈灾", "开仓赈灾", disabled)}
        </div>
      </article>
    `;
  }

  // 军团指令台：优先挑出“此刻能动手”的军团（能攻城/围城/补给的前线军团），
  // 而不是固定取第一支禁军，避免守在京畿的军团让对外战争“看起来玩不了”。
  function armyCommandDeck(strategy, game) {
    const courtArmies = (strategy.armies || []).filter((item) => item.factionId === "court");
    if (!courtArmies.length) return `<article class="strategy-command-card"><h3>军团军令</h3><p>暂无可调军团。</p></article>`;
    const army = pickActionableArmy(courtArmies, strategy);
    const primary = armyPrimaryAction(army, strategy);
    const noCommand = (game.command ?? 0) <= 0;
    const disabled = noCommand ? `disabled data-state-disabled="true"` : "";
    const cityByID = Object.fromEntries(strategy.cities.map((c) => [c.id, c]));
    const here = cityByID[army.location];
    const neighborText = neighborSummary(army, strategy);
    return `
      <article class="strategy-command-card strategy-army-card" data-army-id="${safeAttr(army.id)}" data-strategy-target="${safeAttr(army.id)}">
        <h3>军团军令 · ${safe(army.name)}</h3>
        <p>驻${safe(here?.name || army.location)} · 兵${shortNumber(army.troops)} · 粮${army.grain} · 士${army.morale} · 训${army.training}</p>
        <p class="strategy-command-hint">${safe(neighborText)}</p>
        ${noCommand ? `<p class="strategy-command-hint">御令已尽，推进一季后可再调遣军团。</p>` : ""}
        <div class="strategy-actions">
          ${strategyAction("army_command", "train", army.id, "整训", "整训军团", disabled)}
          ${strategyAction("army_command", "supply", army.id, "转粮", "转运军粮", disabled)}
          ${strategyAction("army_command", primary.mode, primary.target, primary.shortLabel, primary.label, primary.target ? disabled : `disabled data-state-disabled="true"`)}
        </div>
        ${courtArmies.length > 1 ? armySwitcher(courtArmies, army, strategy, disabled) : ""}
      </article>
    `;
  }

  // 多支军团时给出快速切换/直接下令的列表，让玩家能指挥任意一支军团去攻打邻城。
  function armySwitcher(courtArmies, current, strategy, disabled) {
    return `
      <div class="strategy-army-switch">
        ${courtArmies
          .map((army) => {
            const action = armyPrimaryAction(army, strategy);
            const active = army.id === current.id ? "active" : "";
            const canAct = !!action.target;
            const btnDisabled = canAct ? disabled : `disabled data-state-disabled="true"`;
            return `<button type="button" class="${active}" ${btnDisabled} data-strategy-target="${safeAttr(army.id)}" data-army-id="${safeAttr(army.id)}" data-action-kind="army_command" data-action-mode="${safeAttr(action.mode)}" data-action-target="${safeAttr(action.target)}" data-action-label="${safeAttr(action.label)}" title="${safeAttr(action.label)}">${safe(army.name)} · ${safe(action.shortLabel)}</button>`;
          })
          .join("")}
      </div>
    `;
  }

  function neighborSummary(army, strategy) {
    const neighbors = roadNeighbors(strategy, army.location)
      .map((id) => strategy.cities.find((city) => city.id === id))
      .filter(Boolean);
    const hostile = neighbors.filter((city) => city.ownerId !== "court").map((city) => city.name);
    if (hostile.length) return `邻接敌城：${hostile.join("、")}，可挥师攻城。`;
    if (neighbors.length) return `周边皆为朝廷城池，可行军调防或就地整训。`;
    return "孤悬无道路相连，宜整训待命。";
  }

  function pickActionableArmy(courtArmies, strategy) {
    const scored = courtArmies
      .map((army) => ({ army, action: armyPrimaryAction(army, strategy) }))
      .sort((a, b) => armyActionScore(b, strategy) - armyActionScore(a, strategy));
    return scored[0].army;
  }

  function armyActionScore(plan, strategy) {
    const modeScore = { assault: 100, besiege: 90, supply: 70, march: 40, train: 10 };
    const city = strategy.cities.find((item) => item.id === plan.army.location);
    let score = modeScore[plan.action.mode] || 0;
    if (city?.front) score += 18;
    score += Math.min(20, (plan.army.troops || 0) / 1500);
    return score;
  }

  function battleReports(strategy) {
    const battles = (strategy.battles || []).slice(0, 3);
    if (!battles.length) return `<article class="strategy-command-card strategy-battle-card"><h3>最近战报</h3><p>尚无会战记录。外战攻城、围城迫降和边城失守都会记录在这里。</p></article>`;
    return `
      <article class="strategy-command-card strategy-battle-card">
        <h3>最近战报</h3>
        ${battles
          .map(
            (battle) => `
              <div class="strategy-battle-report severity-${riskClass(battle.severity)}">
                <b>${safe(battle.title)}</b>
                <span>${safe(outcomeLabel(battle.outcome))}</span>
                <small>攻损${formatLoss(battle.attackerLoss)} · 守损${formatLoss(battle.defenderLoss)}</small>
                <em>${safe(participantText(battle.participants, strategy))}</em>
                ${battleFactors(battle)}
              </div>
            `,
          )
          .join("")}
      </article>
    `;
  }

  function strategyLogs(strategy) {
    const logs = (strategy.logs || []).slice(0, 3);
    if (!logs.length) return `<article class="strategy-command-card"><h3>战局纪要</h3><p>本季尚无军报。</p></article>`;
    return `
      <article class="strategy-command-card">
        <h3>战局纪要</h3>
        ${logs.map((log) => `<p><b>${safe(log.title)}</b> ${safe(log.summary)}</p>`).join("")}
      </article>
    `;
  }

  function battleFactors(battle) {
    const factors = (battle.factors || []).slice(0, 3);
    if (!factors.length) return "";
    return `<ul>${factors.map((factor) => `<li>${safe(factor)}</li>`).join("")}</ul>`;
  }

  function armyPrimaryAction(army, strategy) {
    if (army.factionId !== "court") return { mode: "train", target: "", label: "敌军行动", shortLabel: "敌" };
    if ((army.grain ?? 99) <= 12) {
      const here = strategy.cities.find((city) => city.id === army.location);
      if (here && here.ownerId === "court") return { mode: "supply", target: army.id, label: "转运军粮", shortLabel: "转粮" };
    }
    const neighbors = roadNeighbors(strategy, army.location);
    const hostile = neighbors.map((id) => strategy.cities.find((city) => city.id === id)).find((city) => city && city.ownerId !== "court");
    if (hostile) return { mode: "assault", target: `${army.id}:${hostile.id}`, label: `攻打${hostile.name}`, shortLabel: "攻城" };
    const friendly = neighbors.map((id) => strategy.cities.find((city) => city.id === id)).find((city) => city && city.ownerId === "court");
    if (friendly) return { mode: "march", target: `${army.id}:${friendly.id}`, label: `行军至${friendly.name}`, shortLabel: "行军" };
    return { mode: "train", target: army.id, label: "整训军团", shortLabel: "整训" };
  }

  function roadNeighbors(strategy, cityID) {
    const result = [];
    for (const road of strategy.roads || []) {
      if (road.from === cityID) result.push(road.to);
      if (road.to === cityID) result.push(road.from);
    }
    return result;
  }

  function strategyAction(kind, mode, target, label, title, disabled) {
    return `<button type="button" ${disabled} data-action-kind="${safeAttr(kind)}" data-action-mode="${safeAttr(mode)}" data-action-target="${safeAttr(target)}" data-action-label="${safeAttr(title)}" title="${safeAttr(title)}">${safe(label)}</button>`;
  }

  function normalizedStrategy(strategy) {
    return {
      cities: strategy?.cities || [],
      roads: strategy?.roads || [],
      factions: strategy?.factions || [],
      armies: strategy?.armies || [],
      logs: strategy?.logs || [],
      battles: strategy?.battles || [],
    };
  }

  function riskClass(value) {
    if ((value || 0) >= 40) return "high";
    if ((value || 0) >= 24) return "mid";
    return "low";
  }

  function shortNumber(value) {
    const number = Number(value) || 0;
    if (number >= 10000) return `${Math.round(number / 1000) / 10}万`;
    return String(number);
  }

  function stat(value) {
    return Number.isFinite(Number(value)) ? Number(value) : 0;
  }

  function formatLoss(value) {
    return String(Math.max(0, Number(value) || 0));
  }

  function outcomeLabel(outcome) {
    const labels = {
      capture: "攻占",
      repelled: "受挫",
      surrender: "迫降",
      enemy_capture: "失守",
    };
    return labels[outcome] || outcome || "未明";
  }

  function participantText(participants, strategy) {
    const armiesByID = Object.fromEntries((strategy.armies || []).map((army) => [army.id, army]));
    return (participants || [])
      .map((id) => {
        const army = armiesByID[id];
        return army?.name ? `${army.name}(${id})` : id;
      })
      .join("、");
  }

  function clamp(value) {
    return Math.max(0, Math.min(100, Number(value) || 0));
  }

  function safe(value) {
    return String(value ?? "").replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;");
  }

  function safeAttr(value) {
    return safe(value).replaceAll("'", "&#39;");
  }

  window.renderStrategyMap = renderStrategyMap;
})();
