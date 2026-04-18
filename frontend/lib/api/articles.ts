/**
 * Articles API
 * 
 * RESTful path-based routing:
 * - GET /api/v1/articles           → root directory listing
 * - GET /api/v1/articles/tech      → directory listing for tech/
 * - GET /api/v1/articles/tech/go.md → article content
 * - GET /api/v1/articles/tech/go.md/timeline → timeline
 * - GET /api/v1/articles/tech/go.md/version  → version
 */

import { apiClient } from './client';
import type { Article, ArticleListResponse, ArticleResponse, CommitInfo, VersionInfo, VersionHistoryEntry } from '@/types';

/**
 * Get directory listing (root or subdirectory)
 * GET /api/v1/articles or GET /api/v1/articles/{dirPath}
 * Returns both directories and files in the given path.
 */
export async function getArticles(dir?: string): Promise<ArticleListResponse> {
  const path = dir ? `/api/v1/articles/${encodeURIComponent(dir)}` : '/api/v1/articles';
  return apiClient.get<ArticleListResponse>(path);
}

/**
 * Get article content by path
 * GET /api/v1/articles/{path}
 */
export async function getArticleContent(path: string): Promise<ArticleResponse> {
  return apiClient.get<ArticleResponse>(`/api/v1/articles/${encodeURIComponent(path)}`);
}

/**
 * Get article raw content (Markdown text)
 * GET /api/v1/articles/{path}
 */
export async function getArticleRaw(path: string): Promise<string> {
  return apiClient.getText(`/api/v1/articles/${encodeURIComponent(path)}`);
}

/**
 * Get article Git timeline (commit history)
 * GET /api/v1/articles/{path}/timeline
 */
export async function getArticleTimeline(path: string): Promise<{ commits: CommitInfo[] }> {
  return apiClient.get(`/api/v1/articles/${encodeURIComponent(path)}/timeline`);
}

/**
 * Get article version info
 * GET /api/v1/articles/{path}/version
 */
export async function getArticleVersion(path: string): Promise<{ version: VersionInfo; history: VersionHistoryEntry[] }> {
  return apiClient.get(`/api/v1/articles/${encodeURIComponent(path)}/version`);
}