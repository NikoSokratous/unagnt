# Data Pipeline Agent Example

Automated ETL (Extract, Transform, Load) agent with environment-aware policies and quality checks.

## Features

- Database query execution (read-only in dev, approval for prod writes)
- Data transformation and validation
- Anomaly detection
- Quality checks
- Alerting on issues
- Environment-aware policies

## Setup

```bash
cd examples/data-pipeline

# Set API key
export OPENAI_API_KEY=sk-...

# Set environment
export AGENT_ENVIRONMENT=development  # or production

# Run pipeline
../../bin/unagnt run \
  --config agent.yaml \
  --goal "Extract user signups from last 24h, check for anomalies" \
  --store pipeline.db
```

## Example Pipelines

### Daily Metrics

```bash
unagnt run --config agent.yaml \
  --goal "Calculate daily active users, compare to last week, alert if drop > 10%"
```

### Data Quality Check

```bash
unagnt run --config agent.yaml \
  --goal "Check users table for null emails, duplicate accounts, invalid dates"
```

### ETL Job

```bash
unagnt run --config agent.yaml \
  --goal "Extract orders from API, transform to analytics format, load to warehouse"
```

## Policy Enforcement

### Development Environment

- All reads: **Auto-approved**
- All writes: **Auto-approved** (safe in dev)

### Production Environment

- Reads: **Auto-approved**
- Writes: **Require approval**
- Destructive ops (DROP, TRUNCATE): **Blocked**

### Quality Gates

- Large queries (>100k rows): **Require approval**
- Anomaly detected: **Pause for review**
- Data quality issues: **Alert operators**

## Architecture

```
Scheduled Trigger / API Call
    ↓
Agent Plans Extraction
    ↓
Execute Query (db_read tool)
    ↓
Transform Data (calc tool)
    ↓
Quality Checks
    ↓
Policy Check (prod write approval)
    ↓
Load to Destination (if approved)
    ↓
Alert on Completion
```

## Monitoring

### Metrics

Track pipeline health:

```bash
curl http://localhost:8080/metrics | grep pipeline
```

Key metrics:
- `Unagnt_tool_executions_total{tool="db_read"}`
- `Unagnt_policy_denials_total{rule="prod-db-write-approval"}`
- `Unagnt_run_duration_seconds`

### Alerts

Configure alerts for:
- Policy denials
- Failed runs
- Anomaly detection
- Long-running pipelines

## Scheduling

### Cron-Based Execution

```yaml
# Schedule in unagntd (future feature)
schedule:
  - cron: "0 */6 * * *"  # Every 6 hours
    agent: data-pipeline
    goal: "Run daily ETL"
```

### Manual Trigger

```bash
# Via API
curl -X POST http://localhost:8080/v1/runs \
  -H "Authorization: Bearer key" \
  -d '{"agent_name":"data-pipeline","goal":"Run ETL"}'
```

## Data Transformations

Example transformations the agent can perform:

```python
# Aggregation
"Calculate sum, avg, count of sales by region"

# Cleaning
"Remove duplicates, fill null values, standardize dates"

# Enrichment
"Join user data with geo data, add timezone info"

# Validation
"Check all emails are valid, all dates are in range"
```

## Observability

### Replay Failed Runs

```bash
# Replay to debug
unagnt replay --run-id <failed-run-id> --store pipeline.db

# Compare successful vs failed
unagnt diff <success-run> <failed-run> --store pipeline.db
```

### Audit Trail

All queries logged for compliance:

```bash
unagnt logs --log-file agent.log --filter "tool=db_write"
```

## Security

### Best Practices

1. **Read-Only Credentials**: Use read-only DB user in dev
2. **Approval Gates**: Always require approval for prod writes
3. **Query Validation**: Block DROP, TRUNCATE, DELETE without WHERE
4. **Data Masking**: Redact PII in logs
5. **Rate Limiting**: Limit queries per run

### Example: Read-Only User

```sql
-- PostgreSQL
CREATE USER pipeline_agent WITH PASSWORD 'secure-password';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO pipeline_agent;
```

## Cost Optimization

- Use GPT-4o-mini for simple transformations
- Cache intermediate results in memory
- Batch operations where possible
- Set aggressive timeouts

## Real-World Use Cases

1. **Nightly ETL**: Extract from production DB, transform, load to warehouse
2. **Data Quality Monitoring**: Hourly checks for anomalies
3. **Reporting**: Generate weekly business intelligence reports
4. **Compliance**: Audit logs for regulatory requirements

## Extending

Add tools for:
- SQL query execution
- CSV/JSON file operations
- Cloud storage (S3, GCS)
- Data validation libraries
- Alert services (PagerDuty, Slack)

## License

MIT
