/**
 * Search API
 * 
 * Handles searching articles and directories by title, file name, and directory name.
 * RESTful v1 API path: GET /api/v1/search?q={keyword}&dir={dir}
 * Returns both articles (type="file") and directories (type="dir").
 */

import { apiClient } from './client';

export interface SearchResult {
  path: string;
  title: string;
  matchedTitle: string;
  type: "file" | "dir";
}

interface SearchParams {
  q: string;
  dir?: string;
}

interface SearchResponse {
  results: SearchResult[];
  total: number;
}

/**
 * Search articles and directories by keyword
 * GET /api/v1/search?q={keyword}&dir={dir}
 */
export async function searchArticles(params: SearchParams): Promise<SearchResponse> {
  const query = new URLSearchParams();
  query.set('q', params.q);
  if (params.dir) query.set('dir', params.dir);
  
  return apiClient.get<SearchResponse>(`/api/v1/search?${query}`);
}