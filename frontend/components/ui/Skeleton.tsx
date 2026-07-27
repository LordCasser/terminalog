/**
 * Brutalist skeleton loading placeholders.
 *
 * Design: no animations, no rounded corners, no pulse - pure solid
 * surface blocks matching the terminal aesthetic. Block characters
 * (░) were considered but rejected because monospace glyph widths
 * vary across fonts; solid div blocks give precise dimension control
 * and stay aligned with the grid.
 *
 * SkeletonTable mirrors ArticleTable's 5-column responsive layout
 * (Created/Editors/Latest Commit hidden on mobile) so the loading
 * state occupies the same footprint as the loaded table.
 */

interface SkeletonProps {
  className?: string;
}

/** Base skeleton block - a solid surface-colored div. */
export function Skeleton({ className = "" }: SkeletonProps) {
  return <div className={`bg-surface-container-high ${className}`} />;
}

/** Table skeleton matching ArticleTable's 5-column structure. */
export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <div className="w-full overflow-x-auto">
      <div className="min-w-[480px] md:min-w-[1000px]">
        {/* Header row - placeholder bars, no labels (loading state is transient) */}
        <div className="flex items-center px-6 py-3 bg-surface-container border-b border-surface-container gap-4">
          <Skeleton className="h-2.5 w-14 hidden md:block" />
          <Skeleton className="h-2.5 w-16" />
          <Skeleton className="h-2.5 w-14 hidden md:block" />
          <Skeleton className="h-2.5 w-28" />
          <Skeleton className="h-2.5 w-20 hidden md:block" />
        </div>
        {/* Data rows */}
        {Array.from({ length: rows }).map((_, i) => (
          <div
            key={i}
            className="flex items-center px-6 py-5 border-b border-surface-container gap-4"
          >
            <Skeleton className="h-3 w-20 hidden md:block" />
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-3 w-24 hidden md:block" />
            <Skeleton className="h-3 flex-1 max-w-[12rem]" />
            <Skeleton className="h-3 w-24 hidden md:block" />
          </div>
        ))}
      </div>
    </div>
  );
}

/** Article content skeleton - header block + paragraph blocks. */
export function SkeletonArticle() {
  return (
    <div className="px-6 max-w-4xl mx-auto">
      {/* Header: tag/version/date row + title */}
      <section className="mb-16">
        <div className="flex items-center gap-4 mb-4">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-20" />
        </div>
        <Skeleton className="h-12 w-3/4" />
      </section>
      {/* Paragraph blocks with breathing gaps */}
      <div className="space-y-6">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-5/6" />
        <div className="h-8" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-4/5" />
        <div className="h-8" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
      </div>
    </div>
  );
}
