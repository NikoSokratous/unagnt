# Incident Runbooks

Operational runbooks for common Agent Runtime incidents. Use these for queue saturation, dead-letter spikes, and replay control.

---

## 1. Queue Saturation

**Symptoms:**
- High `agentruntime_run_queue_depth` (approaching or at capacity)
- Increasing `agentruntime_run_queue_rejected_total`
- API calls to `POST /v1/runs` returning errors or slow responses

**Root causes:**
- Submissions exceeding processing capacity
- Too few workers for current load
- Stuck runs blocking workers

**Response steps:**

1. **Verify queue depth and rejection rate**
   ```bash
   curl -s http://localhost:8080/metrics | grep agentruntime_run_queue
   ```

2. **Scale workers**  
   Increase the runner worker count. If running unagntd, workers are configured at startup (default 2). Consider running multiple replicas behind a load balancer or increasing workers in config.

3. **Increase queue size**  
   For memory backend, increase `--queue-size` (default 256). Restart required.

4. **Switch to durable queue (Redis)**  
   For production, use Redis as queue backend to handle bursts and enable horizontal scaling. Set `QUEUE_BACKEND=redis` and `QUEUE_REDIS_URL`.

5. **Check for stuck runs**  
   Inspect active runs: `GET /v1/runs` and consider canceling stuck runs: `POST /v1/runs/{id}/cancel`.

**Prevention:**
- Monitor `agentruntime_run_queue_depth` and `agentruntime_run_queue_rejected_total`
- Set alerts for queue depth > 80% of capacity or sustained rejections
- Use Redis backend for production workloads

---

## 2. Dead-Letter Spikes

**Symptoms:**
- Rapid increase in `agentruntime_run_dead_letters_total`
- Many entries in `GET /v1/runs/dead-letters`

**Root causes:**
- Upstream agent/model failures (e.g. API errors, timeouts)
- Policy or guardrail rejections
- Misconfiguration (wrong agent name, invalid goal)

**Response steps:**

1. **Inspect recent dead letters**
   ```bash
   curl -s -H "Authorization: Bearer $KEY" http://localhost:8080/v1/runs/dead-letters?limit=20
   ```

2. **Identify source and error**  
   Group by `source` (e.g. webhook, schedule, api) and `error`. Common patterns: timeout, execution_error, rate_limit.

3. **Fix root cause**  
   - Timeout: increase `timeout_ms` or fix slow tool/model
   - API errors: check provider status, API keys, rate limits
   - Policy/guardrail: adjust rules or goal/output constraints

4. **Consider replay**  
   For transient failures, replay specific runs: `POST /v1/runs/dead-letters/{run_id}/replay`. Optionally override goal, max_retries, or timeout.

5. **Enable retention and archival**  
   If dead letters accumulate, enable retention pruning and optional archival to prevent unbounded growth. See [dead-letter-retention.md](dead-letter-retention.md).

**Prevention:**
- Alert on `agentruntime_run_dead_letters_total` rate increase
- Configure retention window and pruning
- Use structured logging to correlate dead letters with upstream events

---

## 3. Replay Control

**When to replay vs purge:**
- **Replay**: Transient failures (network, rate limit, temporary model unavailability) where retrying is safe.
- **Purge / archive**: Non-retryable failures (bad config, policy violation) or when the goal is obsolete.

**Replay workflow:**
- Single run: `POST /v1/runs/dead-letters/{run_id}/replay`
- Body can override `goal`, `max_retries`, `retry_backoff_ms`, `timeout_ms`, `payload`

**Bulk replay considerations:**
- No bulk replay API in v3. Script a loop over `GET /v1/runs/dead-letters` and replay by run_id.
- Rate limit replays to avoid overloading the queue. Stagger requests if replaying many runs.
- Filter by source/agent before bulk replay to avoid replaying irrelevant failures.

**Rate limits:**
- Replays are queued like new runs. Monitor queue depth when replaying many runs.
- If using rate limiting, replays count toward the same limits.
