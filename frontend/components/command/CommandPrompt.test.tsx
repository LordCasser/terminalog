import { describe, it, expect } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useCommandHistory } from './useCommandHistory';

describe('CommandPrompt (minimal)', () => {
  it('useCommandHistory hook is available', () => {
    const { result } = renderHook(() => useCommandHistory());
    expect(result.current.entries).toBeDefined();
    expect(result.current.index).toBe(-1);
  });
});
