import type { LegacyResources } from '../api/types';

export type LegacyImageResource = {
  id: number;
  label: string;
  group: 'battle' | 'campaign' | 'portrait' | 'ui';
  itemCount: number;
  itemLength: number;
};

export type LegacyInventorySummary = {
  available: boolean;
  totalResources: number;
  cityNames: number;
  generalScenarios: number;
  variableStrings: number;
  battleMaps: number;
  knownImageGroups: number;
  presentImageGroups: number;
  imageResourceItems: number;
  imageResources: LegacyImageResource[];
};

const LEGACY_IMAGE_RESOURCES: Array<Omit<LegacyImageResource, 'itemCount' | 'itemLength'>> = [
  { id: 5, label: '兵种', group: 'battle' },
  { id: 7, label: '脚步', group: 'battle' },
  { id: 8, label: '状态', group: 'battle' },
  { id: 9, label: '天气', group: 'battle' },
  { id: 15, label: '数字', group: 'ui' },
  { id: 16, label: '特效', group: 'battle' },
  { id: 26, label: '消息框', group: 'ui' },
  { id: 33, label: '天数', group: 'ui' },
  { id: 34, label: '禁咒', group: 'battle' },
  { id: 44, label: '主菜单', group: 'ui' },
  { id: 45, label: '时期', group: 'ui' },
  { id: 46, label: '存档', group: 'ui' },
  { id: 47, label: '城市', group: 'campaign' },
  { id: 48, label: '头像一', group: 'portrait' },
  { id: 49, label: '头像二', group: 'portrait' },
  { id: 50, label: '头像三', group: 'portrait' },
  { id: 51, label: '头像四', group: 'portrait' },
  { id: 54, label: '城图块', group: 'campaign' },
  { id: 69, label: '地图图标', group: 'campaign' },
  { id: 104, label: '年限一', group: 'ui' },
  { id: 105, label: '年限二', group: 'ui' },
  { id: 106, label: '年限三', group: 'ui' },
  { id: 107, label: '年限四', group: 'ui' },
  { id: 110, label: '战场一', group: 'battle' },
  { id: 111, label: '战场二', group: 'battle' },
  { id: 112, label: '战场三', group: 'battle' },
  { id: 113, label: '战场四', group: 'battle' },
  { id: 114, label: '战场五', group: 'battle' },
  { id: 115, label: '战场六', group: 'battle' },
  { id: 116, label: '战场七', group: 'battle' },
];

export function summarizeLegacyInventory(inventory: LegacyResources | null): LegacyInventorySummary {
  if (!inventory) {
    return emptySummary();
  }

  const resourceByID = new Map(inventory.resources.map((resource) => [resource.id, resource]));
  const imageResources = LEGACY_IMAGE_RESOURCES.flatMap((definition) => {
    const resource = resourceByID.get(definition.id);
    if (!resource) {
      return [];
    }
    return [{
      ...definition,
      itemCount: resource.itemCount,
      itemLength: resource.itemLength,
    }];
  });

  return {
    available: true,
    totalResources: inventory.count,
    cityNames: resourceByID.get(58)?.itemCount ?? 0,
    generalScenarios: resourceByID.get(61)?.itemCount ?? 0,
    variableStrings: resourceByID.get(64)?.itemCount ?? 0,
    battleMaps: inventory.resources.filter((resource) => resource.id >= 110 && resource.id <= 116).length,
    knownImageGroups: LEGACY_IMAGE_RESOURCES.length,
    presentImageGroups: imageResources.length,
    imageResourceItems: imageResources.reduce((total, resource) => total + resource.itemCount, 0),
    imageResources,
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
    knownImageGroups: LEGACY_IMAGE_RESOURCES.length,
    presentImageGroups: 0,
    imageResourceItems: 0,
    imageResources: [],
  };
}
