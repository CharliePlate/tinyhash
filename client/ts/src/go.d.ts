declare class Go {
  importObject: WebAssembly.Imports;
  run: (instance: WebAssembly.Instance) => void;
  constructor();
}

declare function hashSHA256(input: string): string;
declare function countLeadingZeros(hash: string): number;
declare function hashLoop(mhps: number): void;
declare function cancelLoop(): void;
declare function setCurrentMin(min: number): void;
