/**
 * Navbar Component - TopAppBar
 * 
 * Fixed navigation bar with:
 * - Logo/Path display (JetBrains Mono)
 * - Navigation links (POSTS, ABOUTME)
 * - Search button
 */

"use client";

import Link from "next/link";

interface NavbarProps {
  currentPath?: string;
}

export function Navbar({ currentPath = "~/lordcasser" }: NavbarProps) {
  return (
    <header className="fixed top-0 w-full z-50 bg-surface text-primary-container font-mono uppercase tracking-tighter text-sm flex justify-between items-center px-6 py-4">
      <div className="flex items-center gap-6">
        {/* Path Display */}
        <Link href="/" className="text-lg font-bold text-primary normal-case hover:text-secondary transition-colors">
          {currentPath}
        </Link>
        
        {/* Navigation Links */}
        <nav className="hidden md:flex gap-6">
          <Link 
            href="/" 
            className="text-primary-container font-bold border-b-2 border-primary-container pb-1 transition-colors duration-150"
          >
            posts
          </Link>
          <Link 
            href="/aboutme" 
            className="text-outline hover:bg-surface-container-highest hover:text-primary transition-colors duration-150 px-2 py-1"
          >
            ABOUTME
          </Link>
        </nav>
      </div>
      
      {/* Search Button */}
      <div className="flex items-center gap-4">
        <button className="hover:bg-surface-container-highest hover:text-primary p-1 transition-colors duration-150">
          <span className="material-symbols-outlined">search</span>
        </button>
      </div>
    </header>
  );
}