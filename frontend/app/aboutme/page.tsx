/**
 * About Me Page
 * 
 * Displays content from _ABOUTME.md with Brutalist styling.
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { getAboutMe } from "@/lib/api/aboutme";

export default function AboutMePage() {
  const [content, setContent] = useState<string>("");
  const [exists, setExists] = useState<boolean>(false);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await getAboutMe();
        setContent(response.content);
        setExists(response.exists);
      } catch (error) {
        console.error("Failed to fetch aboutme:", error);
        // Use mock content for static export preview
        setContent(mockContent);
        setExists(true);
      } finally {
        setLoading(false);
      }
    };
    
    fetchData();
  }, []);
  
  // Mock content for static export preview
  const mockContent = `# About Me

Hello, I'm a developer who loves terminal aesthetics and Brutalist design philosophy.

## Background

I build systems that prioritize:
- **Clarity** over complexity
- **Precision** over convenience
- **Intent** over aesthetics

## Projects

- **Terminalog**: A terminal-style blog system with Git integration
- **Other Projects**: Various tools and experiments

## Contact

Find me on GitHub or reach out via email.

---

*This page is rendered from \`_ABOUTME.md\` in your blog content directory.*`;
  
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <span className="text-outline font-mono">Loading...</span>
      </div>
    );
  }
  
  if (!exists) {
    return (
      <div className="min-h-screen">
        <header className="fixed top-0 w-full z-50 bg-surface flex justify-between items-center px-6 py-4">
          <Link href="/" className="text-lg font-bold text-secondary font-mono tracking-tight">
            ~/lordcasser/_ABOUTME.md
          </Link>
        </header>
        <main className="pt-24 pb-32 px-6 max-w-4xl mx-auto">
          <div className="text-center">
            <h1 className="font-headline text-4xl text-outline mb-4">No About Me Found</h1>
            <p className="text-on-surface-variant">
              Create <code className="text-tertiary bg-surface-container-low px-2">_ABOUTME.md</code> in your blog directory.
            </p>
          </div>
        </main>
      </div>
    );
  }
  
  return (
    <div className="min-h-screen">
      {/* Navbar */}
      <header className="fixed top-0 w-full z-50 bg-surface flex justify-between items-center px-6 py-4">
        <div className="flex items-center gap-6">
          <Link href="/" className="text-lg font-bold text-secondary font-mono tracking-tight">
            ~/lordcasser/_ABOUTME.md
          </Link>
          <nav className="hidden md:flex items-center gap-8 font-mono tracking-tight">
            <Link href="/" className="text-outline hover:bg-surface-container-highest transition-colors px-2 py-1">
              POSTS
            </Link>
            <Link href="/aboutme" className="text-primary-container font-bold px-2 py-1">
              ABOUTME
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <button className="text-primary-container hover:bg-surface-container-highest transition-colors p-2">
            <span className="material-symbols-outlined">search</span>
          </button>
        </div>
      </header>
      
      {/* Main Content */}
      <main className="pt-24 pb-32 px-6 max-w-4xl mx-auto">
        {/* Content */}
        <article className="space-y-12 text-on-surface-variant leading-relaxed">
          <div className="markdown-body">
            {content.split("\n").map((line, i) => {
              if (line.startsWith("# ")) {
                return <h1 key={i} className="font-headline text-6xl font-bold text-on-surface tracking-tighter mb-8">{line.replace("# ", "").toUpperCase()}</h1>;
              }
              if (line.startsWith("## ")) {
                return <h2 key={i} className="font-headline text-3xl font-bold text-secondary-fixed-dim mb-6">{line.replace("## ", "")}</h2>;
              }
              if (line.startsWith("---")) {
                return <hr key={i} className="border-t border-surface-container my-8" />;
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
              if (line.startsWith("*") && line.endsWith("*")) {
                return <p key={i} className="text-sm text-outline italic mt-4">{line.replace(/\*/g, "")}</p>;
              }
              if (line.includes("`")) {
                const parts = line.split("`");
                return (
                  <p key={i} className="mb-4">
                    {parts.map((part, j) => (
                      j % 2 === 1 ? <code key={j} className="text-tertiary bg-surface-container-low px-1">{part}</code> : <span key={j}>{part}</span>
                    ))}
                  </p>
                );
              }
              if (line.trim() === "") return null;
              return <p key={i} className="text-lg mb-4">{line}</p>;
            })}
          </div>
        </article>
      </main>
    </div>
  );
}