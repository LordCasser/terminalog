/**
 * MarkdownRenderer Component
 * 
 * Renders remote Markdown content fetched from API.
 * Features:
 * - GitHub Flavored Markdown support (remark-gfm)
 * - Code syntax highlighting (rehype-highlight)
 * - Math formula rendering (remark-math + rehype-katex)
 * - Image path transformation (relative → API paths)
 * - Dracula Spectrum styling (aligned with article view prototype HTML)
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
 */
function transformImagePath(src: string, basePath?: string): string {
  if (src.startsWith('http://') || src.startsWith('https://')) {
    return src;
  }
  
  if (src.startsWith('/')) {
    return `/api/assets${src}`;
  }
  
  let normalizedPath = src;
  if (normalizedPath.startsWith('./')) {
    normalizedPath = normalizedPath.slice(2);
  }
  
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
          pre: ({ children, ...props }) => {
            // Extract language from code element's className
            const childCode = Array.isArray(children) ? children[0] : children;
            const codeProps = (childCode as any)?.props || {};
            const className = codeProps.className || '';
            const langMatch = /language-(\w+)/.exec(className);
            const language = langMatch ? langMatch[1] : '';
            
            return (
              <section className="bg-surface-container-lowest p-8 relative">
                {language && (
                  <div className="absolute top-0 right-0 p-4 font-mono text-xs text-outline-variant">
                    {language}
                  </div>
                )}
                <pre className="font-mono text-sm overflow-x-auto text-on-surface leading-loose">
                  <code className={className}>
                    {children}
                  </code>
                </pre>
              </section>
            );
          },
          
          // Inline code (no className) vs code block
          code: ({ className, children, ...props }) => {
            // Inline code (no className)
            if (!className) {
              return (
                <code className="font-mono text-sm bg-surface-container-low text-tertiary px-2 py-0.5">
                  {children}
                </code>
              );
            }
            
            // Code block (has className like language-xxx) - rendered inside pre
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
          
          // Lists - JetBrains Mono font, space-y-4, with tertiary arrow symbols
          ul: ({ children }) => (
            <ul className="space-y-4 font-mono mb-6">
              {children}
            </ul>
          ),
          
          ol: ({ children }) => (
            <ol className="space-y-4 font-mono mb-6 list-decimal list-inside">
              {children}
            </ol>
          ),
          
          li: ({ children }) => (
            <li className="flex items-start gap-4">
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