// Package contracts provides shared data types for inter-service communication.
//
// Types in this package are used by both the producing and consuming side of
// an internal API or Kafka message contract.  Each service may have additional
// local fields (e.g. project-srv adds ProjectStatus); only the common subset
// lives here.
package contracts

// ActivationReadinessCommand identifies which lifecycle transition is being
// evaluated for readiness.
type ActivationReadinessCommand string

const (
	ActivationReadinessCommandActivate ActivationReadinessCommand = "activate"
	ActivationReadinessCommandResume   ActivationReadinessCommand = "resume"
)

// ActivationReadinessCode is a machine-readable error code for a readiness
// blocker.
type ActivationReadinessCode string

const (
	ReadinessCodeDatasourceRequired   ActivationReadinessCode = "DATASOURCE_REQUIRED"
	ReadinessCodePassiveUnconfirmed   ActivationReadinessCode = "PASSIVE_UNCONFIRMED"
	ReadinessCodeTargetDryrunMissing  ActivationReadinessCode = "TARGET_DRYRUN_MISSING"
	ReadinessCodeTargetDryrunFailed   ActivationReadinessCode = "TARGET_DRYRUN_FAILED"
	ReadinessCodeActiveTargetRequired ActivationReadinessCode = "ACTIVE_TARGET_REQUIRED"
	ReadinessCodeDatasourceStatus     ActivationReadinessCode = "DATASOURCE_STATUS_INVALID"
)

// ActivationReadinessError describes one blocker preventing a project
// lifecycle transition.
type ActivationReadinessError struct {
	Code         ActivationReadinessCode `json:"code"`
	Message      string                  `json:"message"`
	DataSourceID string                  `json:"datasource_id,omitempty"`
	TargetID     string                  `json:"target_id,omitempty"`
}

// ActivationReadinessInput is the request to evaluate readiness.
type ActivationReadinessInput struct {
	ProjectID string
	Command   ActivationReadinessCommand
}

// ActivationReadiness is the wire-format payload returned by ingest-srv's
// internal readiness endpoint.  project-srv consumes this via HTTP client.
//
// Services may embed or extend this struct with additional local fields
// (e.g. ProjectStatus).
type ActivationReadiness struct {
	ProjectID                string                     `json:"project_id"`
	DataSourceCount          int                        `json:"datasource_count"`
	HasDatasource            bool                       `json:"has_datasource"`
	PassiveUnconfirmedCount  int                        `json:"passive_unconfirmed_count"`
	MissingTargetDryrunCount int                        `json:"missing_target_dryrun_count"`
	FailedTargetDryrunCount  int                        `json:"failed_target_dryrun_count"`
	CanProceed               bool                       `json:"can_proceed"`
	Errors                   []ActivationReadinessError `json:"errors"`
}

// Readiness message constants for error descriptions.
const (
	ReadinessMsgDatasourceRequired   = "project must have at least one datasource"
	ReadinessMsgPassiveUnconfirmed   = "passive datasource is not confirmed"
	ReadinessMsgTargetDryrunMissing  = "crawl target has never been dry-run"
	ReadinessMsgTargetDryrunFailed   = "crawl target latest dry-run is FAILED"
	ReadinessMsgActiveTargetRequired = "crawl datasource must have at least one active target"
	ReadinessMsgDatasourceStatus     = "datasource status is not eligible for project lifecycle command"
)
