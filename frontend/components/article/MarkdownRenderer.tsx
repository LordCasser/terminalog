/**
 * MarkdownRenderer Component
 * 
 * Renders remote Markdown content fetched from API.
 * Features:
 * - GitHub Flavored Markdown support (remark-gfm)
 * - GitHub Alerts / Callouts (remark-github-blockquote-alert)
 * - Code syntax highlighting (rehype-highlight)
 * - Math formula rendering (remark-math + rehype-katex)
 * - [TOC] / Table of Contents generation (remark-toc)
 * - Heading anchor links with smooth scroll (rehype-slug)
 * - Internal article link routing (next/link)
 * - External links open in new tab
 * - Image path transformation (relative -> /api/v1/assets/ paths)
 * - Dracula Spectrum styling (aligned with article view prototype HTML)
 * 
 * Supported alert types:
 *   > [!NOTE]     -> informational callout
 *   > [!TIP]      -> helpful suggestion
 *   > [!IMPORTANT] -> critical information
 *   > [!WARNING]  -> cautionary advice
 *   > [!CAUTION]  -> potential danger
 * 
 * Usage:
 * <MarkdownRenderer content={markdownContent} basePath="tech/blog" />
 */

/* eslint-disable @next/next/no-img-element */

"use client";

import ReactMarkdown, { type Components } from 'react-markdown';
import remarkGfm from 'remark-gfm';
// Heavy plugins (rehype-highlight ~100KB+, remark-math+rehype-katex+katex CSS ~30KB+)
// are lazy-loaded below based on content detection, not statically imported.
import { remarkAlert } from 'remark-github-blockquote-alert';
import rehypeSlug from 'rehype-slug';
import Link from 'next/link';
import { isValidElement, useMemo, lazy, Suspense, useState, useEffect } from 'react';
import { encodePathForUrl } from '@/lib/utils/path';
import { generateSlug, childrenToText, extractHeadings } from './toc';
import type { TocEntry } from './toc';
import type { ComponentPropsWithoutRef, ReactElement, ReactNode } from 'react';

const MermaidBlock = lazy(() =>
  import('./MermaidBlock').then((m) => ({ default: m.MermaidBlock }))
);

interface MarkdownRendererProps {
  content: string;
  basePath?: string;  // Article directory path for image transformation
}

/**
 * Transform Markdown image paths to API resource paths
 * 
 * Rules (RESTful v1 API):
 * 1. External links (http/https): keep as-is
 * 2. Absolute paths (/...): convert to /api/v1/assets{src}
 * 3. Relative paths starting with .assets: strip .assets layer for cleaner API path
 *    e.g., "./.assets/images/photo.png" + basePath="guides" -> "/api/v1/assets/guides/images/photo.png"
 * 4. Other relative paths: convert to /api/v1/assets/{basePath}/{imagePath}
 */
function transformImagePath(src: string, basePath?: string): string {
  // External links: keep as-is
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  
  // Absolute paths: add /api/v1/assets prefix
  if (src.startsWith('/')) {
    return `/api/v1/assets${src}`;
  }
  
  // Normalize relative path
  let normalizedPath = src;
  if (normalizedPath.startsWith('./')) {
    normalizedPath = normalizedPath.slice(2);
  }
  
  // Special handling for .assets directory: strip the .assets layer
  // This simplifies API path from /api/v1/assets/guides/.assets/images/photo.png
  // to /api/v1/assets/guides/images/photo.png
  if (normalizedPath.startsWith('.assets/')) {
    normalizedPath = normalizedPath.slice(8); // Remove ".assets/" prefix
  }
  
  const fullPath = basePath 
    ? `${basePath}/${normalizedPath}` 
    : normalizedPath;
  
  return `/api/v1/assets/${fullPath}`;
}

function extractLanguageLabel(className: string): string {
  const langMatch = /language-([A-Za-z0-9_+-]+)/.exec(className);
  if (!langMatch) {
    return '';
  }

  const language = langMatch[1].toLowerCase();
  const aliases: Record<string, string> = {
    js: 'JavaScript',
    ts: 'TypeScript',
    tsx: 'TSX',
    jsx: 'JSX',
    sh: 'Shell',
    bash: 'Bash',
    zsh: 'Zsh',
    yml: 'YAML',
    md: 'Markdown',
    plaintext: 'Plain Text',
    text: 'Plain Text',
    golang: 'Go',
    csharp: 'C#',
    cpp: 'C++',
  };

  return aliases[language] ?? language.toUpperCase();
}

/**
 * Process [TOC] markers in markdown content.
 * Only activates when an explicit [TOC] or [toc] marker is present on its own line.
 * 
 * Returns the content split at the [TOC] marker so the TOC can be rendered inline
 * at the correct position, between beforeContent and afterContent.
 */
