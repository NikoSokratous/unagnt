/** Run represents an agent run. */
export interface Run {
  run_id: string;
  agent_name: string;
  goal: string;
  state: string;
  step_count: number;
  created_at: string;
  updated_at: string;
}

/** CreateRunRequest is the request for creating a run. */
export interface CreateRunRequest {
  agent_name: string;
  goal: string;
  max_retries?: number;
  retry_backoff_ms?: number;
  timeout_ms?: number;
}

/** CreateRunResponse is the response from creating a run. */
export interface CreateRunResponse {
  run_id: string;
}

/** ListRunsResponse is the response from listing runs. */
export interface ListRunsResponse {
  run_ids: string[];
}

/** RunEvent represents an event in a run's history. */
export interface RunEvent {
  run_id?: string;
  step_id?: string;
  timestamp?: string;
  type?: string;
  agent?: string;
  data?: Record<string, unknown>;
  model?: ModelMeta;
}

/** ModelMeta captures model metadata. */
export interface ModelMeta {
  provider?: string;
  name?: string;
  version?: string;
}

/** RunEventsResponse is the response from GET /v1/runs/{id}/events. */
export interface RunEventsResponse {
  run_id: string;
  events: RunEvent[];
}

/** StreamChunk represents a Server-Sent Event chunk from run streaming. */
export interface StreamChunk {
  run_id?: string;
  step_id?: string;
  timestamp?: string;
  type?: string;
  agent?: string;
  data?: Record<string, unknown>;
  model?: ModelMeta;
}
