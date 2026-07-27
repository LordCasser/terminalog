import { describe, it, expect, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { cleanup } from '@testing-library/react';
import { MarkdownRenderer } from './MarkdownRenderer';

afterEach(cleanup);

describe('MarkdownRenderer', () => {
  it('renders plain markdown', () => {
    render(<MarkdownRenderer content="# Hello" />);
    expect(screen.getByText('Hello')).toBeDefined();
  });

  it('renders with TOC', () => {
    const content = '# Title\n\n[TOC]\n\n## Section A\n## Section B';
    render(<MarkdownRenderer content={content} />);
    // Title renders as h1 in beforeContent
    expect(screen.getByText('Title')).toBeDefined();
    // Section headings appear in both TOC list and afterContent h2
    expect(screen.getAllByText('Section A').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Section B').length).toBeGreaterThanOrEqual(1);
  });

  it('does not load syntax highlighting when no code blocks', () => {
    // No fenced code -> rehype-highlight is never imported.
    render(<MarkdownRenderer content="# No code here\n\nJust text." />);
    const codeEl = document.querySelector('code.hljs');
    expect(codeEl).toBeNull();
  });

  it('loads syntax highlighting for code blocks', async () => {
    render(<MarkdownRenderer content={'```js\nconst x = 1;\n```'} />);
    // rehype-highlight is lazy-loaded; wait for it to apply hljs class.
    await waitFor(() => {
      const highlighted = document.querySelector('code.hljs');
      expect(highlighted).not.toBeNull();
    }, { timeout: 5000 });
  });

  it('loads katex for math formulas', async () => {
    render(<MarkdownRenderer content={'$E=mc^2$'} />);
    // remark-math + rehype-katex are lazy-loaded; wait for .katex element.
    await waitFor(() => {
      const katex = document.querySelector('.katex');
      expect(katex).not.toBeNull();
    }, { timeout: 5000 });
  });
});
