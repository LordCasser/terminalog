"use client";

import { useState, useCallback } from "react";
import { HISTORY_KEY, MAX_HISTORY_SIZE } from "./utils";

export interface CommandHistory {
  /** Current history entries (newest last). */
  entries: string[];
  /** Current position in history navigation (-1 = not navigating). */
  index: number;

  /** Add a command to history and reset navigation. */
  add: (command: string) => void;
  /** Navigate to the previous command in history. */
  previous: (currentInput: string) => string | null;
  /** Navigate to the next command in history. */
  next: (currentInput: string) => string | null;
  /** Reset navigation position. */
  reset: () => void;
}

/**
 * Manages command-line history with localStorage persistence.
 * Maximum history size is defined by MAX_HISTORY_SIZE (default 100).
 */
export function useCommandHistory(): CommandHistory {
  const [entries, setEntries] = useState<string[]>(() => {
    try {
      if (typeof window !== "undefined") {
        const stored = localStorage.getItem(HISTORY_KEY);
        if (stored) {
          const parsed = JSON.parse(stored);
          if (Array.isArray(parsed)) return parsed;
        }
      }
    } catch (e) {
      console.error("Failed to load command history:", e);
    }
    return [];
  });

  const [index, setIndex] = useState(-1);

  const add = useCallback((command: string) => {
    const trimmed = command.trim();
    if (!trimmed) return;

    setEntries((prev) => {
      // Don't add consecutive duplicates
      if (prev.length > 0 && prev[prev.length - 1] === trimmed) {
        return prev;
      }
      const next = [...prev, trimmed];
      if (next.length > MAX_HISTORY_SIZE) {
        return next.slice(next.length - MAX_HISTORY_SIZE);
      }
      // Persist to localStorage
      try {
        localStorage.setItem(HISTORY_KEY, JSON.stringify(next));
      } catch (e) {
        console.error("Failed to save command history:", e);
      }
      return next;
    });
    setIndex(-1);
  }, []);

  const previous = useCallback(
    (): string | null => {
      if (entries.length === 0) return null;
      const newIndex = index === -1 ? entries.length - 1 : Math.max(0, index - 1);
      setIndex(newIndex);
      return entries[newIndex];
    },
    [entries, index]
  );

  const next = useCallback(
    (): string | null => {
      if (index === -1 || index >= entries.length - 1) {
        setIndex(-1);
        return null; // Return to original input
      }
      const newIndex = index + 1;
      setIndex(newIndex);
      return entries[newIndex];
    },
    [entries, index]
  );

  const reset = useCallback(() => {
    setIndex(-1);
  }, []);

  return { entries, index, add, previous, next, reset };
}
