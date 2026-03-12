package minio

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
)

func validateConfig(cfg *Config) error {
	if cfg.Endpoint == "" {
		return NewInvalidInputError("endpoint is required")
	}
	if cfg.AccessKey == "" {
		return NewInvalidInputError("access key is required")
	}
	if cfg.SecretKey == "" {
		return NewInvalidInputError("secret key is required")
	}
	if cfg.Region == "" {
		return NewInvalidInputError("region is required")
	}
	if cfg.Bucket == "" {
		return NewInvalidInputError("bucket is required")
	}
	if !strings.Contains(cfg.Endpoint, ":") {
		cfg.Endpoint = cfg.Endpoint + DefaultEndpointPort
	}
	return nil
}

func validateUploadRequest(req *UploadRequest) error {
	if req.BucketName == "" {
		return NewInvalidInputError("bucket name is required")
	}
	if req.ObjectName == "" {
		return NewInvalidInputError("object name is required")
	}
	if req.Reader == nil {
		return NewInvalidInputError("reader is required")
	}
	if req.Size <= 0 {
		return NewInvalidInputError("size must be positive")
	}
	if req.ContentType == "" {
		return NewInvalidInputError("content type is required")
	}
	if strings.HasPrefix(req.ObjectName, "/") {
		return NewInvalidInputError("object name cannot start with '/'")
	}
	if strings.HasSuffix(req.ObjectName, "/") {
		return NewInvalidInputError("object name cannot end with '/'")
	}
	if req.Size > MaxFileSizeBytes {
		return NewInvalidInputError("file size cannot exceed 5GB")
	}
	return nil
}

func validateDownloadRequest(req *DownloadRequest) error {
	if req.BucketName == "" {
		return NewInvalidInputError("bucket name is required")
	}
	if req.ObjectName == "" {
		return NewInvalidInputError("object name is required")
	}
	if req.Disposition != "" && req.Disposition != DispositionAuto && req.Disposition != DispositionInline && req.Disposition != DispositionAttachment {
		return NewInvalidInputError("disposition must be 'auto', 'inline', or 'attachment'")
	}
	if req.Range != nil {
		if req.Range.Start < 0 {
			return NewInvalidInputError("range start must be non-negative")
		}
		if req.Range.End < req.Range.Start {
			return NewInvalidInputError("range end must be greater than or equal to start")
		}
	}
	return nil
}
func validateListRequest(req *ListRequest) error {
	if req.BucketName == "" {
		return NewInvalidInputError("bucket name is required")
	}
	if req.MaxKeys <= 0 {
		req.MaxKeys = DefaultListMaxKeys
	}
	if req.MaxKeys > MaxListMaxKeys {
		return NewInvalidInputError("max keys cannot exceed 1000")
	}
	return nil
}

func validatePresignedURLRequest(req *PresignedURLRequest) error {
	if req.BucketName == "" {
		return NewInvalidInputError("bucket name is required")
	}
	if req.ObjectName == "" {
		return NewInvalidInputError("object name is required")
	}
	if req.Method == "" {
		return NewInvalidInputError("method is required")
	}
	if req.Method != MethodGET && req.Method != MethodPUT {
		return NewInvalidInputError("method must be 'GET' or 'PUT'")
	}
	if req.Expiry <= 0 {
		return NewInvalidInputError("expiry must be positive")
	}
	if req.Expiry > MaxPresignedExpiry {
		return NewInvalidInputError("expiry cannot exceed 7 days")
	}
	return nil
}

func validateBucketName(bucketName string) error {
	if bucketName == "" {
		return NewInvalidInputError("bucket name is required")
	}
	if len(bucketName) < 3 {
		return NewInvalidInputError("bucket name must be at least 3 characters")
	}
	if len(bucketName) > 63 {
		return NewInvalidInputError("bucket name cannot exceed 63 characters")
	}
	for _, char := range bucketName {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return NewInvalidInputError("bucket name can only contain lowercase letters, numbers, and hyphens")
		}
	}
	if strings.Contains(bucketName, "--") {
		return NewInvalidInputError("bucket name cannot contain consecutive hyphens")
	}
	if strings.HasPrefix(bucketName, "-") || strings.HasSuffix(bucketName, "-") {
		return NewInvalidInputError("bucket name cannot start or end with hyphen")
	}
	return nil
}

func validateObjectName(objectName string) error {
	if objectName == "" {
		return NewInvalidInputError("object name is required")
	}
	if strings.Contains(objectName, "\\") {
		return NewInvalidInputError("object name cannot contain backslashes")
	}
	return nil
}
func handleMinIOError(err error, operation string) *StorageError {
	if err == nil {
		return nil
	}
	if minioErr, ok := err.(minio.ErrorResponse); ok {
		switch minioErr.Code {
		case "NoSuchBucket":
			return NewBucketNotFoundError("")
		case "NoSuchKey":
			return NewObjectNotFoundError("")
		case "AccessDenied":
			return &StorageError{Code: ErrCodePermission, Message: "Access denied", Operation: operation, Cause: err}
		default:
			return &StorageError{Code: ErrCodeConnection, Message: fmt.Sprintf("MinIO operation failed: %s", minioErr.Code), Operation: operation, Cause: err}
		}
	}
	return NewConnectionError(err)
}

func (m *implMinIO) generateDownloadHeaders(objInfo minio.ObjectInfo, req *DownloadRequest) *DownloadHeaders {
	disposition := m.determineContentDisposition(objInfo.ContentType, req.Disposition)
	originalName := objInfo.UserMetadata["original-name"]
	if originalName == "" {
		originalName = objInfo.Key
	}
	headers := &DownloadHeaders{
		ContentType:        objInfo.ContentType,
		ContentDisposition: fmt.Sprintf("%s; filename=\"%s\"", disposition, originalName),
		ContentLength:      fmt.Sprintf("%d", objInfo.Size),
		LastModified:       objInfo.LastModified.Format(http.TimeFormat),
		ETag:               objInfo.ETag,
		AcceptRanges:       "bytes",
	}
	if disposition == DispositionInline {
		headers.CacheControl = "public, max-age=3600"
	} else {
		headers.CacheControl = "private, no-cache"
	}
	return headers
}

func (m *implMinIO) determineContentDisposition(contentType, requestedDisposition string) string {
	if requestedDisposition == DispositionInline || requestedDisposition == DispositionAttachment {
		return requestedDisposition
	}
	if requestedDisposition == DispositionAuto {
		viewableTypes := []string{"image/", "video/", "audio/", "application/pdf", "text/plain", "text/html", "application/json", "application/xml"}
		for _, viewable := range viewableTypes {
			if strings.HasPrefix(contentType, viewable) {
				return DispositionInline
			}
		}
		return DispositionAttachment
	}
	return DispositionAttachment
}
