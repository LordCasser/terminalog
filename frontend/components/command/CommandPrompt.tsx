/**
 * CommandPrompt Component
 * 
 * Terminal-style command input bar at bottom of page.
 * Features (v1.4):
 * - WebSocket connection for path completion and search (ws://localhost:18085/ws/terminal)
 * - localStorage command history (terminalog_command_history, max 100)
 * - ArrowUp/ArrowDown for history navigation
 * - Tab key auto-completion via WebSocket
 * - Search command via WebSocket
 * - Pure frontend routing (no HTTP API for commands)
 */

"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { SHOW_HELP_MODAL } from "@/components/modal/HelpModal";

// Custom event for search icon click
export const FOCUS_COMMAND_INPUT = "focusCommandInput";

// Available commands for auto-completion
const COMMANDS = ["search", "open", "cd", "help"];

// History storage key and max size
const HISTORY_KEY = "terminalog_command_history";
const MAX_HISTORY_SIZE = 100;

// WebSocket message types
interface CompletionRequest {
  type: "completion_request";
  dir: string;
  prefix: string;
}

interface CompletionResponse {
  type: "completion_response";
  items: string[];
}

interface SearchRequest {
  type: "search_request";
  keyword: string;
}

interface SearchResponse {
  type: "search_response";
  results: Array<{ path: string; title: string }>;
}

interface ErrorResponse {
  type: "error";
  error: string;
}

type WebSocketMessage = CompletionResponse | SearchResponse | ErrorResponse;

