/**
 * Article Detail Page (Dynamic Route)
 * 
 * Displays article content with Brutalist styling.
 * RESTful routing: /article/{path} (path parameter, not query parameter)
 * Uses Next.js catch-all route [...slug] to capture nested paths.
 * Navbar and CommandPrompt are public components in layout.tsx.
 * 
 * Architecture: Split into server component (generateStaticParams) and client component (ArticleContent).
 */

import { ArticleContent } from "./ArticleContent";

// Generate static params for static export
// Returns a fallback route to handle dynamic article paths at runtime
// The actual article content will be fetched from the API when served
export async function generateStaticParams() {
  // Return a fallback that allows any path to be handled at runtime
  return [{ slug: ["_fallback"] }];
}

export default function ArticlePage({ params }: { params: Promise<{ slug: string[] }> }) {
  // In Next.js 15+, params is a Promise
  // For static export, we need to handle the fallback case
  return <ArticleContentWrapper params={params} />;
}

// Async wrapper to handle params Promise
async function ArticleContentWrapper({ params }: { params: Promise<{ slug: string[] }> }) {
  const resolvedParams = await params;
  const articlePath = resolvedParams.slug.join("/");
  
  // Skip fallback route, use actual path
  const actualPath = articlePath === "_fallback" ? "blog.md" : articlePath;
  
  return <ArticleContent articlePath={actualPath} />;
}