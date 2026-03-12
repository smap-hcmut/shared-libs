# MinIO Package

The MinIO package provides object storage operations with distributed tracing integration for SMAP services.

## Features

- **Complete S3 API**: Full MinIO/S3 compatible operations
- **Trace Integration**: Automatic trace_id injection in file metadata
- **Async Uploads**: Background file uploads with progress tracking
- **Presigned URLs**: Generate secure upload/download URLs
- **File Management**: Upload, download, copy, move, delete operations
- **Bucket Management**: Create, delete, list buckets
- **Metadata Management**: Custom file metadata with trace context
- **Stream Support**: Efficient streaming for large files
- **Backward Compatibility**: Drop-in replacement for existing MinIO packages

## Usage

### Basic Configuration

```go
import "github.com/smap-hcmut/shared-libs/go/minio"

// Configuration
cfg := &minio.Config{
    Endpoint:  "localhost:9000",
    AccessKey: "minioadmin",
    SecretKey: "minioadmin",
    Region:    "us-east-1",
    Bucket:    "my-bucket",
    UseSSL:    false,
}

// Create client
client, err := minio.NewMinIO(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Advanced Configuration with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/minio"
    "github.com/smap-hcmut/shared-libs/go/tracing"
)

// Create with custom tracer
tracer := tracing.NewTraceContext()
client, err := minio.NewMinIOWithTracer(cfg, tracer)

// Or create with retry and tracer
client, err := minio.NewMinIOWithRetryAndTracer(cfg, 3, tracer)
```
### Connection Management

```go
// Connect to MinIO
err := client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}

// Connect with retry
err = client.ConnectWithRetry(ctx, 3)

// Health check
err = client.HealthCheck(ctx)
if err != nil {
    log.Printf("MinIO health check failed: %v", err)
}
```

### Bucket Operations

```go
// Create bucket
err := client.CreateBucket(ctx, "my-bucket")

// Check if bucket exists
exists, err := client.BucketExists(ctx, "my-bucket")

// List all buckets
buckets, err := client.ListBuckets(ctx)
for _, bucket := range buckets {
    fmt.Printf("Bucket: %s, Created: %v\n", bucket.Name, bucket.CreationDate)
}

// Delete bucket
err = client.DeleteBucket(ctx, "my-bucket")
```

### File Upload Operations

```go
// Basic file upload
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

stat, _ := file.Stat()
fileInfo, err := client.UploadFile(ctx, &minio.UploadRequest{
    BucketName:   "my-bucket",
    ObjectName:   "documents/document.pdf",
    OriginalName: "document.pdf",
    Reader:       file,
    Size:         stat.Size(),
    ContentType:  "application/pdf",
    Metadata: map[string]string{
        "department": "engineering",
        "project":    "smap",
    },
})

if err != nil {
    log.Fatal(err)
}
fmt.Printf("Uploaded: %s, Size: %d, ETag: %s\n", 
    fileInfo.ObjectName, fileInfo.Size, fileInfo.ETag)
```
### File Download Operations

```go
// Download file
reader, headers, err := client.DownloadFile(ctx, &minio.DownloadRequest{
    BucketName: "my-bucket",
    ObjectName: "documents/document.pdf",
})
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// Save to local file
outFile, err := os.Create("downloaded.pdf")
if err != nil {
    log.Fatal(err)
}
defer outFile.Close()

_, err = io.Copy(outFile, reader)

// Stream file (for web serving)
reader, headers, err := client.StreamFile(ctx, &minio.DownloadRequest{
    BucketName:  "my-bucket",
    ObjectName:  "images/photo.jpg",
    Disposition: minio.DispositionInline,
})

// Partial download with range
reader, headers, err := client.DownloadFile(ctx, &minio.DownloadRequest{
    BucketName: "my-bucket",
    ObjectName: "large-file.zip",
    Range: &minio.ByteRange{
        Start: 0,
        End:   1024, // First 1KB
    },
})
```

### Presigned URLs

```go
// Generate presigned upload URL
uploadURL, err := client.GetPresignedUploadURL(ctx, &minio.PresignedURLRequest{
    BucketName: "my-bucket",
    ObjectName: "uploads/user-file.jpg",
    Expiry:     time.Hour, // Valid for 1 hour
})

fmt.Printf("Upload URL: %s\n", uploadURL.URL)
fmt.Printf("Expires at: %v\n", uploadURL.ExpiresAt)

// Generate presigned download URL
downloadURL, err := client.GetPresignedDownloadURL(ctx, &minio.PresignedURLRequest{
    BucketName: "my-bucket",
    ObjectName: "documents/report.pdf",
    Expiry:     time.Hour * 24, // Valid for 24 hours
})
```
### Async Upload Operations

```go
// Start async upload
taskID, err := client.UploadAsync(ctx, &minio.UploadRequest{
    BucketName:   "my-bucket",
    ObjectName:   "large-files/video.mp4",
    OriginalName: "presentation.mp4",
    Reader:       videoFile,
    Size:         videoSize,
    ContentType:  "video/mp4",
})

if err != nil {
    log.Fatal(err)
}

// Check upload progress
progress, err := client.GetUploadStatus(taskID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Upload progress: %.2f%% (%d/%d bytes)\n", 
    progress.Percentage, progress.BytesUploaded, progress.TotalBytes)

// Wait for upload completion
result, err := client.WaitForUpload(taskID, time.Minute*10)
if err != nil {
    log.Fatal(err)
}

if result.Error != nil {
    log.Printf("Upload failed: %v", result.Error)
} else {
    fmt.Printf("Upload completed: %s\n", result.FileInfo.ObjectName)
    fmt.Printf("Duration: %v\n", result.Duration)
}

// Cancel upload if needed
err = client.CancelUpload(taskID)
```

### File Management

```go
// Get file information
fileInfo, err := client.GetFileInfo(ctx, "my-bucket", "documents/report.pdf")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("File: %s, Size: %d, Modified: %v\n", 
    fileInfo.ObjectName, fileInfo.Size, fileInfo.LastModified)

// Check if file exists
exists, err := client.FileExists(ctx, "my-bucket", "documents/report.pdf")

// Copy file
err = client.CopyFile(ctx, "source-bucket", "file.txt", "dest-bucket", "backup/file.txt")

// Move file
err = client.MoveFile(ctx, "my-bucket", "temp/file.txt", "my-bucket", "final/file.txt")

// Delete file
err = client.DeleteFile(ctx, "my-bucket", "documents/old-report.pdf")
```