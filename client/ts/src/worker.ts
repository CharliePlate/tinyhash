importScripts("/static/js/wasm_exec.js");

const go = new Go();
let threadId: number;

self.onmessage = async function (event) {
  const { wasmPath, action, maxHPS, newHash, id } = event.data;
  if (action === "init") {
    try {
      const response = await fetch(wasmPath);
      if (!response.ok) {
        throw new Error("Failed to fetch WASM file: " + response.statusText);
      }
      const bytes = await response.arrayBuffer();
      const { instance } = await WebAssembly.instantiate(
        bytes,
        go.importObject,
      );
      go.run(instance);
      threadId = id;
      self.postMessage({ type: "initialized" });
    } catch (error: any) {
      console.error("Error initializing WASM:", error);
      self.postMessage({ type: "error", message: error.toString() });
    }
  } else if (action === "startHash") {
    startHash(maxHPS);
  } else if (action === "stopHash") {
    stopHash();
  } else if (action === "updateHash") {
    setCurrentMin(newHash);
  }
};

function startHash(maxHPS: number) {
  hashLoop(maxHPS);
  console.log("Hash calculation started");
  self.postMessage({ type: "startedHash" });
}

function stopHash() {
  cancelLoop();
  console.log("Hash calculation stopped");
  self.postMessage({ type: "stoppedHash" });
}

function updateHash(newHash: string, input: string) {
  self.postMessage({ type: "updateMin", newHash, input });
}

function updateStats(totalHashes: number, hps: number) {
  self.postMessage({
    type: "updateStats",
    threadId,
    totalHashes,
    hashRate: hps,
  });
}
