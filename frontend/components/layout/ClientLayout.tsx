/**
 * ClientLayout Component
 * 
 * Client-side wrapper for layout components that need context access.
 * Contains TerminalConfigProvider for sharing owner and currentDir across components.
 * Includes modal components (HelpModal, SearchResultsModal, PathCompletionModal).
 * Includes FilingFooter for regulatory compliance (invisible to users).
 */

"use client";

import { ReactNode } from "react";
import { TerminalConfigProvider } from "@/lib/hooks/useTerminalConfig";
import { HelpModal, SearchResultsModal, PathCompletionModal } from "@/components/modal";
import { FilingFooter } from "@/components/layout/FilingFooter";

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
      {/* Regulatory Filing Footer (invisible, crawler-detectable) */}
      <FilingFooter />
    </TerminalConfigProvider>
  );
}