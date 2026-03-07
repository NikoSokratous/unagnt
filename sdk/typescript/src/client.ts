import {
  Run,
  CreateRunRequest,
  CreateRunResponse,
  ListRunsResponse,
  RunEventsResponse,
} from './types';
import { APIError, NotFoundError, UnauthorizedError } from './errors';
import { streamRun as streamRunImpl, StreamRunOptions } from './stream';

export interface AgentRuntimeOptions {
  baseUrl?: string;
  apiKey?: string;
  timeout?: number;
}

const DEFAULT_BASE_URL = 'http://localhost:8080';
const DEFAULT_TIMEOUT = 30000;

/** AgentRuntime is the synchronous client for the Agent Runtime API. */
export class AgentRuntime {
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly timeout: number;

  constructor(options: AgentRuntimeOptions = {}) {
    this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/$/, '');
    this.apiKey = options.apiKey ?? '';
    this.timeout = options.timeout ?? DEFAULT_TIMEOUT;
  }

  private headers(): Record<string, string> {
    const h: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.apiKey) {
      h['Authorization'] = `Bearer ${this.apiKey}`;
    }
    return h;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const init: RequestInit = {
        method,
        headers: this.headers(),
        signal: controller.signal,
      };
      if (body !== undefined) {
        init.body = JSON.stringify(body);
      }

      const resp = await fetch(`${this.baseUrl}${path}`, init);
      clearTimeout(timeoutId);

      const text = await resp.text();
      let data: T | undefined;

      if (text) {
        try {
          data = JSON.parse(text) as T;
        } catch {
          // Leave data undefined for non-JSON responses
        }
      }

      if (!resp.ok) {
        if (resp.status === 401) {
          throw new UnauthorizedError();
        }
        if (resp.status === 404) {
          throw new NotFoundError(text);
        }
        throw new APIError(resp.status, text || resp.statusText, text);
      }

      return data as T;
    } catch (err) {
      clearTimeout(timeoutId);
      if (err instanceof APIError || err instanceof NotFoundError || err instanceof UnauthorizedError) {
        throw err;
      }
      throw new APIError(0, (err as Error).message);
    }
  }

  /** Create a new agent run. */
  async createRun(
    agentName: string,
    goal: string,
    opts?: Partial<CreateRunRequest>
  ): Promise<string> {
    const req: CreateRunRequest = {
      agent_name: agentName,
      goal,
      ...opts,
    };
    const resp = await this.request<CreateRunResponse>('POST', '/v1/runs', req);
    return resp.run_id;
  }

  /** Get details of a specific run. */
  async getRun(runId: string): Promise<Run> {
    return this.request<Run>('GET', `/v1/runs/${runId}`);
  }

  /** List recent runs. */
  async listRuns(limit = 100): Promise<string[]> {
    const resp = await this.request<ListRunsResponse>(
      'GET',
      `/v1/runs?limit=${limit}`
    );
    return resp.run_ids;
  }

  /** Cancel an ongoing run. */
  async cancelRun(runId: string): Promise<void> {
    await this.request('POST', `/v1/runs/${runId}/cancel`);
  }

  /** Get run event history. */
  async getRunEvents(runId: string): Promise<RunEventsResponse> {
    return this.request<RunEventsResponse>('GET', `/v1/runs/${runId}/events`);
  }

  /** Wait for a run to complete (poll until terminal state). */
  async waitForRun(
    runId: string,
    pollIntervalMs = 2000,
    timeoutMs?: number
  ): Promise<Run> {
    const start = Date.now();
    while (true) {
      const run = await this.getRun(runId);
      if (['completed', 'failed', 'cancelled'].includes(run.state)) {
        return run;
      }
      if (timeoutMs && Date.now() - start > timeoutMs) {
        throw new APIError(0, `Run ${runId} did not complete within ${timeoutMs}ms`);
      }
      await new Promise((r) => setTimeout(r, pollIntervalMs));
    }
  }

  /** Stream run events via Server-Sent Events. */
  async streamRun(runId: string, options: StreamRunOptions): Promise<void> {
    return streamRunImpl(this.baseUrl, runId, this.apiKey, options);
  }

  /** Check if the service is healthy. */
  async healthCheck(): Promise<boolean> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);
      const resp = await fetch(`${this.baseUrl}/health`, {
        signal: controller.signal,
      });
      clearTimeout(timeoutId);
      return resp.status === 200;
    } catch {
      return false;
    }
  }
}
