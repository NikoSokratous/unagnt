# DB Agent Safety Showcase

Agent that can query SQLite, with policy that **blocks destructive ops** and **restricts sensitive data**. Demonstrates governance for database access.

## What You’ll See

| Scenario | Policy behavior |
|----------|-----------------|
| `SELECT * FROM products` | ✅ Allowed – read from public tables |
| `DROP TABLE products` | ❌ Blocked – destructive SQL denied |
| `SELECT * FROM users` | ❌ Blocked – restricted table |
| `UPDATE products SET price=0` | ⏸️ Requires approval – mutations need human sign-off |

## Setup

```bash
# From project root: create demo DB
cd showcase/db-agent-safety
go run ../../scripts/init-demo-db/

# Run agent (from showcase/db-agent-safety so demo.db is found)
unagnt run -c agent.yaml -g "List all products and their prices"
```

## Demo Script

```bash
# 1. Allowed: read from products
unagnt run -c agent.yaml -g "What products do we have? Show me id, name, and price"

# 2. Blocked: destructive
unagnt run -c agent.yaml -g "Drop the products table"

# 3. Blocked: restricted data
unagnt run -c agent.yaml -g "Show me all users and their passwords"

# 4. Requires approval: mutation (use --approval-webhook or stdin prompt)
unagnt run -c agent.yaml -g "Update Widget A price to 15.99" --approval-webhook http://localhost:9090/request
```

## Policy Rules

- **block-destructive**: Deny DROP, TRUNCATE, DELETE, ALTER, CREATE TABLE
- **block-restricted-data**: Deny access to `users`, `password`, `credentials`
- **require-approval-mutations**: Require approval for INSERT, UPDATE

## Files

| File | Purpose |
|------|---------|
| `agent.yaml` | Agent with sql_query + echo |
| `policy.yaml` | CEL rules for DB safety |
| `demo.db` | SQLite DB (created by init-demo-db) |
