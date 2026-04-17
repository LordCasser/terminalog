/**
 * TerminalConfig Context
 * 
 * Global context for sharing terminal configuration across components.
 * Provides:
 * - owner: Blog owner name (e.g., "lordcasser")
 * - currentDir: Current directory path for path synchronization
 * 
 * Used by:
 * - Navbar: displays ~/owner/currentDir
 * - CommandPrompt: displays guest@blog: ~/currentDir
 */

"use client";

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import { apiClient } from "@/lib/api/client";

// Frontend config response from /api/v1/settings
interface FrontendConfig {
  owner: string;
}

// Context value interface
interface TerminalConfigValue {
  owner: string;
  currentDir: string;
  setCurrentDir: (dir: string) => void;
  isLoading: boolean;
}

// Default values
const DEFAULT_OWNER = "lordcasser";

// Create context with default values
const TerminalConfigContext = createContext<TerminalConfigValue>({
  owner: DEFAULT_OWNER,
  currentDir: "",
  setCurrentDir: () => {},
  isLoading: true,
});

// Provider component
export function TerminalConfigProvider({ children }: { children: ReactNode }) {
  const [owner, setOwner] = useState(DEFAULT_OWNER);
  const [currentDir, setCurrentDir] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  // Fetch config from API on mount
  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const config = await apiClient.get<FrontendConfig>("/api/v1/settings");
        setOwner(config.owner || DEFAULT_OWNER);
      } catch (error) {
        console.error("Failed to fetch config:", error);
        // Use default owner on error
      } finally {
        setIsLoading(false);
      }
    };

    fetchConfig();
  }, []);

  // Get current directory from URL path on mount and URL changes
  // RESTful routing: /dir/{path} (path parameter, not query parameter)
  useEffect(() => {
    const updateCurrentDir = () => {
      if (typeof window !== "undefined") {
        const pathname = window.location.pathname;
        // Extract dir path from /dir/{path} or /dir/{path1}/{path2} routes
        if (pathname.startsWith("/dir/")) {
          setCurrentDir(decodeURIComponent(pathname.slice(5))); // Remove "/dir/" prefix
        } else {
          setCurrentDir(""); // Root directory
        }
      }
    };

    updateCurrentDir();

    // Listen for URL changes (popstate for navigation)
    window.addEventListener("popstate", updateCurrentDir);
    return () => window.removeEventListener("popstate", updateCurrentDir);
  }, []);

  // Memoized setCurrentDir
  const handleSetCurrentDir = useCallback((dir: string) => {
    setCurrentDir(dir);
  }, []);

  return (
    <TerminalConfigContext.Provider
      value={{
        owner,
        currentDir,
        setCurrentDir: handleSetCurrentDir,
        isLoading,
      }}
    >
      {children}
    </TerminalConfigContext.Provider>
  );
}

// Hook to use the context
export function useTerminalConfig() {
  const context = useContext(TerminalConfigContext);
  if (context === undefined) {
    throw new Error("useTerminalConfig must be used within a TerminalConfigProvider");
  }
  return context;
}