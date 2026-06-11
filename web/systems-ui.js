(function () {
  function renderSystemPanels(game, targets, api = {}) {
    if (!game) return;
    renderHarem(game, targets.harem, api);
    renderHeirs(game, targets.heirs, api);
    renderOffices(game, targets.offices, api);
  }

  function renderHarem(game, target, api) {
    if (!target) return;
    const consorts = game.harem || [];
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="system-empty">登基后开启后宫势力、外戚与宠爱博弈。</article>`;
      return;
    }
    target.innerHTML = consorts
      .map((consort) => {
        const portrait = portraitByKey(consort.portrait, api, "consort");
        return `
          <article class="system-row harem-row">
            <span class="portrait-dot system-portrait" style="background-image:url('${portrait}')"></span>
            <span class="system-main">
              <strong>${safe(consort.name)} <em>${safe(consort.rank)}</em></strong>
              <small>${safe(consort.clan)} · ${safe(consort.trait)} · 子嗣 ${childCount(consort)}</small>
              ${miniMeters([
                ["宠", consort.favor],
                ["戚", consort.familyPower],
                ["野", consort.ambition],
                ["势", consort.influence],
              ])}
              ${systemOrderButtons(consortOrders, consort.id, game.command, consort.name)}
            </span>
          </article>
        `;
      })
      .join("");
  }

  function renderHeirs(game, target, api) {
    if (!target) return;
    const heirs = game.heirs || [];
    const succession = game.succession || {};
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="system-empty">皇子阶段的选择会影响登基后的储位格局。</article>`;
      return;
    }
    target.innerHTML = `
      <article class="succession-strip">
        <strong>储位稳定 ${succession.stability ?? 0}</strong>
        <span class="system-track"><i style="width:${clampPercent(succession.stability)}%"></i></span>
        <small>争议 ${succession.dispute ?? 0} · 母族势 ${succession.maternalClanPower ?? 0}</small>
        <em>${safe(succession.lastSuccessionMove || "东宫名分尚需经营。")}</em>
      </article>
      ${heirs
        .map((heir) => {
          const portrait = portraitByKey(heir.portrait, api, heir.age <= 5 ? "infant" : "prince");
          const named = heir.named || succession.namedHeirId === heir.id;
          const disabled = named || (game.command ?? 0) <= 0;
          return `
            <article class="system-row heir-row ${named ? "named" : ""}">
              <span class="portrait-dot system-portrait" style="background-image:url('${portrait}')"></span>
              <span class="system-main">
                <strong>${safe(heir.name)} ${named ? "<em>储君</em>" : ""}</strong>
                <small>${heir.age} 岁 · 母族 ${motherName(game, heir.motherId)} · 康 ${heir.health}</small>
                ${miniMeters([
                  ["才", heir.talent],
                  ["野", heir.ambition],
                  ["拥", heir.support],
                ])}
                <div class="order-buttons">
                  ${orderButton("name_heir", heir.id, "储", "册储：指定此人为继承人", disabled)}
                </div>
              </span>
            </article>
          `;
        })
        .join("")}
    `;
  }

  function renderOffices(game, target, api) {
    if (!target) return;
    const offices = game.offices || [];
    if (game.phase !== "emperor") {
      target.innerHTML = `<article class="system-empty">登基后可任免六大官署，影响财政、军务、改革、外交与暗线。</article>`;
      return;
    }
    target.innerHTML = offices
      .map((office) => {
        const holder = (game.court || []).find((minister) => minister.id === office.holderId);
        const candidates = rankedCandidates(game.court || [], office).slice(0, 3);
        return `
          <article class="office-row domain-${office.domain}">
            <div class="row-head">
              <strong>${safe(office.title)}</strong>
              <small>${domainLabel(office.domain)}</small>
            </div>
            <span>${safe(office.seat)} · ${holder ? `现任 ${safe(holder.name)}` : "暂缺"}</span>
            ${miniMeters([
              ["权", office.authority],
              ["空", office.vacancyRisk],
            ])}
            <div class="office-candidates">
              ${candidates
                .map((minister) =>
                  orderButton(
                    "appoint",
                    `${office.id}:${minister.id}`,
                    minister.name.slice(0, 1),
                    `任官：${minister.name}出任${office.title}`,
                    (game.command ?? 0) <= 0 || minister.id === office.holderId,
                  ),
                )
                .join("")}
              ${orderButton("dismiss", office.id, "罢", `罢官：清空${office.title}`, (game.command ?? 0) <= 0 || !office.holderId)}
            </div>
          </article>
        `;
      })
      .join("");
  }

  function systemOrderButtons(orders, target, command, targetName) {
    if (!Array.isArray(orders)) return "";
    return `
      <div class="order-buttons">
        ${orders.map((order) => orderButton(order.kind, target, order.label, `${targetName} · ${order.title}`, command <= 0)).join("")}
      </div>
    `;
  }

  function orderButton(kind, target, label, title, disabled) {
    const disabledAttrs = disabled ? `disabled data-state-disabled="true"` : "";
    return `<button type="button" ${disabledAttrs} data-order-kind="${safeAttr(kind)}" data-order-target="${safeAttr(target)}" data-order-label="${safeAttr(title)}" title="${safeAttr(title)}">${safe(label)}</button>`;
  }

  function miniMeters(items) {
    return `
      <div class="system-meters">
        ${items
          .map(([label, value]) => `
            <span><b>${safe(label)}</b><i style="width:${clampPercent(value)}%"></i><em>${value ?? 0}</em></span>
          `)
          .join("")}
      </div>
    `;
  }

  function rankedCandidates(court, office) {
    return court
      .slice()
      .sort((a, b) => ministerFitScore(b, office.domain) - ministerFitScore(a, office.domain));
  }

  function ministerFitScore(minister, domain) {
    const base = (minister.ability || 0) + (minister.integrity || 0) / 2 - (minister.stress || 0) / 3;
    const roleBonus = {
      economy: minister.role === "户部尚书" ? 25 : 0,
      military: minister.role === "大将军" ? 25 : 0,
      diplomacy: minister.role === "长公主" ? 22 : 0,
      reform: minister.role === "太傅" ? 20 : 0,
      intrigue: minister.trait === "刚正" ? 12 : 0,
      court: minister.role === "长公主" ? 18 : 0,
    };
    return base + (roleBonus[domain] || 0);
  }

  function portraitByKey(key, api, fallback) {
    const indices = api.portraitIndexByRole || {};
    const role = key === "princess" ? "consort" : key || fallback;
    const index = indices[role] ?? indices[fallback] ?? 0;
    return typeof api.portraitAt === "function" ? api.portraitAt(index) : "";
  }

  function motherName(game, motherID) {
    const mother = (game.harem || []).find((consort) => consort.id === motherID);
    return mother ? mother.name : "未知";
  }

  function childCount(consort) {
    return Array.isArray(consort.children) ? consort.children.length : 0;
  }

  function domainLabel(domain) {
    const labels = {
      domestic: "民政",
      economy: "财政",
      military: "军务",
      diplomacy: "外交",
      reform: "新法",
      intrigue: "暗线",
      court: "宫廷",
    };
    return labels[domain] || "朝政";
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

  window.renderSystemPanels = renderSystemPanels;
})();
