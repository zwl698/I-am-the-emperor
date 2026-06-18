import type { CommandRequest, CreateGameRequest, GameSnapshot, LegacyResources, ScenarioList } from './types';

const API_BASE = import.meta.env.VITE_API_BASE ?? '';

export async function createGame(request: CreateGameRequest): Promise<GameSnapshot> {
  return requestJSON('/api/games', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function getScenarios(): Promise<ScenarioList> {
  return requestJSON('/api/scenarios');
}

export async function getCurrentGame(): Promise<GameSnapshot> {
  return requestJSON('/api/games/current');
}

export async function applyCommand(request: CommandRequest): Promise<GameSnapshot> {
  return requestJSON('/api/games/current/command', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function advanceMonth(): Promise<GameSnapshot> {
  return requestJSON('/api/games/current/advance-month', { method: 'POST' });
}

export async function getLegacyResources(): Promise<LegacyResources> {
  return requestJSON('/api/legacy/resources');
}

async function requestJSON<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed: ${response.status}`);
  }
  return response.json() as Promise<T>;
}
