# GitHub Integration Tool

A shareable tool for GitHub API operations in Unagnt.

## Features

- List repositories
- Create issues
- Get pull requests
- Merge PRs
- Comment on issues/PRs
- Manage labels

## Installation

```bash
unagnt plugin install github-integration
```

## Configuration

```yaml
tools:
  - name: github
    type: github-integration
    config:
      token: ${GITHUB_TOKEN}
      owner: your-org
      repo: your-repo
```

## Usage

```go
result, err := runtime.ExecuteTool(ctx, "github", ToolInput{
    Action: "create_issue",
    Parameters: map[string]interface{}{
        "title": "Bug report",
        "body": "Description of the bug",
        "labels": []string{"bug", "high-priority"},
    },
})
```

## Permissions Required

- `network` - API calls to github.com
- `api` - HTTP client access

## API Methods

### `list_repos`
List repositories for the authenticated user or organization.

### `create_issue`
Create a new issue in a repository.

**Parameters:**
- `title` (string, required)
- `body` (string, optional)
- `labels` ([]string, optional)
- `assignees` ([]string, optional)

### `get_pull_requests`
List pull requests for a repository.

**Parameters:**
- `state` (string, optional): "open", "closed", "all"
- `sort` (string, optional): "created", "updated"

### `merge_pr`
Merge a pull request.

**Parameters:**
- `number` (int, required): PR number
- `method` (string, optional): "merge", "squash", "rebase"

### `add_comment`
Add a comment to an issue or PR.

**Parameters:**
- `number` (int, required): Issue/PR number
- `body` (string, required): Comment text

## Example Workflow

```yaml
name: auto-issue-creator
steps:
  - name: create-issue
    agent: issue-creator
    goal: "Create a GitHub issue for the bug report"
    tools:
      - name: github
        type: github-integration
```

## License

MIT
