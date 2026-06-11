/**
 * mermaid.ts — Configuration module for beautiful-mermaid RenderOptions.
 *
 * Imports THEMES from beautiful-mermaid (lightweight static object, no layout engine).
 * The actual SVG rendering happens via dynamic import() in MermaidBlock.tsx.
 */

import { THEMES } from 'beautiful-mermaid';

const d = THEMES.dracula;

/**
 * Mermaid diagram render options using Dracula dark theme palette.
 *
 * - bg/fg/line/accent/muted from THEMES.dracula
 * - surface and border auto-derive from bg+fg (color-mix fallback)
 * - transparent=true so diagrams blend into the page background,
 *   while Dracula colors color the diagram elements.
 */
export const MERMAID_RENDER_OPTIONS = {
  bg: d.bg,           // "#282a36" — base for surface/border derivation
  fg: d.fg,           // "#f8f8f2" — bright text
  line: d.line,       // "#6272a4" — edges/connectors
  accent: d.accent,   // "#bd93f9" — arrow heads, highlights
  muted: d.muted,     // "#6272a4" — secondary text, edge labels
  transparent: true,   // SVG background transparent (page bg shows through)
} as const;
