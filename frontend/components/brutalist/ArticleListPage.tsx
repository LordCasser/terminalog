/**
 * ArticleListPage - Shared Article List Component
 * 
 * Renders article table for both root (/) and directory (/dir/xxx) views.
 * RESTful routing: directories use /dir/{path}, articles use /article/{path}
 * 
 * API call: GET /api/v1/articles/{dir} (path parameter, not query parameter)
 * In static export mode, the directory path is extracted from the browser URL.
 */

"use client";

import { useState, useEffect } from "react";
import { ArticleTable } from "@/components/brutalist/ArticleTable";
import { getArticles } from "@/lib/api/articles";
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

  useEffect(() => {
    // Use provided currentDir or extract from URL (client-side only)
    const effectiveDir = currentDir ?? extractDirPathFromURL();

    const fetchArticles = async () => {
      try {
        const response = await getArticles(effectiveDir);
        setArticles(response.articles);
      } catch (error) {
        console.error("Failed to fetch articles:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchArticles();
  }, [currentDir]);

  return (
    <div className="min-h-screen">
      <main className="flex-grow w-full bg-surface-lowest px-0 overflow-x-auto">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <span className="text-outline font-mono">Loading...</span>
          </div>
        ) : (
          <ArticleTable articles={articles} />
        )}
      </main>
    </div>
  );
}