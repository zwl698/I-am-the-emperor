(function () {
  function renderDiplomacyIntrigue(game, targets, api = {}) {
    if (!game) return;
    renderForeignStates(game, targets.foreign, api);
    renderPlots(game, targets.plots);
  }

  function renderForeignStates(game, target, api) {
    if (!target) return;
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="strategy-empty">登基后开启外邦诸国、使节往来与长期盟约。</article>`;
      return;
    }
    const states = game.foreignStates || [];
    if (states.length === 0) {
      target.innerHTML = `<article class="strategy-empty">鸿胪寺尚无外邦档案。</article>`;
      return;
    }
    target.innerHTML = states
      .map((foreign) => {
        const pressure = foreign.threat >= 72 ? "danger" : foreign.relation >= 65 || foreign.treaty ? "friendly" : "watching";
        return `
          <article class="diplomacy-row foreign-row ${pressure} ${foreign.treaty ? "treaty" : ""}">
            <span class="diplomacy-portrait" style="background-image:url('${safeAttr(portraitFor(foreign.portrait, api))}')"></span>
            <span class="diplomacy-main">
              <span class="row-head">
                <strong>${safe(foreign.name)}</strong>
                <small>${safe(foreign.attitude || "观望")}</small>
              </span>
              <em>${safe(foreign.ruler)} · ${safe(foreign.envoy)} · ${safe(foreign.treaty || "未缔盟")}</em>
              ${meterGroup([
                ["交", foreign.relation],
                ["威", foreign.threat],
                ["贡", foreign.tribute],
                ["筹", foreign.leverage],
              ])}
              ${foreignButtons(foreign, game.command)}
            </span>
          </article>
        `;
      })
      .join("");
  }

  function renderPlots(game, target) {
    if (!target) return;
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="strategy-empty">登基后开启潜伏阴谋、缇骑侦缉与爆发危机。</article>`;
      return;
    }
    const plots = game.plots || [];
    if (plots.length === 0) {
      target.innerHTML = `<article class="strategy-empty">密档库暂未发现暗线。</article>`;
      return;
    }
    target.innerHTML = plots
      .map((plot) => {
        const stateClass = plot.resolved ? "resolved" : plot.exposed ? "exposed" : plot.progress >= 70 ? "danger" : "hidden";
        return `
          <article class="diplomacy-row plot-row ${stateClass}">
            <span class="plot-sigil">${plot.resolved ? "结" : plot.exposed ? "露" : "密"}</span>
            <span class="diplomacy-main">
              <span class="row-head">
                <strong>${safe(plot.title)}</strong>
                <small>${safe(plot.stage || "潜伏")}</small>
              </span>
              <em>${safe(plot.sponsor)} → ${safe(plot.target)}</em>
              <span>${safe(plot.summary)}</span>
              ${meterGroup([
                ["隐", plot.secrecy],
                ["进", plot.progress],
                ["险", plot.danger],
              ])}
              ${plotButtons(plot, game.command)}
            </span>
          </article>
        `;
      })
      .join("");
  }

  function foreignButtons(foreign, command) {
    return `
      <div class="order-buttons">
        ${foreignOrders
          .map((order) => {
            const blocked = command <= 0 || (order.kind === "treaty" && (foreign.relation < 55 || !!foreign.treaty));
            const reason = order.kind === "treaty" && foreign.relation < 55 ? "关系需至少 55 才能缔约" : order.title;
            return orderButton(order, foreign.id, blocked, reason);
          })
          .join("")}
      </div>
    `;
  }

  function plotButtons(plot, command) {
    return `
      <div class="order-buttons">
        ${plotOrders
          .map((order) => {
            const blocked = command <= 0 || plot.resolved || (order.kind === "suppress_plot" && !plot.exposed);
            const reason = order.kind === "suppress_plot" && !plot.exposed ? "阴谋暴露后才能平谋" : order.title;
            return orderButton(order, plot.id, blocked, reason);
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
      <div class="diplomacy-meters">
        ${items
          .map(([label, value]) => {
            const width = clampPercent(value);
            return `<span><b>${safe(label)}</b><i style="width:${width}%"></i><em>${width}</em></span>`;
          })
          .join("")}
      </div>
    `;
  }

  function portraitFor(role, api) {
    const map = api.portraitIndexByRole || {};
    const key = role || "envoy";
    const index = map[key] ?? map.envoy ?? map.diplomat;
    if (typeof api.portraitAt === "function" && index !== undefined) {
      return api.portraitAt(index);
    }
    return "";
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

  window.renderDiplomacyIntrigue = renderDiplomacyIntrigue;
})();
