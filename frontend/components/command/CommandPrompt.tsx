/**
 * CommandPrompt Component
 * 
 * Terminal-style command input bar at bottom of page.
 * Features:
 * - Path display (guest@blog:~/lordcasser$)
 * - Command input with blinking cursor
 * - Global keyboard focus (press any key to focus)
 * - Search icon click focus
 * - Basic command parsing (help, list, search, open, cd)
 */

"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";

// Custom event for search icon click
export const FOCUS_COMMAND_INPUT = "focusCommandInput";

export function CommandPrompt() {
  const [input, setInput] = useState("");
  const [isFocused, setIsFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();

  // Focus input on custom event (search icon click)
  useEffect(() => {
    const handleFocusEvent = () => {
      inputRef.current?.focus();
    };

    window.addEventListener(FOCUS_COMMAND_INPUT, handleFocusEvent);
    return () => window.removeEventListener(FOCUS_COMMAND_INPUT, handleFocusEvent);
  }, []);

  // Global keyboard listener - focus input on any key press
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if already focused, or if modifier keys pressed
      if (isFocused || e.metaKey || e.ctrlKey || e.altKey) return;
      
      // Ignore navigation keys, function keys, etc.
      const ignoreKeys = [
        "Tab", "Escape", "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
        "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight",
        "PageUp", "PageDown", "Home", "End",
      ];
      if (ignoreKeys.includes(e.key)) return;
      
      // Focus input
      inputRef.current?.focus();
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isFocused]);

  // Handle command execution
  const executeCommand = useCallback((cmd: string) => {
    const trimmedCmd = cmd.trim().toLowerCase();
    
    // Parse command
    if (trimmedCmd === "help") {
      console.log("Available commands: help, list, search <query>, open <article>, cd <path>");
      return;
    }
    
    if (trimmedCmd === "list" || trimmedCmd === "ls") {
      router.push("/");
      return;
    }
    
    if (trimmedCmd.startsWith("search ")) {
      const query = cmd.trim().slice(7);
      // TODO: Implement search functionality
      console.log(`Searching for: ${query}`);
      return;
    }
    
    if (trimmedCmd.startsWith("open ")) {
      const article = cmd.trim().slice(5);
      router.push(`/article?path=${encodeURIComponent(article)}`);
      return;
    }
    
    if (trimmedCmd.startsWith("cd ")) {
      const path = cmd.trim().slice(3);
      router.push(`/?dir=${encodeURIComponent(path)}`);
      return;
    }
    
    // Unknown command
    if (trimmedCmd !== "") {
      console.log(`Unknown command: ${cmd}`);
    }
  }, [router]);

  // Handle form submit
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    executeCommand(input);
    setInput("");
  };

  // Handle focus/blur
  const handleFocus = () => setIsFocused(true);
  const handleBlur = () => setIsFocused(false);

  return (
    <footer className="fixed bottom-0 left-0 w-full z-50 bg-surface shadow-[0_-4px_20px_rgba(0,0,0,0.4)] border-t border-surface-container-highest">
      <form onSubmit={handleSubmit} className="flex items-center gap-3 px-6 h-16 font-mono text-sm">
        {/* Path Display */}
        <span className="text-tertiary font-bold">guest@blog:</span>
        <span className="text-secondary font-bold ml-1">~/lordcasser</span>
        <span className="text-on-surface-variant mx-1">$</span>
        
        {/* Command Input */}
        <div className="flex-1 flex items-center relative">
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onFocus={handleFocus}
            onBlur={handleBlur}
            className="w-full bg-transparent text-on-surface outline-none border-none placeholder-on-surface-variant"
            placeholder="type a command..."
            aria-label="Command input"
          />
          
          {/* Blinking Cursor (when focused and empty) */}
          {isFocused && input === "" && (
            <span className="absolute left-0 w-2.5 h-5 bg-tertiary cursor-blink" />
          )}
        </div>
      </form>
    </footer>
  );
}