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
      return { kind: "war_tactic", mode: (war?.threat || 0) >= 70 ? "campaign" : "mobilize", target: war?.id || "", label: (war?.threat || 0) >= 70 ? "开沙盘决战" : "拨粮整军" };
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
    const disabled = !action?.target || (game.command ?? 0) <= 0;
    const attrs = disabled ? `disabled data-state-disabled="true"` : "";
    return `
      <button class="event-action" type="button" ${attrs} data-action-kind="${safeAttr(action?.kind)}" data-action-mode="${safeAttr(action?.mode)}" data-action-target="${safeAttr(action?.target)}" data-action-label="${safeAttr(action?.label)}">
        ${safe(action?.label || "暂无行动")}
      </button>
    `;
  }

  function mostThreateningWar(game) {
    return (game.wars || []).slice().sort((a, b) => (b.threat || 0) - (a.threat || 0))[0];
  }

  function strategicWarAction(game) {
    const strategy = game.strategy || {};
    const army = (strategy.armies || []).find((item) => item.factionId === "court");
    if (!army) return { kind: "army_command", mode: "train", target: "", label: "整训军团" };
    const hostile = roadNeighbors(strategy, army.location)
      .map((id) => (strategy.cities || []).find((city) => city.id === id))
      .find((city) => city && city.ownerId !== "court");
    if (hostile) return { kind: "army_command", mode: "assault", target: `${army.id}:${hostile.id}`, label: `攻打${hostile.name}` };
    return { kind: "army_command", mode: "train", target: army.id, label: "整训军团" };
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
