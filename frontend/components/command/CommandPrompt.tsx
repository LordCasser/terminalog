/**
 * CommandPrompt Component
 * 
 * Terminal-style command input bar at bottom of page.
 * Features (v1.6):
 * - WebSocket connection for path completion and search
 * - localStorage command history (terminalog_command_history, max 100)
 * - ArrowUp/ArrowDown for history navigation
 * - Tab key: command auto-completion, path completion modal for multiple matches
 * - Search command via WebSocket
 * - Pure frontend routing
 * - Path sync with Navbar via TerminalConfig context
 */

"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { 
  SHOW_HELP_MODAL, 
  SHOW_SEARCH_RESULTS_MODAL, 
  SEARCH_RESULT_SELECTED,
  SHOW_PATH_COMPLETION_MODAL,
  PATH_SELECTED,
  SearchResultsEventDetail,
  PathCompletionEventDetail,
} from "@/components/modal";
import { useTerminalConfig } from "@/lib/hooks/useTerminalConfig";
import { getArticles } from "@/lib/api/articles";
import { searchArticles } from "@/lib/api/search";
import {
  COMMANDS,
  HISTORY_KEY,
  MAX_HISTORY_SIZE,
  encodePathForUrl,
  navigateToPath,
  resolveCdPath,
  resolvePath,
} from "./utils";

// Custom event for search icon click
export const FOCUS_COMMAND_INPUT = "focusCommandInput";

// Custom event for modal visibility (to prevent keyboard conflict)
export const SEARCH_MODAL_VISIBLE = "searchModalVisible";

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

interface ErrorResponse {
  type: "error";
  error: string;
}

type WebSocketMessage = CompletionResponse | ErrorResponse;

// No match hint state (shows for 1 second)
export const NO_MATCH_HINT = "noMatchHint";

