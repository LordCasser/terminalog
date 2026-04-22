/**
 * FilingFooter Component
 * 
 * Invisible footer for Chinese regulatory filing information (ICP备案, 公安备案).
 * The content is present in the DOM for crawler/program detection but visually hidden
 * from users using off-screen positioning.
 * 
 * - Text content is machine-readable (search engines, compliance scanners)
 * - CSS positions the element off-screen so it's invisible to users
 * - Only renders when filing information is configured
 */

"use client";

import { useTerminalConfig } from "@/lib/hooks/useTerminalConfig";

export function FilingFooter() {
  const { icpFiling, icpFilingURL, policeFiling, policeFilingURL } = useTerminalConfig();

  // Only render when at least one filing field is configured
  if (!icpFiling && !policeFiling) {
    return null;
  }

  return (
    <footer
      role="contentinfo"
      className="fixed -left-[9999px] w-1 h-1 overflow-hidden opacity-0 pointer-events-none"
    >
      <p>
        {icpFiling && (
          <span>
            ICP备案:{" "}
            <a href={icpFilingURL || "https://beian.miit.gov.cn/"}>{icpFiling}</a>
          </span>
        )}
        {icpFiling && policeFiling && " | "}
        {policeFiling && (
          <span>
            公安备案:{" "}
            <a href={policeFilingURL || "#"}>{policeFiling}</a>
          </span>
        )}
      </p>
    </footer>
  );
}
