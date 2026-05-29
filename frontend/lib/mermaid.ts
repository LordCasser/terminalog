/**
 * mermaid.ts — Pure configuration module for beautiful-mermaid RenderOptions.
 *
 * DO NOT import 'beautiful-mermaid' here. This module only exports constants.
 * The actual rendering happens via dynamic import() in MermaidBlock.tsx.
 */

export const MERMAID_RENDER_OPTIONS = {
  bg: 'transparent',
  fg: 'var(--color-on-surface)',
  accent: 'var(--color-primary)',
  muted: 'var(--color-outline)',
  surface: 'var(--color-surface-container-high)',
  border: 'var(--color-outline-variant)',
  transparent: true,
} as const;
