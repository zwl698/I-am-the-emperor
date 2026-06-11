(function () {
  function renderSeasonEvents(game, target, api = {}) {
    if (!target) return;
    const events = game?.recentEvents || [];
    if (game?.phase !== "emperor") {
      target.innerHTML = `<article class="event-empty">皇子阶段以主线剧情为主；登基后每季会出现由局势生成的随机奏报。</article>`;
      return;
    }
    if (events.length === 0) {
      target.innerHTML = `<article class="event-empty">本季尚无突发奏报。推进一次朝会大议题后，局势会自行发酵。</article>`;
      return;
    }
    target.innerHTML = events
      .map((event) => {
        const portrait = portraitFor(event.portrait, api);
        const check = event.category === "micro_game" ? checkLine(event) : "";
        return `
          <article class="season-event domain-${event.domain}">
            <span class="portrait-dot event-portrait" style="background-image:url('${portrait}')"></span>
            <span class="event-main">
              <strong>${safe(event.title)} <em>${categoryLabel(event.category)}</em></strong>
              <small>${safe(event.detail)}</small>
              <p>${safe(event.summary)}</p>
              ${check}
              <b>${formatEventEffects(event.effects)}</b>
              <i>${(event.tags || []).map((tag) => `<u>${safe(tag)}</u>`).join("")}</i>
            </span>
          </article>
        `;
      })
      .join("");
  }

  function checkLine(event) {
    const ok = event.success ? "通过" : "失败";
    return `<span class="event-check ${event.success ? "success" : "fail"}">${safe(event.check)} · ${event.roll}/${event.target} · ${ok}</span>`;
  }

  function categoryLabel(category) {
    const labels = {
      story_arc: "剧情",
      system_pressure: "系统",
      micro_game: "检定",
    };
    return labels[category] || "奏报";
  }

  function formatEventEffects(effects = {}) {
    if (typeof window.formatEffects === "function") return window.formatEffects(effects);
    return Object.entries(effects)
      .filter(([, value]) => value)
      .map(([key, value]) => `${key}${value > 0 ? "+" : ""}${value}`)
      .join("、");
  }

  function portraitFor(key, api) {
    const map = api.portraitIndexByRole || {};
    const index = map[key] ?? map.emperor ?? 0;
    return typeof api.portraitAt === "function" ? api.portraitAt(index) : "";
  }

  function safe(value) {
    return String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;");
  }

  window.renderSeasonEvents = renderSeasonEvents;
})();
