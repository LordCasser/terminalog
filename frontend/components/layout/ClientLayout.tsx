/**
 * ClientLayout Component
 * 
 * Client-side wrapper for layout components that need context access.
 * Contains TerminalConfigProvider for sharing owner and currentDir across components.
 */

"use client";

import { ReactNode } from "react";
import { TerminalConfigProvider } from "@/lib/hooks/useTerminalConfig";

interface ClientLayoutProps {
  children: ReactNode;
}

export function ClientLayout({ children }: ClientLayoutProps) {
  return (
    <TerminalConfigProvider>
      {children}
    </TerminalConfigProvider>
  );
}