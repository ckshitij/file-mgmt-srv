package filesrv

import "context"

type errorer interface {
	Err() error
}

type FileService interface {
	InitUpload(ctx context.Context, filename string, totalChunks int, chunkSize int) (string, error)
	UploadChunk(ctx context.Context, sessionID string, chunkNum int, data []byte) error
	FinalizeUpload(ctx context.Context, sessionID string) (string, error)
	AbortUpload(ctx context.Context, sessionID string) error
	DownloadFile(ctx context.Context, filename string) ([]byte, error)
}
