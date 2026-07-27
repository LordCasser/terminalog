/**
 * ArticleListPage - Shared Article List Component
 * 
 * Renders article table for both root (/) and directory (/dir/xxx) views.
 * RESTful routing: directories use /dir/{path}, articles use /article/{path}
 * 
 * API call: GET /api/v1/articles/{dir}?sort=xxx&order=xxx
 * Column header click triggers sort via query parameters.
 * In static export mode, the directory path is extracted from the browser URL.
 */

"use client";

import { useState, useEffect } from "react";
import { ArticleTable } from "@/components/brutalist/ArticleTable";
import { SkeletonTable } from "@/components/ui/Skeleton";
import { ErrorState } from "@/components/ui/ErrorState";
import { Icon } from "@/components/ui/Icon";
import { getArticles, SortField, SortOrder } from "@/lib/api/articles";
import { ApiError } from "@/lib/api/client";
import type { Article } from "@/types";

interface ArticleListPageProps {
  /** Current directory path. If not provided, extracted from browser URL. */
  currentDir?: string;
}

/**
 * Extract directory path from browser URL.
 * URL format: /dir/{path} -> path = "tech/golang"
 * For root path (/), returns empty string.
 */
function extractDirPathFromURL(): string {
  const pathname = window.location.pathname;
  if (pathname === "/" || pathname === "") {
    return "";
  }
  // Remove /dir/ prefix and trailing slash
  const path = pathname.replace(/^\/dir\/+/, "").replace(/\/+$/, "");
  return path;
}

export function ArticleListPage({ currentDir }: ArticleListPageProps) {
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [retryKey, setRetryKey] = useState(0);
  const [sortField, setSortField] = useState<SortField | undefined>(undefined);
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");

  useEffect(() => {
    let cancelled = false;

    // Use provided currentDir or extract from URL (client-side only)
    const effectiveDir = currentDir ?? extractDirPathFromURL();

    const fetchArticles = async () => {
      try {
        setLoading(true);
        setError(null);
        setNotFound(false);
        const response = await getArticles(effectiveDir, sortField, sortOrder);
        if (!cancelled) setArticles(response.articles);
      } catch (err) {
        if (!cancelled) {
          console.error("Failed to fetch articles:", err);
          if (err instanceof ApiError && err.status === 404) {
            setNotFound(true);
            setError("The requested directory does not exist.");
          } else {
            setError("Failed to load directory listing.");
          }
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetchArticles();
    return () => {
      cancelled = true;
    };
  }, [currentDir, sortField, sortOrder, retryKey]);

  /** Handle column header click for sorting */
  const handleSort = (field: SortField) => {
    // If same field, toggle order; otherwise set new field with desc default
    if (sortField === field) {
      setSortOrder(sortOrder === "desc" ? "asc" : "desc");
    } else {
      setSortField(field);
      setSortOrder("desc");
    }
  };

  /** Retry fetch after error */
  const handleRetry = () => setRetryKey((k) => k + 1);

  if (loading) {
    return <SkeletonTable />;
  }

  if (error) {
    return (
      <ErrorState
        title={notFound ? "Not Found" : "Directory Unavailable"}
        message={error}
        onRetry={notFound ? undefined : handleRetry}
      />
    );
  }

  if (articles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-32">
        <Icon name="folder" className="text-4xl text-outline mb-4" />
        <p className="font-mono text-sm text-outline">no entries</p>
      </div>
    );
  }

  return (
    <ArticleTable
      articles={articles}
      sortField={sortField}
      sortOrder={sortOrder}
      onSort={handleSort}
    />
  );
}
