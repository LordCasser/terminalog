/**
 * Home Page - Root Article List
 * 
 * Displays article table for root directory.
 * RESTful routing: / (root directory listing)
 * Uses shared ArticleListPage component.
 */

"use client";

import { ArticleListPage } from "@/components/brutalist/ArticleListPage";

export default function Home() {
  return <ArticleListPage currentDir="" />;
}