const assert = require("assert");
const fs = require("fs");
const path = require("path");
const vm = require("vm");

const root = __dirname;
const context = { console, window: {} };
vm.createContext(context);
vm.runInContext(fs.readFileSync(path.join(root, "talent-pool-ui.js"), "utf8"), context, { filename: "talent-pool-ui.js" });

const target = { innerHTML: "" };
context.window.renderTalentPool(
  {
    phase: "emperor",
    command: 2,
    talentPool: [
      { id: "talent-zhuge-liang-court", name: "诸葛武侯·待诏", role: "翰林待诏", trait: "谨密", specialty: "reform", origin: "中国三国", inspiration: "诸葛亮", school: "法度屯田", loyalty: 84, ability: 95, ambition: 33, integrity: 94, stress: 8 },
      { id: "talent-caesar-frontier", name: "凯撒·边策", role: "边镇参议", trait: "雄辩", specialty: "military", origin: "罗马", inspiration: "尤利乌斯·凯撒", school: "军团政治", loyalty: 42, ability: 96, ambition: 95, integrity: 36, stress: 12 },
    ],
  },
  target,
);

assert.match(target.innerHTML, /天下人才谱/);
assert.match(target.innerHTML, /候选 2/);
assert.match(target.innerHTML, /诸葛武侯/);
assert.match(target.innerHTML, /取法 诸葛亮/);
assert.match(target.innerHTML, /新法/);
assert.match(target.innerHTML, /data-order-kind="recruit_talent"/);
assert.match(target.innerHTML, /data-order-target="talent-zhuge-liang-court"/);

const locked = { innerHTML: "" };
context.window.renderTalentPool({ phase: "prince", talentPool: [] }, locked);
assert.match(locked.innerHTML, /登基后/);
