/**
 * TOC Utility Functions
 *
 * Pure functions for Table of Contents generation from Markdown content.
 * Extracted from MarkdownRenderer.tsx for testability and reuse.
 *
 * Responsibilities:
 * - generateSlug: URL-friendly slug from plain text
 * - childrenToText: Convert React children to plain text string
 * - stripInlineMarkdown: Remove inline formatting from heading text
 * - extractHeadings: Parse markdown to find heading entries for TOC
 */

/**
 * Generates a URL-friendly slug from text content.
 */
export function generateSlug(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/(^-|-$)/g, '')
    .replace(/--+/g, '-');
}

/**
 * Converts React children to a plain text string.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function childrenToText(children: any): string {
  if (typeof children === 'string') return children;
  if (Array.isArray(children)) return children.map(childrenToText).join('');
  if (children && typeof children === 'object' && 'props' in children) {
    return childrenToText(children.props.children);
  }
  return '';
}

/**
 * Strips inline markdown formatting from a heading string
 * (bold, italic, code, links, images).
 */
export function stripInlineMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/__(.+?)__/g, '$1')
    .replace(/\*(.+?)\*/g, '$1')
    .replace(/_(.+?)_/g, '$1')
    .replace(/`(.+?)`/g, '$1')
    .replace(/\[(.+?)\]\(.+?\)/g, '$1')
    .replace(/!\[.*?\]\(.+?\)/g, '');
}

/** A TOC heading entry */
export interface TocEntry {
  level: number;
  text: string;
  slug: string;
}

/**
 * Extracts headings (h1-h6) from markdown string.
 */
export function extractHeadings(markdown: string): TocEntry[] {
  const headingRegex = /^(#{1,6})\s+(.+)$/gm;
  const headings: TocEntry[] = [];

  let match;
  while ((match = headingRegex.exec(markdown)) !== null) {
    const level = match[1].length;
    const rawText = match[2].trim();
    const text = stripInlineMarkdown(rawText);
    const slug = generateSlug(text);
    headings.push({ level, text, slug });
  }

  return headings;
}
