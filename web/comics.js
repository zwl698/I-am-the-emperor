var comicBeats = [
  { id: "birth-snow", when: "scene", scene: "birth-omen", title: "雪窗啼声", caption: "一声啼哭，宫墙外的雪也停了一瞬。", sceneIndex: 0, portrait: "infant" },
  { id: "jade-rattle", when: "scene", scene: "birth-omen", title: "掌中玉铃", caption: "小小手指抓住的不是玩物，是第一道传闻。", sceneIndex: 0, portrait: "infant" },
  { id: "study-lamp", when: "scene", scene: "study-yard", title: "书院灯影", caption: "太傅合上书卷，等你给天下排第一件事。", sceneIndex: 1, portrait: "prince" },
  { id: "brothers-watch", when: "scene", scene: "study-yard", title: "同席暗潮", caption: "兄弟们低头读书，余光都落在你身上。", sceneIndex: 1, portrait: "tutor" },
  { id: "hunt-fall", when: "scene", scene: "winter-hunt", title: "惊马雪线", caption: "马蹄扬起碎雪，少年第一次看见宫斗的牙。", sceneIndex: 2, portrait: "prince" },
  { id: "hunt-banner", when: "scene", scene: "winter-hunt", title: "猎旗回卷", caption: "猎场不是猎兽，是猎名望。", sceneIndex: 2, portrait: "general" },
  { id: "river-break", when: "scene", scene: "flood-memorial", title: "南河决口", caption: "奏章上的水痕，比朱批更早抵达御案。", sceneIndex: 3, portrait: "engineer" },
  { id: "granary-key", when: "domain", domain: "domestic", title: "仓钥落掌", caption: "粮仓打开时，民心也有了喘息。", sceneIndex: 6, portrait: "farmer" },
  { id: "succession-shadow", when: "scene", scene: "succession-night", title: "烛影夺嫡", caption: "谁先进宫门，谁就可能改变天亮后的称呼。", sceneIndex: 4, portrait: "prince" },
  { id: "edict-seal", when: "domain", domain: "court", title: "玉玺压纸", caption: "制度有时比刀更冷，也更稳。", sceneIndex: 5, portrait: "emperor" },
  { id: "court-lines", when: "phase", phase: "emperor", title: "百官分班", caption: "每一列朝臣，都是一条会反咬的线。", sceneIndex: 5, portrait: "emperor" },
  { id: "tax-abacus", when: "domain", domain: "economy", title: "银珠疾响", caption: "算盘越响，地方豪强越安静，也越危险。", sceneIndex: 7, portrait: "minister" },
  { id: "salt-ledger", when: "order", kind: "tax", title: "盐引重发", caption: "账册翻新，旧账不会自己消失。", sceneIndex: 7, portrait: "merchant" },
  { id: "relief-smoke", when: "order", kind: "relief", title: "粥棚烟火", caption: "热粥升起白气，遮住灾民眼里的寒。", sceneIndex: 6, portrait: "farmer" },
  { id: "garrison-torches", when: "order", kind: "garrison", title: "烽燧重明", caption: "边墙灯火一盏盏亮回去。", sceneIndex: 8, portrait: "general" },
  { id: "canal-draft", when: "order", kind: "canal", title: "渠图铺开", caption: "一道水路，能改写几十年的税粮。", sceneIndex: 13, portrait: "engineer" },
  { id: "trade-caravan", when: "order", kind: "trade", title: "互市驼铃", caption: "货物先进关，试探随后而来。", sceneIndex: 15, portrait: "merchant" },
  { id: "spy-lantern", when: "domain", domain: "intrigue", title: "密档夜灯", caption: "灯芯剪短，名单变长。", sceneIndex: 11, portrait: "spy" },
  { id: "inspect-sleeve", when: "order", kind: "inspect", title: "袖中供词", caption: "证词被折成四方，刚好能藏进袖口。", sceneIndex: 11, portrait: "assassin" },
  { id: "appease-banquet", when: "order", kind: "appease", title: "赐宴留阶", caption: "一杯酒，有时能换半季安静。", sceneIndex: 12, portrait: "consort" },
  { id: "purge-hall", when: "order", kind: "purge", title: "廷杖回声", caption: "权势落地有声，怨气无声。", sceneIndex: 11, portrait: "scholar" },
  { id: "reform-archive", when: "domain", domain: "reform", title: "新法成册", caption: "纸上新法很轻，落到官场却重如铁。", sceneIndex: 10, portrait: "reformer" },
  { id: "envoy-pass", when: "domain", domain: "diplomacy", title: "使节出关", caption: "金册和笑容都是真的，刀也是真的。", sceneIndex: 9, portrait: "envoy" },
  { id: "war-map", when: "domain", domain: "military", title: "军图压案", caption: "朱笔沿山川推进，粮草在背后追赶。", sceneIndex: 26, portrait: "general" },
  { id: "mobilize-drum", when: "order", kind: "mobilize", title: "点将击鼓", caption: "战鼓响起之前，户部先听见银库空了一格。", sceneIndex: 21, portrait: "guard" },
  { id: "campaign-charge", when: "order", kind: "campaign", title: "出塞决战", caption: "雪线尽头，骑兵像墨点冲进风里。", sceneIndex: 14, portrait: "general" },
  { id: "fortify-wall", when: "order", kind: "fortify", title: "堡垒连星", caption: "一座座土堡，把恐惧钉进地里。", sceneIndex: 8, portrait: "engineer" },
  { id: "truce-tent", when: "order", kind: "truce", title: "帐中议和", caption: "边风吹动盟书，也吹动武臣的眉头。", sceneIndex: 28, portrait: "diplomat" },
  { id: "office-seal", when: "order", kind: "appoint", title: "官印换手", caption: "一枚官印落下，半个朝堂都要重新站队。", sceneIndex: 19, portrait: "minister" },
  { id: "dismiss-tablet", when: "order", kind: "dismiss", title: "牙牌撤名", caption: "名字从班簿上消失，比廷杖更安静。", sceneIndex: 5, portrait: "scholar" },
  { id: "heir-canon", when: "order", kind: "name_heir", title: "册文入庙", caption: "储君二字写得很端正，宫墙里的心却未必。", sceneIndex: 18, portrait: "prince" },
  { id: "harem-lantern", when: "order", kind: "favor_consort", title: "宫灯偏照", caption: "今夜哪座宫灯更亮，明日哪家外戚就更近。", sceneIndex: 16, portrait: "consort" },
  { id: "marriage-jade", when: "order", kind: "marriage_alliance", title: "玉册联姻", caption: "红绸连起两族，也把筹码系上龙案。", sceneIndex: 12, portrait: "empress" },
  { id: "court-office-draft", when: "domain", domain: "court", title: "差遣重排", caption: "吏部纸面上几行小字，能让朝局换一个重心。", sceneIndex: 19, portrait: "tutor" },
  { id: "event-edict-stack", when: "eventCategory", category: "system_pressure", title: "突发奏报", caption: "平静不是默认状态，只是奏章还没递到殿前。", sceneIndex: 5, portrait: "minister" },
  { id: "event-micro-check", when: "eventCategory", category: "micro_game", title: "御前检定", caption: "这一刻，不是选项在判定你，是此前所有经营在判定你。", sceneIndex: 10, portrait: "reformer" },
  { id: "event-heir-rumor", when: "eventTag", tag: "继承", title: "东宫传闻", caption: "储位的风声，常常比正式册文更快。", sceneIndex: 18, portrait: "prince" },
  { id: "event-office-risk", when: "eventTag", tag: "官职", title: "官署空转", caption: "椅子空着的时候，权力不会空着。", sceneIndex: 19, portrait: "minister" },
  { id: "event-market-freeze", when: "eventTag", tag: "财政", title: "银荒入市", caption: "钱铺关门的声音，能传得比钟声更远。", sceneIndex: 23, portrait: "merchant" },
  { id: "event-war-fog", when: "eventTag", tag: "战争", title: "战雾入京", caption: "远方的马蹄，最后会踩在朝堂的沉默里。", sceneIndex: 14, portrait: "general" },
  { id: "war-pressure", when: "war", title: "敌骑压境", caption: "战线离京畿还远，恐惧已经进城。", sceneIndex: 14, portrait: "khan" },
  { id: "war-supply", when: "warSupply", title: "粮道断续", caption: "军队最怕的不是刀，是空锅。", sceneIndex: 20, portrait: "minister" },
  { id: "court-stress", when: "courtStress", title: "朝臣倦色", caption: "能臣也会疲惫，疲惫会长出野心。", sceneIndex: 19, portrait: "tutor" },
  { id: "low-loyalty", when: "lowLoyalty", title: "班列侧目", caption: "忠诚低到某个位置，沉默就是奏章。", sceneIndex: 5, portrait: "scholar" },
  { id: "crisis-red", when: "crisis", title: "危机红线", caption: "危机钟不是钟，是倒数的刀。", sceneIndex: 22, portrait: "assassin" },
  { id: "temple-oath", when: "stability", title: "太庙告成", caption: "祖宗牌位前，盛世也要低声说话。", sceneIndex: 18, portrait: "emperor" },
  { id: "exam-talent", when: "reformHigh", title: "贡院开榜", caption: "新法若能养出新人，旧党就不再是唯一答案。", sceneIndex: 25, portrait: "poet" },
  { id: "festival-golden", when: "golden", title: "万邦烟火", caption: "灯火连成河，史官终于舍得用盛世二字。", sceneIndex: 29, portrait: "elder" },
];

