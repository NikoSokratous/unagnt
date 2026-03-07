import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { AgentRuntime, APIError, NotFoundError, UnauthorizedError } from '../src';

describe('AgentRuntime', () => {
  const originalFetch = globalThis.fetch;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    globalThis.fetch = mockFetch as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('createRun returns run_id', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ run_id: 'run-123' })),
    });

    const client = new AgentRuntime({ baseUrl: 'http://localhost:8080' });
    const runId = await client.createRun('demo-agent', 'test goal');

    expect(runId).toBe('run-123');
    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/runs',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent_name: 'demo-agent', goal: 'test goal' }),
      })
    );
  });

  it('getRun returns run details', async () => {
    const run = {
      run_id: 'run-123',
      agent_name: 'demo',
      goal: 'task',
      state: 'completed',
      step_count: 2,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:01:00Z',
    };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify(run)),
    });

    const client = new AgentRuntime();
    const result = await client.getRun('run-123');

    expect(result).toEqual(run);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/v1/runs/run-123'),
      expect.objectContaining({ method: 'GET' })
    );
  });

  it('listRuns returns run_ids', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ run_ids: ['a', 'b'] })),
    });

    const client = new AgentRuntime();
    const ids = await client.listRuns(10);

    expect(ids).toEqual(['a', 'b']);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('limit=10'),
      expect.any(Object)
    );
  });

  it('cancelRun succeeds', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve('{}'),
    });

    const client = new AgentRuntime();
    await client.cancelRun('run-123');

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/v1/runs/run-123/cancel'),
      expect.objectContaining({ method: 'POST' })
    );
  });

  it('getRunEvents returns events', async () => {
    const events = { run_id: 'run-123', events: [{ type: 'step', data: {} }] };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify(events)),
    });

    const client = new AgentRuntime();
    const result = await client.getRunEvents('run-123');

    expect(result).toEqual(events);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/v1/runs/run-123/events'),
      expect.any(Object)
    );
  });

  it('includes Bearer token when apiKey is set', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify({ run_ids: [] })),
    });

    const client = new AgentRuntime({ apiKey: 'secret' });
    await client.listRuns();

    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer secret',
        }),
      })
    );
  });

  it('throws NotFoundError on 404', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      text: () => Promise.resolve('run not found'),
    });

    const client = new AgentRuntime();
    await expect(client.getRun('missing')).rejects.toThrow(NotFoundError);
  });

  it('throws UnauthorizedError on 401', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      text: () => Promise.resolve('unauthorized'),
    });

    const client = new AgentRuntime();
    await expect(client.getRun('run-123')).rejects.toThrow(UnauthorizedError);
  });

  it('throws APIError on other HTTP errors', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: () => Promise.resolve('internal error'),
    });

    const client = new AgentRuntime();
    await expect(client.getRun('run-123')).rejects.toThrow(APIError);
  });

  it('waitForRun returns when run completes', async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        text: () =>
          Promise.resolve(
            JSON.stringify({
              run_id: 'r1',
              agent_name: 'a',
              goal: 'g',
              state: 'running',
              step_count: 0,
              created_at: '',
              updated_at: '',
            })
          ),
      })
      .mockResolvedValueOnce({
        ok: true,
        text: () =>
          Promise.resolve(
            JSON.stringify({
              run_id: 'r1',
              agent_name: 'a',
              goal: 'g',
              state: 'completed',
              step_count: 2,
              created_at: '',
              updated_at: '',
            })
          ),
      });

    const client = new AgentRuntime();
    const run = await client.waitForRun('r1', 10);

    expect(run.state).toBe('completed');
  });
});
