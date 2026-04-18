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
import { usePathname } from "next/navigation";
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

/**
 * Extract current directory from URL pathname.
 * - /dir/{path} → path (e.g., /dir/tech → "tech")
 * - /article/{path} → directory part (e.g., /article/tech/frontend/react-guide.md → "tech/frontend")
 * - /article/welcome.md → "" (root-level article)
 * - / or /aboutme → "" (root)
 */
function extractDirFromPathname(pathname: string): string {
  if (pathname.startsWith("/dir/")) {
    return decodeURIComponent(pathname.slice(5));
  } else if (pathname.startsWith("/article/")) {
    const articlePath = decodeURIComponent(pathname.slice(9));
    const lastSlashIndex = articlePath.lastIndexOf("/");
    if (lastSlashIndex > 0) {
      return articlePath.slice(0, lastSlashIndex);
    }
    return ""; // Root-level article
  }
  return ""; // Root or other pages
}

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
  const [isLoading, setIsLoading] = useState(true);
  const pathname = usePathname();
  const currentDir = extractDirFromPathname(pathname);

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

  // Kept for API compatibility with existing consumers. Navigation drives currentDir.
  const handleSetCurrentDir = useCallback((dir: string) => {
    void dir;
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
