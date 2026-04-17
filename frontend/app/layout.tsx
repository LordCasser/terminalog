import type { Metadata } from "next";
import "./globals.css";
import { Navbar } from "@/components/brutalist";
import { CommandPrompt } from "@/components/command";
import { ClientLayout } from "@/components/layout/ClientLayout";

export const metadata: Metadata = {
  title: "Terminalog | Terminal Editorial Blog",
  description: "A Brutalist Compiler-style terminal blog system",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <head>
        {/* Google Fonts - Dracula Spectrum Typography */}
        <link
          href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&display=swap"
          rel="stylesheet"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600&display=swap"
          rel="stylesheet"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap"
          rel="stylesheet"
        />
        {/* Material Symbols */}
        <link
          href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="min-h-screen flex flex-col bg-surface-lowest text-on-surface selection:bg-primary-container selection:text-on-primary-container">
        {/* Terminal Fog Effect */}
        <div className="fixed inset-0 pointer-events-none opacity-5 bg-[radial-gradient(circle_at_50%_50%,#bd93f9_0%,transparent_50%)]" />
        
        {/* Client-side Layout with Config Context */}
        <ClientLayout>
          {/* Top Navigation Bar - Public Component */}
          <Navbar />
          
          {/* Main Content */}
          <main className="flex-grow pb-24 pt-16">
            {children}
          </main>
          
          {/* Bottom Command Prompt - Public Component */}
          <CommandPrompt />
        </ClientLayout>
      </body>
    </html>
  );
}