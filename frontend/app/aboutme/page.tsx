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
import { SkeletonArticle } from "@/components/ui/Skeleton";
import { ErrorState } from "@/components/ui/ErrorState";

export default function AboutMePage() {
  const [content, setContent] = useState<string>("");
  const [exists, setExists] = useState<boolean>(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [retryKey, setRetryKey] = useState(0);

  useEffect(() => {
    let cancelled = false;

    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);
        const response = await getAboutMe();
        if (cancelled) return;
        setContent(response.content);
        setExists(response.exists);
      } catch (err) {
        if (!cancelled) {
          console.error("Failed to fetch aboutme:", err);
          setError("Failed to load About Me.");
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

  // Update document title for the About Me page (restores global title on unmount)
  useEffect(() => {
    if (!loading && !error && exists) {
      document.title = "About Me | Terminalog";
      return () => {
        document.title = "Terminalog | Terminal Editorial Blog";
      };
    }
  }, [content, exists, loading, error]);

  if (loading) {
    return <SkeletonArticle />;
  }

  if (error) {
    return (
      <ErrorState
        title="About Me Unavailable"
        message={error}
        onRetry={() => setRetryKey((k) => k + 1)}
      />
    );
  }

  if (!exists) {
    return (
      <div className="px-6 max-w-4xl mx-auto py-32">
          <div className="text-center">
            <h1 className="font-headline text-4xl text-outline mb-4">No About Me Found</h1>
            <p className="text-on-surface-variant">
              Create <code className="text-tertiary bg-surface-container-low px-2">_ABOUTME.md</code> in your blog directory.
            </p>
          </div>
      </div>
    );
  }

  return (
    <div className="px-6 max-w-4xl mx-auto">
        {/* Content - Use MarkdownRenderer */}
        <article className="space-y-12">
          <MarkdownRenderer content={content} />
        </article>
    </div>
  );
}
