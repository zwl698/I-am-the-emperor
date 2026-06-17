const assert = require("assert");
const fs = require("fs");
const path = require("path");
const vm = require("vm");

const root = __dirname;
const context = { console, window: {} };
vm.createContext(context);
vm.runInContext(fs.readFileSync(path.join(root, "strategy-map-ui.js"), "utf8"), context, { filename: "strategy-map-ui.js" });

const target = { innerHTML: "" };
context.window.renderStrategyMap(
  {
    phase: "emperor",
    command: 3,
    strategy: {
      cities: [
        { id: "capital", name: "京畿", ownerId: "court", x: 50, y: 45, troops: 18000, grain: 80, defense: 70, order: 66, disaster: 10, tags: ["都城"] },
        { id: "north", name: "北境", ownerId: "court", x: 50, y: 20, troops: 15000, grain: 45, defense: 74, order: 52, disaster: 30, front: true, tags: ["边塞"] },
        { id: "snow-ridge", name: "雪岭", ownerId: "beidi", x: 50, y: 8, troops: 22000, grain: 55, defense: 64, order: 54, disaster: 36, front: true, tags: ["关隘"] },
      ],
      roads: [
        { from: "capital", to: "north", terrain: "边道", risk: 28, distance: 3 },
        { from: "north", to: "snow-ridge", terrain: "雪道", risk: 48, distance: 3 },
      ],
      factions: [
        { id: "court", name: "朝廷", color: "#d7a84f", isPlayer: true },
        { id: "beidi", name: "北狄", color: "#7ea8d8" },
      ],
      armies: [
        { id: "northern-banner", name: "北府军", factionId: "court", location: "north", troops: 18000, grain: 54, morale: 66, training: 70, status: "驻防" },
        { id: "beidi-vanguard", name: "黑毡前锋", factionId: "beidi", location: "snow-ridge", troops: 22000, grain: 58, morale: 72, training: 68, status: "压境" },
      ],
      logs: [{ title: "敌军压境", summary: "黑毡前锋在北境外施压。", severity: 78 }],
      battles: [
        {
          title: "雪岭攻城战",
          cityId: "snow-ridge",
          outcome: "capture",
          attackerLoss: 3200,
          defenderLoss: 7600,
          participants: ["northern-banner", "imperial-guard"],
          summary: "北府军攻破雪岭，禁军右营侧翼支援。",
          factors: ["攻势 92000 > 守势 64000", "粮草 54", "支援军 1"],
          severity: 76,
        },
      ],
    },
  },
  target,
);

assert.match(target.innerHTML, /战略地图/);
assert.match(target.innerHTML, /京畿/);
assert.match(target.innerHTML, /北境/);
assert.match(target.innerHTML, /雪岭/);
assert.match(target.innerHTML, /strategy-terrain/);
assert.match(target.innerHTML, /map-road/);
assert.match(target.innerHTML, /map-river/);
assert.match(target.innerHTML, /strategy-army/);
assert.match(target.innerHTML, /data-army-id="northern-banner"/);
assert.match(target.innerHTML, /data-action-kind="city_develop"/);
assert.match(target.innerHTML, /data-action-mode="fortify"/);
assert.match(target.innerHTML, /data-action-kind="army_command"/);
assert.match(target.innerHTML, /data-action-mode="assault"/);
assert.match(target.innerHTML, /data-action-target="northern-banner:snow-ridge"/);
assert.match(target.innerHTML, /最近战报/);
assert.match(target.innerHTML, /雪岭攻城战/);
assert.match(target.innerHTML, /攻占/);
assert.match(target.innerHTML, /损3200/);
assert.match(target.innerHTML, /imperial-guard/);
assert.match(target.innerHTML, /攻势 92000 &gt; 守势 64000/);
assert.match(target.innerHTML, /支援军 1/);

const lockedTarget = { innerHTML: "" };
context.window.renderStrategyMap({ phase: "prince" }, lockedTarget);
assert.match(lockedTarget.innerHTML, /登基后/);
