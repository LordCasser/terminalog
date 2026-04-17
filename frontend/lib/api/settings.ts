/**
 * Settings API
 * 
 * Handles fetching frontend configuration.
 * RESTful v1 API path: GET /api/v1/settings
 * (renamed from /api/config, see docs/api-spec.md)
 */

import { apiClient } from './client';

export interface SettingsResponse {
  owner: string;
  title?: string;
  description?: string;
}

/**
 * Get frontend settings
 * GET /api/v1/settings
 */
export async function getSettings(): Promise<SettingsResponse> {
  return apiClient.get<SettingsResponse>('/api/v1/settings');
}