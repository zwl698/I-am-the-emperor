(function () {
  function renderEventHand(game, target) {
    if (!target) return;
    if (!game || game.phase !== "emperor") {
      target.innerHTML = `<section class="event-hand locked"><strong>事件手牌</strong><span>登基后每季由灾害、战争、财政、后宫、继承和外交压力发牌。</span></section>`;
      return;
    }
    const cards = (game.eventHand || []).slice(0, 5);
    if (!cards.length) {
      target.innerHTML = `<section class="event-hand"><div class="event-hand-head"><strong>事件手牌</strong><small>本季暂未收到新牌，推进一季后刷新。</small></div></section>`;
      return;
    }
    target.innerHTML = `
      <section class="event-hand">
        <div class="event-hand-head">
          <strong>事件手牌</strong>
          <small>${cards.length} 张局势牌 · 按压力动态发牌 · 可直接转入小游戏行动</small>
        </div>
        <div class="event-hand-track">
          ${cards.map((card, index) => eventCard(card, game, index)).join("")}
        </div>
      </section>
    `;
  }

  function eventCard(card, game, index) {
    const action = suggestedAction(card, game);
    return `
      <article class="event-hand-card domain-${safeAttr(card.domain || "court")}" style="--card-index:${index}">
        <div class="event-hand-topline">
          <span>${safe(card.category || "朝局")}</span>
          <em>急 ${clamp(card.urgency)} · 烈 ${clamp(card.severity)}</em>
        </div>
        <h3>${safe(card.title)}</h3>
        <p>${safe(card.summary)}</p>
        <blockquote>${safe(card.hook || card.stage || "")}</blockquote>
        <small>${safe(card.consequence || "处理结果会影响本季局势。")}</small>
        ${actionButton(action, game)}
      </article>
    `;
  }

  function suggestedAction(card, game) {
    const category = card.category || "";
    const domain = card.domain || "";
    if (category.includes("战争") || domain === "military") {
      const action = strategicWarAction(game);
      if (action.target) return action;
      const war = mostThreateningWar(game);
      if (war?.id) {
        return { kind: "war_tactic", mode: (war.threat || 0) >= 70 ? "campaign" : "mobilize", target: war.id, label: (war.threat || 0) >= 70 ? "开沙盘决战" : "拨粮整军" };
      }
      // 边境暂宁、无现成战局时，回退到练兵或边防巡查，保证战争牌始终有入口
      const army = courtArmy(game);
      if (army) return { kind: "army_command", mode: "train", target: army.id, label: "整军备战" };
      const border = borderProvince(game);
      if (border) return { kind: "map_allocation", mode: "garrison", target: border.id, label: "增戍边防" };
      return { kind: "map_allocation", mode: "inspect", target: worstProvince(game)?.id || "", label: "巡阅边镇" };
    }
    if (category.includes("灾害") || domain === "domestic") {
      const city = worstStrategicCity(game);
      if (city) return { kind: "city_develop", mode: "relief", target: city.id, label: "开仓赈灾" };
      const province = worstProvince(game);
      return { kind: "map_allocation", mode: "relief", target: province?.id || "", label: "调度赈灾" };
    }
    if (category.includes("财政") || domain === "economy") {
      const city = richestStrategicCity(game);
      if (city) return { kind: "city_develop", mode: "market", target: city.id, label: "修市筹银" };
      const province = richestProvince(game);
      return { kind: "map_allocation", mode: "tax", target: province?.id || "", label: "清丈筹银" };
    }
    if (category.includes("继承") || category.includes("东宫")) {
      const heir = (game.heirs || [])[0];
      return { kind: "heir_lesson", mode: "study", target: heir?.id || "", label: "召太傅授课" };
    }
    if (category.includes("外交") || category.includes("诸邦") || domain === "diplomacy") {
      const foreign = mostDangerousForeign(game);
      return { kind: "envoy_mission", mode: "embassy", target: foreign?.id || "", label: "遣使修好" };
    }
    if (category.includes("刑狱") || category.includes("密谋") || domain === "intrigue") {
      const legalCase = (game.legalCases || []).find((item) => !item.resolved) || (game.legalCases || [])[0];
      return { kind: "trial_move", mode: "open_trial", target: legalCase?.id || "", label: "开堂追查" };
    }
    if (category.includes("官职") || category.includes("朝堂")) {
      const office = (game.offices || [])[0];
      const minister = (game.court || [])[0];
      const target = office && minister ? `${office.id}:${minister.id}` : "";
      return { kind: "office_assign", mode: "appoint", target, label: "调官补署" };
    }
    const province = worstProvince(game);
    return { kind: "map_allocation", mode: "inspect", target: province?.id || "", label: "派巡按入局" };
  }

  function actionButton(action, game) {
    const noTarget = !action?.target;
    const noCommand = (game.command ?? 0) <= 0;
    const focusAttrs = actionFocusAttrs(action);
    const panel = focusPanelForAction(action);

    // 无可执行目标 / 御令耗尽：不再渲染成无法点击的死按钮，
    // 而是退化为“入面板手动处置”的软入口，避免战争等事件卡死。
    if (noTarget || noCommand) {
      const hint = noCommand ? "御令已尽，下季再行" : "需在面板内手动处置";
      if (panel) {
        return `
          <button class="event-action event-action-soft" type="button"${focusAttrs} data-focus-soft="true" title="${safeAttr(hint)}">
            ${safe(action?.label || "查看局势")} · 入面板
          </button>
        `;
      }
      return `
        <button class="event-action" type="button" disabled data-state-disabled="true" title="${safeAttr(hint)}">
          ${safe(action?.label || "暂无行动")}
        </button>
      `;
    }

    return `
      <button class="event-action" type="button"${focusAttrs} data-action-kind="${safeAttr(action?.kind)}" data-action-mode="${safeAttr(action?.mode)}" data-action-target="${safeAttr(action?.target)}" data-action-label="${safeAttr(action?.label)}">
        ${safe(action?.label || "暂无行动")}
      </button>
    `;
  }

  function actionFocusAttrs(action) {
    const panel = focusPanelForAction(action);
    if (!panel) return "";
    const target = focusTargetForAction(action);
    const targetAttr = target ? ` data-focus-target="${safeAttr(target)}"` : "";
    return ` data-focus-panel="${safeAttr(panel)}"${targetAttr}`;
  }

  function focusPanelForAction(action) {
    switch (action?.kind) {
      case "city_develop":
      case "army_command":
      case "siege_command":
      case "governor_assign":
      case "war_tactic":
        return "strategy-map-panel";
      case "trial_move":
        return "case-list";
      case "office_assign":
        return "office-list";
      case "envoy_mission":
        return "foreign-list";
      case "heir_lesson":
        return "heir-list";
      case "map_allocation":
        return "province-list";
      default:
        return "";
    }
  }

  function focusTargetForAction(action) {
    const [primary, secondary] = String(action?.target || "").split(":");
    if (action?.kind === "army_command" && ["march", "assault", "besiege"].includes(action?.mode)) return secondary || primary;
    if (action?.kind === "siege_command") return secondary || primary;
    if (action?.kind === "war_tactic") return strategicTargetForWar(primary);
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

  function mostThreateningWar(game) {
    return (game.wars || []).slice().sort((a, b) => (b.threat || 0) - (a.threat || 0))[0];
  }

  function courtArmy(game) {
    return (game.strategy?.armies || [])
      .filter((army) => army.factionId === "court")
      .slice()
      .sort((a, b) => (b.troops || 0) - (a.troops || 0))[0];
  }

  function borderProvince(game) {
    const provinces = (game.provinces || []).slice();
    const front = provinces.filter((p) => p.front || p.border || (p.threat || 0) > 0);
    const pool = front.length ? front : provinces;
    return pool.sort((a, b) => (b.threat || 0) - (a.threat || 0))[0];
  }

  function strategicWarAction(game) {
    const strategy = game.strategy || {};
    const plan = strategicWarPlan(strategy);
    if (!plan) return { kind: "army_command", mode: "train", target: "", label: "整训军团" };
    const { army, action } = plan;
    return { kind: "army_command", mode: action.mode, target: action.target, label: action.label };
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
    return { mode: "train", target: army.id, label: "整训军团" };
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

  function worstStrategicCity(game) {
    return (game.strategy?.cities || [])
      .filter((city) => city.ownerId === "court")
      .slice()
      .sort((a, b) => (b.disaster || 0) + (50 - (b.order || 0)) - ((a.disaster || 0) + (50 - (a.order || 0))))[0];
  }

  function richestStrategicCity(game) {
    return (game.strategy?.cities || [])
      .filter((city) => city.ownerId === "court")
      .slice()
      .sort((a, b) => (b.commerce || 0) + (b.gold || 0) - ((a.commerce || 0) + (a.gold || 0)))[0];
  }

  function worstProvince(game) {
    return (game.provinces || []).slice().sort((a, b) => (b.disaster || 0) + (50 - (b.order || 0)) - ((a.disaster || 0) + (50 - (a.order || 0))))[0];
  }

  function richestProvince(game) {
    return (game.provinces || []).slice().sort((a, b) => (b.wealth || 0) - (a.wealth || 0))[0];
  }

  function mostDangerousForeign(game) {
    return (game.foreignStates || []).slice().sort((a, b) => (b.threat || 0) - (a.threat || 0))[0];
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

  window.renderEventHand = renderEventHand;
})();
