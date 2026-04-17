/**
 * About Me API
 * 
 * Handles fetching _ABOUTME.md content.
 * RESTful v1 API path: GET /api/v1/special/aboutme
 */

import { apiClient } from './client';

export interface AboutMeResponse {
  content: string;
  exists: boolean;
}

/**
 * Get About Me page content from _ABOUTME.md
 * GET /api/v1/special/aboutme
 */
export async function getAboutMe(): Promise<AboutMeResponse> {
  return apiClient.get<AboutMeResponse>('/api/v1/special/aboutme');
}