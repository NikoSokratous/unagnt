# MCP (Model Context Protocol) Support

Unagnt can act as an **MCP client**, connecting to MCP-compliant servers and exposing their tools to agents.

## Configuration

Add `mcp_sources` to your agent config:

```yaml
# agent.yaml
name: my-agent
model:
  provider: openai
  name: gpt-4o-mini

mcp_sources:
  - type: stdio
    command: npx
    args:
      - -y
      - @modelcontextprotocol/server-filesystem
      - /tmp
    tool_prefix: mcp_fs_

  # - type: http
  #   url: https://mcp.example.com
  #   tool_prefix: mcp_remote_
```

## Transport Types

- **stdio**: Spawn an MCP server as a subprocess. Common for local tools (filesystem, Git, etc.).
- **http**: Connect to a remote MCP server over Streamable HTTP.

## Tool Prefix

Use `tool_prefix` to avoid name clashes when using multiple MCP sources. Tools will be registered as `{prefix}{name}` (e.g. `mcp_fs_list_files`).

## Example: Filesystem MCP Server

```bash
# Ensure Node.js is installed
npx -y @modelcontextprotocol/server-filesystem /tmp
```

Add to agent config and run:

```bash
unagnt run -c agent.yaml -g "List files in /tmp"
```

## References

- [MCP Specification](https://modelcontextprotocol.io/)
- [mcp-go](https://github.com/mark3labs/mcp-go)
