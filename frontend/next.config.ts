import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export',              // 启用静态导出
  trailingSlash: true,           // URL 带 trailing slash
  images: {
    unoptimized: true,           // 静态导出不支持图片优化
  },
};

export default nextConfig;
