/**
 * Navbar Component - TopAppBar (Public Component)
 * 
 * Fixed navigation bar integrated in layout.tsx for all pages.
 * Features:
 * - Logo/Path display (~/owner/currentDir, JetBrains Mono uppercase)
 * - Navigation links (POSTS, ABOUTME)
 * - Search button (triggers focus on CommandPrompt)
 * - Path sync with CommandPrompt via TerminalConfig context
 */

"use client";

import Link from "next/link";
import { FOCUS_COMMAND_INPUT } from "@/components/command";
import { useTerminalConfig } from "@/lib/hooks/useTerminalConfig";

export function Navbar() {
  // Handle search icon click - focus command input
  const handleSearchClick = () => {
    window.dispatchEvent(new Event(FOCUS_COMMAND_INPUT));
  };

  // Get owner and currentDir from context
  const { owner, currentDir } = useTerminalConfig();

  return (
    <header className="fixed top-0 w-full z-50 bg-surface text-primary-container font-mono uppercase tracking-tighter text-sm flex justify-between items-center px-6 py-4">
      <div className="flex items-center gap-6">
        {/* Path Display - Dynamic path ~/{owner}/{currentDir} */}
        <Link href="/" className="text-lg font-bold text-primary hover:text-secondary transition-colors">
          ~/{owner}{currentDir ? `/${currentDir}` : ""}
        </Link>
        
        {/* Navigation Links - All uppercase */}
        <nav className="hidden md:flex gap-6">
          <Link 
            href="/" 
            className="text-primary-container font-bold border-b-2 border-primary-container pb-1 transition-colors duration-150"
          >
            POSTS
          </Link>
          <Link 
            href="/aboutme" 
            className="text-outline hover:bg-surface-container-highest hover:text-primary transition-colors duration-150 px-2 py-1"
          >
            ABOUTME
          </Link>
        </nav>
      </div>
      
      {/* Search Button - Click to focus command input */}
      <div className="flex items-center gap-4">
        <button 
          onClick={handleSearchClick}
          className="hover:bg-surface-container-highest hover:text-primary p-1 transition-colors duration-150"
          aria-label="Search"
        >
          <span className="material-symbols-outlined">search</span>
        </button>
      </div>
    </header>
  );
}