import { spawn } from "node:child_process";

const procs = [];

function run(name, command, args, cwd) {
  const p = spawn(command, args, {
    cwd,
    stdio: "inherit",
    shell: process.platform === "win32"
  });
  p.on("exit", (code) => {
    if (code !== 0) process.exit(code ?? 1);
  });
  procs.push({ name, p });
}

run("web", "npm", ["run", "dev", "-w", "apps/web"], process.cwd());
run("api", "go", ["run", "."], new URL("../apps/api", import.meta.url).pathname);
run(
  "indexer",
  "go",
  ["run", "."],
  new URL("../apps/indexer", import.meta.url).pathname
);
run(
  "worker",
  "go",
  ["run", "."],
  new URL("../apps/worker", import.meta.url).pathname
);

process.on("SIGINT", () => {
  for (const { p } of procs) p.kill("SIGINT");
  process.exit(0);
});

