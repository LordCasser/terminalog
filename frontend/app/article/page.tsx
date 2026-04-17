/**
 * Article Detail Page
 * 
 * Displays article content with Brutalist styling.
 * Uses query parameter for article path (compatible with static export).
 */

"use client";

import { useState, useEffect, Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { getArticleContent, getArticleTimeline, getArticleVersion } from "@/lib/api/articles";
import { Navbar } from "@/components/brutalist";
import type { Article, CommitInfo, VersionInfo, VersionHistoryEntry } from "@/types";

function ArticleContent() {
  const searchParams = useSearchParams();
  const articlePath = searchParams.get("path") || "blog.md";
  const decodedPath = decodeURIComponent(articlePath);
  
  const [article, setArticle] = useState<Article | null>(null);
  const [content, setContent] = useState<string>("");
  const [commits, setCommits] = useState<CommitInfo[]>([]);
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [versionHistory, setVersionHistory] = useState<VersionHistoryEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [showHistory, setShowHistory] = useState(false);
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch article content
        const articleResponse = await getArticleContent(decodedPath);
        setArticle(articleResponse.article);
        setContent(articleResponse.content || "");
        
        // Fetch timeline
        const timelineResponse = await getArticleTimeline(decodedPath);
        setCommits(timelineResponse.commits);
        
        // Fetch version info
        const versionResponse = await getArticleVersion(decodedPath);
        setVersionInfo(versionResponse.version);
        setVersionHistory(versionResponse.history);
      } catch (error) {
        console.error("Failed to fetch article:", error);
        // Use mock data for static export preview
        setArticle(mockArticle);
        setContent(mockContent);
        setCommits(mockCommits);
        setVersionInfo(mockVersionInfo);
        setVersionHistory(mockVersionHistory);
      } finally {
        setLoading(false);
      }
    };
    
    fetchData();
  }, [decodedPath]);
  
  // Mock data for static export preview
  const mockArticle: Article = {
    path: decodedPath,
    name: decodedPath.split("/").pop() || "blog.md",
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
  
  const mockVersionHistory: VersionHistoryEntry[] = [
    { version: "v2.0.48", hash: "a7f2b91", author: "root", timestamp: "2024-05-24 09:12", message: "Update blog structure", linesAdded: 12, linesDeleted: 3, changeType: "patch" },
    { version: "v2.0.47", hash: "c3d1e42", author: "root", timestamp: "2024-05-20 16:45", message: "Initial draft", linesAdded: 150, linesDeleted: 0, changeType: "minor" },
    { version: "v2.0.40", hash: "f9e8d7c", author: "root", timestamp: "2024-05-18 10:20", message: "Setup repository", linesAdded: 50, linesDeleted: 0, changeType: "major" },
  ];
  
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <span className="text-outline font-mono">Loading...</span>
      </div>
    );
  }
  
  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "2-digit" });
  };
  
  // Extract quote from content (first blockquote or first paragraph)
  const quote = "We are not just building a blog; we are building a high-fidelity documentation of thought.";
  
  return (
    <div className="min-h-screen">
      {/* Navbar */}
      <Navbar currentPath={`~/lordcasser/${decodedPath}`} />
      
      {/* Main Content */}
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
          
          {/* Quote Blockquote */}
          <p className="font-mono text-lg text-primary max-w-2xl border-l-4 border-primary pl-6 py-2 italic bg-surface-container-low">
            "{quote}"
          </p>
        </section>
        
        {/* Content */}
        <article className="space-y-12 text-on-surface-variant leading-relaxed">
          {/* Simple Markdown Rendering */}
          <div className="markdown-body">
            {content.split("\n").map((line, i) => {
              if (line.startsWith("## ")) {
                return <h2 key={i} className="font-headline text-3xl font-bold text-secondary-fixed-dim mb-6">{line.replace("## ", "")}</h2>;
              }
              if (line.startsWith("- **")) {
                const text = line.replace("- **", "").replace("**:", ":").replace("**", "");
                return (
                  <div key={i} className="flex items-start gap-4 mb-4">
                    <span className="text-tertiary">➜</span>
                    <span>
                      {text.split(":").map((part, j) => (
                        j === 0 ? <span key={j} className="text-primary font-bold">{part}</span> : <span key={j}>:{part}</span>
                      ))}
                    </span>
                  </div>
                );
              }
              if (line.trim() === "") return null;
              return <p key={i} className="text-lg mb-4">{line}</p>;
            })}
          </div>
          
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

export default function ArticlePage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <span className="text-outline font-mono">Loading...</span>
      </div>
    }>
      <ArticleContent />
    </Suspense>
  );
}