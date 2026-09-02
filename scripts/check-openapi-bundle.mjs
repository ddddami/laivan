import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const temporaryDirectory = mkdtempSync(join(tmpdir(), "laivan-openapi-"));
const generatedBundle = join(temporaryDirectory, "bundle.yaml");
const pnpm = process.platform === "win32" ? "pnpm.cmd" : "pnpm";

try {
  execFileSync(
    pnpm,
    [
      "exec",
      "redocly",
      "bundle",
      "openapi/openapi.yaml",
      "--output",
      generatedBundle,
      "--ext",
      "yaml",
    ],
    { stdio: "inherit" },
  );

  const committed = readFileSync("openapi/bundle.yaml", "utf8");
  const generated = readFileSync(generatedBundle, "utf8");

  if (committed !== generated) {
    process.stderr.write(
      "openapi/bundle.yaml is stale. Run pnpm contract:generate and commit the result.\n",
    );
    process.exitCode = 1;
  }
} finally {
  rmSync(temporaryDirectory, { recursive: true, force: true });
}
