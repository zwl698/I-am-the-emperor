import { describe, expect, it } from 'vitest';
import { summarizeLegacyInventory } from './legacyInventory';
import type { LegacyResources } from '../api/types';

describe('summarizeLegacyInventory', () => {
  it('summarizes imported legacy archive resources for the HUD', () => {
    const inventory: LegacyResources = {
      source: '../sanguobaye_c-master/src/dat.lib.orig',
      count: 87,
      resources: [
        { address: 192843, length: 392, id: 58, itemCount: 43, itemLength: 10, key: 192, reserved: 0 },
        { address: 65536, length: 12013, id: 61, itemCount: 4, itemLength: 3000, key: 0, reserved: 0 },
        { address: 111412, length: 2914, id: 64, itemCount: 163, itemLength: 0, key: 192, reserved: 0 },
        { address: 181588, length: 1052, id: 110, itemCount: 1, itemLength: 1040, key: 0, reserved: 0 },
        { address: 182640, length: 1052, id: 111, itemCount: 1, itemLength: 1040, key: 0, reserved: 0 },
      ],
    };

    expect(summarizeLegacyInventory(inventory)).toEqual({
      available: true,
      totalResources: 87,
      cityNames: 43,
      generalScenarios: 4,
      variableStrings: 163,
      battleMaps: 2,
    });
  });

  it('returns an unavailable summary when the archive cannot be read', () => {
    expect(summarizeLegacyInventory(null)).toEqual({
      available: false,
      totalResources: 0,
      cityNames: 0,
      generalScenarios: 0,
      variableStrings: 0,
      battleMaps: 0,
    });
  });
});
