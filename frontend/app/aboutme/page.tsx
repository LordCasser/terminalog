/**
 * About Me Page
 * 
 * Displays content from _ABOUTME.md with Brutalist styling.
 * Navbar and CommandPrompt are public components in layout.tsx.
 */

"use client";

import { useState, useEffect } from "react";
import { getAboutMe } from "@/lib/api/aboutme";
import { MarkdownRenderer } from "@/components/article/MarkdownRenderer";

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
        {/* Main Content - Navbar is in layout.tsx */}
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
      {/* Main Content - Navbar is in layout.tsx */}
      <main className="pt-24 pb-32 px-6 max-w-4xl mx-auto">
        {/* Content - Use MarkdownRenderer */}
        <article className="space-y-12">
          <MarkdownRenderer content={content} />
        </article>
      </main>
    </div>
  );
}