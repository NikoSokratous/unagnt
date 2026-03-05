# PostgreSQL Query Tool

Execute SQL queries and manage PostgreSQL databases from Unagnt.

## Features

- Execute SELECT queries
- Run INSERT/UPDATE/DELETE operations
- Transaction support
- Query result formatting
- Connection pooling
- Parameter binding (SQL injection protection)

## Installation

```bash
unagnt plugin install postgres-query
```

## Configuration

```yaml
tools:
  - name: postgres
    type: postgres-query
    config:
      host: localhost
      port: 5432
      database: mydb
      user: ${POSTGRES_USER}
      password: ${POSTGRES_PASSWORD}
      ssl_mode: require
      max_connections: 10
```

## Usage

```go
result, err := runtime.ExecuteTool(ctx, "postgres", ToolInput{
    Action: "query",
    Parameters: map[string]interface{}{
        "sql": "SELECT * FROM users WHERE email = $1",
        "params": []interface{}{"user@example.com"},
    },
})
```

## Permissions Required

- `network` - Database connections
- `database` - Database access

## API Methods

### `query`
Execute a SELECT query and return results.

**Parameters:**
- `sql` (string, required): SQL SELECT statement
- `params` ([]interface{}, optional): Query parameters

**Returns:**
- `rows` ([]map[string]interface{}): Query results
- `count` (int): Number of rows returned

### `execute`
Execute INSERT, UPDATE, or DELETE statements.

**Parameters:**
- `sql` (string, required): SQL statement
- `params` ([]interface{}, optional): Query parameters

**Returns:**
- `affected_rows` (int): Number of affected rows

### `transaction`
Execute multiple statements in a transaction.

**Parameters:**
- `statements` ([]object, required): Array of SQL statements
  - `sql` (string): SQL statement
  - `params` ([]interface{}): Parameters

**Returns:**
- `results` ([]object): Results for each statement

### `schema`
Get database schema information.

**Parameters:**
- `table` (string, optional): Specific table name

**Returns:**
- `tables` ([]object): Table information
- `columns` ([]object): Column definitions

## Example Workflow

```yaml
name: data-analyzer
steps:
  - name: extract-data
    agent: data-extractor
    goal: "Extract data from PostgreSQL"
    tools:
      - name: postgres
        type: postgres-query
  
  - name: analyze
    agent: analyzer
    goal: "Analyze extracted data"
    depends_on:
      - extract-data
```

## Security

- Always use parameterized queries to prevent SQL injection
- Store credentials in environment variables
- Use SSL connections in production
- Limit permissions to minimum required
- Enable connection pooling for performance

## Example Queries

### Simple SELECT
```yaml
action: query
parameters:
  sql: "SELECT id, name, email FROM users"
```

### Parameterized Query
```yaml
action: query
parameters:
  sql: "SELECT * FROM orders WHERE user_id = $1 AND status = $2"
  params: [123, "completed"]
```

### INSERT with RETURNING
```yaml
action: execute
parameters:
  sql: "INSERT INTO logs (message, level) VALUES ($1, $2) RETURNING id"
  params: ["System started", "INFO"]
```

### Transaction Example
```yaml
action: transaction
parameters:
  statements:
    - sql: "UPDATE accounts SET balance = balance - $1 WHERE id = $2"
      params: [100, 1]
    - sql: "UPDATE accounts SET balance = balance + $1 WHERE id = $2"
      params: [100, 2]
```

## License

MIT
