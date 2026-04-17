/**
 * Articles API
 * 
 * Handles fetching article list, article content, and timeline.
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
 */
export async function getArticles(params: GetArticlesParams = {}): Promise<ArticleListResponse> {
  const query = new URLSearchParams();
  
  if (params.dir) query.set('dir', params.dir);
  if (params.sort) query.set('sort', params.sort);
  if (params.order) query.set('order', params.order);
  
  return apiClient.get<ArticleListResponse>(`/api/articles?${query}`);
}

/**
 * Get article content by path
 */
export async function getArticleContent(path: string): Promise<ArticleResponse> {
  return apiClient.get<ArticleResponse>(`/api/articles/${encodeURIComponent(path)}`);
}

/**
 * Get article raw content (Markdown text)
 */
export async function getArticleRaw(path: string): Promise<string> {
  return apiClient.getText(`/api/articles/${encodeURIComponent(path)}`);
}

/**
 * Get article Git timeline (commit history)
 */
export async function getArticleTimeline(path: string): Promise<{ commits: CommitInfo[] }> {
  return apiClient.get(`/api/articles/${encodeURIComponent(path)}/timeline`);
}

/**
 * Get article version info
 */
export async function getArticleVersion(path: string): Promise<{ version: VersionInfo; history: VersionHistoryEntry[] }> {
  return apiClient.get(`/api/articles/${encodeURIComponent(path)}/version`);
}