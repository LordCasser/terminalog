/**
 * mermaid.ts — Pure configuration module for beautiful-mermaid RenderOptions.
 *
 * DO NOT import 'beautiful-mermaid' here. This module only exports constants.
 * The actual rendering happens via dynamic import() in MermaidBlock.tsx.
 */

export const MERMAID_RENDER_OPTIONS = {
  theme_name: 'dracula',
  transparent: true,
} as const;
