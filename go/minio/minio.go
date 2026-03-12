package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

// --- implMinIO: connection with trace integration ---

func (m *implMinIO) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add trace logging
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("MinIO connection attempt", "trace_id", traceID)
	}

	_, err := m.minioClient.ListBuckets(ctx)
	if err != nil {
		m.connected = false
		return handleMinIOError(err, "connect")
	}
	m.connected = true
	return nil
}

func (m *implMinIO) ConnectWithRetry(ctx context.Context, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := m.Connect(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			backoff := time.Duration(1<<uint(i)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}
	return fmt.Errorf("failed to connect after %d retries: %w", maxRetries, lastErr)
}

func (m *implMinIO) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.connected {
		return NewConnectionError(fmt.Errorf("not connected"))
	}
	_, err := m.minioClient.ListBuckets(ctx)
	if err != nil {
		return handleMinIOError(err, "health_check")
	}
	return nil
}

func (m *implMinIO) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	if m.asyncUploadMgr != nil {
		m.asyncUploadMgr.stop()
	}
	return nil
}

// --- implMinIO: bucket operations with trace integration ---

func (m *implMinIO) CreateBucket(ctx context.Context, bucketName string) error {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Creating bucket", "trace_id", traceID, "bucket", bucketName)
	}

	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	exists, err := m.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return handleMinIOError(err, "check_bucket_exists")
	}
	if exists {
		return NewInvalidInputError(fmt.Sprintf("bucket already exists: %s", bucketName))
	}
	err = m.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: m.config.Region})
	if err != nil {
		return handleMinIOError(err, "create_bucket")
	}
	return nil
}

func (m *implMinIO) DeleteBucket(ctx context.Context, bucketName string) error {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Deleting bucket", "trace_id", traceID, "bucket", bucketName)
	}

	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	return handleMinIOError(m.minioClient.RemoveBucket(ctx, bucketName), "delete_bucket")
}

func (m *implMinIO) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	if err := validateBucketName(bucketName); err != nil {
		return false, err
	}
	exists, err := m.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return false, handleMinIOError(err, "check_bucket_exists")
	}
	return exists, nil
}

func (m *implMinIO) ListBuckets(ctx context.Context) ([]*BucketInfo, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Listing buckets", "trace_id", traceID)
	}

	buckets, err := m.minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, handleMinIOError(err, "list_buckets")
	}
	var result []*BucketInfo
	for _, bucket := range buckets {
		result = append(result, &BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
			Region:       m.config.Region,
		})
	}
	return result, nil
}

// --- implMinIO: upload operations with trace integration ---

func (m *implMinIO) UploadFile(ctx context.Context, req *UploadRequest) (*FileInfo, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Uploading file", "trace_id", traceID, "bucket", req.BucketName, "object", req.ObjectName)
	}

	if err := validateUploadRequest(req); err != nil {
		return nil, err
	}
	opts := minio.PutObjectOptions{ContentType: req.ContentType}
	if req.Metadata != nil {
		opts.UserMetadata = req.Metadata
	} else {
		opts.UserMetadata = make(map[string]string)
	}
	if req.OriginalName != "" {
		opts.UserMetadata["original-name"] = req.OriginalName
	}

	// Add trace_id to metadata if available
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		opts.UserMetadata["trace-id"] = traceID
	}

	info, err := m.minioClient.PutObject(ctx, req.BucketName, req.ObjectName, req.Reader, req.Size, opts)
	if err != nil {
		return nil, handleMinIOError(err, "upload_file")
	}
	return &FileInfo{
		BucketName:   req.BucketName,
		ObjectName:   req.ObjectName,
		OriginalName: req.OriginalName,
		Size:         info.Size,
		ContentType:  req.ContentType,
		ETag:         info.ETag,
		LastModified: time.Now(),
		Metadata:     req.Metadata,
	}, nil
}

