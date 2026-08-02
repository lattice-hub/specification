import { cp, mkdir, readdir, rm, writeFile } from "node:fs/promises";
import { basename, dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";

const packageDirectory = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const repositoryDirectory = resolve(packageDirectory, "../..");
const apiDirectory = join(repositoryDirectory, "api");
const protoDirectory = join(packageDirectory, "proto");
const generatedDirectory = join(packageDirectory, "src", "generated");

async function collectProtoFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      files.push(...await collectProtoFiles(path));
    } else if (entry.isFile() && entry.name.endsWith(".proto")) {
      files.push(path);
    }
  }
  return files.sort();
}

await rm(protoDirectory, { recursive: true, force: true });
await rm(generatedDirectory, { recursive: true, force: true });
await mkdir(protoDirectory, { recursive: true });
await mkdir(generatedDirectory, { recursive: true });

const sourceFiles = await collectProtoFiles(apiDirectory);
const seenNames = new Set();
for (const sourceFile of sourceFiles) {
  const name = basename(sourceFile);
  if (seenNames.has(name)) {
    throw new Error(`duplicate proto basename: ${name}`);
  }
  seenNames.add(name);
  await cp(sourceFile, join(protoDirectory, name));
}

for (const licenseName of ["LICENSE-APACHE", "LICENSE-MIT"]) {
  await cp(
    join(repositoryDirectory, "source", "rust", "pole-specification", licenseName),
    join(packageDirectory, licenseName),
  );
}

const generator = join(packageDirectory, "node_modules", ".bin", "proto-loader-gen-types");
const result = spawnSync(generator, [
  "--includeComments",
  "--includeDirs",
  protoDirectory,
  "--grpcLib",
  "@grpc/grpc-js",
  "--outDir",
  generatedDirectory,
  ...[...seenNames].sort().map((name) => join(protoDirectory, name)),
], { stdio: "inherit" });

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

const manifest = [...seenNames].sort();
await writeFile(
  join(packageDirectory, "src", "proto-files.json"),
  JSON.stringify(manifest, null, 2) + "\n",
);
