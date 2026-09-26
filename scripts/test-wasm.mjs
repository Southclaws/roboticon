// Run after build-web.sh. Compare real WASM render output to native goldens.
import {readFile} from "node:fs/promises";
import {resolve} from "node:path";
import {pathToFileURL} from "node:url";
import assert from "node:assert/strict";

await import(pathToFileURL(resolve("web/dist/wasm_exec.js")));
const go = new Go();
const {instance} = await WebAssembly.instantiate(await readFile("web/dist/roboticon.wasm"), go.importObject);
go.run(instance);
for (const seed of ["robot-000", "robot-011"]) {
  for (const format of ["svg", "png"]) {
    const result = globalThis.roboticonRender(seed, 64, format, false);
    assert.equal(result.error, undefined);
    const bytes = Buffer.from(result.data, format === "png" ? "base64" : "utf8");
    assert.deepEqual(bytes, await readFile(`testdata/${seed}.${format}`));
  }
}
assert.match(globalThis.roboticonRender("hello", 0, "png", false).error, /size/);
assert.match(globalThis.roboticonRender("hello", 64, "gif", false).error, /format/);
const transparent = globalThis.roboticonRender("hello", 64, "svg", true).data;
assert.ok(!transparent.includes('<rect x="0" y="0" width="100"'));
console.log("WASM PNG/SVG output matches all four native goldens; options and errors verified.");
process.exit(0);