func (m *implMinIO) GetPresignedUploadURL(ctx context.Context, req *PresignedURLRequest) (*PresignedURLResponse, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Generating presigned upload URL", "trace_id", traceID, "bucket", req.BucketName)
	}

	if err := validatePresignedURLRequest(req); err != nil {
		return nil, err
	}
	url, err := m.minioClient.PresignedPutObject(ctx, req.BucketName, req.ObjectName, req.Expiry)
	if err != nil {
		return nil, handleMinIOError(err, "get_presigned_upload_url")
	}
	return &PresignedURLResponse{
		URL:       url.String(),
		ExpiresAt: time.Now().Add(req.Expiry),
		Method:    MethodPUT,
		Headers:   req.Headers,
	}, nil
}

// --- implMinIO: download operations with trace integration ---

func (m *implMinIO) DownloadFile(ctx context.Context, req *DownloadRequest) (io.ReadCloser, *DownloadHeaders, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Downloading file", "trace_id", traceID, "bucket", req.BucketName, "object", req.ObjectName)
	}

	if err := validateDownloadRequest(req); err != nil {
		return nil, nil, err
	}
	objInfo, err := m.minioClient.StatObject(ctx, req.BucketName, req.ObjectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, nil, handleMinIOError(err, "get_file_info")
	}
	opts := minio.GetObjectOptions{}
	if req.Range != nil {
		opts.SetRange(req.Range.Start, req.Range.End)
	}
	object, err := m.minioClient.GetObject(ctx, req.BucketName, req.ObjectName, opts)
	if err != nil {
		return nil, nil, handleMinIOError(err, "download_file")
	}
	return object, m.generateDownloadHeaders(objInfo, req), nil
}

func (m *implMinIO) StreamFile(ctx context.Context, req *DownloadRequest) (io.ReadCloser, *DownloadHeaders, error) {
	req.Disposition = DispositionInline
	reader, headers, err := m.DownloadFile(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	headers.CacheControl = "public, max-age=86400"
	headers.AcceptRanges = "bytes"
	if req.Range != nil {
		headers.ContentRange = fmt.Sprintf("bytes %d-%d/%s", req.Range.Start, req.Range.End, headers.ContentLength)
	}
	return reader, headers, nil
}

func (m *implMinIO) GetPresignedDownloadURL(ctx context.Context, req *PresignedURLRequest) (*PresignedURLResponse, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Generating presigned download URL", "trace_id", traceID, "bucket", req.BucketName)
	}

	if err := validatePresignedURLRequest(req); err != nil {
		return nil, err
	}
	url, err := m.minioClient.PresignedGetObject(ctx, req.BucketName, req.ObjectName, req.Expiry, nil)
	if err != nil {
		return nil, handleMinIOError(err, "get_presigned_download_url")
	}
	return &PresignedURLResponse{
		URL:       url.String(),
		ExpiresAt: time.Now().Add(req.Expiry),
		Method:    MethodGET,
		Headers:   req.Headers,
	}, nil
}

// --- implMinIO: file management operations with trace integration ---

func (m *implMinIO) GetFileInfo(ctx context.Context, bucketName, objectName string) (*FileInfo, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Getting file info", "trace_id", traceID, "bucket", bucketName, "object", objectName)
	}

	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}
	if err := validateObjectName(objectName); err != nil {
		return nil, err
	}
	objInfo, err := m.minioClient.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, handleMinIOError(err, "get_file_info")
	}
	fileInfo := &FileInfo{
		BucketName:   bucketName,
		ObjectName:   objectName,
		Size:         objInfo.Size,
		ContentType:  objInfo.ContentType,
		ETag:         objInfo.ETag,
		LastModified: objInfo.LastModified,
		Metadata:     objInfo.UserMetadata,
	}
	if originalName, exists := objInfo.UserMetadata["original-name"]; exists {
		fileInfo.OriginalName = originalName
	}
	return fileInfo, nil
}

func (m *implMinIO) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Deleting file", "trace_id", traceID, "bucket", bucketName, "object", objectName)
	}

	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}
	return handleMinIOError(m.minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}), "delete_file")
}

