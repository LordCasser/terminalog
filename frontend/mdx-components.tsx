import type { MDXComponents } from 'mdx/types'

/**
 * Global MDX Components Configuration
 * 
 * Defines custom styles for Markdown elements following Dracula Spectrum design system,
 * aligned with the article view prototype HTML.
 * - Headings: Space Grotesk font, secondary-fixed-dim color
 * - Body text: inherits from article container (text-on-surface-variant leading-relaxed)
 * - Code blocks: wrapped in section with bg-surface-container-lowest, language tag top-right, p-8 padding
 * - Inline code: JetBrains Mono, bg-surface-container-low, text-tertiary
 * - Blockquotes: JetBrains Mono font, border-l-4 border-primary
 * - Lists: JetBrains Mono font, space-y-4, with tertiary arrow symbols
 * - Paragraphs: minimal styling, inherits color/leading from container
 */

const components: MDXComponents = {
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
    // Extract language from className
    const className = (props as any)?.className || '';
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
  
  // Inline code
  code: ({ children }) => (
    <code className="font-mono text-sm bg-surface-container-low text-tertiary px-2 py-0.5">
      {children}
    </code>
  ),
  
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
      )
    }
    return (
      <a href={href} className="text-primary hover:text-secondary transition-colors underline">
        {children}
      </a>
    )
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
}

export function useMDXComponents(): MDXComponents {
  return components
}