/**
 * SearchResultsModal Component
 * 
 * Modal dialog showing search results when multiple matches found.
 * Features (v1.6):
 * - Displays search results with title (left-aligned) and date (right-aligned)
 * - ArrowUp/ArrowDown key navigation for result selection
 * - Enter key to confirm selection and navigate to article
 * - 10-second auto-close timer
 * - Manual close via right-top 'x' button
 * - Uses React onKeyDown with autoFocus to avoid closure issues
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { SEARCH_MODAL_VISIBLE, FOCUS_COMMAND_INPUT } from "@/components/command/CommandPrompt";

// Custom event for showing search results modal
export const SHOW_SEARCH_RESULTS_MODAL = "showSearchResultsModal";

// Custom event for result selection (to avoid React callback error)
export const SEARCH_RESULT_SELECTED = "searchResultSelected";

// Search result item interface
export interface SearchResultItem {
  path: string;
  title: string;
  type: "file" | "dir"; // "file" for articles, "dir" for directories
  lastModified?: string; // Date string (optional)
}

// Event detail interface (no callback - use event instead)
export interface SearchResultsEventDetail {
  results: SearchResultItem[];
}

export function SearchResultsModal() {
  const [isVisible, setIsVisible] = useState(false);
  const [results, setResults] = useState<SearchResultItem[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [timer, setTimer] = useState<NodeJS.Timeout | null>(null);

  // Manual close handler (clear timer)
  const handleClose = useCallback(() => {
    if (timer) {
      clearTimeout(timer);
      setTimer(null);
    }
    setIsVisible(false);
    setResults([]);
    setSelectedIndex(0);
    // Notify CommandPrompt that modal is hidden, and refocus input
    window.dispatchEvent(new CustomEvent(SEARCH_MODAL_VISIBLE, { detail: false }));
    window.dispatchEvent(new Event(FOCUS_COMMAND_INPUT));
  }, [timer]);

  // Handle result selection (Enter key or click) - smart navigation
  const handleSelect = useCallback((result: SearchResultItem) => {
    if (result.type === "dir") {
      window.dispatchEvent(new CustomEvent(SEARCH_RESULT_SELECTED, { detail: result.path }));
    } else {
      window.dispatchEvent(new CustomEvent(SEARCH_RESULT_SELECTED, { detail: result.path }));
    }
    handleClose();
  }, [handleClose]);

  // Listen for custom event to show modal
  useEffect(() => {
    const handleShowModal = (e: CustomEvent<SearchResultsEventDetail>) => {
      const detail = e.detail;
      if (detail.results && detail.results.length > 0) {
        setResults(detail.results);
        setSelectedIndex(0);
        setIsVisible(true);
        
        // Notify CommandPrompt that modal is visible
        window.dispatchEvent(new CustomEvent(SEARCH_MODAL_VISIBLE, { detail: true }));
        
        // Start 10-second auto-close timer
        const autoCloseTimer = setTimeout(() => {
          setIsVisible(false);
          setResults([]);
          setSelectedIndex(0);
          setTimer(null);
          window.dispatchEvent(new CustomEvent(SEARCH_MODAL_VISIBLE, { detail: false }));
          window.dispatchEvent(new Event(FOCUS_COMMAND_INPUT));
        }, 10000);
        
        setTimer(autoCloseTimer);
      }
    };

    window.addEventListener(SHOW_SEARCH_RESULTS_MODAL, handleShowModal as EventListener);
    return () => window.removeEventListener(SHOW_SEARCH_RESULTS_MODAL, handleShowModal as EventListener);
  }, []);

  // Keyboard navigation - window listener in capture phase
  // Uses DOM attribute (data-search-modal) as synchronous visibility flag
  // to avoid React state async update race condition
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check synchronously via DOM attribute instead of React state
      const modalEl = document.querySelector('[data-search-modal]');
      if (!modalEl) return;
      
      if (e.key === "ArrowDown") {
        e.preventDefault();
        e.stopImmediatePropagation(); // Prevent CommandPrompt from handling
        setSelectedIndex(prev => prev < results.length - 1 ? prev + 1 : prev);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        e.stopImmediatePropagation();
        setSelectedIndex(prev => prev > 0 ? prev - 1 : prev);
      } else if (e.key === "Enter") {
        e.preventDefault();
        e.stopImmediatePropagation();
        if (results[selectedIndex]) {
          handleSelect(results[selectedIndex]);
        }
      } else if (e.key === "Escape") {
        e.preventDefault();
        e.stopImmediatePropagation();
        handleClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown, true); // capture phase
    return () => window.removeEventListener("keydown", handleKeyDown, true);
  }, [results, selectedIndex, handleSelect, handleClose]);

  if (!isVisible || results.length === 0) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center outline-none"
      data-search-modal
    >
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-background/60 backdrop-blur-sm"
        onClick={handleClose}
      />
      
      {/* Modal Content - Glass Effect */}
      <div className="relative max-w-xl w-full mx-4 bg-surface-container-high/42 backdrop-blur-lg border border-primary/25 shadow-2xl">
        {/* Header with close button */}
        <div className="flex items-center justify-between p-4 border-b border-surface-container-highest">
          <h2 className="font-mono text-sm uppercase tracking-tighter text-secondary">
            Search Results ({results.length})
          </h2>
          <button
            onClick={handleClose}
            className="text-on-surface-variant hover:text-on-surface transition-colors"
            aria-label="Close search results modal"
          >
            <svg 
              className="w-5 h-5" 
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path 
                strokeLinecap="square" 
                strokeWidth={2} 
                d="M6 18L18 6M6 6l12 12" 
              />
            </svg>
          </button>
        </div>
        
        {/* Results List */}
        <div className="p-4 space-y-1 max-h-[400px] overflow-y-auto">
          {results.map((result, index) => (
            <div 
              key={result.path}
              onClick={() => handleSelect(result)}
              className={`
                flex justify-between items-center px-3 py-2 cursor-pointer
                transition-colors duration-100 font-mono text-sm
                ${index === selectedIndex 
                  ? "bg-primary/20 text-primary border-l-2 border-primary" 
                  : "text-on-surface-variant hover:bg-surface-container-highest/50"
                }
              `}
            >
              <span className="text-left truncate max-w-[70%] flex items-center gap-2">
                {/* Type icon: folder for dir, description for file */}
                <span className="material-symbols-outlined text-xs">
                  {result.type === "dir" ? "folder" : "description"}
                </span>
                {/* Display: title with hierarchical path */}
                <span className="truncate">
                  {result.title}
                  {result.path !== result.title && (
                    <span className="text-outline ml-1 text-xs">
                      ({result.type === "dir" ? result.path : result.path.replace(/\.md$/, "")})
                    </span>
                  )}
                </span>
              </span>
              <span className="text-right text-xs text-outline">
                {result.type === "dir" ? "DIR" : (result.lastModified || "N/A")}
              </span>
            </div>
          ))}
        </div>
        
        {/* Footer hint */}
        <div className="p-3 border-t border-surface-container-highest text-center">
          <span className="font-mono text-xs text-on-surface-variant">
            ↑↓ Navigate · Enter Select · Auto-closes in 10s
          </span>
        </div>
      </div>
    </div>
  );
}