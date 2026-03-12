package minio

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// newAsyncUploadManager creates a new async upload manager
func newAsyncUploadManager(impl *implMinIO, workerPoolSize, queueSize int) *asyncUploadManager {
	if workerPoolSize <= 0 {
		workerPoolSize = DefaultAsyncWorkers
	}
	if queueSize <= 0 {
		queueSize = DefaultAsyncQueueSize
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &asyncUploadManager{
		minio:         impl,
		workerPool:    workerPoolSize,
		uploadQueue:   make(chan *AsyncUploadTask, queueSize),
		statusTracker: newUploadStatusTracker(),
		ctx:           ctx,
		cancel:        cancel,
		started:       false,
	}
}

func (m *asyncUploadManager) cancelUpload(taskID string) error {
	progress, exists := m.statusTracker.getStatus(taskID)
	if !exists {
		return ErrConnectionClosed
	}
	if progress.Status != UploadStatusPending && progress.Status != UploadStatusUploading {
		return ErrConnectionClosed
	}
	m.statusTracker.updateStatus(taskID, &UploadProgress{
		TaskID: taskID, Status: UploadStatusCancelled, UpdatedAt: time.Now(),
	})
	return nil
}

func (m *asyncUploadManager) start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return
	}
	for i := 0; i < m.workerPool; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
	m.wg.Add(1)
	go m.cleanupWorker()
	m.started = true
}

func (m *asyncUploadManager) stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.started {
		return
	}
	m.cancel()
	close(m.uploadQueue)
	m.wg.Wait()
	m.started = false
}

func (m *asyncUploadManager) uploadAsync(ctx context.Context, req *UploadRequest) (string, error) {
	m.mu.RLock()
	if !m.started {
		m.mu.RUnlock()
		return "", ErrConnectionClosed
	}
	m.mu.RUnlock()

	taskID := uuid.New().String()
	taskCtx, taskCancel := context.WithCancel(ctx)
	task := &AsyncUploadTask{
		ID:           taskID,
		Request:      req,
		ResultChan:   make(chan *AsyncUploadResult, 1),
		ProgressChan: make(chan *UploadProgress, 10),
		CreatedAt:    time.Now(),
		ctx:          taskCtx,
		cancel:       taskCancel,
	}
	m.statusTracker.updateStatus(taskID, &UploadProgress{
		TaskID: taskID, TotalBytes: req.Size, Status: UploadStatusPending, UpdatedAt: time.Now(),
	})
	select {
	case m.uploadQueue <- task:
		return taskID, nil
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return "", ErrConnectionClosed
	}
}
func (m *asyncUploadManager) getUploadStatus(taskID string) (*UploadProgress, error) {
	progress, exists := m.statusTracker.getStatus(taskID)
	if !exists {
		return nil, ErrConnectionClosed
	}
	return progress, nil
}

func (m *asyncUploadManager) waitForUpload(taskID string, timeout time.Duration) (*AsyncUploadResult, error) {
	progress, exists := m.statusTracker.getStatus(taskID)
	if !exists {
		return nil, ErrConnectionClosed
	}
	if progress.Status == UploadStatusCompleted || progress.Status == UploadStatusFailed {
		result := m.statusTracker.getResult(taskID)
		if result != nil {
			return result, nil
		}
		return nil, ErrConnectionClosed
	}
	ticker := time.NewTicker(WaitForUploadPollInterval)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()
	for {
		select {
		case <-timeoutTimer.C:
			return nil, ErrConnectionTimeout
		case <-ticker.C:
			progress, exists = m.statusTracker.getStatus(taskID)
			if !exists {
				return nil, ErrConnectionClosed
			}
			if progress.Status == UploadStatusCompleted || progress.Status == UploadStatusFailed {
				result := m.statusTracker.getResult(taskID)
				if result != nil {
					return result, nil
				}
				return nil, ErrConnectionClosed
			}
		}
	}
}
func (m *asyncUploadManager) worker(workerID int) {
	defer m.wg.Done()
	for {
		select {
		case <-m.ctx.Done():
			return
		case task, ok := <-m.uploadQueue:
			if !ok {
				return
			}
			m.processUploadTask(workerID, task)
		}
	}
}

func (m *asyncUploadManager) processUploadTask(workerID int, task *AsyncUploadTask) {
	startTime := time.Now()
	progress, _ := m.statusTracker.getStatus(task.ID)
	if progress != nil && progress.Status == UploadStatusCancelled {
		return
	}
	m.statusTracker.updateStatus(task.ID, &UploadProgress{
		TaskID: task.ID, TotalBytes: task.Request.Size, Status: UploadStatusUploading, UpdatedAt: time.Now(),
	})

	fileInfo, err := m.minio.UploadFile(task.ctx, task.Request)
	duration := time.Since(startTime)
	endTime := time.Now()
	result := &AsyncUploadResult{
		TaskID: task.ID, FileInfo: fileInfo, Error: err, Duration: duration, StartTime: startTime, EndTime: endTime,
	}
	if err != nil {
		m.statusTracker.updateStatus(task.ID, &UploadProgress{
			TaskID: task.ID, Status: UploadStatusFailed, Error: err.Error(), UpdatedAt: time.Now(),
		})
	} else {
		m.statusTracker.updateStatus(task.ID, &UploadProgress{
			TaskID: task.ID, BytesUploaded: task.Request.Size, TotalBytes: task.Request.Size,
			Percentage: 100, Status: UploadStatusCompleted, UpdatedAt: time.Now(),
		})
	}
	m.statusTracker.storeResult(task.ID, result)
	select {
	case task.ResultChan <- result:
	default:
	}
}

func (m *asyncUploadManager) cleanupWorker() {
	defer m.wg.Done()
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.statusTracker.cleanupOldStatuses(CleanupMaxAge)
		}
	}
}
