export type CityState = 'normal' | 'famine';

export type GameDate = {
  year: number;
  month: number;
};

export type Ruler = {
  id: string;
  name: string;
  character: string;
  color: string;
};

export type City = {
  id: string;
  name: string;
  x: number;
  y: number;
  ownerId: string;
  state: CityState;
  farmingLimit: number;
  farming: number;
  commerceLimit: number;
  commerce: number;
  peopleDevotion: number;
  avoidCalamity: number;
  populationLimit: number;
  population: number;
  money: number;
  food: number;
  garrison: number;
};

export type General = {
  id: string;
  name: string;
  ownerId: string;
  cityId: string;
  level: number;
  force: number;
  intellect: number;
  loyalty: number;
  stamina: number;
  soldiers: number;
  armsType: string;
  captive?: boolean;
};

export type Route = {
  from: string;
  to: string;
};

export type GameSnapshot = {
  scenarioId: string;
  playerId: string;
  date: GameDate;
  rulers: Ruler[];
  cities: City[];
  generals: General[];
  routes: Route[];
  log: string[];
};

export type RulerOption = {
  id: string;
  name: string;
  character: string;
  color: string;
  cityCount: number;
};

export type ScenarioOption = {
  id: string;
  period: number;
  name: string;
  year: number;
  rulers: RulerOption[];
  cityMax: number;
};

export type ScenarioList = {
  scenarios: ScenarioOption[];
};

export type CreateGameRequest = {
  scenarioId: string;
  playerId: string;
};

export type CommandRequest = {
  cityId: string;
  generalId: string;
  commandId: string;
  targetCityId?: string;
  targetGeneralId?: string;
};

export type BattleRequest = {
  cityId: string;
  generalId: string;
  targetCityId: string;
};

export type BattleOutcome = {
  won: boolean;
  fromCityId: string;
  fromCityName: string;
  targetCityId: string;
  targetCityName: string;
  generalId: string;
  generalName: string;
  attackerRulerId: string;
  attackerRulerName: string;
  defenderRulerId: string;
  defenderRulerName: string;
  defenderGenerals?: string[];
  attackPower: number;
  defensePower: number;
  attackerLosses: number;
  defenderLosses: number;
  captured: boolean;
  capturedGenerals?: string[];
  message: string;
};

export type BattleResponse = {
  outcome: BattleOutcome;
  snapshot: GameSnapshot;
};

export type LegacyResourceHeader = {
  address: number;
  length: number;
  id: number;
  itemCount: number;
  itemLength: number;
  key: number;
  reserved: number;
};

export type LegacyResources = {
  source: string;
  count: number;
  resources: LegacyResourceHeader[];
};
