import { spawn } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, "..");

const procs = [];

function loadEnv(dir) {
  const envPath = path.join(dir, ".env");
  const env = { ...process.env };
  try {
    const content = readFileSync(envPath, "utf8");
    for (const line of content.split("\n")) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;
      const [key, ...rest] = trimmed.split("=");
      env[key.trim()] = rest.join("=").trim();
    }
  } catch {
    // .env не найден — ничего страшного
  }
  return env;
}

function run(name, command, args, cwd) {
  const p = spawn(command, args, {
    cwd,
    stdio: "inherit",
    shell: true,
    env: loadEnv(cwd),
  });
  p.on("error", (err) => {
    console.error(`[${name}] Failed to start: ${err.message}`);
  });
  p.on("exit", (code) => {
    if (code !== 0) process.exit(code ?? 1);
  });
  procs.push({ name, p });
}

run("web", "npm", ["run", "dev", "-w", "apps/web"], rootDir);
run("api", "go", ["run", "."], path.join(rootDir, "apps", "api"));
run("indexer", "go", ["run", "."], path.join(rootDir, "apps", "indexer"));
run("worker", "go", ["run", "."], path.join(rootDir, "apps", "worker"));

process.on("SIGINT", () => {
  for (const { p } of procs) p.kill("SIGINT");
  process.exit(0);
});
