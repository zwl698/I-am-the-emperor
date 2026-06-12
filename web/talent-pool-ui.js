(function () {
  function renderTalentPool(game, target) {
    if (!target) return;
    if (!game || game.phase !== "emperor") {
      target.innerHTML = `<article class="talent-empty">登基后开启天下人才谱，可征辟各地名士、归化客卿与军政奇才。</article>`;
      return;
    }
    const pool = Array.isArray(game.talentPool) ? game.talentPool : [];
    if (!pool.length) {
      target.innerHTML = `<article class="talent-empty">天下人才已尽入朝堂，剩下的是如何驾驭他们。</article>`;
      return;
    }
    const featured = pool.slice().sort((a, b) => talentScore(b, game) - talentScore(a, game)).slice(0, 8);
    target.innerHTML = `
      <section class="talent-pool">
        <div class="talent-head">
          <strong>天下人才谱</strong>
          <small>候选 ${pool.length} · 优先显示契合当前危机者</small>
        </div>
        ${featured.map((talent) => talentRow(talent, game)).join("")}
      </section>
    `;
  }

  function talentRow(talent, game) {
    const disabled = (game.command ?? 0) <= 0 ? `disabled data-state-disabled="true"` : "";
    return `
      <article class="talent-row domain-${safeAttr(talent.specialty || "court")}">
        <div>
          <strong>${safe(talent.name)}</strong>
          <small>${safe(talent.role)} · ${safe(talent.trait)} · ${domainLabel(talent.specialty)}</small>
          <em>${safe(talent.origin)} · ${safe(talent.school)} · 取法 ${safe(talent.inspiration)}</em>
          <span>忠${stat(talent.loyalty)} · 才${stat(talent.ability)} · 野${stat(talent.ambition)} · 廉${stat(talent.integrity)}</span>
        </div>
        <button type="button" ${disabled} data-order-kind="recruit_talent" data-order-target="${safeAttr(talent.id)}" data-order-label="征辟${safeAttr(talent.name)}" title="征辟${safeAttr(talent.name)}入朝">征</button>
      </article>
    `;
  }

  function talentScore(talent, game) {
    const crisis = game.crisis || {};
    const stats = game.stats || {};
    let score = stat(talent.ability) * 2 + stat(talent.integrity) - stat(talent.ambition) / 3 - stat(talent.stress);
    if (stats.borderThreat >= 65 && talent.specialty === "military") score += 40;
    if (stats.treasury <= 45 && talent.specialty === "economy") score += 35;
    if (stats.populace <= 45 && talent.specialty === "domestic") score += 35;
    if (stats.reform <= 35 && talent.specialty === "reform") score += 28;
    if (crisis.severity >= 65 && (talent.specialty === "intrigue" || talent.specialty === "court")) score += 24;
    return score;
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

  function stat(value) {
    return Math.max(0, Number(value) || 0);
  }

  function safe(value) {
    return String(value ?? "").replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;");
  }

  function safeAttr(value) {
    return safe(value).replaceAll("'", "&#39;");
  }

  window.renderTalentPool = renderTalentPool;
})();
