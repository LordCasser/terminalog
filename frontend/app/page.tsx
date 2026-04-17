/**
 * Home Page - Main Article List
 * 
 * Displays article table with Brutalist styling.
 * Navbar and CommandPrompt are public components in layout.tsx.
 */

"use client";

import { useState, useEffect } from "react";
import { ArticleTable } from "@/components/brutalist/ArticleTable";
import { getArticles } from "@/lib/api/articles";
import type { Article } from "@/types";

export default function Home() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [sortField, setSortField] = useState<string>("edited");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    // Fetch articles on mount
    const fetchArticles = async () => {
      try {
        const response = await getArticles({
          sort: sortField as "created" | "edited",
          order: sortOrder,
        });
        setArticles(response.articles);
      } catch (error) {
        console.error("Failed to fetch articles:", error);
        // Use mock data for static export preview
        setArticles(mockArticles);
      } finally {
        setLoading(false);
      }
    };
    
    fetchArticles();
  }, [sortField, sortOrder]);
  
  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortOrder("desc");
    }
  };
  
  // Mock data for static export preview
  const mockArticles: Article[] = [
    {
      path: "assets",
      name: "assets/",
      title: "",
      type: "dir",
      createdAt: "2023-10-12",
      createdBy: "root",
      editedAt: new Date(Date.now() - 2 * 3600000).toISOString(),
      editedBy: "root",
      contributors: ["root", "dev"],
      latestCommit: "chore: optimize image assets",
    },
    {
      path: "the-brutalist-compiler.md",
      name: "the-brutalist-compiler.md",
      title: "The Brutalist Compiler",
      type: "file",
      createdAt: "2023-11-05",
      createdBy: "root",
      editedAt: new Date(Date.now() - 10 * 60000).toISOString(),
      editedBy: "root",
      contributors: ["root"],
      latestCommit: "feat: add manifesto section",
    },
    {
      path: "syntax-highlighter-setup.md",
      name: "syntax-highlighter-setup.md",
      title: "Syntax Highlighter Setup",
      type: "file",
      createdAt: "2023-11-01",
      createdBy: "guest",
      editedAt: new Date(Date.now() - 24 * 3600000).toISOString(),
      editedBy: "guest",
      contributors: ["guest"],
      latestCommit: "fix: correct dracula hex codes",
    },
    {
      path: "config.yml",
      name: "config.yml",
      title: "Configuration",
      type: "file",
      createdAt: "2023-10-20",
      createdBy: "root",
      editedAt: new Date(Date.now() - 4 * 24 * 3600000).toISOString(),
      editedBy: "root",
      contributors: ["root", "dev"],
      latestCommit: "build: update ci/cd pipelines",
    },
    {
      path: "drafts",
      name: "drafts/",
      title: "",
      type: "dir",
      createdAt: "2023-09-15",
      createdBy: "root",
      editedAt: new Date(Date.now() - 30 * 24 * 3600000).toISOString(),
      editedBy: "root",
      contributors: ["root"],
      latestCommit: "initial commit",
    },
  ];
  
  return (
    <div className="min-h-screen">
      {/* Main Content - Navbar is in layout.tsx */}
      <main className="flex-grow w-full bg-surface-lowest px-0 overflow-x-auto">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <span className="text-outline font-mono">Loading...</span>
          </div>
        ) : (
          <ArticleTable 
            articles={articles}
            onSort={handleSort}
            sortField={sortField}
            sortOrder={sortOrder}
          />
        )}
      </main>
    </div>
  );
}