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
    target.innerHTML = `
      <section class="strategy-map">
        <div class="strategy-map-head">
          <strong>战略地图</strong>
          <small>${strategy.cities.length} 城 · ${strategy.armies.length} 军团 · 对外战争进入地图层</small>
        </div>
        <div class="strategy-board" aria-label="山河战略地图">
          ${strategy.roads.map((road) => roadLine(road, cityByID)).join("")}
          ${strategy.cities.map((city) => cityNode(city, strategy, game)).join("")}
          ${strategy.armies.map((army) => armyPiece(army, cityByID, strategy, game)).join("")}
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

  function roadLine(road, cityByID) {
    const from = cityByID[road.from];
    const to = cityByID[road.to];
    if (!from || !to) return "";
    const dx = to.x - from.x;
    const dy = to.y - from.y;
    const length = Math.sqrt(dx * dx + dy * dy);
    const angle = Math.atan2(dy, dx) * (180 / Math.PI);
    const style = `left:${from.x}%;top:${from.y}%;width:${length}%;transform:rotate(${angle}deg)`;
    return `<span class="strategy-road risk-${riskClass(road.risk)}" style="${style}" title="${safeAttr(road.terrain)} · 风险 ${clamp(road.risk)}"></span>`;
  }

  function cityNode(city, strategy, game) {
    const faction = strategy.factions.find((item) => item.id === city.ownerId);
    const style = `left:${city.x}%;top:${city.y}%;--owner:${safeAttr(faction?.color || "#b74a38")}`;
    const disabled = (game.command ?? 0) <= 0 || city.ownerId !== "court" ? `disabled data-state-disabled="true"` : "";
    return `
      <article class="strategy-city owner-${safeAttr(city.ownerId)} ${city.front ? "front" : ""}" style="${style}" data-city-id="${safeAttr(city.id)}">
        <b>${safe(city.name)}</b>
        <small>兵${shortNumber(city.troops)} · 粮${city.grain} · 防${city.defense}</small>
        <em>${safe(faction?.name || city.ownerId)}</em>
        <button type="button" ${disabled} data-action-kind="city_develop" data-action-mode="${city.disaster >= 40 ? "relief" : "fortify"}" data-action-target="${safeAttr(city.id)}" data-action-label="${city.disaster >= 40 ? "开仓赈灾" : "筑城修垒"}">${city.disaster >= 40 ? "赈" : "筑"}</button>
      </article>
    `;
  }

  function armyPiece(army, cityByID, strategy, game) {
    const city = cityByID[army.location];
    if (!city) return "";
    const faction = strategy.factions.find((item) => item.id === army.factionId);
    const offset = army.factionId === "court" ? -3 : 3;
    const style = `left:${city.x + offset}%;top:${city.y + 5}%;--owner:${safeAttr(faction?.color || "#b74a38")}`;
    const action = armyPrimaryAction(army, strategy);
    const disabled = (game.command ?? 0) <= 0 || army.factionId !== "court" || !action.target ? `disabled data-state-disabled="true"` : "";
    return `
      <button class="strategy-army owner-${safeAttr(army.factionId)}" type="button" ${disabled} style="${style}" data-action-kind="army_command" data-action-mode="${safeAttr(action.mode)}" data-action-target="${safeAttr(action.target)}" data-action-label="${safeAttr(action.label)}" title="${safeAttr(army.name)} · ${safeAttr(army.status)}">
        <b>${safe(army.name.slice(0, 2))}</b>
        <small>${shortNumber(army.troops)}</small>
      </button>
    `;
  }

  function cityCommandDeck(strategy, game) {
    const city = strategy.cities.find((item) => item.ownerId === "court" && item.front) || strategy.cities.find((item) => item.ownerId === "court");
    if (!city) return `<article class="strategy-command-card"><h3>城池经营</h3><p>暂无可治理城池。</p></article>`;
    const disabled = (game.command ?? 0) <= 0 ? `disabled data-state-disabled="true"` : "";
    return `
      <article class="strategy-command-card">
        <h3>城池经营 · ${safe(city.name)}</h3>
        <p>民${stat(city.order)} · 灾${stat(city.disaster)} · 农${stat(city.agriculture)} · 商${stat(city.commerce)}</p>
        <div class="strategy-actions">
          ${strategyAction("city_develop", "farm", city.id, "垦田", "垦田积粮", disabled)}
          ${strategyAction("city_develop", "market", city.id, "修市", "修市开榷", disabled)}
          ${strategyAction("city_develop", "fortify", city.id, "筑城", "筑城修垒", disabled)}
          ${strategyAction("city_develop", "relief", city.id, "赈灾", "开仓赈灾", disabled)}
        </div>
      </article>
    `;
  }

  function armyCommandDeck(strategy, game) {
    const army = strategy.armies.find((item) => item.factionId === "court");
    if (!army) return `<article class="strategy-command-card"><h3>军团军令</h3><p>暂无可调军团。</p></article>`;
    const primary = armyPrimaryAction(army, strategy);
    const disabled = (game.command ?? 0) <= 0 ? `disabled data-state-disabled="true"` : "";
    return `
      <article class="strategy-command-card">
        <h3>军团军令 · ${safe(army.name)}</h3>
        <p>兵${shortNumber(army.troops)} · 粮${army.grain} · 士${army.morale} · 训${army.training}</p>
        <div class="strategy-actions">
          ${strategyAction("army_command", "train", army.id, "整训", "整训军团", disabled)}
          ${strategyAction("army_command", "supply", army.id, "转粮", "转运军粮", disabled)}
          ${strategyAction("army_command", primary.mode, primary.target, primary.shortLabel, primary.label, primary.target ? disabled : `disabled data-state-disabled="true"`)}
        </div>
      </article>
    `;
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
    return `<button type="button" ${disabled} data-action-kind="${safeAttr(kind)}" data-action-mode="${safeAttr(mode)}" data-action-target="${safeAttr(target)}" data-action-label="${safeAttr(title)}">${safe(label)}</button>`;
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
