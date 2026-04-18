/**
 * Article Detail Page (Optional Catch-all Route)
 * 
 * Displays article content with Brutalist styling.
 * RESTful routing: /article/{path} (path parameter, not query parameter)
 * Uses Next.js optional catch-all route [[...slug]] to capture nested paths.
 * 
 * For static export with fallback, ArticleContent extracts the path
 * from the browser URL rather than relying on Next.js params,
 * because the fallback page always has slug=["_fallback"].
 */

import { ArticleContent } from "./ArticleContent";

// Generate static params for static export.
// The empty slug satisfies Next.js export requirements for the route itself.
// The _fallback route is used by the Go static handler for deep links.
export async function generateStaticParams() {
  return [{ slug: [] }, { slug: ["_fallback"] }];
}

export default function ArticlePage() {
  // ArticleContent extracts path from browser URL directly
  // This avoids the static export fallback problem where params
  // are always ["_fallback"] regardless of the actual URL
  return <ArticleContent />;
}
