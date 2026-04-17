/**
 * MarkdownRenderer Component
 * 
 * Renders remote Markdown content fetched from API.
 * Features:
 * - GitHub Flavored Markdown support (remark-gfm)
 * - Code syntax highlighting (rehype-highlight)
 * - Math formula rendering (remark-math + rehype-katex)
 * - Image path transformation (relative → API paths)
 * - Dracula Spectrum styling (consistent with mdx-components.tsx)
 * 
 * Usage:
 * <MarkdownRenderer content={markdownContent} basePath="tech/blog" />
 */

"use client";

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import 'katex/dist/katex.min.css';

interface MarkdownRendererProps {
  content: string;
  basePath?: string;  // Article directory path for image transformation
}

/**
 * Transform Markdown image paths to API resource paths
 * 
 * Rules:
 * 1. External links (http/https): keep as-is
 * 2. Relative paths: convert to /api/assets/{basePath}/{imagePath}
 * 
 * @example
 * transformImagePath("./images/photo.png", "tech/blog")
 * // Returns: "/api/assets/tech/blog/images/photo.png"
 */
function transformImagePath(src: string, basePath?: string): string {
  // External links: keep as-is
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  
  // Absolute path (relative to repo root)
  if (src.startsWith('/')) {
    return `/api/assets${src}`;
  }
  
  // Relative path: convert to API path
  let normalizedPath = src;
  
  // Remove ./ prefix
  if (normalizedPath.startsWith('./')) {
    normalizedPath = normalizedPath.slice(2);
  }
  
  // Combine basePath and image path
  const fullPath = basePath 
    ? `${basePath}/${normalizedPath}` 
    : normalizedPath;
  
  return `/api/assets/${fullPath}`;
}

export function MarkdownRenderer({ content, basePath }: MarkdownRendererProps) {
  return (
    <div className="markdown-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={[rehypeHighlight, rehypeKatex]}
        components={{
          // Headings - Space Grotesk font
          h1: ({ children }) => (
            <h1 className="font-headline text-6xl font-bold text-on-surface tracking-tighter mb-8">
              {children}
            </h1>
          ),
          
          h2: ({ children }) => (
            <h2 className="font-headline text-3xl font-bold text-secondary-fixed-dim mb-6">
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
          
          // Body text - Inter font, text-lg
          p: ({ children }) => (
            <p className="text-lg text-on-surface-variant mb-4 leading-relaxed">
              {children}
            </p>
          ),
          
          // Code blocks - JetBrains Mono font
          pre: ({ children }) => (
            <pre className="font-mono text-sm bg-surface-container-lowest text-on-surface p-4 rounded-none overflow-x-auto mb-6">
              {children}
            </pre>
          ),
          
          code: ({ className, children, ...props }) => {
            // Inline code (no className)
            if (!className) {
              return (
                <code className="font-mono text-sm bg-surface-container-low text-tertiary px-2 py-0.5">
                  {children}
                </code>
              );
            }
            
            // Code block (has className like language-xxx)
            return (
              <code className={className} {...props}>
                {children}
              </code>
            );
          },
          
          // Blockquotes - JetBrains Mono font, border-l-4 border-primary
          blockquote: ({ children }) => (
            <blockquote className="font-mono text-lg text-primary border-l-4 border-primary bg-surface-container-low pl-6 py-2 italic mb-6">
              {children}
            </blockquote>
          ),
          
          // Lists - JetBrains Mono font, with tertiary arrow symbols
          ul: ({ children }) => (
            <ul className="font-mono text-base space-y-2 mb-6">
              {children}
            </ul>
          ),
          
          ol: ({ children }) => (
            <ol className="font-mono text-base space-y-2 mb-6 list-decimal list-inside">
              {children}
            </ol>
          ),
          
          li: ({ children }) => (
            <li className="text-on-surface-variant flex items-start gap-4">
              <span className="text-tertiary">➜</span>
              <span>{children}</span>
            </li>
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
            // Ensure src is string (React img src can be string | Blob)
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