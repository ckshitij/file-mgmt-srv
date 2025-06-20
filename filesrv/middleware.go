// Package filesrv provides middleware for logging operations
// in the FileService, capturing method call details such as execution time,
// errors, and important parameters using Go-Kit's logging infrastructure.
package filesrv

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log"
)

// Middleware defines a service middleware that wraps a FileService
// with additional behavior (e.g., logging, metrics, tracing).
type Middleware func(FileService) FileService

// LoggingMiddleware returns a middleware that logs each method call
// with timing, errors, and input parameters.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next FileService) FileService {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

// loggingMiddleware is a FileService implementation that wraps another
// FileService and logs information about each method invocation.
type loggingMiddleware struct {
	next   FileService // the next service in the chain
	logger log.Logger  // structured logger from Go-Kit
}

// InitUpload logs metadata and duration for InitUpload calls.
func (mw loggingMiddleware) InitUpload(ctx context.Context, filename string, totalChunks int, chunkSize int) (id string, err error) {
	defer func(begin time.Time) {
		if logErr := mw.logger.Log("method", "InitUpload", "fileName", filename, "took", time.Since(begin), "err", err); logErr != nil {
			fmt.Println("log error:", logErr)
		}
	}(time.Now())
	return mw.next.InitUpload(ctx, filename, totalChunks, chunkSize)
}

// UploadChunk logs metadata and duration for UploadChunk calls.
func (mw loggingMiddleware) UploadChunk(ctx context.Context, sessionID string, chunkNum int, data []byte) (err error) {
	defer func(begin time.Time) {
		if logErr := mw.logger.Log("method", "UploadChunk", "sessionID", sessionID, "chunkNum", chunkNum, "took", time.Since(begin), "err", err); logErr != nil {
			fmt.Println("log error:", logErr)
		}
	}(time.Now())
	return mw.next.UploadChunk(ctx, sessionID, chunkNum, data)
}

// FinalizeUpload logs metadata and duration for FinalizeUpload calls.
func (mw loggingMiddleware) FinalizeUpload(ctx context.Context, sessionID string) (fileID string, err error) {
	defer func(begin time.Time) {
		if logErr := mw.logger.Log("method", "FinalizeUpload", "sessionID", sessionID, "fileID", fileID, "took", time.Since(begin), "err", err); logErr != nil {
			fmt.Println("log error:", logErr)
		}
	}(time.Now())
	return mw.next.FinalizeUpload(ctx, sessionID)
}

// AbortUpload logs metadata and duration for AbortUpload calls.
func (mw loggingMiddleware) AbortUpload(ctx context.Context, sessionID string) (err error) {
	defer func(begin time.Time) {
		if logErr := mw.logger.Log("method", "AbortUpload", "sessionID", sessionID, "took", time.Since(begin), "err", err); logErr != nil {
			fmt.Println("log error:", logErr)
		}
	}(time.Now())
	return mw.next.AbortUpload(ctx, sessionID)
}

// DownloadFile logs metadata and duration for DownloadFile calls,
// including the size of the downloaded data.
func (mw loggingMiddleware) DownloadFile(ctx context.Context, filename string) (downloadedData []byte, err error) {
	defer func(begin time.Time) {
		if logErr := mw.logger.Log("method", "DownloadFile", "filename", filename, "downloadedData", len(downloadedData), "took", time.Since(begin), "err", err); logErr != nil {
			fmt.Println("log error:", logErr)
		}
	}(time.Now())
	return mw.next.DownloadFile(ctx, filename)
}
