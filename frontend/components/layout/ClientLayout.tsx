/**
 * ClientLayout Component
 * 
 * Client-side wrapper for layout components that need context access.
 * Contains TerminalConfigProvider for sharing owner and currentDir across components.
 * Includes modal components (HelpModal, SearchResultsModal, PathCompletionModal).
 */

"use client";

import { ReactNode } from "react";
import { TerminalConfigProvider } from "@/lib/hooks/useTerminalConfig";
import { HelpModal, SearchResultsModal, PathCompletionModal } from "@/components/modal";

interface ClientLayoutProps {
  children: ReactNode;
}

export function ClientLayout({ children }: ClientLayoutProps) {
  return (
    <TerminalConfigProvider>
      {children}
      {/* Modal Components */}
      <HelpModal />
      <SearchResultsModal />
      <PathCompletionModal />
    </TerminalConfigProvider>
  );
}