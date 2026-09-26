/* The worker owns Go's runtime; JavaScript never selects or draws traits. */
importScripts("wasm_exec.js");

const ready = (async () => {
  const go = new Go();
  const response = await fetch("roboticon.wasm");
  if (!response.ok) throw new Error(`Could not load renderer (${response.status})`);
  const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
  go.run(instance).catch(error => { throw error; });
  if (typeof self.roboticonRender !== "function") throw new Error("Go renderer did not start");
})();

self.onmessage = async ({data: {id, seed, size, format, transparent}}) => {
  try {
    await ready;
    const result = self.roboticonRender(seed, size, format, transparent);
    self.postMessage({id, ...result});
  } catch (error) {
    self.postMessage({id, error: error.message});
  }
};
