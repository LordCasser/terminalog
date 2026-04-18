/**
 * Articles API
 * 
 * RESTful path-based routing with query parameter sorting:
 * - GET /api/v1/articles           → root directory listing
 * - GET /api/v1/articles/tech      → directory listing for tech/
 * - GET /api/v1/articles/tech?sort=edited&order=desc → sorted listing
 * - GET /api/v1/articles/tech/go.md → article content
 * - GET /api/v1/articles/tech/go.md/timeline → timeline
 * - GET /api/v1/articles/tech/go.md/version  → version
 */

import { apiClient } from './client';
import type { ArticleListResponse, ArticleDetailResponse, CommitInfo, VersionInfo } from '@/types';

/** Sort field options */
export type SortField = 'created' | 'edited' | 'name';

/** Sort order options */
export type SortOrder = 'asc' | 'desc';

/**
 * Encode a path for use in URL path segments.
 * Preserves "/" separators but encodes special characters in each segment.
 */
function encodePath(path: string): string {
  return path.split("/").map(segment => encodeURIComponent(segment)).join("/");
}

/**
 * Get directory listing (root or subdirectory)
 * GET /api/v1/articles or GET /api/v1/articles/{dirPath}
 * Returns both directories and files in the given path.
 * Optional sort/order query parameters control listing order.
 */
export async function getArticles(dir?: string, sort?: SortField, order?: SortOrder): Promise<ArticleListResponse> {
  let path = dir ? `/api/v1/articles/${encodePath(dir)}` : '/api/v1/articles';
  
  // Append sort/order query parameters if provided
  const params = new URLSearchParams();
  if (sort) params.set('sort', sort);
  if (order) params.set('order', order);
  const queryString = params.toString();
  if (queryString) path += `?${queryString}`;
  
  return apiClient.get<ArticleListResponse>(path);
}

/**
 * Get article content by path
 * GET /api/v1/articles/{path}
 */
export async function getArticleContent(path: string): Promise<ArticleDetailResponse> {
  return apiClient.get<ArticleDetailResponse>(`/api/v1/articles/${encodePath(path)}`);
}

/**
 * Get article raw content (Markdown text)
 * GET /api/v1/articles/{path}
 */
export async function getArticleRaw(path: string): Promise<string> {
  return apiClient.getText(`/api/v1/articles/${encodePath(path)}`);
}

/**
 * Get article Git timeline (commit history)
 * GET /api/v1/articles/{path}/timeline
 */
export async function getArticleTimeline(path: string): Promise<{ commits: CommitInfo[] }> {
  return apiClient.get(`/api/v1/articles/${encodePath(path)}/timeline`);
}

/**
 * Get article version info
 * GET /api/v1/articles/{path}/version
 */
export async function getArticleVersion(path: string): Promise<VersionInfo> {
  return apiClient.get<VersionInfo>(`/api/v1/articles/${encodePath(path)}/version`);
}