func (m *implMinIO) CopyFile(ctx context.Context, srcBucket, srcObject, destBucket, destObject string) error {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Copying file", "trace_id", traceID, "src_bucket", srcBucket, "dest_bucket", destBucket)
	}

	_, err := m.minioClient.CopyObject(ctx,
		minio.CopyDestOptions{Bucket: destBucket, Object: destObject},
		minio.CopySrcOptions{Bucket: srcBucket, Object: srcObject})
	return handleMinIOError(err, "copy_file")
}

func (m *implMinIO) MoveFile(ctx context.Context, srcBucket, srcObject, destBucket, destObject string) error {
	if err := m.CopyFile(ctx, srcBucket, srcObject, destBucket, destObject); err != nil {
		return err
	}
	if err := m.DeleteFile(ctx, srcBucket, srcObject); err != nil {
		if cleanupErr := m.DeleteFile(ctx, destBucket, destObject); cleanupErr != nil {
			return fmt.Errorf("move failed: %w, cleanup also failed: %v", err, cleanupErr)
		}
		return fmt.Errorf("move failed: %w", err)
	}
	return nil
}

func (m *implMinIO) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	_, err := m.GetFileInfo(ctx, bucketName, objectName)
	if err != nil {
		if storageErr, ok := err.(*StorageError); ok && storageErr.Code == ErrCodeObjectNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// --- implMinIO: list and metadata operations with trace integration ---

func (m *implMinIO) ListFiles(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Listing files", "trace_id", traceID, "bucket", req.BucketName, "prefix", req.Prefix)
	}

	if err := validateListRequest(req); err != nil {
		return nil, err
	}
	opts := minio.ListObjectsOptions{Prefix: req.Prefix, Recursive: req.Recursive, MaxKeys: req.MaxKeys}
	var files []*FileInfo
	objectCh := m.minioClient.ListObjects(ctx, req.BucketName, opts)
	for object := range objectCh {
		if object.Err != nil {
			return nil, handleMinIOError(object.Err, "list_files")
		}
		files = append(files, &FileInfo{
			BucketName:   req.BucketName,
			ObjectName:   object.Key,
			Size:         object.Size,
			ETag:         object.ETag,
			LastModified: object.LastModified,
			ContentType:  object.ContentType,
		})
	}
	resp := &ListResponse{Files: files, TotalCount: len(files), IsTruncated: len(files) >= req.MaxKeys}
	if resp.IsTruncated && len(files) > 0 {
		resp.NextMarker = files[len(files)-1].ObjectName
	}
	return resp, nil
}

func (m *implMinIO) UpdateMetadata(ctx context.Context, bucketName, objectName string, metadata map[string]string) error {
	if traceID := m.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Updating metadata", "trace_id", traceID, "bucket", bucketName, "object", objectName)
	}

	_, err := m.minioClient.CopyObject(ctx,
		minio.CopyDestOptions{Bucket: bucketName, Object: objectName, UserMetadata: metadata, ReplaceMetadata: true},
		minio.CopySrcOptions{Bucket: bucketName, Object: objectName})
	return handleMinIOError(err, "update_metadata")
}

func (m *implMinIO) GetMetadata(ctx context.Context, bucketName, objectName string) (map[string]string, error) {
	fileInfo, err := m.GetFileInfo(ctx, bucketName, objectName)
	if err != nil {
		return nil, err
	}
	return fileInfo.Metadata, nil
}

// --- implMinIO: async upload operations (delegate to manager) ---

func (m *implMinIO) UploadAsync(ctx context.Context, req *UploadRequest) (string, error) {
	return m.asyncUploadMgr.uploadAsync(ctx, req)
}

func (m *implMinIO) GetUploadStatus(taskID string) (*UploadProgress, error) {
	return m.asyncUploadMgr.getUploadStatus(taskID)
}

func (m *implMinIO) WaitForUpload(taskID string, timeout time.Duration) (*AsyncUploadResult, error) {
	return m.asyncUploadMgr.waitForUpload(taskID, timeout)
}

func (m *implMinIO) CancelUpload(taskID string) error {
	return m.asyncUploadMgr.cancelUpload(taskID)
}