function dynamicComicPanels(game, action) {
  const beat = selectComicBeat(game, action);
  const dynasty = currentDynasty();
  return [
    {
      className: "scene-panel",
      image: game.scene?.art || sceneAt(beat.sceneIndex),
      title: dynasty.name,
      caption: game.scene?.mood ? `${game.scene.year} · ${game.scene.mood}` : dynasty.challenge,
    },
    {
      className: "character-panel",
      image: portraitAt(portraitIndexByRole[beat.portrait] ?? portraitIndexByRole.emperor),
      title: beat.title,
      caption: beat.caption,
    },
    {
      className: `action-panel domain-${action?.domain || beat.domain || "court"}`,
      image: sceneAt(beat.sceneIndex),
      title: action?.title || beat.title,
      caption: action?.summary || beat.caption,
    },
    {
      className: "crisis-panel",
      image: crisisSceneArt(),
      title: game.crisis?.title || "朝局",
      caption: game.crisis ? `烈度 ${game.crisis.severity} · 危机钟 ${game.crisis.clock}/8` : "风暴尚未命名。",
    },
  ];
}

function selectComicBeat(game, action) {
  if (action?.kind) {
    const orderBeats = comicBeats.filter((beat) => beat.when === "order" && beat.kind === action.kind);
    if (orderBeats.length > 0) return orderBeats[game.turn % orderBeats.length];
  }
  if (action?.domain) {
    const domainBeats = comicBeats.filter((beat) => beat.when === "domain" && beat.domain === action.domain);
    if (domainBeats.length > 0) return domainBeats[game.turn % domainBeats.length];
  }
  const candidates = comicBeats.filter((beat) => comicBeatMatches(beat, game, action));
  if (candidates.length > 0) {
    return candidates[(game.turn + (action?.summary?.length || 0)) % candidates.length];
  }
  return comicBeats[game.turn % comicBeats.length];
}

