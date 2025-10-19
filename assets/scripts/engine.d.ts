// Engine-side functions you expose to JS
declare function log(msg: string): void;

// called by the engine
export function onSpawn(id: number): void;
export function onUpdate(id: number, dt: number): void;