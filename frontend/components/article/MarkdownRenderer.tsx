/**
 * MarkdownRenderer Component
 * 
 * Renders remote Markdown content fetched from API.
 * Features:
 * - GitHub Flavored Markdown support (remark-gfm)
 * - GitHub Alerts / Callouts (remark-github-blockquote-alert)
 * - Code syntax highlighting (rehype-highlight)
 * - Math formula rendering (remark-math + rehype-katex)
 * - Image path transformation (relative → /api/v1/assets/ paths)
 * - Dracula Spectrum styling (aligned with article view prototype HTML)
 * 
 * Supported alert types:
 *   > [!NOTE]     → informational callout
 *   > [!TIP]      → helpful suggestion
 *   > [!IMPORTANT] → critical information
 *   > [!WARNING]  → cautionary advice
 *   > [!CAUTION]  → potential danger
 * 
 * Usage:
 * <MarkdownRenderer content={markdownContent} basePath="tech/blog" />
 */

/* eslint-disable @next/next/no-img-element */

"use client";

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import { remarkAlert } from 'remark-github-blockquote-alert';
import 'katex/dist/katex.min.css';
import { isValidElement } from 'react';
import type { ComponentPropsWithoutRef, ReactElement, ReactNode } from 'react';

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
 *    e.g., "./.assets/images/photo.png" + basePath="guides" → "/api/v1/assets/guides/images/photo.png"
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

export function MarkdownRenderer({ content, basePath }: MarkdownRendererProps) {
  return (
    <div className="markdown-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath, remarkAlert]}
        rehypePlugins={[rehypeHighlight, rehypeKatex]}
        components={{
          // Headings - Space Grotesk font
          h1: ({ children }) => (
            <h1 className="font-headline text-4xl md:text-5xl font-bold leading-none text-on-surface tracking-tighter mb-8 mt-12 first:mt-0">
              {children}
            </h1>
          ),
          
          h2: ({ children }) => (
            <h2 className="font-headline text-3xl font-bold text-secondary-fixed-dim">
              {children}
            </h2>
          ),
          
          h3: ({ children }) => (
            <h3 className="font-headline text-2xl font-bold text-secondary-fixed mb-4">
              {children}
            </h3>
          ),
          
          h4: ({ children }) => (
            <h4 className="font-headline text-xl font-bold text-on-surface mb-4">
              {children}
            </h4>
          ),
          
          // Body text - inherits color/leading from article container
          p: ({ children }) => (
            <p className="mb-4">
              {children}
            </p>
          ),
          
          // Code blocks - wrapped in section with bg-surface-container-lowest, language tag, p-8
          pre: ({ children }) => {
            // Extract language from code element's className
            const childCode = (Array.isArray(children) ? children[0] : children) as ReactElement<ComponentPropsWithoutRef<'code'>> | ReactNode;
            const codeProps = isValidElement<ComponentPropsWithoutRef<'code'>>(childCode)
              ? childCode.props
              : undefined;
            const className = codeProps?.className || '';
            const language = extractLanguageLabel(className);
            
            return (
              <section className="markdown-code-block">
                {language && (
                  <div className="markdown-code-block__label">
                    {language}
                  </div>
                )}
                <pre className="markdown-code-block__pre">{children}</pre>
              </section>
            );
          },
          
          // Inline code vs code block detection
          //   - className includes "hljs"/"language-"  → code block (highlight.js)
          //   - no className + multiline content        → code block (plain fenced)
          //   - no className + single-line content      → inline code
          code: ({ className, children, ...props }) => {
            const textContent = typeof children === 'string'
              ? children
              : Array.isArray(children)
                ? children.join('')
                : '';
            const isMultilineCodeBlock = !className && textContent.includes('\n');
            
            if (!className && !isMultilineCodeBlock) {
              return (
                <code className="markdown-inline-code">
                  {children}
                </code>
              );
            }
            
            return (
              <code className={className} {...props}>
                {children}
              </code>
            );
          },
          
          // Blockquotes - cleaner editorial pull quote treatment
          blockquote: ({ children }) => (
            <blockquote className="markdown-blockquote">
              <div className="markdown-blockquote__content">{children}</div>
            </blockquote>
          ),
          
          // Lists - unordered and ordered lists use different visual treatments
          ul: ({ children }) => (
            <ul className="markdown-list markdown-list--unordered">
              {children}
            </ul>
          ),
          
          ol: ({ children }) => (
            <ol className="markdown-list markdown-list--ordered">
              {children}
            </ol>
          ),
          
          li: ({ children }) => (
            <li className="markdown-list__item">
              <div className="markdown-list__content">
                {children}
              </div>
            </li>
          ),

          table: ({ children }) => (
            <div className="markdown-table-shell my-10">
              <div className="markdown-table-scroll">
                <table className="markdown-table">
                  {children}
                </table>
              </div>
            </div>
          ),

          thead: ({ children }) => (
            <thead className="markdown-table-head">
              {children}
            </thead>
          ),

          tbody: ({ children }) => (
            <tbody className="markdown-table-body">
              {children}
            </tbody>
          ),

          tr: ({ children }) => (
            <tr className="markdown-table-row">
              {children}
            </tr>
          ),

          th: ({ children }) => (
            <th className="markdown-table-heading">
              {children}
            </th>
          ),

          td: ({ children }) => (
            <td className="markdown-table-cell">
              {children}
            </td>
          ),
          
          // Links
          a: ({ href, children }) => {
            if (href?.startsWith('http')) {
              return (
                <a href={href} target="_blank" rel="noopener noreferrer" className="text-primary hover:text-secondary transition-colors underline">
                  {children}
                </a>
              );
            }
            return (
              <a href={href} className="text-primary hover:text-secondary transition-colors underline">
                {children}
              </a>
            );
          },
          
          // Images - Transform paths
          img: ({ src, alt, ...props }) => {
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
          
          // Horizontal rule
          hr: () => (
            <hr className="border-t border-surface-container my-8" />
          ),
          
          // Strong and emphasis
          strong: ({ children }) => (
            <strong className="text-primary font-bold">
              {children}
            </strong>
          ),
          
          em: ({ children }) => (
            <em className="text-outline italic">
              {children}
            </em>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
