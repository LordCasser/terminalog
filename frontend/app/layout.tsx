import type { Metadata } from "next";
import { Space_Grotesk, Inter, JetBrains_Mono, Fira_Code } from "next/font/google";
import "./globals.css";
import { Navbar } from "@/components/brutalist";
import { CommandPrompt } from "@/components/command";
import { ClientLayout } from "@/components/layout/ClientLayout";

// Self-hosted fonts via next/font — build-time download, fully offline static export.
// Each generates a CSS variable on <html> that globals.css @theme references.
const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-space-grotesk",
  display: "swap",
});
const inter = Inter({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600"],
  variable: "--font-inter",
  display: "swap",
});
const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500", "700"],
  variable: "--font-jetbrains-mono",
  display: "swap",
});
const firaCode = Fira_Code({
  subsets: ["latin"],
  weight: ["400", "500", "700"],
  variable: "--font-fira-code",
  display: "swap",
});

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
    <html
      lang="en"
      className={`dark ${spaceGrotesk.variable} ${inter.variable} ${jetbrainsMono.variable} ${firaCode.variable}`}
    >
      <body className="min-h-screen flex flex-col bg-surface-lowest text-on-surface selection:bg-primary-container selection:text-on-primary-container">
        {/* Terminal Fog Effect */}
        <div className="fixed inset-0 pointer-events-none opacity-5 bg-[radial-gradient(circle_at_50%_50%,#bd93f9_0%,transparent_50%)]" />

        {/* Client-side Layout with Config Context */}
        <ClientLayout>
          {/* Top Navigation Bar - Public Component */}
          <Navbar />
          
          {/* Main Content - single source of page padding (navbar + command prompt clearance) */}
          <main className="flex-grow pt-20 pb-28">
            {children}
          </main>
          
          {/* Bottom Command Prompt - Public Component */}
          <CommandPrompt />
        </ClientLayout>
      </body>
    </html>
  );
}
