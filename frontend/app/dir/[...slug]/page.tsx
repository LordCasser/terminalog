/**
 * Directory Page - Article List for Subdirectories
 * 
 * RESTful routing: /dir/{path} (path parameter, not query parameter)
 * Uses Next.js catch-all route [...slug] to capture nested paths.
 * Examples:
 *   /dir/tech → shows articles in tech/ directory
 *   /dir/tech/golang → shows articles in tech/golang/ directory
 * 
 * For static export with fallback, ArticleListPage extracts the directory
 * path from the browser URL rather than relying on Next.js params.
 */

import { ArticleListPage } from "@/components/brutalist/ArticleListPage";

// Generate static params for static export
export async function generateStaticParams() {
  return [{ slug: ["_fallback"] }];
}

export default function DirPage() {
  // ArticleListPage extracts path from browser URL directly
  return <ArticleListPage />;
}