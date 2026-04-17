/**
 * Navbar Component - TopAppBar (Public Component)
 * 
 * Fixed navigation bar integrated in layout.tsx for all pages.
 * Features (v1.6.1):
 * - Left: Logo/Path display (~/{owner}/{currentDir}, JetBrains Mono uppercase)
 * - Right: POSTS and ABOUTME navigation links + Search icon (right-aligned)
 * - Selected state: after pseudo-element underline + color change (doesn't affect text baseline)
 * - Search button (triggers focus on CommandPrompt)
 * - Path sync with CommandPrompt via TerminalConfig context
 */

"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { FOCUS_COMMAND_INPUT } from "@/components/command";
import { useTerminalConfig } from "@/lib/hooks/useTerminalConfig";

// Navigation link component with selected state
// Uses after pseudo-element for underline so it doesn't affect text baseline alignment
function NavLink({ 
  href, 
  label, 
  isSelected 
}: { 
  href: string; 
  label: string; 
  isSelected: boolean;
}) {
  return (
    <Link 
      href={href} 
      className={`
        font-bold transition-colors duration-150 relative
        after:absolute after:left-0 after:right-0 after:-bottom-1.5 after:h-0.5
        ${isSelected 
          ? "text-primary-container after:content-[''] after:bg-primary-container" 
          : "text-outline hover:bg-surface-container-highest hover:text-primary px-2 py-1"
        }
      `}
    >
      {label}
    </Link>
  );
}

export function Navbar() {
  // Handle search icon click - focus command input
  const handleSearchClick = () => {
    window.dispatchEvent(new Event(FOCUS_COMMAND_INPUT));
  };

  // Get owner and currentDir from context
  const { owner, currentDir } = useTerminalConfig();
  
  // Get current pathname for selected state
  const pathname = usePathname();
  const isPostsSelected = pathname === "/" || pathname.startsWith("/?dir=");
  const isAboutMeSelected = pathname === "/aboutme" || pathname === "/aboutme/";

  return (
    <header className="fixed top-0 w-full z-50 bg-surface text-primary-container font-mono uppercase tracking-tighter text-sm flex justify-between items-center px-6 py-4">
      {/* Left: Path Display - Dynamic path ~/{owner}/{currentDir} */}
      <Link href="/" className="text-lg font-bold text-primary hover:text-secondary transition-colors">
        ~/{owner}{currentDir ? `/${currentDir}` : ""}
      </Link>
      
      {/* Right: Navigation Links + Search Button */}
      <div className="flex items-center gap-6">
        {/* Navigation Links - All uppercase, right-aligned */}
        <nav className="hidden md:flex gap-6">
          <NavLink href="/" label="POSTS" isSelected={isPostsSelected} />
          <NavLink href="/aboutme" label="ABOUTME" isSelected={isAboutMeSelected} />
        </nav>
        
        {/* Search Button - Click to focus command input */}
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