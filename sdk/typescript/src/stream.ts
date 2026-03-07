import { StreamChunk } from './types';

export interface StreamRunOptions {
  onEvent?: (chunk: StreamChunk) => void;
  onError?: (err: Error) => void;
}

/** Stream run events via Server-Sent Events. */
export async function streamRun(
  baseUrl: string,
  runId: string,
  apiKey: string,
  options: StreamRunOptions
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/v1/runs/${runId}/stream`;
  const headers: Record<string, string> = { Accept: 'text/event-stream' };
  if (apiKey) {
    headers['Authorization'] = `Bearer ${apiKey}`;
  }

  const resp = await fetch(url, { headers });
  if (!resp.ok) {
    throw new Error(`Stream failed: status ${resp.status}`);
  }

  const reader = resp.body?.getReader();
  if (!reader) {
    throw new Error('No response body');
  }

  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split(/\r?\n/);
      buffer = lines.pop() ?? '';

      for (const line of lines) {
        if (!line || line.startsWith(':')) continue;
        if (line.startsWith('data: ')) {
          const json = line.slice(6);
          try {
            const chunk = JSON.parse(json) as StreamChunk;
            options.onEvent?.(chunk);
          } catch (err) {
            options.onError?.(err as Error);
          }
        }
      }
    }
  } finally {
    reader.releaseLock();
  }
}
