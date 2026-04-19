import { readdir, readFile, rm, stat, writeFile } from "node:fs/promises";
import path from "node:path";
import { brotliCompress, constants, gzip } from "node:zlib";
import { promisify } from "node:util";

const gzipAsync = promisify(gzip);
const brotliCompressAsync = promisify(brotliCompress);

const outputDir = path.resolve("out");
const minSizeBytes = 1024;
const compressibleExtensions = new Set([
  ".css",
  ".html",
  ".js",
  ".json",
  ".map",
  ".md",
  ".svg",
  ".txt",
  ".xml",
]);

async function walk(dir) {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = await Promise.all(
    entries.map(async (entry) => {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        return walk(fullPath);
      }
      return [fullPath];
    }),
  );

  return files.flat();
}

function isCompressible(filePath) {
  const ext = path.extname(filePath).toLowerCase();
  return compressibleExtensions.has(ext);
}

async function writeCompressedVariant(filePath, extension, content) {
  const targetPath = `${filePath}${extension}`;
  if (content.length === 0) {
    await rm(targetPath, { force: true });
    return;
  }

  const sourceStats = await stat(filePath);
  if (content.length >= sourceStats.size) {
    await rm(targetPath, { force: true });
    return;
  }

  await writeFile(targetPath, content);
}

async function compressFile(filePath) {
  if (!isCompressible(filePath)) {
    return;
  }

  const sourceStats = await stat(filePath);
  if (sourceStats.size < minSizeBytes) {
    return;
  }

  const source = await readFile(filePath);
  const [brotliBuffer, gzipBuffer] = await Promise.all([
    brotliCompressAsync(source, {
      params: {
        [constants.BROTLI_PARAM_MODE]: constants.BROTLI_MODE_TEXT,
        [constants.BROTLI_PARAM_QUALITY]: 11,
      },
    }),
    gzipAsync(source, { level: constants.Z_BEST_COMPRESSION }),
  ]);

  await Promise.all([
    writeCompressedVariant(filePath, ".br", brotliBuffer),
    writeCompressedVariant(filePath, ".gz", gzipBuffer),
  ]);
}

async function main() {
  const files = await walk(outputDir);
  await Promise.all(files.map((filePath) => compressFile(filePath)));
}

main().catch((error) => {
  console.error("Failed to precompress exported assets:", error);
  process.exitCode = 1;
});
