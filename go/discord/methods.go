package discord

import (
	"context"
	"fmt"
	"time"
)

// SendMessage sends a simple text message with trace context.
func (d *discordImpl) SendMessage(ctx context.Context, content string) error {
	if err := d.validateMessageLength(content); err != nil {
		return err
	}

	d.logOperation(ctx, "SendMessage", fmt.Sprintf("sending message (length: %d)", len(content)))

	payload := &WebhookPayload{
		Content:   content,
		Username:  d.config.DefaultUsername,
		AvatarURL: d.config.DefaultAvatarURL,
	}

	return d.sendWithRetry(ctx, payload)
}

// SendEmbed sends an embed message with options and trace context.
func (d *discordImpl) SendEmbed(ctx context.Context, options MessageOptions) error {
	embed := &Embed{
		Title:       d.truncateString(options.Title, MaxTitleLen),
		Description: d.truncateString(options.Description, MaxDescriptionLen),
		Color:       d.getColorForType(options.Type),
		Fields:      options.Fields,
		Footer:      options.Footer,
		Author:      options.Author,
		Thumbnail:   options.Thumbnail,
		Image:       options.Image,
	}

	if !options.Timestamp.IsZero() {
		embed.Timestamp = d.formatTimestamp(options.Timestamp)
	}

	if err := d.validateEmbedLength(embed); err != nil {
		return err
	}

	d.logOperation(ctx, "SendEmbed", fmt.Sprintf("sending embed (type: %s, title: %s)", options.Type, options.Title))

	payload := &WebhookPayload{
		Embeds:    []Embed{*embed},
		Username:  options.Username,
		AvatarURL: options.AvatarURL,
	}

	if payload.Username == "" {
		payload.Username = d.config.DefaultUsername
	}
	if payload.AvatarURL == "" {
		payload.AvatarURL = d.config.DefaultAvatarURL
	}

	return d.sendWithRetry(ctx, payload)
}

// SendError sends an error message with trace context.
func (d *discordImpl) SendError(ctx context.Context, title, description string, err error) error {
	fields := []EmbedField{}
	if err != nil {
		fields = append(fields, EmbedField{
			Name:   "Error",
			Value:  d.truncateString(err.Error(), MaxFieldValueLen),
			Inline: false,
		})
	}

	// Add trace_id to error fields if available
	if traceID := d.tracer.GetTraceID(ctx); traceID != "" {
		fields = append(fields, EmbedField{
			Name:   "Trace ID",
			Value:  traceID,
			Inline: true,
		})
	}

	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeError,
		Level:       LevelHigh,
		Title:       title,
		Description: description,
		Fields:      fields,
		Timestamp:   time.Now(),
	})
}

// SendSuccess sends a success message with trace context.
func (d *discordImpl) SendSuccess(ctx context.Context, title, description string) error {
	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeSuccess,
		Level:       LevelNormal,
		Title:       title,
		Description: description,
		Timestamp:   time.Now(),
	})
}

// SendWarning sends a warning message with trace context.
func (d *discordImpl) SendWarning(ctx context.Context, title, description string) error {
	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeWarning,
		Level:       LevelNormal,
		Title:       title,
		Description: description,
		Timestamp:   time.Now(),
	})
}

// SendInfo sends an info message with trace context.
func (d *discordImpl) SendInfo(ctx context.Context, title, description string) error {
	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeInfo,
		Level:       LevelNormal,
		Title:       title,
		Description: description,
		Timestamp:   time.Now(),
	})
}

// ReportBug sends a bug report with trace context.
func (d *discordImpl) ReportBug(ctx context.Context, message string) error {
	if len(message) > ReportBugDescLen {
		message = message[:ReportBugDescLen-3] + "..."
	}

	fields := []EmbedField{}

	// Add trace_id to bug report if available
	if traceID := d.tracer.GetTraceID(ctx); traceID != "" {
		fields = append(fields, EmbedField{
			Name:   "Trace ID",
			Value:  traceID,
			Inline: true,
		})
	}

	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeError,
		Level:       LevelUrgent,
		Title:       ReportBugTitle,
		Description: fmt.Sprintf("```%s```", message),
		Fields:      fields,
		Timestamp:   time.Now(),
	})
}

// SendNotification sends a notification with fields and trace context.
func (d *discordImpl) SendNotification(ctx context.Context, title, description string, fields map[string]string) error {
	var embedFields []EmbedField

	for name, value := range fields {
		embedFields = append(embedFields, EmbedField{
			Name:   d.truncateString(name, MaxTitleLen),
			Value:  d.truncateString(value, MaxFieldValueLen),
			Inline: true,
		})
	}

	// Add trace_id to notification if available
	if traceID := d.tracer.GetTraceID(ctx); traceID != "" {
		embedFields = append(embedFields, EmbedField{
			Name:   "Trace ID",
			Value:  traceID,
			Inline: true,
		})
	}

	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeInfo,
		Level:       LevelNormal,
		Title:       title,
		Description: description,
		Fields:      embedFields,
		Timestamp:   time.Now(),
	})
}

// SendActivityLog sends an activity log with trace context.
func (d *discordImpl) SendActivityLog(ctx context.Context, action, user, details string) error {
	fields := []EmbedField{
		{Name: "Action", Value: action, Inline: true},
		{Name: "User", Value: user, Inline: true},
	}

	if details != "" {
		fields = append(fields, EmbedField{
			Name:   "Details",
			Value:  details,
			Inline: false,
		})
	}

	// Add trace_id to activity log if available
	if traceID := d.tracer.GetTraceID(ctx); traceID != "" {
		fields = append(fields, EmbedField{
			Name:   "Trace ID",
			Value:  traceID,
			Inline: true,
		})
	}

	return d.SendEmbed(ctx, MessageOptions{
		Type:        MessageTypeInfo,
		Level:       LevelLow,
		Title:       ActivityLogTitle,
		Description: fmt.Sprintf("**%s** performed **%s**", user, action),
		Fields:      fields,
		Timestamp:   time.Now(),
	})
}
