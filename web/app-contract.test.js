const assert = require("assert");
const fs = require("fs");
const path = require("path");

const root = __dirname;
const html = fs.readFileSync(path.join(root, "index.html"), "utf8");
const app = fs.readFileSync(path.join(root, "app.js"), "utf8");

assert.match(html, /id="event-hand-panel"/);
assert.match(html, /src="\/event-hand-ui.js"/);
assert.match(html, /src="\/panel-renderers.js"/);
assert.match(html, /href="\/gameplay.css"/);
assert.match(app, /eventHand:\s*document\.querySelector\("#event-hand-panel"\)/);
assert.match(app, /eventHand:\s*game\.eventHand \|\| \[\]/);
assert.match(app, /function issueAction/);
assert.match(app, /\/actions`/);
assert.match(app, /data-action-kind/);
assert.match(app, /renderExternalPanelsIfReady/);
