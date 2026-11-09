import { readFile, writeFile } from "node:fs/promises";

const version = process.argv[2];

if (!version || !/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(version)) {
  console.error("Usage: node .github/scripts/set-wails-version.mjs <major.minor.patch>");
  process.exit(1);
}

const configUrl = new URL("../../wails.json", import.meta.url);
const config = JSON.parse(await readFile(configUrl, "utf8"));

config.info = {
  ...(config.info ?? {}),
  productVersion: version,
};

await writeFile(configUrl, `${JSON.stringify(config, null, 2)}\n`, "utf8");
console.log(`Set wails.json productVersion to ${version}`);
