package minio

import (
	"time"
)

// newUploadStatusTracker creates a new upload status tracker
func newUploadStatusTracker() *uploadStatusTracker {
	return &uploadStatusTracker{
		statuses: make(map[string]*UploadProgress),
		results:  make(map[string]*AsyncUploadResult),
	}
}

func (t *uploadStatusTracker) updateStatus(taskID string, progress *UploadProgress) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, exists := t.statuses[taskID]; exists {
		if progress.BytesUploaded > 0 {
			existing.BytesUploaded = progress.BytesUploaded
		}
		if progress.TotalBytes > 0 {
			existing.TotalBytes = progress.TotalBytes
		}
		if progress.Percentage > 0 {
			existing.Percentage = progress.Percentage
		}
		if progress.Status != "" {
			existing.Status = progress.Status
		}
		if progress.Error != "" {
			existing.Error = progress.Error
		}
		existing.UpdatedAt = progress.UpdatedAt
	} else {
		t.statuses[taskID] = progress
	}
}

func (t *uploadStatusTracker) getStatus(taskID string) (*UploadProgress, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	progress, exists := t.statuses[taskID]
	if !exists {
		return nil, false
	}
	progressCopy := *progress
	return &progressCopy, true
}

func (t *uploadStatusTracker) storeResult(taskID string, result *AsyncUploadResult) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.results[taskID] = result
}

func (t *uploadStatusTracker) getResult(taskID string) *AsyncUploadResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.results[taskID]
}

func (t *uploadStatusTracker) cleanupOldStatuses(maxAge time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for taskID, progress := range t.statuses {
		if progress.Status == UploadStatusCompleted || progress.Status == UploadStatusFailed || progress.Status == UploadStatusCancelled {
			if now.Sub(progress.UpdatedAt) > maxAge {
				delete(t.statuses, taskID)
				delete(t.results, taskID)
			}
		}
	}
}
