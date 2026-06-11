import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MarkdownRenderer } from './MarkdownRenderer';

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
});
