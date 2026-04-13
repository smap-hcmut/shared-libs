package constants

// Kafka topic names used across SMAP microservices.
//
// Naming convention: <domain>.<entity>.<event>
//
// Flow:
//   ingest-srv  --(TopicCollectorOutput)--> analysis-srv
//   analysis-srv --(TopicAnalytics*)--> knowledge-srv
//   project-srv  --(TopicProjectEvents)--> (project-srv consumer)
//   identity-srv --(TopicAuditEvents)-->  (identity-srv consumer)

const (
	// TopicCollectorOutput is produced by ingest-srv after UAP normalisation
	// and consumed by analysis-srv for NLP pipeline processing.
	TopicCollectorOutput = "smap.collector.output"

	// TopicAnalyticsBatchCompleted is produced by analysis-srv after a batch
	// of posts has been analysed (Layer 3) and consumed by knowledge-srv for
	// Qdrant indexing.
	TopicAnalyticsBatchCompleted = "analytics.batch.completed"

	// TopicAnalyticsInsightsPublished is produced by analysis-srv per insight
	// card (Layer 2) and consumed by knowledge-srv.
	TopicAnalyticsInsightsPublished = "analytics.insights.published"

	// TopicAnalyticsReportDigest is produced by analysis-srv as a final
	// summary (Layer 1) and consumed by knowledge-srv to trigger NotebookLM
	// export.
	TopicAnalyticsReportDigest = "analytics.report.digest"

	// TopicProjectEvents is produced by project-srv on lifecycle transitions
	// (activate, pause, resume, archive).
	TopicProjectEvents = "project.events"

	// TopicAuditEvents is produced and consumed by identity-srv for audit
	// logging.
	TopicAuditEvents = "audit.events"
)

// Dead-letter queue topics used by knowledge-srv when processing fails.
const (
	TopicInsightsPublishedDLQ = "analytics.insights.published.dlq"
	TopicReportDigestDLQ      = "analytics.report.digest.dlq"
)
