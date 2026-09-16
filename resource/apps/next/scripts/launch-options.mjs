export function applyLaunchOptions(args = process.argv.slice(2)) {
  const options = parseArgs(args);
  if (options.help) {
    console.log(`Usage: node start.mjs [options]

Options:
  --goose-origin <url>  GooseForum backend origin
  --hostname <host>     Listen host (default: 127.0.0.1)
  --port <port>         Listen port (default: 3012)
  --help                Show this help`);
    return false;
  }
  if (options.gooseOrigin) {
    const origin = new URL(options.gooseOrigin);
    if (origin.protocol !== "http:" && origin.protocol !== "https:")
      throw new Error("--goose-origin must use http or https");
    process.env.GOOSEFORUM_ORIGIN = origin.origin;
  }
  process.env.HOSTNAME = options.hostname || process.env.HOSTNAME || "127.0.0.1";
  process.env.PORT = options.port || process.env.PORT || "3012";
  return true;
}

function parseArgs(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--help") {
      options.help = true;
      continue;
    }
    const [name, inlineValue] = argument.split("=", 2);
    if (!["--goose-origin", "--hostname", "--port"].includes(name))
      throw new Error(`Unknown option: ${name}`);
    const value = inlineValue || args[++index];
    if (!value || value.startsWith("--"))
      throw new Error(`Missing value for ${name}`);
    if (name === "--goose-origin") options.gooseOrigin = value;
    else if (name === "--hostname") options.hostname = value;
    else {
      const port = Number(value);
      if (!Number.isInteger(port) || port < 1 || port > 65535)
        throw new Error("--port must be an integer between 1 and 65535");
      options.port = String(port);
    }
  }
  return options;
}