function useTocProcessing(content: string): {
  beforeContent: string;
  afterContent: string;
  tocHeadings: TocEntry[] | null;
} {
  return useMemo(() => {
    const tocRegex = /^\[toc\]$/im;
    if (!tocRegex.test(content)) {
      return { beforeContent: content, afterContent: '', tocHeadings: null };
    }
    const parts = content.split(tocRegex, 2);
    // Remove fenced code blocks to avoid false heading matches
    const withoutCode = content.replace(/```[\s\S]*?```/g, '');
    const allHeadings = extractHeadings(withoutCode);
    // Filter: skip h1 (article title) and the TOC heading itself
    const tocHeadingPattern = /^(目录|table\s*of\s*contents|contents)$/i;
    const headings = allHeadings.filter(h =>
      h.level > 1 && !tocHeadingPattern.test(h.text)
    );
    return {
      beforeContent: parts[0],
      afterContent: parts[1] || '',
      tocHeadings: headings,
    };
  }, [content]);
}

/**
 * Handle anchor link click: smooth scroll + URL hash update.
 * Decodes URL-encoded href to match raw Unicode heading IDs generated by rehype-slug.
 */
function handleAnchorClick(e: React.MouseEvent, href: string) {
  e.preventDefault();
  const id = decodeURIComponent(href.slice(1));
  const target = document.getElementById(id);
  if (target) {
    target.scrollIntoView({ behavior: 'smooth' });
    window.history.pushState(null, '', href);
  }
}

/**
 * Shared markdown component overrides used by both ReactMarkdown instances.
 */
