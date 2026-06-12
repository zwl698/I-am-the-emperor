const assert = require("assert");
const fs = require("fs");
const path = require("path");
const vm = require("vm");

const root = __dirname;
const context = { console, window: {} };
vm.createContext(context);

for (const file of ["game-data.js", "mini-games.js"]) {
  vm.runInContext(fs.readFileSync(path.join(root, file), "utf8"), context, { filename: file });
}

const target = { innerHTML: "" };
context.window.renderMiniGames(
  {
    phase: "emperor",
    command: 3,
    strategy: {
      cities: [
        { id: "north", name: "北境", ownerId: "court", x: 50, y: 20 },
        { id: "snow-ridge", name: "雪岭", ownerId: "beidi", x: 50, y: 8 },
      ],
      roads: [{ from: "north", to: "snow-ridge" }],
      armies: [{ id: "northern-banner", name: "北府军", factionId: "court", location: "north", troops: 18000, grain: 54, morale: 66, training: 70, status: "驻防" }],
    },
    wars: [{ id: "north", name: "雪岭攻防", enemy: "北狄", stage: "相持", threat: 72, supply: 41, morale: 56, progress: 33 }],
    legalCases: [{ id: "case-a", title: "宫印误用案", defendant: "失宠外戚", heat: 76, evidence: 52, resolved: false }],
    offices: [{ id: "censorate", title: "都察院左都御史", domain: "intrigue", holderId: "", authority: 46, vacancyRisk: 64 }],
    court: [
      { id: "gu", name: "顾衡", role: "太傅", ability: 82, integrity: 76, stress: 18 },
      { id: "huo", name: "霍骁", role: "大将军", ability: 78, integrity: 52, stress: 35 },
    ],
  },
  target,
);

assert.match(target.innerHTML, /御前操作台/);
assert.match(target.innerHTML, /兵棋沙盘/);
assert.match(target.innerHTML, /三司会审/);
assert.match(target.innerHTML, /六部调度/);
assert.match(target.innerHTML, /data-action-kind="army_command"/);
assert.match(target.innerHTML, /data-action-mode="assault"/);
assert.match(target.innerHTML, /data-action-target="northern-banner:snow-ridge"/);
assert.match(target.innerHTML, /data-action-kind="trial_move"/);
assert.match(target.innerHTML, /data-action-mode="open_trial"/);
assert.match(target.innerHTML, /data-action-kind="office_assign"/);
assert.match(target.innerHTML, /data-action-target="censorate:gu"|data-action-target="censorate:huo"/);

const lockedTarget = { innerHTML: "" };
context.window.renderMiniGames({ phase: "prince" }, lockedTarget);
assert.match(lockedTarget.innerHTML, /登基后/);
