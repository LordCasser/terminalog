/**
 * HelpModal Component
 * 
 * Modal dialog showing available commands when user types 'help' or '?'.
 * Features:
 * - Displays all available commands with descriptions
 * - 3-second auto-close timer
 * - Manual close via right-top 'x' button
 * - Enter key to close
 * - Dracula Spectrum styling with Glass effect
 * - Width: max-w-xl (ensures each command description fits in one line)
 */

"use client";

import { useEffect, useState, useCallback } from "react";

// Custom event for showing help modal
export const SHOW_HELP_MODAL = "showHelpModal";

// Available commands for display
const COMMANDS_INFO = [
  { cmd: "cd <path>", desc: "Change directory to specified path" },
  { cmd: "cd ..", desc: "Go back to parent directory" },
  { cmd: "cd .", desc: "Refresh current directory" },
  { cmd: "open <filename>", desc: "Open article for viewing" },
  { cmd: "search <keyword>", desc: "Search articles by title keyword" },
  { cmd: "help / ?", desc: "Show this help dialog" },
];

export function HelpModal() {
  const [isVisible, setIsVisible] = useState(false);
  const [timer, setTimer] = useState<NodeJS.Timeout | null>(null);

  // Manual close handler (clear timer)
  const handleClose = useCallback(() => {
    if (timer) {
      clearTimeout(timer);
      setTimer(null);
    }
    setIsVisible(false);
  }, [timer]);

  // Listen for custom event to show modal
  useEffect(() => {
    const handleShowModal = () => {
      setIsVisible(true);
      
      // Start 3-second auto-close timer
      const autoCloseTimer = setTimeout(() => {
        setIsVisible(false);
        setTimer(null);
      }, 3000);
      
      setTimer(autoCloseTimer);
    };

    window.addEventListener(SHOW_HELP_MODAL, handleShowModal);
    return () => window.removeEventListener(SHOW_HELP_MODAL, handleShowModal);
  }, []);

  // Enter key to close modal
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (isVisible && e.key === "Enter") {
        handleClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isVisible, handleClose]);

  if (!isVisible) return null;

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
            Available Commands
          </h2>
          <button
            onClick={handleClose}
            className="text-on-surface-variant hover:text-on-surface transition-colors"
            aria-label="Close help modal"
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
        
        {/* Commands List */}
        <div className="p-6 space-y-3">
          {COMMANDS_INFO.map((item, index) => (
            <div key={index} className="flex gap-4 font-mono text-sm">
              <span className="text-tertiary font-bold min-w-[140px]">
                {item.cmd}
              </span>
              <span className="text-on-surface-variant">
                {item.desc}
              </span>
            </div>
          ))}
        </div>
        
        {/* Footer hint */}
        <div className="p-4 border-t border-surface-container-highest text-center">
          <span className="font-mono text-xs text-on-surface-variant">
            Auto-closes in 3 seconds, press Enter or click × to close
          </span>
        </div>
      </div>
    </div>
  );
}