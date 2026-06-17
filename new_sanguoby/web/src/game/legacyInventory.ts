import type { LegacyResources } from '../api/types';

export type LegacyInventorySummary = {
  available: boolean;
  totalResources: number;
  cityNames: number;
  generalScenarios: number;
  variableStrings: number;
  battleMaps: number;
};

export function summarizeLegacyInventory(inventory: LegacyResources | null): LegacyInventorySummary {
  if (!inventory) {
    return emptySummary();
  }

  const resourceByID = new Map(inventory.resources.map((resource) => [resource.id, resource]));
  return {
    available: true,
    totalResources: inventory.count,
    cityNames: resourceByID.get(58)?.itemCount ?? 0,
    generalScenarios: resourceByID.get(61)?.itemCount ?? 0,
    variableStrings: resourceByID.get(64)?.itemCount ?? 0,
    battleMaps: inventory.resources.filter((resource) => resource.id >= 110 && resource.id <= 116).length,
  };
}

function emptySummary(): LegacyInventorySummary {
  return {
    available: false,
    totalResources: 0,
    cityNames: 0,
    generalScenarios: 0,
    variableStrings: 0,
    battleMaps: 0,
  };
}
