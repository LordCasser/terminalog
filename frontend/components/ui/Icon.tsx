/**
 * Inline SVG icon set — replaces Material Symbols Outlined font.
 *
 * Uses 1em sizing so icons inherit from Tailwind text-size classes
 * (text-xs, text-sm, text-[18px], etc.) exactly like the old font icons.
 * Stroke-only rendering matches the Material Symbols "outlined" variant
 * and the brutalist terminal aesthetic.
 */

export type IconName = "search" | "folder" | "file" | "chevron-down";

const ICON_PATHS: Record<IconName, string> = {
  // Magnifying glass — circle + handle
  search: "M11 19a8 8 0 100-16 8 8 0 000 16zM21 21l-4.35-4.35",
  // Folder outline
  folder: "M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z",
  // Document with text lines (was Material "description")
  file: "M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8",
  // Downward chevron (was Material "expand_more")
  "chevron-down": "M6 9l6 6 6-6",
};

interface IconProps {
  name: IconName;
  className?: string;
}

export function Icon({ name, className }: IconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="1em"
      height="1em"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden="true"
    >
      <path d={ICON_PATHS[name]} />
    </svg>
  );
}
