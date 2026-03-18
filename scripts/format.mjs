import { spawn } from "node:child_process";

const p = spawn("npm", ["run", "-ws", "format"], {
  stdio: "inherit",
  shell: process.platform === "win32"
});
p.on("exit", (code) => process.exit(code ?? 1));

