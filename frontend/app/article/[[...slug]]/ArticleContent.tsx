/**
 * Article Content Client Component
 * 
 * Client-side rendering for article content.
 * Fetches article data from API and renders with Brutalist styling.
 * 
 * In static export mode, the article path is extracted from the browser URL
 * rather than Next.js params, because the fallback page always has
 * slug=["_fallback"] regardless of the actual URL path.
 */

"use client";

import { useState, useEffect } from "react";
import { getArticleContent, getArticleTimeline, getArticleVersion } from "@/lib/api/articles";
import { MarkdownRenderer } from "@/components/article/MarkdownRenderer";
import { Icon } from "@/components/ui/Icon";
import { SkeletonArticle } from "@/components/ui/Skeleton";
import { ErrorState } from "@/components/ui/ErrorState";
import { ApiError } from "@/lib/api/client";
import type { Article, CommitInfo, VersionInfo } from "@/types";

/**
 * Extract article path from browser URL.
 * URL format: /article/{path} -> path = "tech/golang/go-guide.md"
 */
function extractArticlePathFromURL(): string {
  const pathname = window.location.pathname;
  // Remove /article/ prefix and trailing slash
  const path = pathname.replace(/^\/article\/+/, "").replace(/\/+$/, "");
  return path || "";
}

function extractLeadingMarkdownTitle(markdown: string): string | null {
  const match = markdown.match(/^\s*#\s+(.+?)\s*(?:\r?\n|$)/);
  return match?.[1]?.trim() || null;
}

function stripLeadingMarkdownTitle(markdown: string): string {
  return markdown.replace(/^\s*#\s+.+?\s*(?:\r?\n){1,2}/, "");
}

export function ArticleContent() {
  const [article, setArticle] = useState<Article | null>(null);
  const [content, setContent] = useState<string>("");
  const [commits, setCommits] = useState<CommitInfo[]>([]);
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [retryKey, setRetryKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    // Extract article path from browser URL (only in client)
    const articlePath = extractArticlePathFromURL();

    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);
        setNotFound(false);
        const [articleResponse, timelineResponse, versionResponse] = await Promise.all([
          getArticleContent(articlePath),
          getArticleTimeline(articlePath),
          getArticleVersion(articlePath),
        ]);
        if (cancelled) return;

        setArticle({
          path: articleResponse.path,
          name: articlePath.split("/").pop() || articleResponse.title,
          title: articleResponse.title,
          type: "file",
          createdAt: articleResponse.createdAt,
          createdBy: articleResponse.createdBy,
          editedAt: articleResponse.editedAt,
          editedBy: articleResponse.editedBy,
          contributors: articleResponse.contributors,
          latestCommit: "",
        });
        setContent(articleResponse.content || "");
        setCommits(timelineResponse.commits);
        setVersionInfo(versionResponse);
      } catch (err) {
        if (!cancelled) {
          console.error("Failed to fetch article:", err);
          if (err instanceof ApiError && err.status === 404) {
            setNotFound(true);
            setError("The requested article does not exist.");
          } else {
            setError("Failed to load article.");
          }
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetchData();
    return () => {
      cancelled = true;
    };
  }, [retryKey]);
  
  // Extract base path for image transformation (client-side only)
  const basePath = (article?.path || "").replace(/\/[^\/]+\.md$/, '');
  const leadingMarkdownTitle = extractLeadingMarkdownTitle(content);
  const articleHeading = leadingMarkdownTitle || article?.title || "UNTITLED";
  const renderedContent = leadingMarkdownTitle ? stripLeadingMarkdownTitle(content) : content;

  // Update document title to reflect the current article (restores global title on unmount)
  useEffect(() => {
    if (article && !loading && !error) {
      document.title = `${articleHeading} | Terminalog`;
      return () => {
        document.title = "Terminalog | Terminal Editorial Blog";
      };
    }
  }, [articleHeading, article, loading, error]);

  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "2-digit" });
  };

  if (loading) {
    return <SkeletonArticle />;
  }

  if (error || !article) {
    return (
      <ErrorState
        title={notFound ? "Article Not Found" : "Article Unavailable"}
        message={error || "The requested article could not be loaded."}
        onRetry={notFound ? undefined : () => setRetryKey((k) => k + 1)}
      />
    );
  }
  
  return (
    <div className="px-6 max-w-4xl mx-auto">
        {/* Post Header */}
        <section className="mb-16">
          {/* Tags + Version + Date */}
          <div className="flex items-center gap-4 mb-4">
            <span className="font-mono text-xs uppercase tracking-widest text-secondary bg-secondary-container px-2 py-0.5">
              Article
            </span>
            <span className="font-mono text-xs text-outline">
              {versionInfo?.currentVersion || "v1.0.0"}
            </span>
            <span className="font-mono text-xs text-outline">
              {formatDate(article.editedAt)}
            </span>
          </div>
          
          {/* Title */}
          <h1 className="font-article font-bold text-5xl leading-none text-on-surface tracking-tighter mb-8">
            {articleHeading}
          </h1>
        </section>
        
        {/* Content - Use MarkdownRenderer */}
        <article className="font-article space-y-12 text-on-surface-variant leading-relaxed">
          <MarkdownRenderer content={renderedContent} basePath={basePath} />
          
          {/* EOF Section */}
          <section className="border-t border-surface-container pt-12">
            <div className="font-article text-sm space-y-4">
              <p className="text-sm font-article text-outline uppercase tracking-[0.2em] mb-6">
                EOF (END OF FILE)
              </p>
              
              <h3 className="text-outline-variant text-xs uppercase tracking-widest mb-4">
                File History (git log --oneline)
              </h3>
              
              {/* History Items */}
              <div className="space-y-2 border-l border-surface-container-highest pl-4">
                {commits.slice(0, 3).map((commit, i) => (
                  <div key={i} className="flex flex-wrap items-center gap-x-4">
                    <span className="text-outline">{formatDate(commit.timestamp)}</span>
                    <span className="text-primary-container">{commit.hash}</span>
                    <span className="text-on-surface">{commit.message}</span>
                    <span className="text-tertiary">@{commit.author}</span>
                  </div>
                ))}
              </div>
              
              {/* Expand History Button */}
              {commits.length > 3 && (
                <button 
                  className="text-xs text-primary hover:underline flex items-center gap-1 mt-2"
                  onClick={() => setShowHistory(!showHistory)}
                >
                  <Icon name="chevron-down" className="text-sm" />
                  View full history
                </button>
              )}
            </div>
          </section>
        </article>
    </div>
  );
}
