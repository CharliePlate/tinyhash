let ready = false;
let wasmReadyPromise = new Promise<void>((resolve) => {
  (self as any).wasmReady = () => {
    ready = true;
    resolve();
  };
});

let threadId: number;

self.onmessage = async function (event: MessageEvent<any>) {
  const { wasmPath, action, maxHPS, newHash, id } = event.data;

  if (action === "init") {
    try {
      importScripts("/static/js/wasm_exec.js");
      const go = new (self as any).Go();
      const result = await WebAssembly.instantiateStreaming(
        fetch(wasmPath!),
        go.importObject,
      );
      go.run(result.instance);
      threadId = id!;
      (self as any).wasmReady();
      self.postMessage({ type: "initialized" });
    } catch (error) {
      console.error("Error initializing WASM:", error);
      self.postMessage({ type: "error", message: (error as Error).toString() });
    }
  } else {
    if (!ready) {
      console.log("Waiting for WASM to be ready");
      await wasmReadyPromise;
    }

    switch (action) {
      case "startHash":
        startHash(maxHPS!);
        break;
      case "stopHash":
        stopHash();
        break;
      case "updateHash":
        if (typeof (self as any).setCurrentMin === "function") {
          (self as any).setCurrentMin(newHash);
        } else {
          console.error("setCurrentMin is not available");
        }
        break;
      default:
        console.error("Unknown action:", action);
    }
  }
};

function startHash(maxHPS: number): void {
  (self as any).hashLoop(maxHPS);
  console.log("Hash calculation started");
  self.postMessage({ type: "startedHash" });
}

function stopHash(): void {
  (self as any).cancelLoop();
  console.log("Hash calculation stopped");
  self.postMessage({ type: "stoppedHash" });
}

function updateHash(newHash: string, input: string): void {
  self.postMessage({ type: "updateMin", newHash, input });
}

function updateStats(totalHashes: number, hps: number): void {
  self.postMessage({
    type: "updateStats",
    threadId,
    totalHashes,
    hashRate: hps,
  });
}
