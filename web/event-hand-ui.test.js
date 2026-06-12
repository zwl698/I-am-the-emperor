const assert = require("assert");
const fs = require("fs");
const path = require("path");
const vm = require("vm");

const root = __dirname;
const context = { console, window: {} };
vm.createContext(context);
vm.runInContext(fs.readFileSync(path.join(root, "event-hand-ui.js"), "utf8"), context, { filename: "event-hand-ui.js" });

const target = { innerHTML: "" };
context.window.renderEventHand(
  {
    phase: "emperor",
    command: 3,
    eventHand: [
      {
        id: "domestic-1",
        title: "河堤决口",
        category: "内政灾害",
        domain: "domestic",
        stage: "急奏",
        summary: "江南堤岸崩坏，灾民沿官道北上。",
        hook: "工部请银，户部称仓廪已紧。",
        consequence: "拖延会推高灾害和民怨。",
        severity: 78,
        urgency: 84,
      },
      {
        id: "war-1",
        title: "雪夜奇袭",
        category: "对外战争",
        domain: "military",
        stage: "边报",
        summary: "北狄前锋试探烽燧。",
        hook: "大将军请准出塞反击。",
        consequence: "败则军心动摇，胜则边患大降。",
        severity: 72,
        urgency: 69,
      },
      {
        id: "heir-1",
        title: "太子伴读",
        category: "继承东宫",
        domain: "court",
        stage: "宫闱",
        summary: "东宫伴读牵出母族押注。",
        hook: "太傅建议加开经筵。",
        consequence: "储君成长会改变群臣站队。",
        severity: 58,
        urgency: 61,
      },
    ],
    provinces: [{ id: "south", name: "江南", disaster: 72, order: 41 }],
    wars: [{ id: "north", name: "雪岭攻防", threat: 77, progress: 31 }],
    heirs: [{ id: "crown", name: "萧承曜", talent: 64, legitimacy: 70 }],
  },
  target,
);

assert.match(target.innerHTML, /事件手牌/);
assert.match(target.innerHTML, /河堤决口/);
assert.match(target.innerHTML, /雪夜奇袭/);
assert.match(target.innerHTML, /太子伴读/);
assert.match(target.innerHTML, /data-action-kind="map_allocation"/);
assert.match(target.innerHTML, /data-action-mode="relief"/);
assert.match(target.innerHTML, /data-action-kind="war_tactic"/);
assert.match(target.innerHTML, /data-action-mode="campaign"/);
assert.match(target.innerHTML, /data-action-kind="heir_lesson"/);
assert.match(target.innerHTML, /data-action-mode="study"/);

const lockedTarget = { innerHTML: "" };
context.window.renderEventHand({ phase: "prince" }, lockedTarget);
assert.match(lockedTarget.innerHTML, /登基后/);