export function CommandPrompt() {
  const [input, setInput] = useState("");
  const [isFocused, setIsFocused] = useState(false);
  const [searchModalVisible, setSearchModalVisible] = useState(false);
  const [showNoMatchHint, setShowNoMatchHint] = useState(false);
  const [noMatchHintType, setNoMatchHintType] = useState<"completion" | "search" | "command" | "searchTab">("completion");
  // Initialize history from localStorage (lazy initialization)
  const [history, setHistory] = useState<string[]>(() => {
    try {
      if (typeof window !== "undefined") {
        const stored = localStorage.getItem(HISTORY_KEY);
        if (stored) {
          const parsed = JSON.parse(stored);
          if (Array.isArray(parsed)) {
            return parsed;
          }
        }
      }
    } catch (e) {
      console.error("Failed to load command history:", e);
    }
    return [];
  });
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [cursorPosition, setCursorPosition] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const cursorMeasureRef = useRef<HTMLSpanElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const router = useRouter();
  
  // Calculate cursor position for no-match hint positioning
  useEffect(() => {
    if (cursorMeasureRef.current) {
      const textWidth = cursorMeasureRef.current.offsetWidth;
      // Position hint at the cursor position, but ensure it doesn't overflow right
      const maxWidth = cursorMeasureRef.current.parentElement?.offsetWidth || 500;
      const hintWidth = 120; // approximate hint span width
      const pos = Math.min(textWidth, maxWidth - hintWidth);
      setCursorPosition(Math.max(0, pos));
    }
  }, [input]);

  // Get owner and currentDir from TerminalConfig context
  const { owner, currentDir } = useTerminalConfig();

  const showTransientNoMatchHint = useCallback((type: "completion" | "search" | "command" | "searchTab") => {
    setNoMatchHintType(type);
    setShowNoMatchHint(true);
    setTimeout(() => setShowNoMatchHint(false), 1000);
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
    // In debug mode, use API_BASE from env, otherwise use window.location.host
    const apiBase = process.env.NEXT_PUBLIC_API_BASE || '';
    const wsHost = apiBase ? apiBase.replace(/^https?:\/\//, '').replace(/^http:\/\//, '') : (window.location.host || "localhost:18085");
    const wsUrl = `${wsProtocol}//${wsHost}/ws/terminal`;

    const connectWebSocket = () => {
      try {
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
        // WebSocket connected
      };

      ws.onerror = () => {
        // WebSocket error
      };

      ws.onclose = () => {
        // WebSocket disconnected, attempt reconnect after 3 seconds
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
      // Auto-fill "search " command when triggered from search icon
      setInput("search ");
    };

    window.addEventListener(FOCUS_COMMAND_INPUT, handleFocusEvent);
    return () => window.removeEventListener(FOCUS_COMMAND_INPUT, handleFocusEvent);
  }, []);

  // Listen for search modal visibility changes
  useEffect(() => {
    const handleModalVisible = (e: CustomEvent<boolean>) => {
      setSearchModalVisible(e.detail);
    };

    window.addEventListener(SEARCH_MODAL_VISIBLE, handleModalVisible as EventListener);
    return () => window.removeEventListener(SEARCH_MODAL_VISIBLE, handleModalVisible as EventListener);
  }, []);

// Listen for search result selection from SearchResultsModal
  useEffect(() => {
    const handleResultSelected = (e: CustomEvent<string>) => {
      const path = e.detail;
      // Smart navigation: .md → article, otherwise → directory
      navigateToPath(router, path);
    };

    window.addEventListener(SEARCH_RESULT_SELECTED, handleResultSelected as EventListener);
    return () => window.removeEventListener(SEARCH_RESULT_SELECTED, handleResultSelected as EventListener);
  }, [router]);

  // Listen for path selection from PathCompletionModal
  useEffect(() => {
    const handlePathSelected = (e: CustomEvent<{ path: string; command: string }>) => {
      const { path, command } = e.detail;
      setInput(`${command} ${path}`);
      // Refocus input after selection
      inputRef.current?.focus();
    };

    window.addEventListener(PATH_SELECTED, handlePathSelected as EventListener);
    return () => window.removeEventListener(PATH_SELECTED, handlePathSelected as EventListener);
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
    <T extends WebSocketMessage>(message: CompletionRequest): Promise<T> => {
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
      } else if (matchingCommands.length === 0) {
        showTransientNoMatchHint("completion");
      }
      return;
    }
    
    // Path completion for commands after the first word
    if (parts.length === 2 && parts[1].length > 0) {
      const cmd = parts[0].toLowerCase();
      
      // search command: Tab completion is NOT supported — show hint
      if (cmd === "search") {
        showTransientNoMatchHint("searchTab");
        return;
      }
      
      const partialPath = parts[1];
      
      // open/cd commands: search within current directory
      const dir = currentDir || "/";
      
      try {
        const response = await sendWebSocketMessage<CompletionResponse>({
          type: "completion_request",
          dir: dir,
          prefix: partialPath,
        });

        if (response.type === "completion_response") {
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
            // Single match: auto-fill
            setInput(`${cmd} ${matchingItems[0]}`);
          } else if (matchingItems.length > 1) {
            // Multiple matches: show modal
            const event = new CustomEvent<PathCompletionEventDetail>(
              SHOW_PATH_COMPLETION_MODAL,
              {
                detail: {
                  paths: matchingItems.map(path => ({
                    path,
                    isDirectory: path.endsWith("/"),
                  })),
                  command: cmd,
                },
              }
            );
            window.dispatchEvent(event);
          } else {
            showTransientNoMatchHint("completion");
          }
        }
      } catch {
        showTransientNoMatchHint("completion");
      }
    }
  }, [input, currentDir, sendWebSocketMessage, showTransientNoMatchHint]);

  // Handle ArrowUp/ArrowDown for history navigation
  const handleHistoryNavigation = useCallback((e: React.KeyboardEvent) => {
    // Don't navigate history when search modal is open
    if (searchModalVisible) return;
    
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
  }, [history, historyIndex, searchModalVisible]);

  // Handle keyboard events
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    // When search results modal is open (DOM check, not React state),
    // don't process ArrowUp/ArrowDown/Enter to avoid conflicts
    const searchModalOpen = document.querySelector('[data-search-modal]');
    if (searchModalOpen && (e.key === "ArrowUp" || e.key === "ArrowDown" || e.key === "Enter")) {
      return;
    }
    
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
      
      // Always perform keyword search via REST API
      // Path-style queries like "linux/Qemu的安装和配置(MIPS).md" are also
      // valid search terms — the search will match against file paths and titles.
      try {
        const searchResponse = await searchArticles({ q: query });

        if (searchResponse.results.length > 0) {
          // Always show modal for selection (no single-result auto-navigation)
          const event = new CustomEvent<SearchResultsEventDetail>(
            SHOW_SEARCH_RESULTS_MODAL,
            {
              detail: {
                results: searchResponse.results.map(r => ({
                  path: r.path,
                  title: r.title,
                  type: r.type,
                  lastModified: undefined,
                })),
              },
            }
          );
          window.dispatchEvent(event);
        } else {
          showTransientNoMatchHint("search");
        }
      } catch (error) {
        console.error("Search API error:", error);
        showTransientNoMatchHint("search");
      }
      return;
    }
    
    if (trimmedCmd.startsWith("open ")) {
      const target = cmd.trim().slice(5);
      // Resolve path relative to current directory
      const resolvedPath = resolvePath(currentDir, target);
      // Navigate to article (no validation - article page handles 404)
      router.push(`/article/${encodePathForUrl(resolvedPath)}`);
      return;
    }
    
    // cd without arguments → go to root (same as cd /)
    if (trimmedCmd === "cd") {
      router.push("/");
      return;
    }
    
    if (trimmedCmd.startsWith("cd ")) {
      const target = cmd.trim().slice(3);
      
      // Resolve path: handle relative paths (.., subdir) and absolute paths
      const resolvedPath = resolveCdPath(currentDir, target);
      
      // Validate directory exists by trying to fetch its listing
      try {
        await getArticles(resolvedPath);
        // Directory exists - navigate to it
        if (resolvedPath) {
          router.push(`/dir/${encodePathForUrl(resolvedPath)}`);
        } else {
          router.push("/");
        }
      } catch {
        showTransientNoMatchHint("completion");
      }
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
      showTransientNoMatchHint("command");
    }
  }, [router, currentDir, showTransientNoMatchHint]);

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
        <span className="text-secondary font-bold ml-1">~/{owner}{currentDir ? `/${currentDir}` : ""}</span>
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
            className="w-full bg-transparent text-on-surface outline-none border-none placeholder:text-on-surface-variant/15"
            placeholder="type a command..."
            aria-label="Command input"
          />
          
          {/* Hidden span to measure input text width for cursor positioning */}
          <span
            ref={cursorMeasureRef}
            className="absolute invisible whitespace-pre font-mono text-sm left-0 top-0"
            aria-hidden="true"
          >
            {input}
          </span>
          
          {/* No Match Hint (above input at cursor position, shows for 1 second) */}
          {showNoMatchHint && (
            <span 
              className="absolute bottom-full mb-2 px-2 py-1 bg-surface-container-high text-error font-mono text-xs rounded"
              style={{ left: `${cursorPosition}px` }}
            >
              {noMatchHintType === "searchTab"
                ? "search 不可补全"
                : noMatchHintType === "search"
                  ? "没有搜索结果"
                  : noMatchHintType === "command"
                    ? "命令不存在"
                    : "没有匹配内容"}
            </span>
          )}
        </div>
      </form>
    </footer>
  );
}
