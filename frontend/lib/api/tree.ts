/**
 * Tree API
 * 
 * Handles fetching directory tree structure.
 */

import { apiClient } from './client';
import type { TreeNode } from '@/types';

interface GetTreeParams {
  dir?: string;
}

interface TreeResponse {
  tree: TreeNode;
}

/**
 * Get directory tree structure
 */
export async function getTree(params: GetTreeParams = {}): Promise<TreeResponse> {
  const query = params.dir ? `?dir=${encodeURIComponent(params.dir)}` : '';
  return apiClient.get<TreeResponse>(`/api/tree${query}`);
}