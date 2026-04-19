import { PHASE_DEVELOPMENT_SERVER } from "next/constants.js";

export default function nextConfig(phase) {
  const isDevServer = phase === PHASE_DEVELOPMENT_SERVER;

  /** @type {import('next').NextConfig} */
  return {
    compress: true,
    // Keep static export for production builds, but disable it in `next dev`
    // so dynamic routes can be debugged directly.
    output: isDevServer ? undefined : "export",
    trailingSlash: true,
    images: {
      unoptimized: true,
    },
    pageExtensions: ["js", "jsx", "ts", "tsx"],
  };
}
