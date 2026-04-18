/**
 * Tree API
 * 
 * Handles fetching directory tree structure.
 * RESTful v1 API path: GET /api/v1/tree
 */

import { apiClient } from './client';
import type { TreeNode } from '@/types';

interface GetTreeParams {
  dir?: string;
}

interface TreeResponse {
  root: TreeNode;
  currentDir: string;
}

/**
 * Get directory tree structure
 * GET /api/v1/tree?dir={dir}
 */
export async function getTree(params: GetTreeParams = {}): Promise<TreeResponse> {
  const query = params.dir ? `?dir=${encodeURIComponent(params.dir)}` : '';
  return apiClient.get<TreeResponse>(`/api/v1/tree${query}`);
}
