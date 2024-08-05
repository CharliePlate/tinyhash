enum WorkerAction {
  Init = "initialized",
  StartedHash = "startedHash",
  StoppedHash = "stoppedHash",
  UpdateMin = "updateMin",
  UpdateStats = "updateStats",
  Error = "error",
}

const workerPath = "/static/js/worker.js";
let workers: Worker[] = [];

async function createWorker(wasmPath: string, maxHPS: number, id: number) {
  const worker = new Worker(workerPath);

  worker.onmessage = function (event: MessageEvent) {
    switch (event.data.type) {
      case WorkerAction.Init:
        console.log("WASM initialized in worker");
        worker.postMessage({ action: "startHash", maxHPS });
        setupThreadStats();
        break;
      case WorkerAction.StartedHash:
        console.log("Hash calculation started");
        break;
      case WorkerAction.StoppedHash:
        console.log("Hash calculation stopped");
        break;
      case WorkerAction.UpdateStats:
        handleReportStats(
          event.data.totalHashes,
          event.data.hashRate,
          event.data.threadId,
        );
        break;
      case WorkerAction.UpdateMin:
        updateHashInputs(event.data.newHash, event.data.input);
        break;
      case WorkerAction.Error:
        console.error("Error in worker", event.data.error);
        break;
      default:
        console.error("Unknown message type", event.data);
    }
  };

  worker.postMessage({ action: "init", wasmPath: wasmPath, id });
  return worker;
}

async function initWorkers(
  threadCount: number,
  wasmPath: string,
  maxHPS: number,
) {
  stopAllWorkers();
  for (let i = 0; i < threadCount; i++) {
    const worker = await createWorker(wasmPath, maxHPS, i);
    workers.push(worker);
  }
}

function stopAllWorkers() {
  workers.forEach((worker) => {
    worker.postMessage({ action: "stopHash" });
    worker.terminate();
  });
  workers = [];
}

function updateHashInputs(newMin: string, input: string) {
  const hashElement = document.getElementById("hash");
  if (hashElement) {
    hashElement.innerText = newMin;
  }

  const hashInput = document.getElementById("hashInput");
  if (hashInput) {
    hashInput.innerText = input;
  }

  workers.forEach((worker) => {
    worker.postMessage({ action: "updateHash", min: newMin });
  });
}

async function startHashButton() {
  const threadCountInput = document.getElementById(
    "threadCount",
  ) as HTMLInputElement;
  const threadCount = parseInt(threadCountInput.value, 10);

  const maxHPSInput = document.getElementById("maxHPS") as HTMLInputElement;
  const maxMHPS = parseInt(maxHPSInput.value, 10);

  await initWorkers(threadCount, "/static/wasm/main.wasm", maxMHPS);
}

function stopHashButton() {
  stopAllWorkers();
}

function setupThreadStats() {
  const statsDiv = document.getElementById("threadStats");
  if (statsDiv) {
    statsDiv.textContent = "Thread Stats:";
    for (let i = 0; i < workers.length; i++) {
      const threadStatDiv = document.createElement("div");
      threadStatDiv.id = `threadStat${i}`;
      statsDiv.appendChild(threadStatDiv);
    }
  } else {
    console.error("Could not find threadStats");
  }
}

function handleReportStats(
  totalHashes: number,
  hashRate: number,
  threadId: number,
) {
  const threadStatDiv = document.getElementById(`threadStat${threadId}`);
  if (threadStatDiv) {
    threadStatDiv.innerText = `Thread ${threadId}: ${totalHashes} hashes at ${hashRate} H/s`;
  }
}

globalThis.updateHashInputs = updateHashInputs;

document.addEventListener("DOMContentLoaded", () => {
  const calcHashButton = document.getElementById("calcHash");
  if (calcHashButton) {
    calcHashButton.addEventListener("click", startHashButton);
  }

  const cancelHashButton = document.getElementById("cancelHash");
  if (cancelHashButton) {
    cancelHashButton.addEventListener("click", stopHashButton);
  }
});
