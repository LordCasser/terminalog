import type { Metadata } from "next";
import "./globals.css";

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
        
        {/* Main Content */}
        <main className="flex-grow pb-24">
          {children}
        </main>
        
        {/* Bottom Command Prompt */}
        <footer className="fixed bottom-0 left-0 w-full z-50 bg-surface shadow-[0_-4px_20px_rgba(0,0,0,0.4)] border-t border-surface-container-highest">
          <div className="flex items-center gap-3 px-6 h-16 font-mono text-sm">
            <span className="text-tertiary font-bold">guest@blog:</span>
            <span className="text-secondary font-bold ml-1">~/lordcasser</span>
            <span className="text-on-surface-variant mx-1">$</span>
            <span className="w-2.5 h-5 bg-tertiary cursor-blink inline-block ml-1" />
          </div>
        </footer>
      </body>
    </html>
  );
}