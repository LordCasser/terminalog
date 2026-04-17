/**
 * Articles API
 * 
 * Handles fetching article list, article content, and timeline.
 * RESTful v1 API paths (see docs/api-spec.md)
 */

import { apiClient } from './client';
import type { Article, ArticleListResponse, ArticleResponse, CommitInfo, VersionInfo, VersionHistoryEntry } from '@/types';

interface GetArticlesParams {
  dir?: string;
  sort?: 'created' | 'edited';
  order?: 'asc' | 'desc';
}

/**
 * Get list of articles in a directory
 * GET /api/v1/articles
 */
export async function getArticles(params: GetArticlesParams = {}): Promise<ArticleListResponse> {
  const query = new URLSearchParams();
  
  if (params.dir) query.set('dir', params.dir);
  if (params.sort) query.set('sort', params.sort);
  if (params.order) query.set('order', params.order);
  
  return apiClient.get<ArticleListResponse>(`/api/v1/articles?${query}`);
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