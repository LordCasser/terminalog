import type { MDXComponents } from 'mdx/types'

/**
 * Global MDX Components Configuration
 * 
 * Defines custom styles for Markdown elements following Dracula Spectrum design system.
 * - Headings: Space Grotesk font, secondary-fixed-dim color
 * - Body text: Inter font, text-lg size
 * - Code blocks: JetBrains Mono font, bg-surface-container-lowest
 * - Blockquotes: JetBrains Mono font, border-l-4 border-primary
 * - Lists: JetBrains Mono font, with tertiary arrow symbols
 */

const components: MDXComponents = {
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