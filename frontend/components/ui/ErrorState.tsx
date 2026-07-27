/**
 * Reusable error state with optional retry button.
 *
 * Brutalist terminal aesthetic: monospace message, bordered button,
 * no rounded corners. The retry button triggers a caller-provided
 * callback - typically incrementing a retryKey state that re-triggers
 * the data-fetching useEffect.
 */

interface ErrorStateProps {
  title: string;
  message: string;
  onRetry?: () => void;
}

export function ErrorState({ title, message, onRetry }: ErrorStateProps) {
  return (
    <div className="px-6 max-w-4xl mx-auto py-32">
      <div className="text-center">
        <h1 className="font-headline text-4xl text-outline mb-4">{title}</h1>
        <p className="text-on-surface-variant mb-6 font-mono text-sm">{message}</p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="font-mono text-sm text-primary border border-primary px-4 py-2 hover:bg-primary hover:text-surface transition-colors"
          >
            retry
          </button>
        )}
      </div>
    </div>
  );
}
