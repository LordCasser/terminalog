/**
 * MermaidBlock Component
 *
 * Renders a Mermaid diagram within an article. Loads beautiful-mermaid lazily
 * via dynamic import() to keep it out of the main bundle.
 *
 * States:
 *   - loading  → Skeleton placeholder
 *   - success  → SVG rendered via dangerouslySetInnerHTML
 *   - error    → Red error card with raw code fallback
 *
 * Props:
 *   code — raw Mermaid source text
 */

'use client';

import { useState, useEffect } from 'react';
import { MERMAID_RENDER_OPTIONS } from '@/lib/mermaid';

interface MermaidBlockProps {
  code: string;
}

export function MermaidBlock({ code }: MermaidBlockProps) {
  const [state, setState] = useState<'loading' | 'done' | 'error'>('loading');
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function render() {
      try {
        const { renderMermaidSVG } = await import('beautiful-mermaid');
        if (cancelled) return;
        const result = renderMermaidSVG(code.trim(), { ...MERMAID_RENDER_OPTIONS });
        if (cancelled) return;
        setSvg(result);
        setState('done');
      } catch (err: unknown) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : String(err);
        setError(message);
        setState('error');
      }
    }

    render();

    return () => {
      cancelled = true;
    };
  }, [code]);

  // --- Loading ---
  if (state === 'loading') {
    return (
      <div
        role="status"
        aria-label="Rendering diagram..."
        className="mermaid-loading"
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '48px 24px',
          borderRadius: 12,
          background: 'var(--color-surface-container-low)',
        }}
      >
        <div
          className="mermaid-loading__skeleton"
          style={{
            width: '60%',
            height: 12,
            borderRadius: 6,
            background: 'var(--color-surface-container-highest)',
            marginBottom: 8,
          }}
        />
        <div
          className="mermaid-loading__skeleton"
          style={{
            width: '80%',
            height: 12,
            borderRadius: 6,
            background: 'var(--color-surface-container-highest)',
            marginBottom: 8,
          }}
        />
        <div
          className="mermaid-loading__skeleton"
          style={{
            width: '50%',
            height: 12,
            borderRadius: 6,
            background: 'var(--color-surface-container-highest)',
            marginBottom: 16,
          }}
        />
        <span style={{ color: 'var(--color-outline)', fontSize: '0.875rem' }}>
          Rendering diagram...
        </span>
      </div>
    );
  }

  // --- Error ---
  if (state === 'error') {
    return (
      <div
        role="alert"
        className="mermaid-error"
        style={{
          padding: '20px 24px',
          borderRadius: 12,
          border: '1px solid var(--color-error)',
          background: 'var(--color-surface-container-low)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginBottom: 12,
            color: 'var(--color-error)',
            fontWeight: 600,
            fontSize: '0.875rem',
          }}
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <circle cx="8" cy="8" r="7" stroke="currentColor" strokeWidth="1.5" />
            <path d="M8 4.5v4M8 11.5h.01" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
          <span>Diagram render error</span>
        </div>
        {error && (
          <p
            style={{
              margin: '0 0 12px 0',
              fontSize: '0.8rem',
              color: 'var(--color-outline)',
              fontFamily: 'monospace',
            }}
          >
            {error}
          </p>
        )}
        <pre
          style={{
            margin: 0,
            padding: '12px 16px',
            borderRadius: 8,
            background: 'var(--color-surface-lowest)',
            color: 'var(--color-on-surface)',
            fontSize: '0.8rem',
            overflowX: 'auto',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          <code>{code}</code>
        </pre>
      </div>
    );
  }

  // --- Success ---
  return (
    <div
      role="img"
      aria-label="Mermaid diagram"
      className="mermaid-diagram"
      style={{
        display: 'flex',
        justifyContent: 'center',
        padding: '16px 0',
      }}
      dangerouslySetInnerHTML={{ __html: svg ?? '' }}
    />
  );
}
