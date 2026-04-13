// Package constants provides shared constant definitions used across SMAP microservices.
//
// Platform, Kafka topic, and RabbitMQ queue names are defined here as the
// single source of truth so that producer and consumer services stay in sync.
package constants

// Platform represents a social media platform supported by the SMAP system.
// Canonical form is lowercase.  Services that store platforms in UPPER case
// (e.g. ingest-srv SQL enum) should map at their application layer.
type Platform string

const (
	PlatformTikTok    Platform = "tiktok"
	PlatformFacebook  Platform = "facebook"
	PlatformYouTube   Platform = "youtube"
	PlatformInstagram Platform = "instagram"
)

// AllPlatforms returns every crawl-supported platform.
func AllPlatforms() []Platform {
	return []Platform{PlatformTikTok, PlatformFacebook, PlatformYouTube, PlatformInstagram}
}

// CrawlPlatforms returns platforms that currently have a scapper-srv handler.
// Instagram is excluded because no crawler implementation exists yet.
func CrawlPlatforms() []Platform {
	return []Platform{PlatformTikTok, PlatformFacebook, PlatformYouTube}
}
