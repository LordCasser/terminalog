/**
 * Search API
 * 
 * Handles searching articles by title.
 * RESTful v1 API path: GET /api/v1/articles/search
 * (merged into Articles resource, see docs/api-spec.md)
 */

import { apiClient } from './client';

export interface SearchResult {
  path: string;
  title: string;
  matchedTitle: string;
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
 * Search articles by title
 * GET /api/v1/articles/search?q={keyword}&dir={dir}
 */
export async function searchArticles(params: SearchParams): Promise<SearchResponse> {
  const query = new URLSearchParams();
  query.set('q', params.q);
  if (params.dir) query.set('dir', params.dir);
  
  return apiClient.get<SearchResponse>(`/api/v1/articles/search?${query}`);
}