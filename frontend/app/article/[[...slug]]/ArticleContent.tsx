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

export function ArticleContent() {
  const [article, setArticle] = useState<Article | null>(null);
  const [content, setContent] = useState<string>("");
  const [commits, setCommits] = useState<CommitInfo[]>([]);
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showHistory, setShowHistory] = useState(false);
  
  useEffect(() => {
    // Extract article path from browser URL (only in client)
    const articlePath = extractArticlePathFromURL();
    
    const fetchData = async () => {
      try {
        // Fetch article content
        const articleResponse = await getArticleContent(articlePath);
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
        
        // Fetch timeline
        const timelineResponse = await getArticleTimeline(articlePath);
        setCommits(timelineResponse.commits);
        
        // Fetch version info
        const versionResponse = await getArticleVersion(articlePath);
        setVersionInfo(versionResponse);
      } catch (error) {
        console.error("Failed to fetch article:", error);
        setError("Failed to load article.");
      } finally {
        setLoading(false);
      }
    };
    
    fetchData();
  }, []);
  
  // Extract base path for image transformation (client-side only)
  const basePath = (article?.path || "").replace(/\/[^\/]+\.md$/, '');
  
  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "2-digit" });
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <span className="text-outline font-mono">Loading...</span>
      </div>
    );
  }

  if (error || !article) {
    return (
      <div className="min-h-screen">
        <main className="pt-24 pb-32 px-6 max-w-4xl mx-auto">
          <div className="text-center">
            <h1 className="font-headline text-4xl text-outline mb-4">Article Unavailable</h1>
            <p className="text-on-surface-variant">{error || "The requested article could not be loaded."}</p>
          </div>
        </main>
      </div>
    );
  }
  
  return (
    <div className="min-h-screen">
      {/* Main Content - Navbar is in layout.tsx */}
      <main className="pt-24 pb-32 px-6 max-w-4xl mx-auto">
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
          <h1 className="font-headline font-bold text-4xl leading-none text-on-surface tracking-tighter mb-8">
            {article.title.toUpperCase() || "UNTITLED"}
          </h1>
        </section>
        
        {/* Content - Use MarkdownRenderer */}
        <article className="space-y-12 text-on-surface-variant leading-relaxed">
          <MarkdownRenderer content={content} basePath={basePath} />
          
          {/* EOF Section */}
          <section className="border-t border-surface-container pt-12">
            <div className="font-mono text-sm space-y-4">
              <p className="text-sm font-mono text-outline uppercase tracking-[0.2em] mb-6">
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
                  <span className="material-symbols-outlined text-sm">expand_more</span>
                  View full history
                </button>
              )}
            </div>
          </section>
        </article>
      </main>
    </div>
  );
}
