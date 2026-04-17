/**
 * PathCompletionModal Component
 * 
 * Modal dialog showing path completion options when multiple matches found.
 * Features (v1.6):
 * - Displays path options with type indicator (file/folder)
 * - ArrowUp/ArrowDown key navigation for selection
 * - Enter key to confirm and fill path
 * - 10-second auto-close timer
 * - Manual close via ESC or backdrop click
 */

"use client";

import { useEffect, useState, useCallback } from "react";

// Custom event for showing path completion modal
export const SHOW_PATH_COMPLETION_MODAL = "showPathCompletionModal";

// Custom event for path selection
export const PATH_SELECTED = "pathSelected";

// Path completion item interface
export interface PathCompletionItem {
  path: string;
  isDirectory: boolean; // true if folder (ends with /)
}

// Event detail interface
export interface PathCompletionEventDetail {
  paths: PathCompletionItem[];
  command: string; // e.g., "open", "cd"
}

export function PathCompletionModal() {
  const [isVisible, setIsVisible] = useState(false);
  const [paths, setPaths] = useState<PathCompletionItem[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [command, setCommand] = useState("");
  const [timer, setTimer] = useState<NodeJS.Timeout | null>(null);

  // Manual close handler (clear timer)
  const handleClose = useCallback(() => {
    if (timer) {
      clearTimeout(timer);
      setTimer(null);
    }
    setIsVisible(false);
    setPaths([]);
    setSelectedIndex(0);
    setCommand("");
  }, [timer]);

  // Handle path selection (Enter key or click)
  const handleSelect = useCallback((item: PathCompletionItem) => {
    window.dispatchEvent(new CustomEvent(PATH_SELECTED, { 
      detail: { path: item.path, command } 
    }));
    handleClose();
  }, [command, handleClose]);

  // Listen for custom event to show modal
  useEffect(() => {
    const handleShowModal = (e: CustomEvent<PathCompletionEventDetail>) => {
      const detail = e.detail;
      if (detail.paths && detail.paths.length > 0) {
        setPaths(detail.paths);
        setSelectedIndex(0);
        setCommand(detail.command || "");
        setIsVisible(true);
        
        // Start 10-second auto-close timer
        const autoCloseTimer = setTimeout(() => {
          setIsVisible(false);
          setPaths([]);
          setSelectedIndex(0);
          setTimer(null);
          setCommand("");
        }, 10000);
        
        setTimer(autoCloseTimer);
      }
    };

    window.addEventListener(SHOW_PATH_COMPLETION_MODAL, handleShowModal as EventListener);
    return () => window.removeEventListener(SHOW_PATH_COMPLETION_MODAL, handleShowModal as EventListener);
  }, []);

  // Keyboard navigation - window listener in capture phase
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const modalEl = document.querySelector('[data-path-modal]');
      if (!modalEl) return;
      
      if (e.key === "ArrowDown") {
        e.preventDefault();
        e.stopImmediatePropagation();
        setSelectedIndex(prev => prev < paths.length - 1 ? prev + 1 : prev);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        e.stopImmediatePropagation();
        setSelectedIndex(prev => prev > 0 ? prev - 1 : prev);
      } else if (e.key === "Enter") {
        e.preventDefault();
        e.stopImmediatePropagation();
        if (paths[selectedIndex]) {
          handleSelect(paths[selectedIndex]);
        }
      } else if (e.key === "Escape" || e.key === "Tab") {
        e.preventDefault();
        e.stopImmediatePropagation();
        handleClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown, true);
    return () => window.removeEventListener("keydown", handleKeyDown, true);
  }, [paths, selectedIndex, handleSelect, handleClose]);

  if (!isVisible || paths.length === 0) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center outline-none"
      data-path-modal
    >
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-background/60 backdrop-blur-sm"
        onClick={handleClose}
      />
      
      {/* Modal Content - Glass Effect */}
      <div className="relative max-w-md w-full mx-4 bg-surface-container-high/42 backdrop-blur-lg border border-primary/25 shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-surface-container-highest">
          <h2 className="font-mono text-sm uppercase tracking-tighter text-secondary">
            {command.toUpperCase()} ({paths.length})
          </h2>
          <button
            onClick={handleClose}
            className="text-on-surface-variant hover:text-on-surface transition-colors"
            aria-label="Close path completion modal"
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
        
        {/* Paths List */}
        <div className="p-4 space-y-1 max-h-[300px] overflow-y-auto">
          {paths.map((item, index) => (
            <div 
              key={item.path}
              onClick={() => handleSelect(item)}
              className={`
                flex items-center gap-2 px-3 py-2 cursor-pointer
                transition-colors duration-100 font-mono text-sm
                ${index === selectedIndex 
                  ? "bg-primary/20 text-primary border-l-2 border-primary" 
                  : "text-on-surface-variant hover:bg-surface-container-highest/50"
                }
              `}
            >
              {/* Type indicator */}
              <span className="text-xs">
                {item.isDirectory ? "📁" : "📄"}
              </span>
              <span className="truncate">
                {item.path}
              </span>
            </div>
          ))}
        </div>
        
        {/* Footer hint */}
        <div className="p-3 border-t border-surface-container-highest text-center">
          <span className="font-mono text-xs text-on-surface-variant">
            ↑↓ Navigate · Enter Select · Tab/ESC Cancel
          </span>
        </div>
      </div>
    </div>
  );
}