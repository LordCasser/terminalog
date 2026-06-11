import { describe, it, expect } from 'vitest';
import { generateSlug, stripInlineMarkdown, extractHeadings, childrenToText } from './toc';

describe('generateSlug', () => {
  it('converts simple text to slug', () => {
    expect(generateSlug('Hello World')).toBe('hello-world');
  });
  it('handles special characters', () => {
    expect(generateSlug('C++ & Rust')).toBe('c-rust');
  });
  it('handles Chinese characters (removes them)', () => {
    expect(generateSlug('你好 World')).toBe('world');
  });
});

describe('stripInlineMarkdown', () => {
  it('strips bold', () => {
    expect(stripInlineMarkdown('**hello** world')).toBe('hello world');
  });
  it('strips inline code', () => {
    expect(stripInlineMarkdown('`code` example')).toBe('code example');
  });
  it('strips links', () => {
    expect(stripInlineMarkdown('[link](url) text')).toBe('link text');
  });
  it('returns plain text unchanged', () => {
    expect(stripInlineMarkdown('plain text')).toBe('plain text');
  });
});

describe('extractHeadings', () => {
  it('extracts headings from markdown', () => {
    const md = '# Title\n## Section\nContent\n### Subsection';
    const headings = extractHeadings(md);
    expect(headings).toHaveLength(3);
    expect(headings[0]).toEqual({ level: 1, text: 'Title', slug: 'title' });
    expect(headings[2]).toEqual({ level: 3, text: 'Subsection', slug: 'subsection' });
  });
  it('handles inline formatting in headings', () => {
    const md = '## **Bold** and `code`';
    const headings = extractHeadings(md);
    expect(headings[0].text).toBe('Bold and code');
  });
});

describe('childrenToText', () => {
  it('returns string as-is', () => {
    expect(childrenToText('hello')).toBe('hello');
  });
  it('joins array of strings', () => {
    expect(childrenToText(['a', 'b'])).toBe('ab');
  });
});
