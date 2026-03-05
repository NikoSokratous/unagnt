# Slack Notifier Tool

Send notifications to Slack channels from Unagnt workflows.

## Features

- Send messages to channels
- Post to threads
- Upload files
- Send direct messages
- Rich message formatting (blocks)
- Interactive messages

## Installation

```bash
unagnt plugin install slack-notifier
```

## Configuration

```yaml
tools:
  - name: slack
    type: slack-notifier
    config:
      webhook_url: ${SLACK_WEBHOOK_URL}
      token: ${SLACK_BOT_TOKEN}
      default_channel: "#general"
```

## Usage

```go
result, err := runtime.ExecuteTool(ctx, "slack", ToolInput{
    Action: "send_message",
    Parameters: map[string]interface{}{
        "channel": "#alerts",
        "text": "Deployment completed successfully!",
        "blocks": []map[string]interface{}{
            {
                "type": "section",
                "text": map[string]string{
                    "type": "mrkdwn",
                    "text": "*Status:* :white_check_mark: Success",
                },
            },
        },
    },
})
```

## Permissions Required

- `network` - API calls to slack.com
- `api` - HTTP client access

## API Methods

### `send_message`
Send a message to a Slack channel.

**Parameters:**
- `channel` (string, required): Channel ID or name
- `text` (string, required): Message text
- `blocks` ([]object, optional): Rich message blocks
- `thread_ts` (string, optional): Thread timestamp to reply to

### `upload_file`
Upload a file to Slack.

**Parameters:**
- `channels` ([]string, required): Target channels
- `filename` (string, required): File name
- `content` (string, required): File content
- `title` (string, optional): File title

### `send_dm`
Send a direct message to a user.

**Parameters:**
- `user_id` (string, required): User ID
- `text` (string, required): Message text

### `update_message`
Update an existing message.

**Parameters:**
- `channel` (string, required): Channel ID
- `ts` (string, required): Message timestamp
- `text` (string, required): New message text

## Example Workflow

```yaml
name: build-notifier
steps:
  - name: build
    agent: builder
    goal: "Build the application"
    
  - name: notify-success
    agent: notifier
    goal: "Notify team of successful build"
    condition: "outputs.build.status == 'success'"
    tools:
      - name: slack
        type: slack-notifier
```

## Message Formatting

Slack supports markdown-like formatting:
- `*bold*` - **bold** text
- `_italic_` - *italic* text
- `~strike~` - ~~strikethrough~~ text
- `` `code` `` - inline code
- ` ```code block``` ` - code block
- `<url|text>` - hyperlink

## License

MIT