export function CommandPrompt() {
  const [input, setInput] = useState("");
  const [isFocused, setIsFocused] = useState(false);
  const [history, setHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [currentDir, setCurrentDir] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const router = useRouter();
  
  // Get current directory from URL params (client-side only)
  useEffect(() => {
    if (typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search);
      setCurrentDir(params.get("dir") || "");
    }
  }, []);

  // Load history from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(HISTORY_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed)) {
          setHistory(parsed);
        }
      }
    } catch (e) {
      console.error("Failed to load command history:", e);
    }
  }, []);

  // Save history to localStorage
  const saveHistory = useCallback((newHistory: string[]) => {
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(newHistory));
    } catch (e) {
      console.error("Failed to save command history:", e);
    }
  }, []);

  // Initialize WebSocket connection
  useEffect(() => {
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsHost = window.location.host || "localhost:18085";
    const wsUrl = `${wsProtocol}//${wsHost}/ws/terminal`;

    const connectWebSocket = () => {
      try {
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          console.log("WebSocket connected");
        };

        ws.onerror = (error) => {
          console.error("WebSocket error:", error);
        };

        ws.onclose = () => {
          console.log("WebSocket disconnected");
          // Attempt reconnect after 3 seconds
          setTimeout(connectWebSocket, 3000);
        };
      } catch (e) {
        console.error("Failed to create WebSocket:", e);
      }
    };

    connectWebSocket();

    // Cleanup on unmount
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

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
        "Escape", "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
        "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight",
        "PageUp", "PageDown", "Home", "End",
      ];
      if (ignoreKeys.includes(e.key)) return;
      
      // Focus input (including Tab key)
      if (e.key === "Tab") {
        e.preventDefault(); // Disable browser default Tab behavior
        inputRef.current?.focus();
        return;
      }
      
      // Focus input on other keys
      inputRef.current?.focus();
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isFocused]);

  // Send WebSocket message and wait for response
  const sendWebSocketMessage = useCallback(
    <T extends WebSocketMessage>(message: CompletionRequest | SearchRequest): Promise<T> => {
      return new Promise((resolve, reject) => {
        if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
          reject(new Error("WebSocket not connected"));
          return;
        }

        const handleMessage = (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data) as T;
            wsRef.current?.removeEventListener("message", handleMessage);
            resolve(data);
          } catch (e) {
            reject(e);
          }
        };

        wsRef.current.addEventListener("message", handleMessage);
        wsRef.current.send(JSON.stringify(message));

        // Timeout after 5 seconds
        setTimeout(() => {
          wsRef.current?.removeEventListener("message", handleMessage);
          reject(new Error("WebSocket timeout"));
        }, 5000);
      });
    },
    []
  );

  // Handle Tab key auto-completion via WebSocket
  const handleTabCompletion = useCallback(async (e: React.KeyboardEvent) => {
    if (e.key !== "Tab") return;
    
    e.preventDefault(); // Always prevent default Tab behavior
    
    const trimmedInput = input.trim();
    const parts = trimmedInput.split(/\s+/);
    
    // Auto-complete commands
    if (parts.length === 1 && parts[0].length > 0) {
      const partialCmd = parts[0].toLowerCase();
      const matchingCommands = COMMANDS.filter(cmd => cmd.startsWith(partialCmd));
      
      if (matchingCommands.length === 1) {
        setInput(matchingCommands[0] + " ");
      }
      return;
    }
    
    // Auto-complete paths via WebSocket
    if (parts.length === 2 && parts[1].length > 0) {
      const cmd = parts[0].toLowerCase();
      const partialPath = parts[1];
      
      try {
        const response = await sendWebSocketMessage<CompletionResponse>({
          type: "completion_request",
          dir: currentDir || "/",
          prefix: partialPath,
        });

        if (response.type === "completion_response" && response.items.length > 0) {
          // Filter items based on command type
          let matchingItems = response.items;
          
          if (cmd === "open") {
            // Files only (no trailing slash)
            matchingItems = response.items.filter(item => !item.endsWith("/"));
          } else if (cmd === "cd") {
            // Directories only (trailing slash)
            matchingItems = response.items.filter(item => item.endsWith("/"));
          }
          
          if (matchingItems.length === 1) {
            setInput(`${cmd} ${matchingItems[0]}`);
          } else if (matchingItems.length > 1) {
            console.log("Matching items:", matchingItems.join(", "));
          }
        }
      } catch (error) {
        console.error("WebSocket completion error:", error);
      }
    }
  }, [input, currentDir, sendWebSocketMessage]);

  // Handle ArrowUp/ArrowDown for history navigation
  const handleHistoryNavigation = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "ArrowUp") {
      e.preventDefault();
      if (history.length > 0 && historyIndex < history.length - 1) {
        const newIndex = historyIndex + 1;
        setHistoryIndex(newIndex);
        setInput(history[history.length - 1 - newIndex]);
      }
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      if (historyIndex > 0) {
        const newIndex = historyIndex - 1;
        setHistoryIndex(newIndex);
        setInput(history[history.length - 1 - newIndex]);
      } else if (historyIndex === 0) {
        setHistoryIndex(-1);
        setInput("");
      }
    }
  }, [history, historyIndex]);

  // Handle keyboard events
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "Tab") {
      handleTabCompletion(e);
    } else if (e.key === "ArrowUp" || e.key === "ArrowDown") {
      handleHistoryNavigation(e);
    }
  }, [handleTabCompletion, handleHistoryNavigation]);

  // Handle command execution
  const executeCommand = useCallback(async (cmd: string) => {
    const trimmedCmd = cmd.trim().toLowerCase();
    
    // Parse command
    if (trimmedCmd.startsWith("search ")) {
      const query = cmd.trim().slice(7);
      if (!query) return;
      
      // Search via WebSocket
      try {
        const response = await sendWebSocketMessage<SearchResponse>({
          type: "search_request",
          keyword: query,
        });

        if (response.type === "search_response" && response.results.length > 0) {
          // Jump to first result
          router.push(`/article?path=${encodeURIComponent(response.results[0].path)}`);
        }
      } catch (error) {
        console.error("WebSocket search error:", error);
      }
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
    
    // Help command - show help modal
    if (trimmedCmd === "help" || trimmedCmd === "?") {
      window.dispatchEvent(new Event(SHOW_HELP_MODAL));
      return;
    }
    
    // Unknown command
    if (trimmedCmd !== "") {
      console.log(`Unknown command: ${cmd}`);
    }
  }, [router, sendWebSocketMessage]);

  // Handle form submit
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const cmd = input.trim();
    if (cmd) {
      // Add to history
      const newHistory = [...history, cmd];
      if (newHistory.length > MAX_HISTORY_SIZE) {
        newHistory.shift();
      }
      setHistory(newHistory);
      saveHistory(newHistory);
      setHistoryIndex(-1); // Reset history index
      
      executeCommand(input);
    }
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
        <span className="text-secondary font-bold ml-1">~{currentDir || "/"}</span>
        <span className="text-on-surface-variant mx-1">$</span>
        
        {/* Command Input */}
        <div className="flex-1 flex items-center relative">
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            onFocus={handleFocus}
            onBlur={handleBlur}
            className="w-full bg-transparent text-on-surface outline-none border-none placeholder-on-surface-variant placeholder-opacity-50"
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