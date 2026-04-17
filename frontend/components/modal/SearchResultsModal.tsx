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
 * - Dracula Spectrum styling with Glass effect (reused from HelpModal)
 */

"use client";

import { useEffect, useState, useCallback } from "react";

// Custom event for showing search results modal
export const SHOW_SEARCH_RESULTS_MODAL = "showSearchResultsModal";

// Search result item interface
export interface SearchResultItem {
  path: string;
  title: string;
  lastModified?: string; // Date string (optional)
}

// Event detail interface (includes optional onSelect callback)
export interface SearchResultsEventDetail {
  results: SearchResultItem[];
  onSelect?: (path: string) => void; // Navigation callback
}

export function SearchResultsModal() {
  const [isVisible, setIsVisible] = useState(false);
  const [results, setResults] = useState<SearchResultItem[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [timer, setTimer] = useState<NodeJS.Timeout | null>(null);
  const [onSelectCallback, setOnSelectCallback] = useState<((path: string) => void) | null>(null);

  // Manual close handler (clear timer)
  const handleClose = useCallback(() => {
    if (timer) {
      clearTimeout(timer);
      setTimer(null);
    }
    setIsVisible(false);
    setResults([]);
    setSelectedIndex(0);
    setOnSelectCallback(null);
  }, [timer]);

  // Handle result selection (Enter key or click)
  const handleSelect = useCallback((result: SearchResultItem) => {
    if (onSelectCallback) {
      onSelectCallback(result.path);
    }
    handleClose();
  }, [onSelectCallback, handleClose]);

  // Listen for custom event to show modal
  useEffect(() => {
    const handleShowModal = (e: CustomEvent<SearchResultsEventDetail>) => {
      const detail = e.detail;
      if (detail.results && detail.results.length > 0) {
        setResults(detail.results);
        setSelectedIndex(0);
        setIsVisible(true);
        
        // Store callback if provided
        if (detail.onSelect) {
          setOnSelectCallback(detail.onSelect);
        }
        
        // Start 10-second auto-close timer
        const autoCloseTimer = setTimeout(() => {
          setIsVisible(false);
          setResults([]);
          setSelectedIndex(0);
          setOnSelectCallback(null);
          setTimer(null);
        }, 10000);
        
        setTimer(autoCloseTimer);
      }
    };

    window.addEventListener(SHOW_SEARCH_RESULTS_MODAL, handleShowModal as EventListener);
    return () => window.removeEventListener(SHOW_SEARCH_RESULTS_MODAL, handleShowModal as EventListener);
  }, []);

  // Keyboard navigation handler
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isVisible) return;
      
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex(prev => 
          prev < results.length - 1 ? prev + 1 : prev
        );
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex(prev => 
          prev > 0 ? prev - 1 : prev
        );
      } else if (e.key === "Enter") {
        e.preventDefault();
        if (results[selectedIndex]) {
          handleSelect(results[selectedIndex]);
        }
      } else if (e.key === "Escape") {
        e.preventDefault();
        handleClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isVisible, results, selectedIndex, handleSelect, handleClose]);

  if (!isVisible || results.length === 0) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
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
              <span className="text-left truncate max-w-[60%]">
                {result.title}
              </span>
              <span className="text-right text-xs text-outline">
                {result.lastModified || "N/A"}
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