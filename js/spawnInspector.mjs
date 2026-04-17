/**
 * Start the gotunnel inspector binary on localhost (for nodetunnel or any Node tunnel).
 * Build once: go build -o inspector ./cmd/inspector
 *
 * @param {object} opts
 * @param {string|number} [opts.port=4040] listen port
 * @param {string} [opts.bin] path to inspector executable (default: GOTUNNEL_INSPECTOR_BIN or "inspector" on PATH)
 * @returns {{ stop: () => void, child: import('child_process').ChildProcessWithoutNullStreams }}
 */
import { spawn } from 'node:child_process';

export function spawnInspector(opts = {}) {
  const port = opts.port != null ? String(opts.port) : '4040';
  const bin = opts.bin || process.env.GOTUNNEL_INSPECTOR_BIN || 'inspector';
  const child = spawn(bin, [port], {
    stdio: 'inherit',
    env: process.env,
  });
  const stop = () => {
    try {
      child.kill('SIGTERM');
    } catch {
      /* ignore */
    }
  };
  return { stop, child };
}
