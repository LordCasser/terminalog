/**
 * ArticleTable Component
 * 
 * 5-column table for articles and directories:
 * - Created | Updated | Editors | Filename | Latest Commit
 * 
 * Features:
 * - File type icons (folder for dirs, description for files)
 * - RESTful routing: /dir/{path} for directories, /article/{path} for files
 * - Clickable column headers for sorting (Created, Updated, Filename)
 */

"use client";

import type { Article } from "@/types";
import type { SortField, SortOrder } from "@/lib/api/articles";
import Link from "next/link";

interface ArticleTableProps {
  articles: Article[];
  sortField?: SortField;
  sortOrder?: SortOrder;
  onSort?: (field: SortField) => void;
}

/**
 * Get file icon based on type/extension
 */
function getFileIcon(article: Article): string {
  if (article.type === "dir") {
    return "folder";
  }
  return "description";
}

/**
 * Get icon color based on file type
 */
function getIconColor(article: Article): string {
  if (article.type === "dir") {
    return "text-primary";
  }
  return "text-tertiary";
}

/**
 * Format relative time
 */
function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);
  const diffMonths = Math.floor(diffDays / 30);
  
  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 30) return `${diffDays}d ago`;
  if (diffMonths < 12) return `${diffMonths}mo ago`;
  return date.toLocaleDateString("en-US", { year: "numeric", month: "short" });
}

/**
 * Format date for Created column
 */
function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", { year: "numeric", month: "2-digit", day: "2-digit" }).replace(",", "");
}

/**
 * Sortable column header component
 */
function SortableHeader({ 
  label, 
  field, 
  currentSort, 
  currentOrder, 
  onSort 
}: { 
  label: string; 
  field: SortField; 
  currentSort?: SortField; 
  currentOrder?: SortOrder; 
  onSort?: (field: SortField) => void;
}) {
  const isActive = currentSort === field;
  const arrow = isActive ? (currentOrder === "desc" ? "↓" : "↑") : "";
  
  return (
    <th 
      className={`px-6 py-3 font-medium cursor-pointer select-none hover:text-secondary transition-colors ${isActive ? "text-secondary" : ""}`}
      onClick={() => onSort?.(field)}
    >
      {label} {arrow}
    </th>
  );
}

export function ArticleTable({ articles, sortField, sortOrder, onSort }: ArticleTableProps) {
  return (
    <div className="w-full overflow-x-auto">
      <table className="w-full text-left border-collapse min-w-[1000px]">
        <thead>
          <tr className="bg-surface-container border-none text-outline uppercase text-[10px] tracking-[0.2em] font-bold">
            <SortableHeader label="Created" field="created" currentSort={sortField} currentOrder={sortOrder} onSort={onSort} />
            <SortableHeader label="Updated" field="edited" currentSort={sortField} currentOrder={sortOrder} onSort={onSort} />
            <th className="px-6 py-3 font-medium">
              Editors
            </th>
            <SortableHeader label="Filename" field="name" currentSort={sortField} currentOrder={sortOrder} onSort={onSort} />
            <th className="px-6 py-3 font-medium">
              Latest Commit
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-transparent">
          {articles.map((article, index) => (
            <tr key={index} className="hover:bg-surface-container-high transition-colors group cursor-pointer">
              {/* Created */}
              <td className="px-6 py-5 text-on-surface-variant text-sm whitespace-nowrap">
                {article.createdAt ? formatDate(article.createdAt) : "—"}
              </td>
              
              {/* Updated */}
              <td className="px-6 py-5 text-on-surface-variant text-sm whitespace-nowrap">
                {article.editedAt ? formatRelativeTime(article.editedAt) : "—"}
              </td>
              
              {/* Editors */}
              <td className="px-6 py-5 whitespace-nowrap">
                {article.contributors?.map((contributor, i) => (
                  <span 
                    key={i} 
                    className={`tag ${i === 0 ? 'tag-primary' : 'tag-secondary'} mr-1`}
                  >
                    {contributor}
                  </span>
                )) || "—"}
              </td>
              
              {/* Filename */}
              <td className="px-6 py-5 whitespace-nowrap">
                <Link 
                  href={article.type === "dir" ? `/dir/${article.path}` : `/article/${article.path}`}
                  className="flex items-center gap-3 font-bold transition-colors"
                >
                  <span className={`material-symbols-outlined text-[18px] ${getIconColor(article)}`}>
                    {getFileIcon(article)}
                  </span>
                  <span className={`${getIconColor(article)} group-hover:text-secondary transition-colors`}>
                    {article.name}
                  </span>
                </Link>
              </td>
              
              {/* Latest Commit */}
              <td className="px-6 py-5 text-outline text-sm italic whitespace-nowrap">
                {article.latestCommit || "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}