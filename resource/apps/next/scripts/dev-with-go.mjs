import { spawn } from "node:child_process";
import path from "node:path";

const appRoot = path.resolve(import.meta.dirname, "..");
const command = process.platform === "win32" ? "pnpm.cmd" : "pnpm";
const child = spawn(command, ["dev"], {
  cwd: appRoot,
  env: {
    ...process.env,
    GOOSEFORUM_ORIGIN:
      process.env.GOOSEFORUM_ORIGIN || "http://127.0.0.1:5234",
  },
  stdio: "inherit",
});

child.on("exit", (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exitCode = code ?? 1;
});
