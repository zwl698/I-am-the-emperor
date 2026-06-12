(function () {
  function renderJusticePanels(game, targets) {
    if (!game) return;
    renderOpinion(game, targets.opinion);
    renderCases(game, targets.cases);
  }

  function renderOpinion(game, target) {
    if (!target) return;
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="strategy-empty">登基后开启舆论、法度、禁谣与判词榜示。</article>`;
      return;
    }
    const opinion = game.publicOpinion || {};
    target.innerHTML = `
      <article class="opinion-card ${opinion.rumor >= 70 ? "rumor-hot" : ""} ${opinion.justice >= 70 ? "lawful" : ""}">
        <div class="row-head">
          <strong>京城风声</strong>
          <small>${opinion.rumor >= 70 ? "流言滚沸" : opinion.fear >= 65 ? "人人噤声" : "可控"}</small>
        </div>
        <p>${safe(opinion.lastEdict || "法司尚未递上新案，坊间还在等第一张榜文。")}</p>
        ${meterGroup([
          ["民望", opinion.popular],
          ["士论", opinion.elite],
          ["谣言", opinion.rumor],
          ["畏惧", opinion.fear],
          ["法度", opinion.justice],
        ])}
      </article>
    `;
  }

  function renderCases(game, target) {
    if (!target) return;
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="strategy-empty">登基后刑部、大理寺、都察院会把朝局矛盾变成案卷。</article>`;
      return;
    }
    const cases = game.legalCases || [];
    if (cases.length === 0) {
      target.innerHTML = `<article class="strategy-empty">刑部案架暂空。</article>`;
      return;
    }
    target.innerHTML = cases
      .map((item) => {
        const stateClass = item.resolved ? "resolved" : item.heat >= 75 ? "danger" : "open";
        return `
          <article class="case-row domain-${safeAttr(item.domain)} ${stateClass}">
            <span class="case-sigil">${item.resolved ? "判" : item.heat >= 75 ? "急" : "案"}</span>
            <span class="case-main">
              <span class="row-head">
                <strong>${safe(item.title)}</strong>
                <small>${safe(item.resolved ? item.verdict || "已结" : "待审")}</small>
              </span>
              <em>${safe(item.accuser)} 诉 ${safe(item.defendant)} · ${safe(item.charge)}</em>
              <span>${safe(item.stakes)}</span>
              ${meterGroup([
                ["热", item.heat],
                ["证", item.evidence],
                ["派", item.factionPressure],
                ["民", item.publicPressure],
              ])}
              ${caseButtons(item, game.command)}
            </span>
          </article>
        `;
      })
      .join("");
  }

  function caseButtons(item, command) {
    return `
      <div class="order-buttons">
        ${justiceOrders
          .map((order) => {
            const blocked =
              command <= 0 ||
              ((order.kind === "open_trial" || order.kind === "clemency") && item.resolved) ||
              (order.kind === "proclaim_verdict" && !item.resolved);
            const title = order.kind === "proclaim_verdict" && !item.resolved ? "先明审或宽赦，形成判词后才能宣判" : order.title;
            return orderButton(order, item.id, blocked, title);
          })
          .join("")}
      </div>
    `;
  }

  function orderButton(order, target, disabled, title) {
    const disabledAttrs = disabled ? `disabled data-state-disabled="true"` : "";
    return `<button type="button" ${disabledAttrs} data-order-kind="${safeAttr(order.kind)}" data-order-target="${safeAttr(target)}" data-order-label="${safeAttr(title)}" title="${safeAttr(title)}">${safe(order.label)}</button>`;
  }

  function meterGroup(items) {
    return `
      <div class="justice-meters">
        ${items
          .map(([label, value]) => {
            const width = clampPercent(value);
            return `<span><b>${safe(label)}</b><i style="width:${width}%"></i><em>${width}</em></span>`;
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

  window.renderJusticePanels = renderJusticePanels;
})();
