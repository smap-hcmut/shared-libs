package constants

// RabbitMQ queue names shared between ingest-srv (producer) and scapper-srv (consumer).
//
// Flow:
//   ingest-srv  --(task queues)--> scapper-srv   (crawl tasks)
//   scapper-srv --(completion queues)--> ingest-srv  (results)

// Task queues: ingest-srv publishes crawl tasks, scapper-srv consumes.
const (
	QueueTikTokTasks   = "tiktok_tasks"
	QueueFacebookTasks = "facebook_tasks"
	QueueYouTubeTasks  = "youtube_tasks"
)

// Completion queues: scapper-srv publishes results, ingest-srv consumes.
const (
	QueueIngestTaskCompletions   = "ingest_task_completions"
	QueueIngestDryrunCompletions = "ingest_dryrun_completions"
)

// Exchange names used for task dispatch routing.
const (
	ExchangeTikTokTasks   = "ingest_tiktok_tasks_exc"
	ExchangeFacebookTasks = "ingest_facebook_tasks_exc"
	ExchangeYouTubeTasks  = "ingest_youtube_tasks_exc"
)

// QueueTasksForPlatform returns the task queue name for a given platform.
func QueueTasksForPlatform(p Platform) string {
	switch p {
	case PlatformTikTok:
		return QueueTikTokTasks
	case PlatformFacebook:
		return QueueFacebookTasks
	case PlatformYouTube:
		return QueueYouTubeTasks
	default:
		return ""
	}
}

// ExchangeForPlatform returns the exchange name for a given platform.
func ExchangeForPlatform(p Platform) string {
	switch p {
	case PlatformTikTok:
		return ExchangeTikTokTasks
	case PlatformFacebook:
		return ExchangeFacebookTasks
	case PlatformYouTube:
		return ExchangeYouTubeTasks
	default:
		return ""
	}
}
