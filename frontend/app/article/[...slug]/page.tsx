/**
 * Article Detail Page (Dynamic Route)
 * 
 * Displays article content with Brutalist styling.
 * RESTful routing: /article/{path} (path parameter, not query parameter)
 * Uses Next.js catch-all route [...slug] to capture nested paths.
 * 
 * For static export with fallback, ArticleContent extracts the path
 * from the browser URL rather than relying on Next.js params,
 * because the fallback page always has slug=["_fallback"].
 */

import { ArticleContent } from "./ArticleContent";

// Generate static params for static export
// Returns a fallback route to handle dynamic article paths at runtime
export async function generateStaticParams() {
  return [{ slug: ["_fallback"] }];
}

export default function ArticlePage() {
  // ArticleContent extracts path from browser URL directly
  // This avoids the static export fallback problem where params
  // are always ["_fallback"] regardless of the actual URL
  return <ArticleContent />;
}