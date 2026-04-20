export const COMMANDS = ["search", "open", "cd", "help"] as const;
export const HISTORY_KEY = "terminalog_command_history";
export const MAX_HISTORY_SIZE = 100;

type RouterLike = {
  push: (href: string) => void;
};

/**
 * Encode a path for use in URL path segments.
 * Preserves "/" separators but encodes special characters in each segment.
 * 
 * Note: encodeURIComponent() does not encode () which are unsafe in URL paths.
 * We manually encode parentheses as %28 and %29 to ensure proper URL handling.
 */
export function encodePathForUrl(path: string): string {
  return path
    .split("/")
    .map(segment => encodeURIComponent(segment).replace(/\(/g, '%28').replace(/\)/g, '%29'))
    .join("/");
}

export function navigateToPath(router: RouterLike, path: string) {
  if (path.endsWith(".md")) {
    router.push(`/article/${encodePathForUrl(path)}`);
    return;
  }

  router.push(`/dir/${encodePathForUrl(path)}`);
}

export function resolvePath(currentDir: string, target: string): string {
  const normalizedTarget = target.trim();
  const segments = normalizedTarget.split("/");
  let currentSegments = currentDir ? currentDir.split("/") : [];

  for (const seg of segments) {
    if (seg === "..") {
      if (currentSegments.length > 0) {
        currentSegments = currentSegments.slice(0, -1);
      }
      continue;
    }

    if (seg === "." || seg === "") {
      continue;
    }

    currentSegments.push(seg);
  }

  return currentSegments.join("/");
}

export function resolveCdPath(currentDir: string, target: string): string {
  const normalizedTarget = target.trim();

  if (normalizedTarget === "." || normalizedTarget === "") {
    return currentDir;
  }

  if (normalizedTarget.startsWith("/")) {
    const absolutePath = normalizedTarget.slice(1);
    return absolutePath || "";
  }

  return resolvePath(currentDir, normalizedTarget);
}
