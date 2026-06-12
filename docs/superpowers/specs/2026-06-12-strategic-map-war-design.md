# Strategic Map War Design

## Goal

将《我是皇帝》的登基后玩法升级为“三国志式战略层”：对外战争进入战略地图/战役层，玩家围绕城池、道路、军团、粮道、将领和敌对势力做决策，而不是只在选择题或单条战役进度条里处理战争。

## Player Fantasy

玩家是皇帝，不是普通诸侯。玩家的核心权力是调度全国资源：任命太守、派遣将领、调粮调兵、筑城赈灾、外交施压、压制叛乱。每道御令都会在地图上留下后果：边镇坐大、粮道吃紧、城市归属变化、敌方势力推进、民心和朝局波动。

## Core Loop

登基后每季按以下顺序推进：

1. 城池根据农业、商业、秩序、灾害产出粮草和税银。
2. 军团消耗粮草；缺粮会掉士气、损兵。
3. 敌对势力根据战略倾向、兵力、威胁和道路连接推进。
4. 行军、围城、攻城、招降等军事状态结算。
5. 事件手牌根据真实地图压力发牌，例如敌军压境、粮道断续、城池失守。
6. 玩家用御令执行城池经营、军团命令、攻城策略、任命太守、外交施压。
7. 大朝会选择推进一季，进入下一轮压力。

## Strategic Entities

### StrategicCity

城池节点是地图的基本单位。

- `id`: 稳定标识。
- `name`: 城池名。
- `region`: 所属区域。
- `ownerId`: 当前归属势力。
- `governorId`: 太守或镇守官员。
- `x`, `y`: 前端地图坐标百分比。
- `population`: 人口。
- `commerce`: 商业。
- `agriculture`: 农业。
- `defense`: 城防。
- `order`: 治安。
- `disaster`: 灾害。
- `troops`: 城防兵。
- `grain`: 城内粮草。
- `gold`: 城内府库。
- `front`: 是否处在前线。
- `tags`: 漕运、边塞、都城、关隘、海贸等标签。

### StrategicRoad

道路决定军团能否移动和敌方能否推进。

- `from`, `to`: 两端城池。
- `terrain`: 平原、山道、雪岭、漕河、海路等。
- `risk`: 粮道和伏击风险。
- `distance`: 行军消耗。

### StrategicFaction

战略势力既包括朝廷，也包括外敌和叛乱。

- 朝廷：玩家势力。
- 北狄诸部：骑兵侵攻，雪岭和北境压力最高。
- 旧朝残部：西陲和河东骚扰。
- 流寇叛军：灾害和低治安城市会滋生。
- 南岭盟寨：可战可和，重视外交和互市。
- 东海诸岛：贸易与海防压力。

字段包含 `relation`, `threat`, `strategy`, `capitalCityId`, `color`, `isPlayer`。

### ArmyGroup

军团代表地图上的军事力量。

- `id`, `name`, `factionId`
- `location`: 当前所在城。
- `target`: 目标城。
- `commanderId`: 主将。
- `troops`, `grain`, `morale`, `training`
- `siege`: 围城进度。
- `status`: 驻防、行军、围城、整训、溃退。

## Player Actions

### City Develop

`city_develop`

- `farm`: 垦田，提升农业和粮草。
- `market`: 修市，提升商业和府库。
- `fortify`: 筑城，提升城防和前线韧性。
- `relief`: 赈灾，降低灾害，提升民心。
- `patrol`: 巡查，提升治安，压低叛乱。
- `levy`: 征兵，增加城防兵，降低治安和民心。

### Army Command

`army_command`

- `train`: 整训，提升士气和训练。
- `supply`: 调粮，增加军团粮草。
- `march`: 沿道路移动到相邻城。
- `assault`: 攻击相邻敌城，按兵力、士气、训练、主将能力、城防结算。
- `recruit`: 在己方城池征募城防兵。

### Siege Command

`siege_command`

- `besiege`: 围城，降低敌城粮草和治安，积累围城进度。
- `cut_supply`: 断粮，增加道路风险和敌军缺粮。
- `persuade`: 招降，依赖外交、魅力、敌城治安和守军士气。

### Governor Assign

`governor_assign`

- `appoint`: 任命太守或镇守官。官员能力影响城市增长，忠诚和野心影响风险。

## Foreign War Entry Rule

任何对外战争相关入口都进入战略地图层：

- 事件手牌中的“对外战争”“外交诸邦”“边境勒书”会指向地图上的敌军或前线城。
- 原“兵棋沙盘”按钮保留，但它操作 `ArmyGroup` 和 `StrategicCity`，不再只改 `WarCampaign.Progress`。
- 大朝会的军事选择可以产生地图压力或给军团加成，但真正战果由地图结算。

## Frontend Design

新增战略地图面板：

- 中央地图显示城池节点和道路。
- 城池颜色代表归属势力。
- 城池徽章显示兵、粮、防、民心或危机状态。
- 军团以棋子形式显示在城池上。
- 点击城池显示经营动作。
- 点击军团显示行军、整训、攻城动作。
- 外战事件牌点击后滚动/聚焦到对应前线。

现阶段使用 DOM + CSS 实现地图，避免引入大型引擎。后续若做更强交互，可将地图 playfield 迁到 Canvas/Phaser，但模拟状态仍归 Go 后端。

## Testing Strategy

后端 TDD：

- 地图初始化至少 12 城、14 路、5 势力、4 军团。
- 每个朝代的前线和敌方压力不同。
- 城市经营动作消耗御令并改变城市数值。
- 军团行军只能沿道路移动。
- 攻城能根据战力占领敌城或造成损失。
- 季度 AI 会推进敌方威胁、移动军团或压迫前线。

前端 TDD：

- 战略地图能渲染城池、道路、军团。
- 城市动作按钮带 `data-action-kind="city_develop"`。
- 军团动作按钮带 `data-action-kind="army_command"`。
- 外战事件牌能产生进入地图的动作目标。

## Implementation Boundaries

- 每个新文件低于 1000 行。
- 不替换现有事件牌系统，而是让事件牌指向地图。
- 不删除旧 `WarCampaign`，先做兼容：地图战果会同步影响 `WarCampaign` 和 `BorderThreat`。
- 不引入网络依赖或新前端构建工具。
