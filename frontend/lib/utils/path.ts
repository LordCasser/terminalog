/**
 * Path encoding utilities for URL construction.
 *
 * These functions ensure that file paths containing special characters
 * (parentheses, spaces, CJK characters, etc.) are safely represented
 * in URL path segments.
 *
 * encodeURIComponent() alone is insufficient for URL paths because it
 * does not encode certain characters that are technically allowed by
 * RFC 3986 but can cause problems in practice:
 *
 *   - Parentheses () are sub-delims that some servers/proxies misinterpret
 *   - The exclamation mark ! is a sub-delim with similar issues
 *   - Single quotes ' can interfere with URL parsing
 *   - Asterisks * are sometimes problematic
 *
 * For maximum safety in URL path segments, we encode everything except
 * the unreserved characters defined in RFC 3986 (A-Z, a-z, 0-9, - _ . ~)
 * and the forward slash / which is our segment separator.
 */

/**
 * Encode a path for use in URL path segments.
 * Preserves "/" separators but encodes all other special characters
 * in each segment, including characters that encodeURIComponent() leaves
 * unencoded but which may cause issues in URL paths.
 */
export function encodePathForUrl(path: string): string {
  return path
    .split("/")
    .map(segment => {
      // Start with standard encodeURIComponent, then re-encode characters
      // it leaves unencoded but which are unsafe in URL paths.
      // encodeURIComponent does NOT encode: A-Z a-z 0-9 - _ . ! ~ * ' ( )
      // Of these, the following are safe in URL paths: A-Z a-z 0-9 - _ . ~
      // The following are NOT safe and need encoding: ! * ' ( )
      return encodeURIComponent(segment)
        .replace(/!/g, '%21')   // exclamation mark
        .replace(/\*/g, '%2A')  // asterisk
        .replace(/'/g, '%27')   // single quote
        .replace(/\(/g, '%28')  // left parenthesis
        .replace(/\)/g, '%29'); // right parenthesis
    })
    .join("/");
}