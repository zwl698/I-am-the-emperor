(function () {
  /* ── Tier labels & colours ── */
  var tierMeta = {
    1: { label: "基础工程", color: "#c9a84d", bg: "rgba(201,168,77,0.10)" },
    2: { label: "进阶工程", color: "#a82021", bg: "rgba(168,32,33,0.10)" },
    3: { label: "鼎新工程", color: "#6a3a9e", bg: "rgba(106,58,158,0.12)" },
  };

  var domainIcons = { story: "卷", domestic: "民", economy: "财", military: "兵", diplomacy: "使", court: "宫", reform: "法", intrigue: "密" };

  function renderGrandStrategy(game, targets) {
    if (!game) return;
    renderProjectsAndPolicies(game, targets.strategy);
    renderRelations(game, targets.relations);
  }

  function renderProjectsAndPolicies(game, target) {
    if (!target) return;
    if (game.phase !== "emperor") {
      target.innerHTML = '<article class="strategy-empty">登基后开启多年工程与常驻国策。</article>';
      return;
    }
    var projects = game.projects || [];
    var policies = game.policies || [];

    var tiers = { 1: [], 2: [], 3: [] };
    projects.forEach(function (p) {
      var t = p.tier || 1;
      if (!tiers[t]) tiers[t] = [];
      tiers[t].push(p);
    });

    var projectMap = {};
    projects.forEach(function (p) { projectMap[p.id] = p; });

    var html = '<div class="reform-tree">';

    [1, 2, 3].forEach(function (tier) {
      var meta = tierMeta[tier] || tierMeta[1];
      var list = tiers[tier] || [];
      html += '<div class="reform-tier" data-tier="' + tier + '">';
      html += '<div class="reform-tier-label" style="color:' + meta.color + ';border-left-color:' + meta.color + '">' + meta.label + '</div>';
      html += '<div class="reform-tier-nodes">';

      list.forEach(function (project) {
        var stateClass = project.completed ? "project-completed" : project.locked ? "project-locked" : "project-active";
        var icon = domainIcons[project.domain] || "策";
        var pct = clampPercent(project.progress);

        // Prereq info
        var prereqHTML = "";
        if (project.prereqs && project.prereqs.length > 0) {
          var prereqNames = project.prereqs.map(function (pid) {
            var pp = projectMap[pid];
            var done = pp && pp.completed;
            return '<span class="prereq-tag ' + (done ? "prereq-met" : "prereq-unmet") + '">' + (pp ? pp.name : pid) + (done ? " ✓" : " ✗") + '</span>';
          });
          prereqHTML = '<div class="project-prereqs">前置：' + prereqNames.join(" + ") + '</div>';
        }

        // Synergy hint
        var synergyHTML = "";
        if (project.synergy) {
          synergyHTML = '<div class="project-synergy-hint" title="' + safeAttr(project.synergy) + '">协同：' + safe(project.synergy.split("：")[0] || project.synergy) + '</div>';
        }

        // Unlock info
        var unlockHTML = "";
        if (project.unlocks && project.unlocks.length > 0 && !project.completed) {
          var unlockNames = project.unlocks.map(function (uid) {
            var up = projectMap[uid];
            return up ? up.name : uid;
          });
          unlockHTML = '<div class="project-unlocks">可解锁：' + unlockNames.join("、") + '</div>';
        }

        html += '<article class="reform-node ' + stateClass + ' domain-' + project.domain + '" data-strategy-target="' + safeAttr(project.id) + '">';
        html += '<div class="reform-node-head">';
        html += '<span class="reform-node-icon">' + icon + '</span>';
        html += '<strong>' + safe(project.name) + '</strong>';
        if (project.completed) {
          html += '<span class="project-badge badge-completed">告成</span>';
        } else if (project.locked) {
          html += '<span class="project-badge badge-locked">锁定</span>';
        } else {
          html += '<span class="project-badge badge-active">' + safe(project.stage) + '</span>';
        }
        html += '</div>';

        html += '<div class="reform-node-desc">' + safe(project.description) + '</div>';

        if (!project.completed) {
          html += '<div class="reform-node-meter">';
          html += '<div class="meter-track"><i style="width:' + pct + '%"></i></div>';
          html += '<em>' + project.progress + '/100 · 风险 ' + project.risk + '</em>';
          html += '</div>';
        }

        html += '<div class="reform-node-reward">' + safe(project.reward) + '</div>';
        html += prereqHTML;
        html += synergyHTML;
        html += unlockHTML;

        if (!project.completed && !project.locked) {
          html += strategyButtons(projectOrders, project.id, game.command, false);
        }
        html += '</article>';
      });

      html += '</div></div>';

      // Connection arrow between tiers
      if (tier < 3 && tiers[tier + 1] && tiers[tier + 1].length > 0) {
        html += '<div class="reform-tier-connector" aria-hidden="true"></div>';
      }
    });

    html += '</div>';

    // Policies section
    html += '<div class="reform-policies-section">';
    html += '<div class="reform-subtitle">常驻政策</div>';
    html += '<div class="reform-policy-grid">';
    policies.forEach(function (policy) {
      html += '<article class="reform-policy-card ' + (policy.active ? "policy-active" : "") + ' domain-' + policy.domain + '">';
      html += '<div class="policy-card-head">';
      html += '<span class="reform-node-icon">' + (domainIcons[policy.domain] || "策") + '</span>';
      html += '<strong>' + safe(policy.name) + '</strong>';
      html += '<span class="policy-status ' + (policy.active ? "status-active" : "status-inactive") + '">' + (policy.active ? "施行中" : "待诏") + '</span>';
      html += '</div>';
      html += '<div class="policy-card-desc">' + safe(policy.description) + '</div>';
      html += '<div class="policy-card-cost">维护 ' + policy.upkeep + ' · 阻力 ' + policy.strain + '</div>';
      html += strategyButtons(policyOrders, policy.id, game.command, false);
      html += '</article>';
    });
    html += '</div></div>';

    target.innerHTML = html;
  }

  function renderRelations(game, target) {
    if (!target) return;
    var relations = game.relations || [];
    if (game.phase !== "emperor") {
      target.innerHTML = '<article class="strategy-empty">成长阶段的选择会影响登基后的关系底色。</article>';
      return;
    }
    target.innerHTML = relations
      .map(function (relation) {
        return '<article class="relation-row">' +
          '<strong>' + safe(relation.from) + ' ↔ ' + safe(relation.to) + '</strong>' +
          '<span>' + safe(relation.bond) + ' · ' + safe(relation.description) + '</span>' +
          '<div class="relation-bars">' +
            '<small>信 ' + relation.trust + '</small><i class="trust" style="width:' + clampPercent(relation.trust) + '%"></i>' +
            '<small>怨 ' + relation.tension + '</small><i class="tension" style="width:' + clampPercent(relation.tension) + '%"></i>' +
          '</div>' +
        '</article>';
      })
      .join("");
  }

  function strategyButtons(orders, target, command, disabled) {
    return '<div class="order-buttons">' +
      orders.map(function (order) {
        var off = disabled || command <= 0;
        var attrs = off ? 'disabled data-state-disabled="true"' : "";
        return '<button type="button" ' + attrs + ' data-order-kind="' + safeAttr(order.kind) + '" data-order-target="' + safeAttr(target) + '" data-order-label="' + safeAttr(order.title) + '" title="' + safeAttr(order.title) + '">' + safe(order.label) + '</button>';
      }).join("") +
    '</div>';
  }

  function clampPercent(value) {
    return Math.max(0, Math.min(100, Number(value) || 0));
  }

  function safe(value) {
    return String(value == null ? "" : value)
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
