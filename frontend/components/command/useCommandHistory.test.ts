import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useCommandHistory } from './useCommandHistory';

describe('useCommandHistory', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('starts with empty history', () => {
    const { result } = renderHook(() => useCommandHistory());
    expect(result.current.entries).toEqual([]);
    expect(result.current.index).toBe(-1);
  });

  it('adds commands to history', () => {
    const { result } = renderHook(() => useCommandHistory());
    act(() => { result.current.add('search test'); });
    expect(result.current.entries).toEqual(['search test']);
    expect(result.current.index).toBe(-1); // Reset after add
  });

  it('does not add consecutive duplicates', () => {
    const { result } = renderHook(() => useCommandHistory());
    act(() => { result.current.add('search test'); });
    act(() => { result.current.add('search test'); });
    expect(result.current.entries).toEqual(['search test']);
  });

  it('navigates history with ArrowUp/ArrowDown', () => {
    const { result } = renderHook(() => useCommandHistory());
    act(() => { result.current.add('cmd1'); });
    act(() => { result.current.add('cmd2'); });
    act(() => { result.current.add('cmd3'); });

    // Arrow up from no navigation → last entry
    let entry: string | null = null;
    act(() => { entry = result.current.previous(''); });
    expect(entry).toBe('cmd3');

    act(() => { entry = result.current.previous(''); });
    expect(entry).toBe('cmd2');

    act(() => { entry = result.current.previous(''); });
    expect(entry).toBe('cmd1');

    // Arrow down → back to cmd2
    act(() => { entry = result.current.next(''); });
    expect(entry).toBe('cmd2');

    // Arrow down → back to cmd3
    act(() => { entry = result.current.next(''); });
    expect(entry).toBe('cmd3');

    // Arrow down past end → null (return to original)
    act(() => { entry = result.current.next(''); });
    expect(entry).toBeNull();
  });

  it('persists to localStorage', () => {
    const { result } = renderHook(() => useCommandHistory());
    act(() => { result.current.add('persisted cmd'); });

    // New hook instance should load persisted data
    const { result: result2 } = renderHook(() => useCommandHistory());
    expect(result2.current.entries).toEqual(['persisted cmd']);
  });
});
