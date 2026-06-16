(function () {
  function renderMiniGames(game, target) {
    if (!target) return;
    if (!game || game.phase !== "emperor") {
      target.innerHTML = `<section class="playdesk locked"><strong>御前操作台</strong><span>登基后开启兵棋沙盘、三司会审和六部调度。</span></section>`;
      return;
    }
    target.innerHTML = `
      <section class="playdesk">
        <div class="playdesk-head">
          <strong>御前操作台</strong>
          <small>${game.command ?? 0} 道御令 · 点击棋格/案牌/官署筹码直接执行</small>
        </div>
        <div class="playdesk-grid">
          ${warMiniGame(game)}
          ${trialMiniGame(game)}
          ${officeMiniGame(game)}
        </div>
      </section>
    `;
  }

  function warMiniGame(game) {
    const strategic = strategyWarMiniGame(game);
    if (strategic) return strategic;
    const war = (game.wars || []).find((item) => item.stage !== "凯旋") || (game.wars || [])[0];
    if (!war) {
      return `<article class="mini-game-card"><h3>兵棋沙盘</h3><p>边境暂宁，暂无可推演战局。</p></article>`;
    }
    return `
      <article class="mini-game-card war-game">
        <h3>兵棋沙盘</h3>
        <p>${safe(war.name)} · ${safe(war.enemy)} · ${safe(war.stage)}</p>
        <div class="tactical-board">
          ${tacticButton("mobilize", war.id, "粮道", war.supply, "补粮道，稳士气", (game.command ?? 0) <= 0)}
          ${tacticButton("campaign", war.id, "突击", war.progress, "推进战果，压敌势", (game.command ?? 0) <= 0)}
          ${tacticButton("fortify", war.id, "筑垒", 100 - war.threat, "固边墙，拖垮敌势", (game.command ?? 0) <= 0)}
        </div>
        <small>敌势 ${war.threat} · 粮道 ${war.supply} · 士气 ${war.morale} · 战果 ${war.progress}</small>
      </article>
    `;
  }

  function strategyWarMiniGame(game) {
    const strategy = game.strategy || {};
    if (!(strategy.cities || []).length || !(strategy.armies || []).length) return "";
    const plan = strategicWarPlan(strategy);
    if (!plan) return "";
    const { army, action } = plan;
    return `
      <article class="mini-game-card war-game">
        <h3>兵棋沙盘</h3>
        <p>${safe(army.name)} · ${safe(cityName(strategy, army.location))} · ${safe(army.status)}</p>
        <div class="tactical-board">
          ${actionButton("army_command", "train", army.id, "整", "闭营整训，提升士气训练", (game.command ?? 0) <= 0)}
          ${actionButton("army_command", "supply", army.id, "粮", "转运军粮，稳住粮道", (game.command ?? 0) <= 0)}
          ${actionButton("army_command", action.mode, action.target, action.label.slice(0, 1), action.label, (game.command ?? 0) <= 0 || !action.target)}
        </div>
        <small>兵 ${army.troops} · 粮 ${army.grain} · 士气 ${army.morale} · 训练 ${army.training}</small>
      </article>
    `;
  }

  function trialMiniGame(game) {
    const legalCase = (game.legalCases || []).find((item) => !item.resolved) || (game.legalCases || [])[0];
    if (!legalCase) {
      return `<article class="mini-game-card"><h3>三司会审</h3><p>刑部案架暂空。</p></article>`;
    }
    const canProclaim = !!legalCase.resolved;
    const disabledTrial = legalCase.resolved || (game.command ?? 0) <= 0;
    return `
      <article class="mini-game-card trial-game">
        <h3>三司会审</h3>
        <p>${safe(legalCase.title)} · ${safe(legalCase.defendant)} · 热度 ${legalCase.heat}</p>
        <div class="verdict-cards">
          ${actionButton("trial_move", "open_trial", legalCase.id, "明审", `公开审理：证据 ${legalCase.evidence}`, disabledTrial)}
          ${actionButton("trial_move", "clemency", legalCase.id, "宽赦", "从轻发落，换取短期稳定", disabledTrial)}
          ${actionButton("trial_move", "censor_rumor", legalCase.id, "禁谣", "压下传帖，牺牲名望", (game.command ?? 0) <= 0)}
          ${actionButton("trial_move", "proclaim_verdict", legalCase.id, "宣判", "榜示判词，转化口碑", !canProclaim || (game.command ?? 0) <= 0)}
        </div>
      </article>
    `;
  }

  function officeMiniGame(game) {
    const office = (game.offices || []).slice().sort((a, b) => (b.vacancyRisk || 0) - (a.vacancyRisk || 0))[0];
    if (!office) {
      return `<article class="mini-game-card"><h3>六部调度</h3><p>暂无官署差遣。</p></article>`;
    }
    const candidates = (game.court || []).slice().sort((a, b) => fitScore(b, office.domain) - fitScore(a, office.domain)).slice(0, 3);
    return `
      <article class="mini-game-card office-game">
        <h3>六部调度</h3>
        <p>${safe(office.title)} · 权威 ${office.authority} · 空转 ${office.vacancyRisk}</p>
        <div class="minister-chips">
          ${candidates
            .map((minister) =>
              actionButton(
                "office_assign",
                "appoint",
                `${office.id}:${minister.id}`,
                minister.name.slice(0, 2),
                `${minister.name}接掌${office.title} · 适配 ${Math.round(fitScore(minister, office.domain))}`,
                (game.command ?? 0) <= 0 || minister.id === office.holderId,
              ),
            )
            .join("")}
        </div>
      </article>
    `;
  }

  function tacticButton(kind, target, label, value, title, disabled) {
    const attrs = disabled ? `disabled data-state-disabled="true"` : "";
    const focusTarget = strategicTargetForWar(target);
    return `
      <button class="tactic-cell" type="button" ${attrs} data-focus-panel="strategy-map-panel" data-focus-target="${safeAttr(focusTarget)}" data-action-kind="war_tactic" data-action-mode="${safeAttr(kind)}" data-action-target="${safeAttr(target)}" data-action-label="${safeAttr(title)}" title="${safeAttr(title)}">
        <b>${safe(label)}</b><i style="height:${clamp(value)}%"></i><em>${clamp(value)}</em>
      </button>
    `;
  }

  function actionButton(kind, mode, target, label, title, disabled) {
    const attrs = disabled ? `disabled data-state-disabled="true"` : "";
    return `<button class="mini-order" type="button" ${attrs}${actionFocusAttrs(kind, mode, target)} data-action-kind="${safeAttr(kind)}" data-action-mode="${safeAttr(mode)}" data-action-target="${safeAttr(target)}" data-action-label="${safeAttr(title)}" title="${safeAttr(title)}">${safe(label)}</button>`;
  }

  function actionFocusAttrs(kind, mode, target) {
    const panel = focusPanelForAction(kind);
    if (!panel) return "";
    const focusTarget = focusTargetForAction(kind, mode, target);
    const targetAttr = focusTarget ? ` data-focus-target="${safeAttr(focusTarget)}"` : "";
    return ` data-focus-panel="${safeAttr(panel)}"${targetAttr}`;
  }

  function focusPanelForAction(kind) {
    switch (kind) {
      case "army_command":
      case "city_develop":
      case "siege_command":
      case "governor_assign":
        return "strategy-map-panel";
      case "trial_move":
        return "case-list";
      case "office_assign":
        return "office-list";
      case "envoy_mission":
        return "foreign-list";
      case "heir_lesson":
        return "heir-list";
      default:
        return "";
    }
  }

  function focusTargetForAction(kind, mode, target) {
    const [primary, secondary] = String(target || "").split(":");
    if (kind === "army_command" && ["march", "assault", "besiege"].includes(mode)) return secondary || primary;
    if (kind === "siege_command") return secondary || primary;
    return primary;
  }

  function strategicTargetForWar(warID) {
    const map = {
      "snow-ridge": "snow-ridge",
      "western-oath": "jade-pass",
      "river-bandits": "river-east",
      "jade-pass": "jade-pass",
      north: "snow-ridge",
    };
    return map[warID] || warID || "";
  }

  function fitScore(minister, domain) {
    const base = (minister.ability || 0) + (minister.integrity || 0) / 2 - (minister.stress || 0) / 3;
    const roleBonus = {
      economy: minister.role === "户部尚书" ? 24 : 0,
      military: minister.role === "大将军" ? 24 : 0,
      diplomacy: minister.role === "长公主" ? 20 : 0,
      reform: minister.role === "太傅" ? 20 : 0,
      intrigue: minister.integrity >= 70 ? 14 : 0,
      court: minister.role === "长公主" ? 18 : 0,
    };
    return base + (roleBonus[domain] || 0);
  }

  function armyPrimaryAction(army, strategy) {
    if ((army.grain ?? 99) <= 10) return { mode: "supply", target: army.id, label: "转运军粮" };
    const hostile = roadNeighbors(strategy, army.location)
      .map((id) => (strategy.cities || []).find((city) => city.id === id))
      .find((city) => city && city.ownerId !== "court");
    if (hostile) return { mode: "assault", target: `${army.id}:${hostile.id}`, label: `攻打${hostile.name}` };
    const friendly = roadNeighbors(strategy, army.location)
      .map((id) => (strategy.cities || []).find((city) => city.id === id))
      .find((city) => city && city.ownerId === "court");
    if (friendly) return { mode: "march", target: `${army.id}:${friendly.id}`, label: `行军至${friendly.name}` };
    return { mode: "train", target: army.id, label: "整训" };
  }

  function strategicWarPlan(strategy) {
    const candidates = (strategy.armies || [])
      .filter((army) => army.factionId === "court")
      .map((army) => ({ army, action: armyPrimaryAction(army, strategy) }))
      .filter((plan) => plan.action.target);
    if (!candidates.length) return null;
    candidates.sort((a, b) => warPlanScore(b, strategy) - warPlanScore(a, strategy));
    return candidates[0];
  }

  function warPlanScore(plan, strategy) {
    const modeScore = { supply: 120, assault: 100, march: 58, train: 32 };
    const city = (strategy.cities || []).find((item) => item.id === plan.army.location);
    let score = modeScore[plan.action.mode] || 0;
    if (city?.front) score += 18;
    score += Math.min(20, (plan.army.troops || 0) / 1000);
    score += Math.min(12, plan.army.morale || 0) / 2;
    if ((plan.army.grain ?? 99) <= 10) score += 35;
    return score;
  }

  function roadNeighbors(strategy, cityID) {
    const neighbors = [];
    for (const road of strategy.roads || []) {
      if (road.from === cityID) neighbors.push(road.to);
      if (road.to === cityID) neighbors.push(road.from);
    }
    return neighbors;
  }

  function cityName(strategy, cityID) {
    return (strategy.cities || []).find((city) => city.id === cityID)?.name || cityID;
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

  window.renderMiniGames = renderMiniGames;
})();
