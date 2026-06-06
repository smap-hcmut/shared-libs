// Package uapurl extracts the public-facing post/comment URL from a UAP
// metadata blob. Multiple producers (ingest-srv, analysis-srv, knowledge-srv,
// notification-srv) used to maintain their own ordered list of candidate
// keys; centralising the list keeps platform additions in one place.
package uapurl

import "strings"

// CandidateKeys is the ordered set of metadata keys we probe for a URL. The
// first non-empty value wins.
var CandidateKeys = []string{
	"post_url",
	"url",
	"permalink_url",
	"original_url",
	"source_url",
	"web_url",
	"comment_url",
	"share_url",
	"parent_post_url",
}

// FirstFromMap walks CandidateKeys against the given metadata map and returns
// the first non-empty string value, trimmed.
func FirstFromMap(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return ""
	}
	for _, key := range CandidateKeys {
		raw, ok := metadata[key]
		if !ok {
			continue
		}
		s, ok := raw.(string)
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
