const assert = require("assert");
const fs = require("fs");
const path = require("path");
const vm = require("vm");

const root = __dirname;
const context = {
  console,
  window: {},
};
vm.createContext(context);

for (const file of ["game-data.js", "diplomacy-ui.js"]) {
  const source = fs.readFileSync(path.join(root, file), "utf8");
  vm.runInContext(source, context, { filename: file });
}

const targets = {
  foreign: { innerHTML: "" },
  plots: { innerHTML: "" },
};

context.window.renderDiplomacyIntrigue(
  {
    phase: "emperor",
    command: 2,
    foreignStates: [
      {
        id: "xiyu",
        name: "西域诸国",
        ruler: "龟兹王女",
        attitude: "礼厚可交",
        relation: 62,
        threat: 31,
        tribute: 38,
        leverage: 50,
        treaty: "",
        envoy: "胡商译官",
        portrait: "diplomat",
      },
      {
        id: "beidi",
        name: "北狄诸部",
        ruler: "阿史那乌勒",
        attitude: "磨刀观望",
        relation: 33,
        threat: 81,
        tribute: 15,
        leverage: 18,
        treaty: "",
        envoy: "黑毡使",
        portrait: "khan",
      },
    ],
    plots: [
      {
        id: "palace-poison",
        title: "宫酒疑云",
        sponsor: "失宠外戚",
        target: "东宫",
        stage: "暴露",
        summary: "内廷酒食采买中多出陌生印记。",
        secrecy: 28,
        progress: 58,
        danger: 69,
        exposed: true,
        resolved: false,
      },
      {
        id: "silk-ledger",
        title: "丝账暗线",
        sponsor: "漕运商帮",
        target: "户部",
        stage: "潜伏",
        summary: "商帮用旧账牵住几名户部郎官。",
        secrecy: 60,
        progress: 34,
        danger: 44,
        exposed: false,
        resolved: false,
      },
    ],
  },
  targets,
  {
    portraitAt: (index) => `/portrait-${index}.png`,
    portraitIndexByRole: { diplomat: 21, khan: 19 },
  },
);

assert.match(targets.foreign.innerHTML, /外邦诸国|西域诸国|北狄诸部/);
assert.match(targets.foreign.innerHTML, /data-order-kind="embassy"/);
assert.match(targets.foreign.innerHTML, /data-order-kind="treaty"/);
assert.match(targets.foreign.innerHTML, /data-order-target="xiyu"/);
assert.match(targets.foreign.innerHTML, /data-order-target="beidi"[^>]*disabled|disabled[^>]*data-order-target="beidi"/);

assert.match(targets.plots.innerHTML, /宫酒疑云|丝账暗线/);
assert.match(targets.plots.innerHTML, /data-order-kind="investigate_plot"/);
assert.match(targets.plots.innerHTML, /data-order-kind="suppress_plot"/);
assert.match(targets.plots.innerHTML, /data-order-target="palace-poison"/);
assert.match(targets.plots.innerHTML, /data-order-target="silk-ledger"[^>]*disabled|disabled[^>]*data-order-target="silk-ledger"/);

const lockedTargets = {
  foreign: { innerHTML: "" },
  plots: { innerHTML: "" },
};
context.window.renderDiplomacyIntrigue({ phase: "prince" }, lockedTargets);
assert.match(lockedTargets.foreign.innerHTML, /登基后/);
assert.match(lockedTargets.plots.innerHTML, /登基后/);