function comicBeatMatches(beat, game, action) {
  if (beat.when === "scene") return game.scene?.id === beat.scene;
  if (beat.when === "phase") return game.phase === beat.phase;
  if (beat.when === "domain") return action?.domain === beat.domain || game.scene?.choices?.some((choice) => choice.domain === beat.domain);
  if (beat.when === "order") return action?.kind === beat.kind;
  if (beat.when === "eventCategory") return (game.recentEvents || []).some((event) => event.category === beat.category);
  if (beat.when === "eventTag") return (game.recentEvents || []).some((event) => (event.tags || []).includes(beat.tag));
  if (beat.when === "war") return (game.wars || []).some((war) => war.threat >= 70);
  if (beat.when === "warSupply") return (game.wars || []).some((war) => war.supply <= 30);
  if (beat.when === "courtStress") return (game.court || []).some((minister) => minister.stress >= 70);
  if (beat.when === "lowLoyalty") return (game.court || []).some((minister) => minister.loyalty <= 35);
  if (beat.when === "crisis") return (game.crisis?.severity || 0) >= 70 || (game.crisis?.clock || 0) >= 6;
  if (beat.when === "stability") return (game.stats?.stability || 0) >= 80;
  if (beat.when === "reformHigh") return (game.stats?.reform || 0) >= 60;
  if (beat.when === "golden") return game.ending?.kind === "golden_age";
  return false;
}

if (typeof window !== "undefined") {
  window.comicBeats = comicBeats;
  window.dynamicComicPanels = dynamicComicPanels;
}
