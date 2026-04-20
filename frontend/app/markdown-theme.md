# Markdown Theme System

The markdown rendering system uses a fully modular CSS variable architecture. All visual properties of markdown content are controlled by `--md-*` CSS custom properties defined in `globals.css` under the `:root` selector.

## Architecture

```
:root {
  --md-* variables    ← Theme tokens (colors, spacing, fonts)
}

.markdown-body        ← Container using theme tokens
.markdown-code-block  ← Code block component
.markdown-blockquote  ← Blockquote component
.markdown-list*       ← List components
.markdown-table*      ← Table components
.markdown-inline-code← Inline code component
.hljs*               ← Syntax highlighting tokens
```

## How to Change Themes

Create a theme file (e.g., `frontend/app/markdown-theme-light.css`) that redefines the `--md-*` variables:

```css
/* Example: Light theme override */
:root {
  --md-color-text: #1a1a2e;
  --md-color-heading: #0d0d1a;
  --md-color-heading-2: #6b2fa0;
  --md-code-block-bg: #f5f5f5;
  --md-inline-code-bg: #e8e8f0;
  --md-inline-code-color: #2d5a2d;
  --md-inline-code-border: 1px solid #c0c0d0;
  --md-blockquote-border: 4px solid #6b2fa0;
  --md-blockquote-color: #4a1a7a;
  --md-hl-comment: #6a737d;
  --md-hl-keyword: #d73a49;
  --md-hl-string: #032f62;
  --md-hl-number: #005cc5;
}
```

Then import it **after** `globals.css` in `layout.tsx`:

```tsx
import "./globals.css";
import "./markdown-theme-light.css";  // ← Override
```

## Variable Reference

### Typography
| Variable | Default | Description |
|---|---|---|
| `--md-font-size` | `1.05rem` | Body text size |
| `--md-line-height` | `1.9` | Body line height |
| `--md-font-body` | `var(--font-body)` | Body font family |
| `--md-font-headline` | `var(--font-headline)` | Heading font family |
| `--md-font-mono` | `var(--font-mono)` | Code font family |

### Colors — Text
| Variable | Default | Description |
|---|---|---|
| `--md-color-text` | `var(--color-on-surface-variant)` | Body text color |
| `--md-color-heading` | `var(--color-on-surface)` | H1/H3/H4 heading color |
| `--md-color-heading-2` | `var(--color-secondary-fixed-dim)` | H2 heading color |
| `--md-color-strong` | `var(--color-primary-fixed)` | Bold text color |
| `--md-color-link` | `var(--color-primary)` | Link color |
| `--md-color-link-hover` | `var(--color-secondary)` | Link hover color |
| `--md-color-emphasis` | `var(--color-outline)` | Italic text color |

### Code Blocks
| Variable | Default | Description |
|---|---|---|
| `--md-code-block-bg` | `var(--color-surface-lowest)` | Code block background |
| `--md-code-block-text` | `var(--color-on-surface)` | Code text color |
| `--md-code-block-label-color` | `var(--color-outline-variant)` | Language label color |
| `--md-code-block-padding` | `1.25rem` | Code block inner padding |
| `--md-code-block-margin` | `1.5rem 0` | Code block margin |
| `--md-code-font-size` | `0.875rem` | Code font size |
| `--md-code-line-height` | `1.85` | Code line height |

### Inline Code
| Variable | Default | Description |
|---|---|---|
| `--md-inline-code-bg` | `rgba(39,41,53,0.92)` | Inline code background |
| `--md-inline-code-color` | `var(--color-tertiary-fixed)` | Inline code text color |
| `--md-inline-code-border` | `1px solid rgba(74,68,81,0.7)` | Inline code border |
| `--md-inline-code-padding` | `0.125rem 0.45rem` | Inline code padding |
| `--md-inline-code-font-size` | `0.9rem` | Inline code font size |
| `--md-inline-code-shadow` | `inset 0 -1px 0 rgba(225,225,241,0.06)` | Inline code shadow |

### Blockquotes
| Variable | Default | Description |
|---|---|---|
| `--md-blockquote-border` | `4px solid var(--color-primary)` | Left border |
| `--md-blockquote-bg` | `linear-gradient(...)` | Background |
| `--md-blockquote-color` | `var(--color-primary)` | Text color |
| `--md-blockquote-padding` | `1rem 0 1rem 1.5rem` | Inner padding |
| `--md-blockquote-margin` | `1.5rem 0` | Outer margin |

### Syntax Highlighting
| Variable | Default | Description |
|---|---|---|
| `--md-hl-comment` | `var(--color-outline)` | Comments & quotes |
| `--md-hl-keyword` | `var(--color-secondary-fixed-dim)` | Keywords & literals |
| `--md-hl-title` | `var(--color-primary-fixed)` | Titles & attributes |
| `--md-hl-string` | `#ffd6a5` | Strings & regexps |
| `--md-hl-number` | `#ff9e64` | Numbers & symbols |
| `--md-hl-builtin` | `#8be9fd` | Built-ins & types |
| `--md-hl-function` | `var(--color-tertiary-fixed)` | Functions & properties |
| `--md-hl-meta` | `#7dcfff` | Meta & tags |

### Tables
| Variable | Default | Description |
|---|---|---|
| `--md-table-bg` | `var(--color-surface-lowest)` | Table background |
| `--md-table-border` | `2px solid var(--color-outline-variant)` | Border style |
| `--md-table-header-bg` | `var(--color-surface-container-highest)` | Header row bg |
| `--md-table-header-color` | `var(--color-secondary-fixed-dim)` | Header text color |
| `--md-table-row-odd` | `rgba(29,31,43,0.92)` | Odd row bg |
| `--md-table-row-even` | `rgba(11,14,24,0.98)` | Even row bg |
| `--md-table-row-hover` | `rgba(50,52,64,0.98)` | Hover row bg |

## Bug Fixes Applied

### CSS Specificity Fix
The `.markdown-body pre` selector (specificity 0-1-1) was overriding `.markdown-code-block__pre` (0-1-0), causing double padding (2rem from section + 2rem from pre = 64px). Fixed by resetting `.markdown-body pre` to `padding: 0; margin: 0;` and letting `.markdown-code-block` handle all spacing.

### Code Block Detection Fix
Code blocks without a language specification (e.g., bare triple backticks) were incorrectly receiving the `markdown-inline-code` class, which added a visible 1px border inside the code block. The `MarkdownRenderer.tsx` `code` component now detects multiline content as code blocks.