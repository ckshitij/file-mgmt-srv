package filesrv

import (
	"context"
	"time"

	"github.com/go-kit/log"
)

type Middleware func(FileService) FileService

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next FileService) FileService {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   FileService
	logger log.Logger
}

func (mw loggingMiddleware) InitUpload(ctx context.Context, filename string, totalChunks int, chunkSize int) (id string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "InitUpload", "fileName", filename, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.InitUpload(ctx, filename, totalChunks, chunkSize)
}

func (mw loggingMiddleware) UploadChunk(ctx context.Context, sessionID string, chunkNum int, data []byte) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "UploadChunk", "sessionID", sessionID, "chunkNum", chunkNum, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.UploadChunk(ctx, sessionID, chunkNum, data)
}

func (mw loggingMiddleware) FinalizeUpload(ctx context.Context, sessionID string) (fileID string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "FinalizeUpload", "sessionID", sessionID, "fileID", fileID, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.FinalizeUpload(ctx, sessionID)
}
func (mw loggingMiddleware) AbortUpload(ctx context.Context, sessionID string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "AbortUpload", "sessionID", sessionID, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.AbortUpload(ctx, sessionID)
}

func (mw loggingMiddleware) DownloadFile(ctx context.Context, filename string) (downloadedData []byte, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DownloadFile", "filename", filename, "downloadedData", len(downloadedData), "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DownloadFile(ctx, filename)
}
