// TypeScript 类型定义 - Terminalog 前端

// 文章类型
export interface Article {
  path: string;
  name: string;
  title: string;
  type: 'file' | 'dir';
  createdAt: string;
  createdBy: string;
  editedAt: string;
  editedBy: string;
  contributors: string[];
  latestCommit: string;
}

// 文章列表响应
export interface ArticleListResponse {
  articles: Article[];
  total: number;
  currentDir: string;
}

// 文章详情响应
export interface ArticleDetailResponse {
  path: string;
  title: string;
  content: string;
  createdAt: string;
  createdBy: string;
  editedAt: string;
  editedBy: string;
  contributors: string[];
}

// Commit 信息
export interface CommitInfo {
  hash: string;
  author: string;
  timestamp: string;
  message: string;
}

// 目录树节点
export interface TreeNode {
  name: string;
  type: 'file' | 'dir';
  path: string;
  children?: TreeNode[];
}

// 命令类型
export interface Command {
  name: string;
  args: string[];
  flags: Record<string, string | boolean>;
  raw: string;
}

// 输出行
export interface OutputLine {
  id: string;
  type: 'command' | 'result' | 'error' | 'info';
  content: string;
  timestamp?: Date;
}

// 终端状态
export interface TerminalState {
  currentDir: string;
  history: string[];
  output: OutputLine[];
  mode: 'list' | 'view';
  viewingArticle?: Article;
  isLoading: boolean;
  error?: string;
}

// 排序状态
export interface SortState {
  field: 'created' | 'edited' | 'name';
  order: 'asc' | 'desc';
}

// 版本信息
export interface VersionInfo {
  currentVersion: string;
  history: VersionHistoryEntry[];
}

// 版本历史条目
export interface VersionHistoryEntry {
  version: string;
  hash: string;
  author: string;
  timestamp: string;
  linesChanged: number;
  changeType: 'major' | 'minor' | 'patch';
}

// About Me 响应
export interface AboutMeResponse {
  content: string;
  exists: boolean;
}
