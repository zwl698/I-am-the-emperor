const assert = require("assert");
const fs = require("fs");
const path = require("path");
const vm = require("vm");

const root = __dirname;
const context = { console, window: {} };
vm.createContext(context);

for (const file of ["game-data.js", "justice-ui.js"]) {
  vm.runInContext(fs.readFileSync(path.join(root, file), "utf8"), context, { filename: file });
}

const targets = {
  opinion: { innerHTML: "" },
  cases: { innerHTML: "" },
};

context.window.renderJusticePanels(
  {
    phase: "emperor",
    command: 2,
    publicOpinion: {
      popular: 58,
      elite: 47,
      rumor: 71,
      fear: 32,
      justice: 54,
      lastEdict: "三司会审将开。",
    },
    legalCases: [
      {
        id: "case-palace",
        title: "宫印误用案",
        domain: "court",
        accuser: "内廷总管",
        defendant: "失宠外戚",
        charge: "伪传懿旨",
        stakes: "后宫名分与储位流言交织。",
        heat: 66,
        evidence: 50,
        factionPressure: 72,
        publicPressure: 59,
        resolved: false,
        verdict: "",
      },
      {
        id: "case-salt",
        title: "盐引私售案",
        domain: "economy",
        accuser: "户部给事中",
        defendant: "漕运商帮",
        charge: "私卖盐引",
        stakes: "财政回血会触动商帮。",
        heat: 24,
        evidence: 68,
        factionPressure: 44,
        publicPressure: 35,
        resolved: true,
        verdict: "明正典刑",
      },
    ],
  },
  targets,
);

assert.match(targets.opinion.innerHTML, /民望|士论|谣言|畏惧|法度/);
assert.match(targets.opinion.innerHTML, /三司会审将开/);
assert.match(targets.cases.innerHTML, /宫印误用案|盐引私售案/);
assert.match(targets.cases.innerHTML, /data-order-kind="open_trial"/);
assert.match(targets.cases.innerHTML, /data-order-kind="clemency"/);
assert.match(targets.cases.innerHTML, /data-order-kind="censor_rumor"/);
assert.match(targets.cases.innerHTML, /data-order-kind="proclaim_verdict"/);
assert.match(targets.cases.innerHTML, /data-order-target="case-salt"[^>]*data-order-kind="open_trial"[^>]*disabled|data-order-kind="open_trial"[^>]*data-order-target="case-salt"[^>]*disabled|disabled[^>]*data-order-kind="open_trial"[^>]*data-order-target="case-salt"/);
assert.match(targets.cases.innerHTML, /data-order-target="case-palace"[^>]*data-order-kind="proclaim_verdict"[^>]*disabled|data-order-kind="proclaim_verdict"[^>]*data-order-target="case-palace"[^>]*disabled|disabled[^>]*data-order-kind="proclaim_verdict"[^>]*data-order-target="case-palace"/);

const lockedTargets = {
  opinion: { innerHTML: "" },
  cases: { innerHTML: "" },
};
context.window.renderJusticePanels({ phase: "prince" }, lockedTargets);
assert.match(lockedTargets.opinion.innerHTML, /登基后/);
assert.match(lockedTargets.cases.innerHTML, /登基后/);
