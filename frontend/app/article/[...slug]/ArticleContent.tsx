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

// Mock data for static export preview
function getMockData(articlePath: string) {
  const mockArticle: Article = {
    path: articlePath,
    name: articlePath.split("/").pop() || "blog.md",
    title: "The Brutalist Compiler",
    type: "file",
    createdAt: "2024-05-18",
    createdBy: "root",
    editedAt: new Date().toISOString(),
    editedBy: "root",
    contributors: ["root"],
    latestCommit: "Update blog structure for v2.0.48",
  };
  
  const mockContent = `## 01. The Core Philosophy

In an era of rounded corners and soft gradients, the **Terminal Editorial** system stands as a monolith of intent. It rejects the standard "boxed" template in favor of intentional asymmetry and monolithic layering.

## 02. Syntax and Structure

The "Compiler" aspect comes from our use of high-contrast "syntax highlighting" colors to guide the eye through complex information architecture.

- **Surface Lowest:** The deepest background level behind the main content.
- **Surface Container:** Used for sidebar navigation or secondary meta-information.
- **Active States:** Reserved for highlighted code blocks or focused terminal lines.

## 03. The Glass & Gradient Rule

To add visual "soul," we implement what we call **"Terminal Fog"**. This is the subtle interplay between sharp edges and soft backdrop blurs that creates a sense of spatial depth without traditional shadows.`;
  
  const mockCommits: CommitInfo[] = [
    { hash: "a7f2b91", author: "root", message: "Update blog structure for v2.0.48", timestamp: "2024-05-24 09:12", linesAdded: 12, linesDeleted: 3 },
    { hash: "c3d1e42", author: "root", message: "Initial draft of The Brutalist Compiler", timestamp: "2024-05-20 16:45", linesAdded: 150, linesDeleted: 0 },
    { hash: "f9e8d7c", author: "root", message: "Setup repository structure", timestamp: "2024-05-18 10:20", linesAdded: 50, linesDeleted: 0 },
  ];
  
  const mockVersionInfo: VersionInfo = {
    version: "v2.0.48",
    changeType: "patch",
    baseLines: 200,
    currentLines: 212,
    changePercent: 6,
  };
  
  return { mockArticle, mockContent, mockCommits, mockVersionInfo };
}

export function ArticleContent() {
  const [article, setArticle] = useState<Article | null>(null);
  const [content, setContent] = useState<string>("");
  const [commits, setCommits] = useState<CommitInfo[]>([]);
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [showHistory, setShowHistory] = useState(false);
  
  useEffect(() => {
    // Extract article path from browser URL (only in client)
    const articlePath = extractArticlePathFromURL();
    
    const fetchData = async () => {
      try {
        // Fetch article content
        const articleResponse = await getArticleContent(articlePath);
        setArticle(articleResponse.article);
        setContent(articleResponse.content || "");
        
        // Fetch timeline
        const timelineResponse = await getArticleTimeline(articlePath);
        setCommits(timelineResponse.commits);
        
        // Fetch version info
        const versionResponse = await getArticleVersion(articlePath);
        setVersionInfo(versionResponse.version);
      } catch (error) {
        console.error("Failed to fetch article:", error);
        // Use mock data for static export preview
        const { mockArticle, mockContent, mockCommits, mockVersionInfo } = getMockData(articlePath);
        setArticle(mockArticle);
        setContent(mockContent);
        setCommits(mockCommits);
        setVersionInfo(mockVersionInfo);
      } finally {
        setLoading(false);
      }
    };
    
fetchData();
  }, []);
  
  // Extract base path for image transformation (client-side only)
  const basePath = (article?.path || "").replace(/\/[^\/]+\.md$/, '');
  
  // Extract quote from content (first blockquote)
  const extractQuote = (markdown: string): string => {
    const match = markdown.match(/^> (.+)$/m);
    return match ? match[1] : "";
  };
  const quote = extractQuote(content);
  
  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "2-digit" });
  };
  
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
              {versionInfo?.version || "v1.0.0"}
            </span>
            <span className="font-mono text-xs text-outline">
              {formatDate(article?.editedAt || new Date().toISOString())}
            </span>
          </div>
          
          {/* Title */}
          <h1 className="font-headline font-bold text-4xl leading-none text-on-surface tracking-tighter mb-8">
            {article?.title?.toUpperCase() || "UNTITLED"}
          </h1>
          
          {/* Quote Blockquote - only show if quote exists */}
          {quote && (
            <p className="font-mono text-lg text-primary max-w-2xl border-l-4 border-primary pl-6 py-2 italic bg-surface-container-low">
              "{quote}"
            </p>
          )}
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