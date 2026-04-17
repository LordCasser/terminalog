/**
 * About Me API
 * 
 * Handles fetching _ABOUTME.md content.
 */

import { apiClient } from './client';

export interface AboutMeResponse {
  content: string;
  exists: boolean;
}

/**
 * Get About Me page content from _ABOUTME.md
 */
export async function getAboutMe(): Promise<AboutMeResponse> {
  return apiClient.get<AboutMeResponse>('/api/aboutme');
}