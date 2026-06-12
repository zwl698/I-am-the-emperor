(function () {
  function renderGrandStrategy(game, targets) {
    if (!game) return;
    renderProjectsAndPolicies(game, targets.strategy);
    renderRelations(game, targets.relations);
  }

  function renderProjectsAndPolicies(game, target) {
    if (!target) return;
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="strategy-empty">登基后开启多年工程与常驻国策。</article>`;
      return;
    }
    const projects = game.projects || [];
    const policies = game.policies || [];
    target.innerHTML = `
      <div class="strategy-subtitle">多年工程</div>
      ${projects
        .map((project) => `
          <article class="strategy-row domain-${project.domain}">
            <div class="row-head">
              <strong>${safe(project.name)}</strong>
              <small>${safe(project.stage)}</small>
            </div>
            <span>${safe(project.description)}</span>
            <div class="strategy-meter"><i style="width:${clampPercent(project.progress)}%"></i><em>${project.progress}/100 · 风险 ${project.risk}</em></div>
            <small>${safe(project.reward)}</small>
            ${strategyButtons(projectOrders, project.id, game.command, project.completed)}
          </article>
        `)
        .join("")}
      <div class="strategy-subtitle">常驻政策</div>
      ${policies
        .map((policy) => `
          <article class="strategy-row policy-row ${policy.active ? "active" : ""} domain-${policy.domain}">
            <div class="row-head">
              <strong>${safe(policy.name)}</strong>
              <small>${policy.active ? "施行中" : "待诏"}</small>
            </div>
            <span>${safe(policy.description)}</span>
            <small>维护 ${policy.upkeep} · 阻力 ${policy.strain}</small>
            ${strategyButtons(policyOrders, policy.id, game.command, false)}
          </article>
        `)
        .join("")}
    `;
  }

  function renderRelations(game, target) {
    if (!target) return;
    const relations = game.relations || [];
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="strategy-empty">成长阶段的选择会影响登基后的关系底色。</article>`;
      return;
    }
    target.innerHTML = relations
      .map((relation) => `
        <article class="relation-row">
          <strong>${safe(relation.from)} ↔ ${safe(relation.to)}</strong>
          <span>${safe(relation.bond)} · ${safe(relation.description)}</span>
          <div class="relation-bars">
            <small>信 ${relation.trust}</small><i class="trust" style="width:${clampPercent(relation.trust)}%"></i>
            <small>怨 ${relation.tension}</small><i class="tension" style="width:${clampPercent(relation.tension)}%"></i>
          </div>
        </article>
      `)
      .join("");
  }

  function strategyButtons(orders, target, command, disabled) {
    return `
      <div class="order-buttons">
        ${orders
          .map((order) => {
            const off = disabled || command <= 0;
            const attrs = off ? `disabled data-state-disabled="true"` : "";
            return `<button type="button" ${attrs} data-order-kind="${safeAttr(order.kind)}" data-order-target="${safeAttr(target)}" data-order-label="${safeAttr(order.title)}" title="${safeAttr(order.title)}">${safe(order.label)}</button>`;
          })
          .join("")}
      </div>
    `;
  }

  function clampPercent(value) {
    return Math.max(0, Math.min(100, Number(value) || 0));
  }

  function safe(value) {
    return String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;");
  }

  function safeAttr(value) {
    return safe(value).replaceAll("'", "&#39;");
  }

  window.renderGrandStrategy = renderGrandStrategy;
})();
