# Discord Package

Unified Discord webhook client with distributed tracing support.

## Features

- **Trace Integration**: Automatic trace_id injection in Discord messages
- **Multiple Message Types**: Info, Success, Warning, Error with color coding
- **Rich Embeds**: Support for fields, thumbnails, images, and formatting
- **Retry Logic**: Configurable retry with exponential backoff
- **Validation**: Message length and embed validation
- **Logging**: Comprehensive logging with trace context

## Usage

```go
import "github.com/smap-hcmut/shared-libs/go/discord"

// Create Discord client
client, err := discord.New(logger, "https://discord.com/api/webhooks/ID/TOKEN")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Send simple message
err = client.SendMessage(ctx, "Hello from SMAP!")

// Send error with trace context
err = client.SendError(ctx, "Database Error", "Connection failed", err)

// Send notification with fields
fields := map[string]string{
    "Service": "identity-srv",
    "Status":  "healthy",
}
err = client.SendNotification(ctx, "Health Check", "Service status update", fields)
```

## Trace Integration

All Discord messages automatically include trace_id when available in context:
- Error messages include trace_id in embed fields
- Activity logs include trace_id for debugging
- HTTP requests include X-Trace-Id header
- All operations are logged with trace context