function markdownComponents(basePath?: string): Components {
  return {
    // ----- Headings with anchor links -----
    // Visual styling (size/color/font) is owned by .markdown-body h1..h6 in
    // globals.css via the --md-* theme vars. Here we only attach the id, a
    // `group` hook for anchor hover, and the anchor link itself.
    h1: ({ children, id }: { children?: ReactNode; id?: string }) => {
      const slug = id || generateSlug(childrenToText(children));
      return (
        <h1 id={slug} className="group">
          <a href={`#${slug}`}
             onClick={(e) => handleAnchorClick(e, `#${slug}`)}
             className="anchor-link focus-visible:opacity-100
                        mr-2 text-secondary no-underline select-none
                        transition-opacity duration-150"
             aria-hidden="true">#</a>
          {children}
        </h1>
      );
    },

    h2: ({ children, id }: { children?: ReactNode; id?: string }) => {
      const slug = id || generateSlug(childrenToText(children));
      return (
        <h2 id={slug} className="group">
          <a href={`#${slug}`}
             onClick={(e) => handleAnchorClick(e, `#${slug}`)}
             className="anchor-link focus-visible:opacity-100
                        mr-2 text-secondary no-underline select-none
                        transition-opacity duration-150"
             aria-hidden="true">#</a>
          {children}
        </h2>
      );
    },

    h3: ({ children, id }: { children?: ReactNode; id?: string }) => {
      const slug = id || generateSlug(childrenToText(children));
      return (
        <h3 id={slug} className="group">
          <a href={`#${slug}`}
             onClick={(e) => handleAnchorClick(e, `#${slug}`)}
             className="anchor-link focus-visible:opacity-100
                        mr-2 text-secondary no-underline select-none
                        transition-opacity duration-150"
             aria-hidden="true">#</a>
          {children}
        </h3>
      );
    },

    h4: ({ children, id }: { children?: ReactNode; id?: string }) => {
      const slug = id || generateSlug(childrenToText(children));
      return (
        <h4 id={slug} className="group">
          <a href={`#${slug}`}
             onClick={(e) => handleAnchorClick(e, `#${slug}`)}
             className="anchor-link focus-visible:opacity-100
                        mr-2 text-secondary no-underline select-none
                        transition-opacity duration-150"
             aria-hidden="true">#</a>
          {children}
        </h4>
      );
    },

    h5: ({ children, id }: { children?: ReactNode; id?: string }) => {
      const slug = id || generateSlug(childrenToText(children));
      return (
        <h5 id={slug} className="group">
          <a href={`#${slug}`}
             onClick={(e) => handleAnchorClick(e, `#${slug}`)}
             className="anchor-link focus-visible:opacity-100
                        mr-2 text-secondary no-underline select-none
                        transition-opacity duration-150"
             aria-hidden="true">#</a>
          {children}
        </h5>
      );
    },

    h6: ({ children, id }: { children?: ReactNode; id?: string }) => {
      const slug = id || generateSlug(childrenToText(children));
      return (
        <h6 id={slug} className="group">
          <a href={`#${slug}`}
             onClick={(e) => handleAnchorClick(e, `#${slug}`)}
             className="anchor-link focus-visible:opacity-100
                        mr-2 text-secondary no-underline select-none
                        transition-opacity duration-150"
             aria-hidden="true">#</a>
          {children}
        </h6>
      );
    },

    // ----- Body text -----
    p: ({ children }: { children?: ReactNode }) => (
      <p className="mb-4">{children}</p>
    ),

    // ----- Code blocks -----
    pre: ({ children }: { children?: ReactNode }) => {
      const childCode = (Array.isArray(children) ? children[0] : children) as ReactElement<ComponentPropsWithoutRef<'code'>> | ReactNode;
      const codeProps = isValidElement<ComponentPropsWithoutRef<'code'>>(childCode)
        ? childCode.props
        : undefined;
      const className = codeProps?.className || '';

      // Mermaid diagrams are self-contained — skip the code-block shell
      if (className?.includes('language-mermaid')) {
        return <>{children}</>;
      }

      const language = extractLanguageLabel(className);

      return (
        <section className="markdown-code-block">
          {language && (
            <div className="markdown-code-block__label">{language}</div>
          )}
          <pre className="markdown-code-block__pre">{children}</pre>
        </section>
      );
    },

    code: ({ className, children, ...props }: ComponentPropsWithoutRef<'code'>) => {
      // --- Mermaid diagram detection (must precede other code-block logic) ---
      if (className?.includes('language-mermaid')) {
        const mermaidCode = typeof children === 'string'
          ? children
          : Array.isArray(children)
            ? children.join('')
            : '';
        return (
          <Suspense fallback={
            <div style={{
              padding: '48px 24px',
              borderRadius: 12,
              background: 'var(--color-surface-container-low)',
              color: 'var(--color-outline)',
              textAlign: 'center',
            }}>
              Loading diagram...
            </div>
          }>
            <MermaidBlock code={mermaidCode} />
          </Suspense>
        );
      }

      const textContent = typeof children === 'string'
        ? children
        : Array.isArray(children)
          ? children.join('')
          : '';
      const isMultilineCodeBlock = !className && textContent.includes('\n');

      if (!className && !isMultilineCodeBlock) {
        return (
          <code>{children}</code>
        );
      }

      return (
        <code className={className} {...props}>{children}</code>
      );
    },

    // ----- Blockquotes -----
    blockquote: ({ children }: { children?: ReactNode }) => (
      <blockquote className="markdown-blockquote">
        <div className="markdown-blockquote__content">{children}</div>
      </blockquote>
    ),

    // ----- Lists -----
    ul: ({ children }: { children?: ReactNode }) => (
      <ul className="markdown-list markdown-list--unordered">{children}</ul>
    ),

    ol: ({ children }: { children?: ReactNode }) => (
      <ol className="markdown-list markdown-list--ordered">{children}</ol>
    ),

    li: ({ children }: { children?: ReactNode }) => (
      <li className="markdown-list__item">
        <div className="markdown-list__content">{children}</div>
      </li>
    ),

    // ----- Tables -----
    table: ({ children }: { children?: ReactNode }) => (
      <div className="markdown-table-shell my-10">
        <div className="markdown-table-scroll">
          <table className="markdown-table">{children}</table>
        </div>
      </div>
    ),

    thead: ({ children }: { children?: ReactNode }) => (
      <thead className="markdown-table-head">{children}</thead>
    ),

    tbody: ({ children }: { children?: ReactNode }) => (
      <tbody className="markdown-table-body">{children}</tbody>
    ),

    tr: ({ children }: { children?: ReactNode }) => (
      <tr className="markdown-table-row">{children}</tr>
    ),

    th: ({ children }: { children?: ReactNode }) => (
      <th className="markdown-table-heading">{children}</th>
    ),

    td: ({ children }: { children?: ReactNode }) => (
      <td className="markdown-table-cell">{children}</td>
    ),

    // ----- Links (3 routing modes) -----
    a: ({ href, children }: { href?: string; children?: ReactNode }) => {
      // 1. External links -> new tab
      if (href?.startsWith('http')) {
        return (
          <a href={href} target="_blank" rel="noopener noreferrer"
             className="text-primary hover:text-secondary transition-colors underline">
            {children}
          </a>
        );
      }

      // 2. Anchor links (including TOC links) -> smooth scroll.
      //    Decode URI component to match rehype-slug's raw Unicode heading IDs.
      if (href?.startsWith('#')) {
        const decodedHref = '#' + decodeURIComponent(href.slice(1));
        return (
          <a href={decodedHref}
             onClick={(e) => handleAnchorClick(e, decodedHref)}
             className="text-primary hover:text-secondary transition-colors no-underline hover:underline cursor-pointer">
            {children}
          </a>
        );
      }

      // 3. Internal article links -> SPA navigation via next/link
      let articlePath = href || '';
      articlePath = articlePath.replace(/\.md$/i, '');
      articlePath = articlePath.replace(/^\.\.?\//, '');
      const targetPath = `/article/${encodePathForUrl(articlePath)}`;

      return (
        <Link href={targetPath}
              className="text-primary hover:text-secondary transition-colors underline">
          {children}
        </Link>
      );
    },

    // ----- Images -----
    img: ({ src, alt, ...props }: ComponentPropsWithoutRef<'img'>) => {
      if (!src) return null;
      const srcString = typeof src === 'string' ? src : '';
      const transformedSrc = transformImagePath(srcString, basePath);
      return (
        <img
          src={transformedSrc}
          alt={alt}
          loading="lazy"
          className="max-w-full h-auto"
          {...props}
        />
      );
    },

    // ----- Horizontal rule -----
    hr: () => (
      <hr className="border-t border-surface-container my-8" />
    ),

    // ----- Inline formatting -----
    strong: ({ children }: { children?: ReactNode }) => (
      <strong className="text-primary font-bold">{children}</strong>
    ),

    em: ({ children }: { children?: ReactNode }) => (
      <em className="text-outline italic">{children}</em>
    ),
  };
}

export function MarkdownRenderer({ content, basePath }: MarkdownRendererProps) {
  // Pre-process [TOC] markers. Content is split at the marker so the TOC renders
  // inline at its actual position in the document (not forcibly at the top).
  const { beforeContent, afterContent, tocHeadings } = useTocProcessing(content);

  const tocMinLevel = tocHeadings ? Math.min(...tocHeadings.map(h => h.level)) : 0;
  const components = markdownComponents(basePath);

  // --- Lazy-load heavy markdown plugins based on content detection ---
  // Avoids loading highlight.js (~100KB+) and KaTeX (~30KB+) when the article
  // contains no code blocks or math. Plugins load asynchronously; until ready,
  // content renders in a degraded-but-readable form (plain code, raw $...$).
  const hasMath = useMemo(() => /\$/.test(content), [content]);
  const hasCode = useMemo(() => /```/.test(content), [content]);

  const [mathPlugins, setMathPlugins] = useState<{ remark: unknown; rehype: unknown } | null>(null);
  const [highlightPlugin, setHighlightPlugin] = useState<unknown>(null);

  useEffect(() => {
    if (!hasMath) {
      setMathPlugins(null);
      return;
    }
    let cancelled = false;
    Promise.all([
      import('remark-math'),
      import('rehype-katex'),
      import('katex/dist/katex.min.css'),
    ]).then(([rm, rk]) => {
      if (!cancelled) setMathPlugins({ remark: rm.default, rehype: rk.default });
    });
    return () => { cancelled = true; };
  }, [hasMath]);

  useEffect(() => {
    if (!hasCode) {
      setHighlightPlugin(null);
      return;
    }
    let cancelled = false;
    import('rehype-highlight').then(({ default: rh }) => {
      // Wrap in arrow so React doesn't treat the plugin function as a state updater.
      if (!cancelled) setHighlightPlugin(() => rh);
    });
    return () => { cancelled = true; };
  }, [hasCode]);

  const remarkPlugins = useMemo(() => {
    const arr: unknown[] = [remarkGfm, remarkAlert];
    if (mathPlugins) arr.push(mathPlugins.remark);
    return arr;
  }, [mathPlugins]);

  const rehypePlugins = useMemo(() => {
    const arr: unknown[] = [rehypeSlug];
    if (highlightPlugin) arr.push(highlightPlugin);
    if (mathPlugins) arr.push(mathPlugins.rehype);
    return arr;
  }, [highlightPlugin, mathPlugins]);

  return (
    <div className="markdown-body">
      {/* Content before [TOC] marker */}
      <ReactMarkdown
        remarkPlugins={remarkPlugins as never}
        rehypePlugins={rehypePlugins as never}
        components={components}
      >
        {beforeContent}
      </ReactMarkdown>

      {/* Inline Table of Contents - rendered at exact [TOC] marker position */}
      {tocHeadings && tocHeadings.length > 0 && (
        <ul className="markdown-toc">
          {tocHeadings.map((h) => (
            <li key={h.slug} style={{ paddingLeft: `${(h.level - tocMinLevel) * 1.5}em` }}>
              <a href={`#${h.slug}`}
                 onClick={(e) => handleAnchorClick(e, `#${h.slug}`)}
                 className="text-primary hover:text-secondary transition-colors no-underline hover:underline cursor-pointer">
                {h.text}
              </a>
            </li>
          ))}
        </ul>
      )}

      {/* Content after [TOC] marker */}
      <ReactMarkdown
        remarkPlugins={remarkPlugins as never}
        rehypePlugins={rehypePlugins as never}
        components={components}
      >
        {afterContent}
      </ReactMarkdown>
    </div>
  );
}
