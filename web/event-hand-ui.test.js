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
    strategy: {
      cities: [
        { id: "south", name: "江南", ownerId: "court", x: 58, y: 68, disaster: 72, order: 41 },
        { id: "north", name: "北境", ownerId: "court", x: 50, y: 20 },
        { id: "snow-ridge", name: "雪岭", ownerId: "beidi", x: 50, y: 8 },
      ],
      roads: [{ from: "north", to: "snow-ridge" }],
      armies: [{ id: "northern-banner", name: "北府军", factionId: "court", location: "north" }],
    },
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
assert.match(target.innerHTML, /data-action-kind="city_develop"/);
assert.match(target.innerHTML, /data-action-mode="relief"/);
assert.match(target.innerHTML, /data-action-kind="army_command"/);
assert.match(target.innerHTML, /data-action-mode="assault"/);
assert.match(target.innerHTML, /data-action-target="northern-banner:snow-ridge"/);
assert.match(target.innerHTML, /data-focus-panel="strategy-map-panel"/);
assert.match(target.innerHTML, /data-action-kind="heir_lesson"/);
assert.match(target.innerHTML, /data-action-mode="study"/);

const lowSupplyTarget = { innerHTML: "" };
context.window.renderEventHand(
  {
    phase: "emperor",
    command: 2,
    eventHand: [
      {
        id: "war-low-supply",
        title: "粮道断续",
        category: "对外战争",
        domain: "military",
        summary: "北府军粮车迟迟不至。",
        severity: 80,
        urgency: 90,
      },
    ],
    strategy: {
      cities: [
        { id: "north", name: "北境", ownerId: "court" },
        { id: "snow-ridge", name: "雪岭", ownerId: "beidi" },
      ],
      roads: [{ from: "north", to: "snow-ridge" }],
      armies: [{ id: "northern-banner", name: "北府军", factionId: "court", location: "north", grain: 5, troops: 18000, morale: 66, training: 70 }],
    },
  },
  lowSupplyTarget,
);
assert.match(lowSupplyTarget.innerHTML, /data-action-kind="army_command"/);
assert.match(lowSupplyTarget.innerHTML, /data-action-mode="supply"/);
assert.match(lowSupplyTarget.innerHTML, /data-action-target="northern-banner"/);

const urgentArmyTarget = { innerHTML: "" };
context.window.renderEventHand(
  {
    phase: "emperor",
    command: 2,
    eventHand: [
      {
        id: "war-urgent-army",
        title: "决战请命",
        category: "对外战争",
        domain: "military",
        summary: "北境主将请准攻城。",
        severity: 74,
        urgency: 82,
      },
    ],
    strategy: {
      cities: [
        { id: "capital", name: "京畿", ownerId: "court" },
        { id: "north", name: "北境", ownerId: "court", front: true },
        { id: "snow-ridge", name: "雪岭", ownerId: "beidi" },
      ],
      roads: [
        { from: "capital", to: "north" },
        { from: "north", to: "snow-ridge" },
      ],
      armies: [
        { id: "imperial-guard", name: "禁军右营", factionId: "court", location: "capital", grain: 70, troops: 16000, morale: 68, training: 62 },
        { id: "northern-banner", name: "北府军", factionId: "court", location: "north", grain: 54, troops: 18000, morale: 66, training: 70 },
      ],
    },
  },
  urgentArmyTarget,
);
assert.match(urgentArmyTarget.innerHTML, /data-action-mode="assault"/);
assert.match(urgentArmyTarget.innerHTML, /data-action-target="northern-banner:snow-ridge"/);

const lockedTarget = { innerHTML: "" };
context.window.renderEventHand({ phase: "prince" }, lockedTarget);
assert.match(lockedTarget.innerHTML, /登基后/);
